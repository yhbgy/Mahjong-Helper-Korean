package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
	"sort"
	"time"
)

type majsoulMessage struct {
	Action string `json:"_action"`

	// 서버 사용자 데이터베이스의 ID. 값이 작을수록 가입 시점이 빠르다
	AccountID int `json:"account_id"`

	// 친구 목록
	Friends lq.FriendList `json:"friends"`

	// 새로 가져온 패보 기본 정보 목록
	RecordBaseInfoList []*majsoulRecordBaseInfo `json:"record_list"`

	// 공유된 패보 기본 정보
	SharedRecordBaseInfo *majsoulRecordBaseInfo `json:"shared_record_base_info"`

	// 현재 보고 있는 패보의 UUID
	CurrentRecordUUID string `json:"current_record_uuid"`

	// 현재 보고 있는 패보의 모든 조작
	RecordActions []*majsoulRecordAction `json:"record_actions"`

	// 웹에서 플레이어가 수행한 클릭 조작(웹이 응답한 것)
	RecordClickAction      string `json:"record_click_action"`
	RecordClickActionIndex int    `json:"record_click_action_index"`
	FastRecordTo           int    `json:"fast_record_to"` // 설명

	// 관전
	LiveBaseInfo   *majsoulLiveRecordBaseInfo `json:"live_head"`
	LiveFastAction *majsoulRecordAction       `json:"live_fast_action"`
	LiveAction     *majsoulRecordAction       `json:"live_action"`

	// 좌석 변경
	ChangeSeatTo *int `json:"change_seat_to"`

	// 게임 재접속 시 받은 데이터
	SyncGameActions []*majsoulRecordAction `json:"sync_game_actions"`

	// ResAuthGame
	// {"seat_list":[x,x,x,x],"is_game_start":false,"game_config":{"category":1,"mode":{"mode":1,"ai":true,"detail_rule":{"time_fixed":60,"time_add":0,"dora_count":3,"shiduan":1,"init_point":25000,"fandian":30000,"bianjietishi":true,"ai_level":1,"fanfu":1}},"meta":{"room_id":18269}},"ready_id_list":[0,0,0]}
	IsGameStart *bool              `json:"is_game_start"` // 설명
	SeatList    []int              `json:"seat_list"`
	ReadyIDList []int              `json:"ready_id_list"`
	GameConfig  *majsoulGameConfig `json:"game_config"`

	// NotifyPlayerLoadGameReady
	//ReadyIDList []int `json:"ready_id_list"`

	// ActionNewRound
	// {"chang":0,"ju":0,"ben":0,"tiles":["1m","3m","7m","3p","6p","7p","6s","1z","1z","2z","3z","4z","7z"],"dora":"6m","scores":[25000,25000,25000,25000],"liqibang":0,"al":false,"md5":"","left_tile_count":69}
	MD5      string      `json:"md5"`
	Chang    *int        `json:"chang"`
	Ju       *int        `json:"ju"`
	Ben      *int        `json:"ben"`
	Tiles    interface{} `json:"tiles"` // 설명
	Dora     string      `json:"dora"`
	Scores   interface{} `json:"scores"`
	Liqibang *int        `json:"liqibang"`

	// RecordNewRound
	Tiles0 []string `json:"tiles0"`
	Tiles1 []string `json:"tiles1"`
	Tiles2 []string `json:"tiles2"`
	Tiles3 []string `json:"tiles3"`

	// ActionDealTile
	// {"seat":1,"tile":"5m","left_tile_count":23,"operation":{"seat":1,"operation_list":[{"type":1}],"time_add":0,"time_fixed":60000},"zhenting":false}
	// 타가 안깡 뒤의 쯔모
	// {"seat":1,"left_tile_count":3,"doras":["7m","0p"],"zhenting":false}
	Seat          *int     `json:"seat"`
	Tile          string   `json:"tile"`
	Doras         []string `json:"doras"` // 설명
	LeftTileCount *int     `json:"left_tile_count"`

	// ActionDiscardTile
	// {"seat":0,"tile":"5z","is_liqi":false,"moqie":true,"zhenting":false,"is_wliqi":false}
	// {"seat":0,"tile":"1z","is_liqi":false,"operation":{"seat":1,"operation_list":[{"type":3,"combination":["1z|1z"]}],"time_add":0,"time_fixed":60000},"moqie":false,"zhenting":false,"is_wliqi":false}
	// 치, 퐁, 화료
	// {"seat":0,"tile":"6p","is_liqi":false,"operation":{"seat":1,"operation_list":[{"type":2,"combination":["7p|8p"]},{"type":3,"combination":["6p|6p"]},{"type":9}],"time_add":0,"time_fixed":60000},"moqie":false,"zhenting":true,"is_wliqi":false}
	IsLiqi    *bool     `json:"is_liqi"`
	IsWliqi   *bool     `json:"is_wliqi"`
	Moqie     *bool     `json:"moqie"`
	Operation *struct{} `json:"operation"`

	// ActionChiPengGang || ActionAnGangAddGang
	// 타가 치 {"seat":0,"type":0,"tiles":["2s","3s","4s"],"froms":[0,0,3],"zhenting":false}
	// 타가 퐁 {"seat":1,"type":1,"tiles":["1z","1z","1z"],"froms":[1,1,0],"operation":{"seat":1,"operation_list":[{"type":1,"combination":["1z"]}],"time_add":0,"time_fixed":60000},"zhenting":false,"tingpais":[{"tile":"4m","zhenting":false,"infos":[{"tile":"6s","haveyi":true},{"tile":"6p","haveyi":true}]},{"tile":"7m","zhenting":false,"infos":[{"tile":"6s","haveyi":true},{"tile":"6p","haveyi":true}]}]}
	// 타가 대명깡 {"seat":2,"type":2,"tiles":["3z","3z","3z","3z"],"froms":[2,2,2,0],"zhenting":false}
	// 타가 가깡 {"seat":2,"type":2,"tiles":"3z"}
	// 타가 안깡 {"seat":2,"type":3,"tiles":"3s"}
	Type  int   `json:"type"`
	Froms []int `json:"froms"`

	// ActionLiqi
	LiqiFailed *bool `json:"failed"`

	// ActionHule
	Hules []struct {
		Seat          int  `json:"seat"`
		Zimo          bool `json:"zimo"`
		PointRong     int  `json:"point_rong"`
		PointZimoQin  int  `json:"point_zimo_qin"`
		PointZimoXian int  `json:"point_zimo_xian"`
	} `json:"hules"`

	// ActionLiuJu
	LiuJuType *int `json:"liuju_type"`
	// {"liujumanguan":false,"players":[{"tingpai":true,"hand":["3s","3s","4s","5s","6s","1z","1z","7z","7z","7z"],"tings":[{"tile":"1z","haveyi":true},{"tile":"3s","haveyi":true}]},{"tingpai":false},{"tingpai":false},{"tingpai":true,"hand":["4m","0m","6m","6m","6m","4s","4s","4s","5s","7s"],"tings":[{"tile":"6s","haveyi":true}]}],"scores":[{"old_scores":[23000,29000,24000,24000],"delta_scores":[1500,-1500,-1500,1500]}],"gameend":false}
	//Liujumanguan *bool `json:"liujumanguan"`
	//Players *struct{ } `json:"players"`
	//Gameend      *bool `json:"gameend"`

	// ActionBabei
}

