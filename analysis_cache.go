package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/fatih/color"
)

type analysisOpType int

const (
	analysisOpTypeTsumo     analysisOpType = iota
	analysisOpTypeChiPonKan                // 치, 퐁, 명깡
	analysisOpTypeKan                      // 가깡, 안깡
)

// TODO: "여기서는 후로해야 하며 넘기면 안 됨"을 안내

type analysisCache struct {
	analysisOpType analysisOpType

	selfDiscardTile int
	//isSelfDiscardRedFive bool
	selfDiscardTileRisk float64
	isRiichiWhenDiscard bool
	meldType            int

	// 손패의 어떤 패로 후로할지. 비어 있으면 후로하지 않고 넘긴다.
	selfOpenTiles []int

	aiAttackDiscardTile      int
	aiDefenceDiscardTile     int
	aiAttackDiscardTileRisk  float64
	aiDefenceDiscardTileRisk float64

	tenpaiRate []float64 // TODO: 상대 세 명의 텐파이율
}

type roundAnalysisCache struct {
	isStart bool
	isEnd   bool
	cache   []*analysisCache

	analysisCacheBeforeChiPon *analysisCache
}

func (rc *roundAnalysisCache) print() {
	const (
		baseInfo  = "도우미가 추천 타패를 계산 중입니다. 잠시만 기다려 주세요... (계산 결과는 참고용입니다)"
		emptyInfo = "--"
		sep       = "  "
	)

	done := rc != nil && rc.isEnd
	if !done {
		color.HiGreen(baseInfo)
	} else {
		// 마지막 동작이 쯔모인지 확인하고, 쯔모라면 추천을 제거한다.
		if len(rc.cache) > 0 {
			latestCache := rc.cache[len(rc.cache)-1]
			if latestCache.selfDiscardTile == -1 {
				latestCache.aiAttackDiscardTile = -1
				latestCache.aiDefenceDiscardTile = -1
			}
		}
	}

	fmt.Print("순목　　")
	if done {
		for i := range rc.cache {
			fmt.Printf("%s%2d", sep, i+1)
		}
	}
	fmt.Println()

	printTileInfo := func(tile int, risk float64, suffix string) {
		info := emptyInfo
		if tile != -1 {
			info = util.Mahjong[tile]
		}
		fmt.Print(sep)
		if info == emptyInfo || risk < 5 {
			fmt.Print(info)
		} else {
			color.New(getNumRiskColor(risk)).Print(info)
		}
		fmt.Print(suffix)
	}

	fmt.Print("자신 타패")
	if done {
		for i, c := range rc.cache {
			suffix := ""
			if c.isRiichiWhenDiscard {
				suffix = "[리치]"
			} else if c.selfDiscardTile == -1 && i == len(rc.cache)-1 {
				//suffix = "[쯔모]"
				// TODO: 유국
			}
			printTileInfo(c.selfDiscardTile, c.selfDiscardTileRisk, suffix)
		}
	}
	fmt.Println()

	fmt.Print("공격 추천")
	if done {
		for _, c := range rc.cache {
			printTileInfo(c.aiAttackDiscardTile, c.aiAttackDiscardTileRisk, "")
		}
	}
	fmt.Println()

	fmt.Print("수비 추천")
	if done {
		for _, c := range rc.cache {
			printTileInfo(c.aiDefenceDiscardTile, c.aiDefenceDiscardTileRisk, "")
		}
	}
	fmt.Println()

	fmt.Println()
}

// 실제 버림패(쯔모 후 또는 후로 후)
func (rc *roundAnalysisCache) addSelfDiscardTile(tile int, risk float64, isRiichiWhenDiscard bool) {
	latestCache := rc.cache[len(rc.cache)-1]
	latestCache.selfDiscardTile = tile
	latestCache.selfDiscardTileRisk = risk
	latestCache.isRiichiWhenDiscard = isRiichiWhenDiscard
}

// 쯔모 시 타패 추천
func (rc *roundAnalysisCache) addAIDiscardTileWhenDrawTile(attackTile int, defenceTile int, attackTileRisk float64, defenceDiscardTileRisk float64) {
	// 쯔모, 순목 +1
	rc.cache = append(rc.cache, &analysisCache{
		analysisOpType:           analysisOpTypeTsumo,
		selfDiscardTile:          -1,
		aiAttackDiscardTile:      attackTile,
		aiDefenceDiscardTile:     defenceTile,
		aiAttackDiscardTileRisk:  attackTileRisk,
		aiDefenceDiscardTileRisk: defenceDiscardTileRisk,
	})
	rc.analysisCacheBeforeChiPon = nil
}

// 가깡, 안깡
func (rc *roundAnalysisCache) addKan(meldType int) {
	// latestCache는 쯔모 상태다.
	latestCache := rc.cache[len(rc.cache)-1]
	latestCache.analysisOpType = analysisOpTypeKan
	latestCache.meldType = meldType
	// 깡 후 다시 쯔모하므로 순목 +1
}

