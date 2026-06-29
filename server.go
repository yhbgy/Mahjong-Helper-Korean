package main

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
	"github.com/EndlessCheng/mahjong-helper/platform/tenhou"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/debug"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
	"github.com/golang/protobuf/proto"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	stdLog "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const defaultPort = 12121

func newLogFilePath() (filePath string, err error) {
	const logDir = "log"
	if err = os.MkdirAll(logDir, os.ModePerm); err != nil {
		return
	}
	fileName := fmt.Sprintf("gamedata-%s.log", time.Now().Format("20060102-150405"))
	filePath = filepath.Join(logDir, fileName)
	return filepath.Abs(filePath)
}

type mjHandler struct {
	log echo.Logger

	analysing bool

	tenhouMessageReceiver *tenhou.MessageReceiver
	tenhouRoundData       *tenhouRoundData

	majsoulMessageQueue chan []byte
	majsoulRawQueue     chan []byte
	majsoulRawRequests  map[uint16]*majsoulRawRequest
	majsoulRoundData    *majsoulRoundData

	majsoulRecordMap                map[string]*majsoulRecordBaseInfo
	majsoulCurrentRecordUUID        string
	majsoulCurrentRecordActionsList []majsoulRoundActions
	majsoulCurrentRoundIndex        int
	majsoulCurrentActionIndex       int

	majsoulCurrentRoundActions majsoulRoundActions
}

func (h *mjHandler) logError(err error) {
	fmt.Fprintln(os.Stderr, err)
	if !debugMode {
		h.log.Error(err)
	}
}