const (
	majsoulMeldTypeChi = iota
	majsoulMeldTypePon
	majsoulMeldTypeMinkanOrKakan
	majsoulMeldTypeAnkan
)

type majsoulRoundData struct {
	*roundData

	originJSON string
	msg        *majsoulMessage

	selfSeat int // 설명
}

func (d *majsoulRoundData) fatalParse(info string, msg string) {
	panic(fmt.Sprintln(info, len(msg), msg, []byte(msg)))
}

func (d *majsoulRoundData) normalTiles(tiles interface{}) (majsoulTiles []string) {
	_tiles, ok := tiles.([]interface{})
	if !ok {
		_tile, ok := tiles.(string)
		if !ok {
			panic(fmt.Sprintln("[normalTiles] 파싱 오류", tiles))
		}
		return []string{_tile}
	}

	majsoulTiles = make([]string, len(_tiles))
	for i, _tile := range _tiles {
		_t, ok := _tile.(string)
		if !ok {
			panic(fmt.Sprintln("[normalTiles] 파싱 오류", tiles))
		}
		majsoulTiles[i] = _t
	}
	return majsoulTiles
}

func (d *majsoulRoundData) parseWho(seat int) int {
	// 0=자신, 1=하가, 2=대가, 3=상가로 변환한다
	// 3인마작과 4인마작 모두에 적용된다
	who := (seat - d.selfSeat + 4) % 4
	return who
}

func (d *majsoulRoundData) mustParseMajsoulTile(humanTile string) (tile34 int, isRedFive bool) {
	tile34, isRedFive, err := util.StrToTile34(humanTile)
	if err != nil {
		panic(err)
	}
	return
}