// 치, 퐁, 명깡
func (rc *roundAnalysisCache) addChiPonKan(meldType int) {
	if meldType == meldTypeMinkan {
		// 명깡은 일단 무시한다. 순목은 여기서 올리지 않고 쯔모 때 +1
		return
	}
	// 순목 +1
	var newCache *analysisCache
	if rc.analysisCacheBeforeChiPon != nil {
		newCache = rc.analysisCacheBeforeChiPon // addPossibleChiPonKan 참고
		newCache.analysisOpType = analysisOpTypeChiPonKan
		newCache.meldType = meldType
		rc.analysisCacheBeforeChiPon = nil
	} else {
		// 이 코드는 보통 실행되지 않아야 한다.
		if debugMode {
			panic("rc.analysisCacheBeforeChiPon == nil")
		}
		newCache = &analysisCache{
			analysisOpType:       analysisOpTypeChiPonKan,
			selfDiscardTile:      -1,
			aiAttackDiscardTile:  -1,
			aiDefenceDiscardTile: -1,
			meldType:             meldType,
		}
	}
	rc.cache = append(rc.cache, newCache)
}

// 치, 퐁, 깡, 넘기기
func (rc *roundAnalysisCache) addPossibleChiPonKan(attackTile int, attackTileRisk float64) {
	rc.analysisCacheBeforeChiPon = &analysisCache{
		analysisOpType:          analysisOpTypeChiPonKan,
		selfDiscardTile:         -1,
		aiAttackDiscardTile:     attackTile,
		aiDefenceDiscardTile:    -1,
		aiAttackDiscardTileRisk: attackTileRisk,
	}
}

//

type gameAnalysisCache struct {
	// 국 수, 본장 수
	wholeGameCache [][]*roundAnalysisCache

	majsoulRecordUUID string

	selfSeat int
}

func newGameAnalysisCache(majsoulRecordUUID string, selfSeat int) *gameAnalysisCache {
	cache := make([][]*roundAnalysisCache, 3*4) // 최대 서4국
	for i := range cache {
		cache[i] = make([]*roundAnalysisCache, 100) // 연장 최대치
	}
	return &gameAnalysisCache{
		wholeGameCache:    cache,
		majsoulRecordUUID: majsoulRecordUUID,
		selfSeat:          selfSeat,
	}
}

//

// TODO: struct로 리팩터링
var (
	_analysisCacheList = make([]*gameAnalysisCache, 4)
	_currentSeat       int
)

func resetAnalysisCache() {
	_analysisCacheList = make([]*gameAnalysisCache, 4)
}

func setAnalysisCache(analysisCache *gameAnalysisCache) {
	_analysisCacheList[analysisCache.selfSeat] = analysisCache
	_currentSeat = analysisCache.selfSeat
}

func getAnalysisCache(seat int) *gameAnalysisCache {
	if seat == -1 {
		return nil
	}
	return _analysisCacheList[seat]
}

func getCurrentAnalysisCache() *gameAnalysisCache {
	return getAnalysisCache(_currentSeat)
}

func (c *gameAnalysisCache) runMajsoulRecordAnalysisTask(actions majsoulRoundActions) error {
	// 첫 action에서 국과 본장을 꺼낸다.
	if len(actions) == 0 {
		return fmt.Errorf("데이터 이상: 이 국의 데이터가 비어 있습니다")
	}

	newRoundAction := actions[0]
	data := newRoundAction.Action
	roundNumber := 4*(*data.Chang) + *data.Ju
	ben := *data.Ben
	roundCache := c.wholeGameCache[roundNumber][ben] // TODO: 원자적 연산 사용 검토
	if roundCache == nil {
		roundCache = &roundAnalysisCache{isStart: true}
		if debugMode {
			fmt.Println("도우미가 추천 타패를 계산 중입니다... roundCache 생성")
		}
		c.wholeGameCache[roundNumber][ben] = roundCache
	} else if roundCache.isStart {
		if debugMode {
			fmt.Println("다시 계산할 필요 없음")
		}
		return nil
	}

	// 자신의 버림패를 순회하며 버리기 직전 동작을 찾는다.
	// 쯔모 동작이면 해당 시점의 AI 공격 타패와 수비 타패를 계산한다.
	// 후로 동작이면 해당 시점의 AI 공격 타패를 계산하고(없으면 -1), 수비 타패는 -1로 둔다.
	// TODO: 플레이어는 넘겼지만 AI는 후로해야 한다고 판단하는 경우?
	majsoulRoundData := &majsoulRoundData{selfSeat: c.selfSeat} // 새 majsoulRoundData로 계산하므로 데이터 충돌이 없다.
	majsoulRoundData.roundData = newGame(majsoulRoundData)
	majsoulRoundData.roundData.gameMode = gameModeRecordCache
	majsoulRoundData.skipOutput = true
	for i, action := range actions[:len(actions)-1] {
		if c.majsoulRecordUUID != getMajsoulCurrentRecordUUID() {
			if debugMode {
				fmt.Println("사용자가 해당 패보를 나갔습니다")
			}
			// 불필요한 계산을 줄이기 위해 조기 종료
			return nil
		}
		if debugMode {
			fmt.Println("도우미가 추천 타패를 계산 중입니다... action", i)
		}
		majsoulRoundData.msg = action.Action
		majsoulRoundData.analysis()
	}
	roundCache.isEnd = true

	if c.majsoulRecordUUID != getMajsoulCurrentRecordUUID() {
		if debugMode {
			fmt.Println("사용자가 해당 패보를 나갔습니다")
		}
		return nil
	}

	clearConsole()
	roundCache.print()

	return nil
}