// 디버그용
func (h *mjHandler) index(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		h.log.Error("[mjHandler.index.ioutil.ReadAll]", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	fmt.Println(data, string(data))
	h.log.Info(data)
	return c.String(http.StatusOK, time.Now().Format("2006-01-02 15:04:05"))
}

// 한 장 버리고 한 장 뽑는 분석기
func (h *mjHandler) analysis(c echo.Context) error {
	if h.analysing {
		return c.NoContent(http.StatusForbidden)
	}

	h.analysing = true
	defer func() { h.analysing = false }()

	d := struct {
		Reset bool   `json:"reset"`
		Tiles string `json:"tiles"`
	}{}
	if err := c.Bind(&d); err != nil {
		fmt.Println(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	if _, err := analysisHumanTiles(model.NewSimpleHumanTilesInfo(d.Tiles)); err != nil {
		fmt.Println(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// 천봉 WebSocket 데이터 분석
func (h *mjHandler) analysisTenhou(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		h.logError(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	h.tenhouMessageReceiver.Put(data)
	return c.NoContent(http.StatusOK)
}
func (h *mjHandler) runAnalysisTenhouMessageTask() {
	if !debugMode {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("내부 오류:", err)
			}
		}()
	}

	for {
		msg := h.tenhouMessageReceiver.Get()
		d := tenhouMessage{}
		if err := json.Unmarshal(msg, &d); err != nil {
			h.logError(err)
			continue
		}

		originJSON := string(msg)
		if h.log != nil {
			h.log.Info(originJSON)
		}

		h.tenhouRoundData.msg = &d
		h.tenhouRoundData.originJSON = originJSON
		if err := h.tenhouRoundData.analysis(); err != nil {
			h.logError(err)
		}
	}
}

// 작혼 WebSocket 데이터 분석
func (h *mjHandler) analysisMajsoul(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		h.logError(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	h.majsoulMessageQueue <- data
	return c.NoContent(http.StatusOK)
}
func (h *mjHandler) analysisMajsoulRaw(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		h.logError(err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if len(data) == 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	h.majsoulRawQueue <- data
	return c.NoContent(http.StatusOK)
}

type majsoulRawRequest struct {
	name         string
	responseType reflect.Type
}

func unwrapMajsoulWrapper(rawData []byte) (string, []byte, error) {
	wrapper := lq.Wrapper{}
	if err := proto.Unmarshal(rawData, &wrapper); err != nil {
		return "", nil, err
	}
	return wrapper.GetName(), wrapper.GetData(), nil
}

func shortProtoMessage(message proto.Message) string {
	if message == nil {
		return ""
	}
	text := proto.CompactTextString(message)
	if len(text) > 300 {
		return text[:300] + "..."
	}
	return text
}

func (h *mjHandler) decodeMajsoulRawRequest(data []byte) {
	if len(data) < 4 {
		fmt.Printf("[majsoul-raw] request len=%d too short\n", len(data))
		return
	}

	messageIndex := binary.LittleEndian.Uint16(data[1:3])
	rawMethodName, payload, err := unwrapMajsoulWrapper(data[3:])
	if err != nil {
		fmt.Printf("[majsoul-raw] request #%d unwrap error: %v\n", messageIndex, err)
		return
	}

	methodName := strings.TrimPrefix(rawMethodName, ".")
	request := &majsoulRawRequest{name: methodName}
	defer func() {
		h.majsoulRawRequests[messageIndex] = request
	}()

	parts := strings.Split(methodName, ".")
	if len(parts) < 3 {
		fmt.Printf("[majsoul-raw] request #%d %s payload=%d\n", messageIndex, methodName, len(payload))
		return
	}

	clientName, rpcName := parts[1], parts[2]
	if clientName != "Lobby" && clientName != "FastTest" {
		fmt.Printf("[majsoul-raw] request #%d %s payload=%d\n", messageIndex, methodName, len(payload))
		return
	}

	methodType := lq.FindMethod(clientName, rpcName)
	if methodType == nil || methodType.NumIn() < 2 || methodType.NumOut() < 1 {
		fmt.Printf("[majsoul-raw] request #%d %s payload=%d\n", messageIndex, methodName, len(payload))
		return
	}

	reqType := methodType.In(1)
	request.responseType = methodType.Out(0)
	reqMessage := reflect.New(reqType.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(payload, reqMessage); err != nil {
		fmt.Printf("[majsoul-raw] request #%d %s decode error: %v\n", messageIndex, methodName, err)
		return
	}

	fmt.Printf("[majsoul-raw] request #%d %s %s\n", messageIndex, methodName, shortProtoMessage(reqMessage))
}

func (h *mjHandler) decodeMajsoulRawResponse(data []byte) {
	if len(data) < 4 {
		fmt.Printf("[majsoul-raw] response len=%d too short\n", len(data))
		return
	}

	messageIndex := binary.LittleEndian.Uint16(data[1:3])
	request, ok := h.majsoulRawRequests[messageIndex]
	if ok {
		delete(h.majsoulRawRequests, messageIndex)
	}

	if !ok || request.responseType == nil {
		name, payload, err := unwrapMajsoulWrapper(data[3:])
		if err != nil {
			fmt.Printf("[majsoul-raw] response #%d unwrap error: %v\n", messageIndex, err)
			return
		}
		if name == "" && ok {
			name = request.name
		}
		fmt.Printf("[majsoul-raw] response #%d %s payload=%d\n", messageIndex, name, len(payload))
		return
	}

	respMessage := reflect.New(request.responseType.Elem()).Interface().(proto.Message)
	_, payload, err := unwrapMajsoulWrapper(data[3:])
	if err != nil {
		fmt.Printf("[majsoul-raw] response #%d %s unwrap error: %v\n", messageIndex, request.name, err)
		return
	}
	if err := proto.Unmarshal(payload, respMessage); err != nil {
		fmt.Printf("[majsoul-raw] response #%d %s decode error: %v\n", messageIndex, request.name, err)
		return
	}

	fmt.Printf("[majsoul-raw] response #%d %s %s\n", messageIndex, request.name, shortProtoMessage(respMessage))
}

func (h *mjHandler) decodeMajsoulRawNotify(data []byte) {
	name, payload, err := unwrapMajsoulWrapper(data[1:])
	if err != nil {
		fmt.Printf("[majsoul-raw] notify unwrap error: %v\n", err)
		return
	}

	messageName := strings.TrimPrefix(name, ".")
	messageType := proto.MessageType(messageName)
	if messageType == nil {
		fmt.Printf("[majsoul-raw] notify %s payload=%d\n", messageName, len(payload))
		return
	}

	notifyMessage := reflect.New(messageType.Elem()).Interface().(proto.Message)
	if err := proto.Unmarshal(payload, notifyMessage); err != nil {
		fmt.Printf("[majsoul-raw] notify %s decode error: %v\n", messageName, err)
		return
	}
	fmt.Printf("[majsoul-raw] notify %s %s\n", messageName, shortProtoMessage(notifyMessage))
}

func (h *mjHandler) runAnalysisMajsoulRawMessageTask() {
	for data := range h.majsoulRawQueue {
		if len(data) == 0 {
			continue
		}

		switch data[0] {
		case 1:
			h.decodeMajsoulRawNotify(data)
		case 2:
			h.decodeMajsoulRawRequest(data)
		case 3:
			h.decodeMajsoulRawResponse(data)
		default:
			fmt.Printf("[majsoul-raw] unknown type=%d len=%d first=%v\n", data[0], len(data), data[:minInt(len(data), 12)])
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (h *mjHandler) runAnalysisMajsoulMessageTask() {
	if !debugMode {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("내부 오류:", err)
			}
		}()
	}

	for msg := range h.majsoulMessageQueue {
		d := &majsoulMessage{}
		if err := json.Unmarshal(msg, d); err != nil {
			h.logError(err)
			continue
		}

		originJSON := string(msg)
		if h.log != nil && debug.Lo == 0 {
			h.log.Info(originJSON)
		} else {
			if len(originJSON) > 500 {
				originJSON = originJSON[:500]
			}
			fmt.Println(originJSON)
		}

		switch {
		case len(d.Friends) > 0:
			// 친구 목록
			fmt.Println(d.Friends)
		case len(d.RecordBaseInfoList) > 0:
			// 패보 기본 정보 목록
			for _, record := range d.RecordBaseInfoList {
				h.majsoulRecordMap[record.UUID] = record
			}
			color.HiGreen("작혼 패보 %2d개 수신(총 %d개 수집). 웹페이지에서 [보기]를 클릭하세요", len(d.RecordBaseInfoList), len(h.majsoulRecordMap))
		case d.SharedRecordBaseInfo != nil:
			// 공유된 패보 기본 정보 처리
			// FIXME: 자신의 패보를 볼 때도 d.SharedRecordBaseInfo가 생긴다
			record := d.SharedRecordBaseInfo
			h.majsoulRecordMap[record.UUID] = record
			if err := h._loadMajsoulRecordBaseInfo(record.UUID); err != nil {
				h.logError(err)
				break
			}
		case d.CurrentRecordUUID != "":
			// 특정 패보 로드
			resetAnalysisCache()
			h.majsoulCurrentRecordActionsList = nil

			if err := h._loadMajsoulRecordBaseInfo(d.CurrentRecordUUID); err != nil {
				// 공유 패보를 보는 경우(CurrentRecordUUID와 AccountID를 먼저 받고, 이후 SharedRecordBaseInfo를 받음)
				// 또는 대회장 패보인 경우
				// 주 시점 ID를 기록한다(0일 수 있음)
				gameConf.setMajsoulAccountID(d.AccountID)
				break
			}

			// 자신의 패보를 보는 경우
			// 현재 사용할 계정을 갱신한다
			gameConf.addMajsoulAccountID(d.AccountID)
			if gameConf.currentActiveMajsoulAccountID != d.AccountID {
				fmt.Println()
				printAccountInfo(d.AccountID)
				gameConf.setMajsoulAccountID(d.AccountID)
			}
		case len(d.RecordActions) > 0:
			if h.majsoulCurrentRecordActionsList != nil {
				// TODO: 웹에서 더 적절한 정보를 보내도록 할까?
				break
			}

			if h.majsoulCurrentRecordUUID == "" {
				h.logError(fmt.Errorf("오류: 보고 있는 작혼 패보의 UUID를 받지 못했습니다"))
				break
			}

			baseInfo, ok := h.majsoulRecordMap[h.majsoulCurrentRecordUUID]
			if !ok {
				h.logError(fmt.Errorf("오류: 작혼 패보 %s 를 찾을 수 없습니다", h.majsoulCurrentRecordUUID))
				break
			}

			selfAccountID := gameConf.currentActiveMajsoulAccountID
			if selfAccountID == -1 {
				h.logError(fmt.Errorf("오류: 현재 작혼 계정이 비어 있습니다"))
				break
			}

			h.majsoulRoundData.newGame()
			h.majsoulRoundData.gameMode = gameModeRecord

			// 주 시점의 초기 좌석을 가져와 설정한다
			selfSeat, err := baseInfo.getSelfSeat(selfAccountID)
			if err != nil {
				h.logError(err)
				break
			}
			h.majsoulRoundData.selfSeat = selfSeat

			// 분석 준비...
			majsoulCurrentRecordActions, err := parseMajsoulRecordAction(d.RecordActions)
			if err != nil {
				h.logError(err)
				break
			}
			h.majsoulCurrentRecordActionsList = majsoulCurrentRecordActions
			h.majsoulCurrentRoundIndex = 0
			h.majsoulCurrentActionIndex = 0

			actions := h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex]

			// 분석 작업 생성
			analysisCache := newGameAnalysisCache(h.majsoulCurrentRecordUUID, selfSeat)
			setAnalysisCache(analysisCache)
			go analysisCache.runMajsoulRecordAnalysisTask(actions)

			// 첫 국의 시작 정보를 분석한다
			data := actions[0].Action
			h._analysisMajsoulRoundData(data, originJSON)
		case d.RecordClickAction != "":
			// 웹 패보 클릭 처리: 이전 국/특정 국 이동/다음 국/이전 순/특정 순 이동/다음 순/이전 단계/재생/일시정지/다음 단계/탁자 클릭
			// 아직 타가 손패는 분석할 수 없다
			h._onRecordClick(d.RecordClickAction, d.RecordClickActionIndex, d.FastRecordTo)
		case d.LiveBaseInfo != nil:
			// 관전
			gameConf.setMajsoulAccountID(1) // TODO: 설명 보완
			h.majsoulRoundData.newGame()
			h.majsoulRoundData.selfSeat = 0 // 설명
			h.majsoulRoundData.gameMode = gameModeLive
			clearConsole()
			fmt.Printf("대전 불러오는 중: %s", d.LiveBaseInfo.String())
		case d.LiveFastAction != nil:
			if err := h._loadLiveAction(d.LiveFastAction, true); err != nil {
				h.logError(err)
				break
			}
		case d.LiveAction != nil:
			if err := h._loadLiveAction(d.LiveAction, false); err != nil {
				h.logError(err)
				break
			}
		case d.ChangeSeatTo != nil:
			// 좌석 전환
			changeSeatTo := *(d.ChangeSeatTo)
			h.majsoulRoundData.selfSeat = changeSeatTo
			if debugMode {
				fmt.Println("좌석 전환:", changeSeatTo)
			}

			var actions majsoulRoundActions
			if h.majsoulRoundData.gameMode == gameModeLive { // 설명
				actions = h.majsoulCurrentRoundActions
			} else { // 설명
				fullActions := h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex]
				actions = fullActions[:h.majsoulCurrentActionIndex+1]
				analysisCache := getAnalysisCache(changeSeatTo)
				if analysisCache == nil {
					analysisCache = newGameAnalysisCache(h.majsoulCurrentRecordUUID, changeSeatTo)
				}
				setAnalysisCache(analysisCache)
				// 분석 작업 생성
				go analysisCache.runMajsoulRecordAnalysisTask(fullActions)
			}

			h._fastLoadActions(actions)
		case len(d.SyncGameActions) > 0:
			h._fastLoadActions(d.SyncGameActions)
		default:
			// 기타: AI 분석
			h._analysisMajsoulRoundData(d, originJSON)
		}
	}
}

func (h *mjHandler) _loadMajsoulRecordBaseInfo(majsoulRecordUUID string) error {
	baseInfo, ok := h.majsoulRecordMap[majsoulRecordUUID]
	if !ok {
		return fmt.Errorf("오류: 작혼 패보 %s 를 찾을 수 없습니다", majsoulRecordUUID)
	}

	// 현재 보고 있는 패보를 표시한다
	h.majsoulCurrentRecordUUID = majsoulRecordUUID
	clearConsole()
	fmt.Printf("작혼 패보 분석 중: %s", baseInfo.String())

	// 고역 모드를 표시한다
	isGuyiMode := baseInfo.Config.isGuyiMode()
	util.SetConsiderOldYaku(isGuyiMode)
	if isGuyiMode {
		fmt.Println()
		color.HiGreen("고역 모드가 켜졌습니다")
	}

	return nil
}

func (h *mjHandler) _loadLiveAction(action *majsoulRecordAction, isFast bool) error {
	if debugMode {
		fmt.Println("[_loadLiveAction] 수신", action, isFast)
	}

	newActions, err := h.majsoulCurrentRoundActions.append(action)
	if err != nil {
		return err
	}
	h.majsoulCurrentRoundActions = newActions

	h.majsoulRoundData.skipOutput = isFast
	h._analysisMajsoulRoundData(action.Action, "")
	return nil
}

func (h *mjHandler) _analysisMajsoulRoundData(data *majsoulMessage, originJSON string) {
	//if originJSON == "{}" {
	//	return
	//}
	h.majsoulRoundData.msg = data
	h.majsoulRoundData.originJSON = originJSON
	if err := h.majsoulRoundData.analysis(); err != nil {
		h.logError(err)
	}
}

func (h *mjHandler) _fastLoadActions(actions []*majsoulRecordAction) {
	if len(actions) == 0 {
		return
	}
	fastRecordEnd := util.MaxInt(0, len(actions)-3)
	h.majsoulRoundData.skipOutput = true
	// 마지막 세 번의 갱신을 남겨 두어 화면 갱신을 보장한다
	for _, action := range actions[:fastRecordEnd] {
		h._analysisMajsoulRoundData(action.Action, "")
	}
	h.majsoulRoundData.skipOutput = false
	for _, action := range actions[fastRecordEnd:] {
		h._analysisMajsoulRoundData(action.Action, "")
	}
}

func (h *mjHandler) _onRecordClick(clickAction string, clickActionIndex int, fastRecordTo int) {
	if debugMode {
		fmt.Println("[_onRecordClick] 수신", clickAction, clickActionIndex, fastRecordTo)
	}

	analysisCache := getCurrentAnalysisCache()

	switch clickAction {
	case "nextStep", "update":
		newActionIndex := h.majsoulCurrentActionIndex + 1
		if newActionIndex >= len(h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex]) {
			return
		}
		h.majsoulCurrentActionIndex = newActionIndex
	case "nextRound":
		h.majsoulCurrentRoundIndex = (h.majsoulCurrentRoundIndex + 1) % len(h.majsoulCurrentRecordActionsList)
		h.majsoulCurrentActionIndex = 0
		go analysisCache.runMajsoulRecordAnalysisTask(h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex])
	case "preRound":
		h.majsoulCurrentRoundIndex = (h.majsoulCurrentRoundIndex - 1 + len(h.majsoulCurrentRecordActionsList)) % len(h.majsoulCurrentRecordActionsList)
		h.majsoulCurrentActionIndex = 0
		go analysisCache.runMajsoulRecordAnalysisTask(h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex])
	case "jumpRound":
		h.majsoulCurrentRoundIndex = clickActionIndex % len(h.majsoulCurrentRecordActionsList)
		h.majsoulCurrentActionIndex = 0
		go analysisCache.runMajsoulRecordAnalysisTask(h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex])
	case "nextXun", "preXun", "jumpXun", "preStep", "jumpToLastRoundXun":
		if clickAction == "jumpToLastRoundXun" {
			h.majsoulCurrentRoundIndex = (h.majsoulCurrentRoundIndex - 1 + len(h.majsoulCurrentRecordActionsList)) % len(h.majsoulCurrentRecordActionsList)
			go analysisCache.runMajsoulRecordAnalysisTask(h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex])
		}

		h.majsoulRoundData.skipOutput = true
		currentRoundActions := h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex]
		startActionIndex := 0
		endActionIndex := fastRecordTo
		if clickAction == "nextXun" {
			startActionIndex = h.majsoulCurrentActionIndex + 1
		}
		if debugMode {
			fmt.Printf("패보 동작 빠른 처리: 국 %d 동작 %d-%d\n", h.majsoulCurrentRoundIndex, startActionIndex, endActionIndex)
		}
		for i, action := range currentRoundActions[startActionIndex : endActionIndex+1] {
			if debugMode {
				fmt.Printf("패보 동작 빠른 처리: 국 %d 동작 %d\n", h.majsoulCurrentRoundIndex, startActionIndex+i)
			}
			h._analysisMajsoulRoundData(action.Action, "")
		}
		h.majsoulRoundData.skipOutput = false

		h.majsoulCurrentActionIndex = endActionIndex + 1
	default:
		return
	}

	if debugMode {
		fmt.Printf("패보 동작 처리: 국 %d 동작 %d\n", h.majsoulCurrentRoundIndex, h.majsoulCurrentActionIndex)
	}
	action := h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex][h.majsoulCurrentActionIndex]
	h._analysisMajsoulRoundData(action.Action, "")

	if action.Name == "RecordHule" || action.Name == "RecordLiuJu" || action.Name == "RecordNoTile" {
		// 화료/유국 애니메이션을 재생하고 다음 국으로 넘어가거나 종료 애니메이션을 표시한다
		h.majsoulCurrentRoundIndex++
		h.majsoulCurrentActionIndex = 0
		if h.majsoulCurrentRoundIndex == len(h.majsoulCurrentRecordActionsList) {
			h.majsoulCurrentRoundIndex = 0
			return
		}

		time.Sleep(time.Second)

		actions := h.majsoulCurrentRecordActionsList[h.majsoulCurrentRoundIndex]
		go analysisCache.runMajsoulRecordAnalysisTask(actions)
		// 다음 국의 시작 정보를 분석한다
		data := actions[h.majsoulCurrentActionIndex].Action
		h._analysisMajsoulRoundData(data, "")
	}
}