func (d *majsoulRoundData) normalInts(values interface{}) (ints []int) {
	switch typed := values.(type) {
	case nil:
		return nil
	case []interface{}:
		ints = make([]int, 0, len(typed))
		for _, value := range typed {
			switch v := value.(type) {
			case float64:
				ints = append(ints, int(v))
			case int:
				ints = append(ints, v)
			}
		}
	case []int:
		ints = append(ints, typed...)
	}
	return
}

func (d *majsoulRoundData) ParseRoundState() (scores []int, liqibang int, ok bool) {
	msg := d.msg
	scores = d.normalInts(msg.Scores)
	if msg.Liqibang != nil {
		liqibang = *msg.Liqibang
	}
	return scores, liqibang, len(scores) > 0 || msg.Liqibang != nil
}

func (d *majsoulRoundData) mustParseMajsoulTiles(majsoulTiles []string) (tiles []int, numRedFive int) {
	tiles = make([]int, len(majsoulTiles))
	for i, majsoulTile := range majsoulTiles {
		var isRedFive bool
		tiles[i], isRedFive = d.mustParseMajsoulTile(majsoulTile)
		if isRedFive {
			numRedFive++
		}
	}
	return
}

func (d *majsoulRoundData) isNewDora(doras []string) bool {
	return len(doras) > len(d.doraIndicators)
}

func (d *majsoulRoundData) GetDataSourceType() int {
	return dataSourceTypeMajsoul
}

func (d *majsoulRoundData) GetSelfSeat() int {
	return d.selfSeat
}

func (d *majsoulRoundData) GetMessage() string {
	return d.originJSON
}

func (d *majsoulRoundData) SkipMessage() bool {
	msg := d.msg

	// 계정이 없으면 skip
	if gameConf.currentActiveMajsoulAccountID == -1 {
		return true
	}

	// TODO: 리팩터링
	if msg.SeatList != nil {
		// 고역 모드 특수 처리
		isGuyiMode := msg.GameConfig.isGuyiMode()
		util.SetConsiderOldYaku(isGuyiMode)
		if isGuyiMode {
			color.HiGreen("고역 모드가 켜졌습니다")
			time.Sleep(2 * time.Second)
		}
	} else {
		// msg.SeatList는 nil이어야 한다
		if msg.ReadyIDList != nil {
			// 준비 정보를 출력한다
			fmt.Printf("플레이어 준비 대기 (%d/%d) %v\n", len(msg.ReadyIDList), d.playerNumber, msg.ReadyIDList)
		}
	}

	return false
}

func (d *majsoulRoundData) IsLogin() bool {
	msg := d.msg
	return msg.AccountID > 0 || msg.SeatList != nil
}

func (d *majsoulRoundData) HandleLogin() {
	msg := d.msg

	if accountID := msg.AccountID; accountID > 0 {
		gameConf.addMajsoulAccountID(accountID)
		if accountID != gameConf.currentActiveMajsoulAccountID {
			printAccountInfo(accountID)
			gameConf.setMajsoulAccountID(accountID)
		}
		return
	}

	// 대전 ID 목록에서 계정 ID를 가져온다
	if seatList := msg.SeatList; seatList != nil {
		// 캐시된 계정 ID를 찾아본다
		for _, accountID := range seatList {
			if accountID > 0 && gameConf.isIDExist(accountID) {
				// 찾았으면 현재 사용할 계정 ID를 갱신한다
				if gameConf.currentActiveMajsoulAccountID != accountID {
					printAccountInfo(accountID)
					gameConf.setMajsoulAccountID(accountID)
				}
				return
			}
		}

		// 캐시 ID를 찾지 못함
		if gameConf.currentActiveMajsoulAccountID > 0 {
			color.HiRed("아직 계정 ID를 가져오지 못했습니다. 웹페이지를 새로고침하거나 CPU전을 한 판 시작하세요. (오류 정보: 계정 ID %d가 대전 목록 %v 안에 없습니다)", gameConf.currentActiveMajsoulAccountID, msg.SeatList)
			return
		}

		// AI전인지 판단하고, AI전이면 계정 ID를 가져온다
		if !util.InInts(0, msg.SeatList) {
			return
		}
		for _, accountID := range msg.SeatList {
			if accountID > 0 {
				gameConf.addMajsoulAccountID(accountID)
				printAccountInfo(accountID)
				gameConf.setMajsoulAccountID(accountID)
				return
			}
		}
	}
}

func (d *majsoulRoundData) IsInit() bool {
	msg := d.msg
	// ResAuthGame || ActionNewRound RecordNewRound
	return msg.IsGameStart != nil || msg.MD5 != ""
}

func (d *majsoulRoundData) ParseInit() (roundNumber int, benNumber int, dealer int, doraIndicators []int, handTiles []int, numRedFives []int) {
	msg := d.msg

	if playerNumber := len(msg.SeatList); playerNumber >= 3 {
		d.playerNumber = playerNumber
		// 자신의 초기 좌석을 가져온다: 0-첫 국 동가, 1-첫 국 남가, 2-첫 국 서가, 3-첫 국 북가
		for i, accountID := range msg.SeatList {
			if accountID == gameConf.currentActiveMajsoulAccountID {
				d.selfSeat = i
				break
			}
		}
		// dealer: 0=자신, 1=하가, 2=대가, 3=상가
		dealer = (4 - d.selfSeat) % 4
		return
	} else if len(msg.Tiles2) > 0 {
		if len(msg.Tiles3) > 0 {
			d.playerNumber = 4
		} else {
			d.playerNumber = 3
		}
	}
	dealer = -1

	roundNumber = 4*(*msg.Chang) + *msg.Ju
	benNumber = *msg.Ben
	if msg.Dora != "" {
		doraIndicator, _ := d.mustParseMajsoulTile(msg.Dora)
		doraIndicators = append(doraIndicators, doraIndicator)
	} else {
		for _, dora := range msg.Doras {
			doraIndicator, _ := d.mustParseMajsoulTile(dora)
			doraIndicators = append(doraIndicators, doraIndicator)
		}
	}
	numRedFives = make([]int, 3)

	var majsoulTiles []string
	if msg.Tiles != nil { // 설명
		majsoulTiles = d.normalTiles(msg.Tiles)
	} else { // 설명
		majsoulTiles = [][]string{msg.Tiles0, msg.Tiles1, msg.Tiles2, msg.Tiles3}[d.selfSeat]
	}
	for _, majsoulTile := range majsoulTiles {
		tile, isRedFive := d.mustParseMajsoulTile(majsoulTile)
		handTiles = append(handTiles, tile)
		if isRedFive {
			numRedFives[tile/9]++
		}
	}

	return
}

func (d *majsoulRoundData) IsSelfDraw() bool {
	msg := d.msg
	// ActionDealTile RecordDealTile
	return msg.Seat != nil && msg.Tile != "" && msg.Moqie == nil && d.parseWho(*msg.Seat) == 0
}

func (d *majsoulRoundData) ParseSelfDraw() (tile int, isRedFive bool, kanDoraIndicator int) {
	msg := d.msg
	tile, isRedFive = d.mustParseMajsoulTile(msg.Tile)
	kanDoraIndicator = -1
	if d.isNewDora(msg.Doras) {
		kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	}
	return
}

func (d *majsoulRoundData) IsDiscard() bool {
	msg := d.msg
	// ActionDiscardTile RecordDiscardTile
	return msg.IsLiqi != nil
}

func (d *majsoulRoundData) ParseDiscard() (who int, discardTile int, isRedFive bool, isTsumogiri bool, isReach bool, canBeMeld bool, kanDoraIndicator int) {
	msg := d.msg
	who = d.parseWho(*msg.Seat)
	discardTile, isRedFive = d.mustParseMajsoulTile(msg.Tile)
	isTsumogiri = *msg.Moqie
	isReach = *msg.IsLiqi
	if msg.IsWliqi != nil && !isReach { // 설명
		isReach = *msg.IsWliqi
	}
	canBeMeld = msg.Operation != nil // 설명
	kanDoraIndicator = -1
	if d.isNewDora(msg.Doras) {
		kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	}
	return
}

func (d *majsoulRoundData) IsOpen() bool {
	msg := d.msg
	// ActionChiPengGang RecordChiPengGang || ActionAnGangAddGang RecordAnGangAddGang
	return msg.Tiles != nil && len(d.normalTiles(msg.Tiles)) <= 4
}