var h *mjHandler

func getMajsoulCurrentRecordUUID() string {
	return h.majsoulCurrentRecordUUID
}

func runServer(isHTTPS bool, port int) (err error) {
	e := echo.New()

	// echo.Echo와 http.Server가 콘솔에 출력하는 정보를 제거한다
	e.HideBanner = true
	e.HidePort = true
	e.StdLogger = stdLog.New(ioutil.Discard, "", 0)

	// 기본값은 log.ERROR
	e.Logger.SetLevel(log.INFO)

	// 로그를 log/gamedata-xxx.log로 출력하도록 설정한다
	filePath, err := newLogFilePath()
	if err != nil {
		return
	}
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	e.Logger.SetOutput(logFile)

	e.Logger.Info("============================================================================================")
	e.Logger.Info("서비스 시작")

	h = &mjHandler{
		log: e.Logger,

		tenhouMessageReceiver: tenhou.NewMessageReceiver(),
		tenhouRoundData:       &tenhouRoundData{isRoundEnd: true},
		majsoulMessageQueue:   make(chan []byte, 100),
		majsoulRawQueue:       make(chan []byte, 100),
		majsoulRawRequests:    map[uint16]*majsoulRawRequest{},
		majsoulRoundData:      &majsoulRoundData{selfSeat: -1},
		majsoulRecordMap:      map[string]*majsoulRecordBaseInfo{},
	}
	h.tenhouRoundData.roundData = newGame(h.tenhouRoundData)
	h.majsoulRoundData.roundData = newGame(h.majsoulRoundData)

	go h.runAnalysisTenhouMessageTask()
	go h.runAnalysisMajsoulMessageTask()
	go h.runAnalysisMajsoulRawMessageTask()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.GET("/", h.index)
	e.POST("/debug", h.index)
	e.POST("/analysis", h.analysis)
	e.POST("/tenhou", h.analysisTenhou)
	e.POST("/majsoul", h.analysisMajsoul)
	e.POST("/majsoul-raw", h.analysisMajsoulRaw)

	// code.js도 이 포트를 사용한다
	if port == 0 {
		port = defaultPort
	}
	addr := ":" + strconv.Itoa(port)
	if !isHTTPS {
		e.POST("/", h.analysisTenhou)
		err = e.Start(addr)
	} else {
		e.POST("/", h.analysisMajsoul)
		err = startTLS(e, addr)
	}
	if err != nil {
		// 포트 점유 오류인지 확인한다
		if opErr, ok := err.(*net.OpError); ok && opErr.Op == "listen" {
			if syscallErr, ok := opErr.Err.(*os.SyscallError); ok && syscallErr.Syscall == "bind" {
				color.HiRed(addr + " 포트가 이미 사용 중이라 프로그램을 시작할 수 없습니다. 이미 실행 중인지 확인하세요.")
			}
		}
		return
	}
	return nil
}

func startTLS(e *echo.Echo, address string) (err error) {
	s := e.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.Certificates = make([]tls.Certificate, 1)
	s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair("res/selfsigned.crt", "res/selfsigned.key")
	if err != nil {
		return
	}

	s.Addr = address
	if !e.DisableHTTP2 {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, "h2")
	}
	return e.StartServer(e.TLSServer)
}