func (d *majsoulRoundData) ParseOpen() (who int, meld *model.Meld, kanDoraIndicator int) {
	msg := d.msg

	who = d.parseWho(*msg.Seat)

	kanDoraIndicator = -1
	if d.isNewDora(msg.Doras) { // 설명
		kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	}

	var meldType, calledTile int

	majsoulTiles := d.normalTiles(msg.Tiles)
	isSelfKan := len(majsoulTiles) == 1 // 설명
	if isSelfKan {
		majsoulTile := majsoulTiles[0]
		majsoulTiles = []string{majsoulTile, majsoulTile, majsoulTile, majsoulTile}
	}
	meldTiles, numRedFive := d.mustParseMajsoulTiles(majsoulTiles)
	containRedFive := numRedFive > 0
	if len(majsoulTiles) == 4 && meldTiles[0] < 27 && meldTiles[0]%9 == 4 {
		// 설명
		containRedFive = true
	}

	if isSelfKan {
		calledTile = meldTiles[0]
		// msg.Type으로 가깡인지 안깡인지 판단한다
		// 관련 퐁 후로가 있는지로도 가깡/안깡을 판단할 수 있다
		if msg.Type == majsoulMeldTypeMinkanOrKakan {
			meldType = meldTypeKakan // 설명
		} else if msg.Type == majsoulMeldTypeAnkan {
			meldType = meldTypeAnkan // 설명
		}
		meld = &model.Meld{
			MeldType:       meldType,
			Tiles:          meldTiles,
			CalledTile:     calledTile,
			ContainRedFive: containRedFive,
		}
		return
	}

	var rawCalledTile string
	for i, seat := range msg.Froms {
		fromWho := d.parseWho(seat)
		if fromWho != who {
			rawCalledTile = majsoulTiles[i]
		}
	}
	if rawCalledTile == "" {
		panic("데이터 파싱 이상: rawCalledTile을 찾을 수 없음")
	}
	calledTile, redFiveFromOthers := d.mustParseMajsoulTile(rawCalledTile)

	if len(meldTiles) == 3 {
		if meldTiles[0] == meldTiles[1] {
			meldType = meldTypePon // 설명
		} else {
			meldType = meldTypeChi // 설명
			sort.Ints(meldTiles)
		}
	} else if len(meldTiles) == 4 {
		meldType = meldTypeMinkan // 설명
	} else {
		panic("후로 데이터 파싱 실패!")
	}
	meld = &model.Meld{
		MeldType:          meldType,
		Tiles:             meldTiles,
		CalledTile:        calledTile,
		ContainRedFive:    containRedFive,
		RedFiveFromOthers: redFiveFromOthers,
	}
	return
}

func (d *majsoulRoundData) IsReach() bool {
	msg := d.msg
	return msg.Action == "ActionLiqi" && (msg.LiqiFailed == nil || !*msg.LiqiFailed) && msg.Seat != nil
}

func (d *majsoulRoundData) ParseReach() (who int) {
	return d.parseWho(*d.msg.Seat)
}

func (d *majsoulRoundData) IsFuriten() bool {
	return false
}

func (d *majsoulRoundData) IsRoundWin() bool {
	msg := d.msg
	// ActionHule RecordHule
	return msg.Hules != nil
}

func (d *majsoulRoundData) ParseRoundWin() (whos []int, points []int) {
	msg := d.msg

	for _, result := range msg.Hules {
		who := d.parseWho(result.Seat)
		whos = append(whos, d.parseWho(result.Seat))
		point := result.PointRong
		if result.Zimo {
			if who == d.dealer {
				point = 3 * result.PointZimoXian
			} else {
				point = result.PointZimoQin + 2*result.PointZimoXian
			}
			if d.playerNumber == 3 {
				// 쯔모 손실(자식 한 명)
				point -= result.PointZimoXian
			}
		}
		points = append(points, point)
	}
	return
}

func (d *majsoulRoundData) IsRyuukyoku() bool {
	msg := d.msg
	return msg.Action == "ActionNoTile" || msg.Action == "ActionLiuJu"
}

func (d *majsoulRoundData) ParseRyuukyoku() (type_ int, whos []int, points []int) {
	if d.msg.LiuJuType != nil {
		type_ = *d.msg.LiuJuType
	}
	return
}

// 북도라
func (d *majsoulRoundData) IsNukiDora() bool {
	msg := d.msg
	// ActionBaBei RecordBaBei
	return msg.Seat != nil && msg.Moqie != nil && msg.Tile == ""
}

func (d *majsoulRoundData) ParseNukiDora() (who int, isTsumogiri bool) {
	msg := d.msg
	return d.parseWho(*msg.Seat), *msg.Moqie
}

// 이 항목은 마지막에 처리한다
func (d *majsoulRoundData) IsNewDora() bool {
	msg := d.msg
	// ActionDealTile
	return d.isNewDora(msg.Doras)
}

func (d *majsoulRoundData) ParseNewDora() (kanDoraIndicator int) {
	msg := d.msg

	kanDoraIndicator, _ = d.mustParseMajsoulTile(msg.Doras[len(msg.Doras)-1])
	return
}
