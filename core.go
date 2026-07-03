package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
)

type DataParser interface {
	// 설명
	GetDataSourceType() int

	// 설명
	// 설명
	GetSelfSeat() int

	// 설명
	GetMessage() string

	// 설명
	SkipMessage() bool

	// 설명
	IsLogin() bool
	HandleLogin()

	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	IsInit() bool
	ParseInit() (roundNumber int, benNumber int, dealer int, doraIndicators []int, handTiles []int, numRedFives []int)

	// 설명
	// tile: 0-33
	// 설명
	// 설명
	IsSelfDraw() bool
	ParseSelfDraw() (tile int, isRedFive bool, kanDoraIndicator int)

	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	IsDiscard() bool
	ParseDiscard() (who int, discardTile int, isRedFive bool, isTsumogiri bool, isReach bool, canBeMeld bool, kanDoraIndicator int)

	// 설명
	// 설명
	IsOpen() bool
	ParseOpen() (who int, meld *model.Meld, kanDoraIndicator int)

	// 설명
	IsReach() bool
	ParseReach() (who int)

	// 설명
	IsFuriten() bool

	// 설명
	IsRoundWin() bool
	ParseRoundWin() (whos []int, points []int)

	// 설명
	// 설명
	// 설명
	IsRyuukyoku() bool
	ParseRyuukyoku() (type_ int, whos []int, points []int)

	// 설명
	IsNukiDora() bool
	ParseNukiDora() (who int, isTsumogiri bool)

	// 설명
	// 설명
	// kanDoraIndicator: 0-33
	IsNewDora() bool
	ParseNewDora() (kanDoraIndicator int)
}

type roundStateParser interface {
	ParseRoundState() (scores []int, liqibang int, ok bool)
}

type playerInfo struct {
	name string // 설명

	selfWindTile int // 설명

	melds                []*model.Meld // 설명
	meldDiscardsAtGlobal []int
	meldDiscardsAt       []int
	isNaki               bool // 설명

	// 설명
	discardTiles          []int // 설명
	latestDiscardAtGlobal int   // 설명
	earlyOutsideTiles     []int // 설명

	isReached  bool // 설명
	canIppatsu bool // 설명

	reachTileAtGlobal int // 설명
	reachTileAt       int // 설명

	nukiDoraNum int // 설명
}

func newPlayerInfo(name string, selfWindTile int) *playerInfo {
	return &playerInfo{
		name:                  name,
		selfWindTile:          selfWindTile,
		latestDiscardAtGlobal: -1,
		reachTileAtGlobal:     -1,
		reachTileAt:           -1,
	}
}

func modifySanninPlayerInfoList(lst []*playerInfo, roundNumber int) []*playerInfo {
	windToIdxMap := map[int]int{}
	for i, pi := range lst {
		windToIdxMap[pi.selfWindTile] = i
	}

	idxS, idxW, idxN := windToIdxMap[28], windToIdxMap[29], windToIdxMap[30]
	switch roundNumber % 4 {
	case 0:
	case 1:
		// 설명
		lst[idxN].selfWindTile, lst[idxW].selfWindTile = lst[idxW].selfWindTile, lst[idxN].selfWindTile
	case 2:
		// 설명
		lst[idxN].selfWindTile, lst[idxW].selfWindTile, lst[idxS].selfWindTile = lst[idxW].selfWindTile, lst[idxS].selfWindTile, lst[idxN].selfWindTile
	default:
		panic("[modifySanninPlayerInfoList] 코드 오류")
	}
	return lst
}

func (p *playerInfo) doraNum(doraList []int) (doraCount int) {
	for _, meld := range p.melds {
		for _, tile := range meld.Tiles {
			for _, doraTile := range doraList {
				if tile == doraTile {
					doraCount++
				}
			}
		}
		if meld.ContainRedFive {
			doraCount++
		}
	}
	if p.nukiDoraNum > 0 {
		doraCount += p.nukiDoraNum
		// 설명
		for _, doraTile := range doraList {
			if doraTile == 30 {
				doraCount += p.nukiDoraNum
			}
		}
	}
	return
}

//

type roundData struct {
	parser DataParser

	gameMode gameMode

	skipOutput bool

	// 설명
	playerNumber int

	// 설명
	roundNumber int

	// 설명
	benNumber int

	// 설명
	roundWindTile int

	// 설명
	// 설명
	dealer int

	scores   []int
	liqibang int

	pendingSelfNukiDoraReplacement bool
	pendingSelfDrawnNorth          bool

	// 설명
	doraIndicators []int

	// 설명
	counts []int

	// 설명
	// 설명
	numRedFives []int

	// 설명
	leftCounts []int

	// 설명
	// 설명
	// 설명
	globalDiscardTiles []int

	// 설명
	players []*playerInfo
}

func newRoundData(parser DataParser, roundNumber int, benNumber int, dealer int) *roundData {
	// 설명
	const playerNumber = 4
	roundWindTile := 27 + roundNumber/playerNumber
	playerWindTile := make([]int, playerNumber)
	for i := 0; i < playerNumber; i++ {
		playerWindTile[i] = 27 + (playerNumber-dealer+i)%playerNumber
	}
	return &roundData{
		parser:      parser,
		roundNumber: roundNumber,
		benNumber:   benNumber,

		roundWindTile:      roundWindTile,
		dealer:             dealer,
		counts:             make([]int, 34),
		leftCounts:         util.InitLeftTiles34(),
		globalDiscardTiles: []int{},
		players: []*playerInfo{
			newPlayerInfo("자신", playerWindTile[0]),
			newPlayerInfo("하가", playerWindTile[1]),
			newPlayerInfo("대면", playerWindTile[2]),
			newPlayerInfo("상가", playerWindTile[3]),
		},
	}
}

func newGame(parser DataParser) *roundData {
	return newRoundData(parser, 0, 0, 0)
}

// 설명
func (d *roundData) reset(roundNumber int, benNumber int, dealer int) {
	skipOutput := d.skipOutput
	gameMode := d.gameMode
	playerNumber := d.playerNumber
	newData := newRoundData(d.parser, roundNumber, benNumber, dealer)
	newData.skipOutput = skipOutput
	newData.gameMode = gameMode
	newData.playerNumber = playerNumber
	if playerNumber == 3 {
		// 설명
		for i := 1; i <= 7; i++ {
			newData.leftCounts[i] = 0
		}
		newData.players = modifySanninPlayerInfoList(newData.players, roundNumber)
	}
	*d = *newData
}

func (d *roundData) updateRoundStateFromParser() {
	parser, ok := d.parser.(roundStateParser)
	if !ok {
		return
	}
	scores, liqibang, ok := parser.ParseRoundState()
	if !ok {
		return
	}
	if len(scores) > 0 {
		d.scores = scores
	}
	d.liqibang = liqibang
}

func (d *roundData) newGame() {
	d.reset(0, 0, 0)
}

func (d *roundData) descLeftCounts(tile int) {
	d.leftCounts[tile]--
	if d.leftCounts[tile] < 0 {
		info := fmt.Sprintf("데이터 이상: %s 개수가 %d입니다", util.MahjongZH[tile], d.leftCounts[tile])
		if debugMode {
			panic(info)
		} else {
			fmt.Println(info)
		}
	}
}

// 설명
func (d *roundData) newDora(kanDoraIndicator int) {
	d.doraIndicators = append(d.doraIndicators, kanDoraIndicator)
	d.descLeftCounts(kanDoraIndicator)

	if d.skipOutput {
		return
	}

	color.Yellow("깡도라 표시패는 %s", util.MahjongZH[kanDoraIndicator])
}

// 설명
func (d *roundData) doraList() (dl []int) {
	return model.DoraList(d.doraIndicators, d.playerNumber == 3)
}

func (d *roundData) printDiscards() {
	d.printRoundState()
	// 설명
	for i := len(d.players) - 1; i >= 1; i-- {
		if player := d.players[i]; d.playerNumber != 3 || player.selfWindTile != 30 {
			player.printDiscards()
		}
	}
}

// 설명
// 설명
func (d *roundData) analysisTilesRisk() (riList riskInfoList) {
	riList = make(riskInfoList, len(d.players))
	for who := range riList {
		riList[who] = &riskInfo{
			playerNumber: d.playerNumber,
			safeTiles34:  make([]bool, 34),
		}
	}

	// 설명
	for who, player := range d.players {
		if who == 0 {
			// TODO: 설명 보완
			continue
		}

		// 설명
		for _, tile := range normalDiscardTiles(player.discardTiles) {
			riList[who].safeTiles34[tile] = true
		}
		if player.reachTileAtGlobal != -1 {
			// 설명
			for _, tile := range normalDiscardTiles(d.globalDiscardTiles[player.reachTileAtGlobal:]) {
				riList[who].safeTiles34[tile] = true
			}
		} else if player.latestDiscardAtGlobal != -1 {
			// 설명
			// 설명
			for _, tile := range normalDiscardTiles(d.globalDiscardTiles[player.latestDiscardAtGlobal:]) {
				riList[who].safeTiles34[tile] = true
			}
		}

		// 설명
		// 설명
		for _, meld := range player.melds {
			if meld.IsKan() {
				riList[who].safeTiles34[meld.Tiles[0]] = true
			}
		}
	}

	// 설명
	for who, player := range d.players {
		if who == 0 {
			// TODO: 설명 보완
			continue
		}

		// 설명
		turns := util.MinInt(len(player.discardTiles), util.MaxTurns)
		if turns == 0 {
			turns = 1
		}

		// TODO: 설명 보완
		if player.isReached {
			riList[who].tenpaiRate = 100.0
			if player.reachTileAtGlobal < len(d.globalDiscardTiles) { // 설명
				riList[who].isTsumogiriRiichi = d.globalDiscardTiles[player.reachTileAtGlobal] < 0
			}
		} else {
			rate := util.CalcTenpaiRate(player.melds, player.discardTiles, player.meldDiscardsAt)
			if d.playerNumber == 3 {
				rate = util.GetTenpaiRate3(rate)
			}
			riList[who].tenpaiRate = rate
		}

		// 설명
		var ronPoint float64
		switch {
		case player.canIppatsu:
			// 설명
			ronPoint = util.RonPointRiichiIppatsu
		case player.isReached:
			// 설명
			ronPoint = util.RonPointRiichiHiIppatsu
		case player.isNaki:
			// 설명
			doraCount := player.doraNum(d.doraList())
			ronPoint = util.RonPointOtherNakiWithDora(doraCount)
		default:
			// 설명
			ronPoint = util.RonPointDama
		}
		// 설명
		if who == d.dealer {
			ronPoint *= 1.5
		}
		riList[who]._ronPoint = ronPoint

		// 설명
		risk34 := util.CalculateRiskTiles34(turns, riList[who].safeTiles34, d.leftCounts, d.doraList(), d.roundWindTile, player.selfWindTile).
			FixWithEarlyOutside(player.earlyOutsideTiles).
			FixWithPoint(ronPoint)
		riList[who].riskTable = riskTable(risk34)

		// 설명
		if len(player.melds) < 4 {
			riList[who].leftNoSujiTiles = util.CalculateLeftNoSujiTiles(riList[who].safeTiles34, d.leftCounts)
		} else {
			// 설명
		}
	}

	return riList
}

// TODO: 설명 보완
func (d *roundData) isPlayerDaburii(who int) bool {
	// 설명
	for _, p := range d.players {
		if len(p.melds) > 0 {
			return false
		}
		// 설명
		if p.nukiDoraNum > 0 {
			return false
		}
	}
	return d.players[who].reachTileAt == 0
}

// 설명
func (d *roundData) newModelPlayerInfo() *model.PlayerInfo {
	const wannpaiTilesCount = 14
	leftDrawTilesCount := util.CountOfTiles34(d.leftCounts) - (wannpaiTilesCount - len(d.doraIndicators))
	for _, player := range d.players[1:] {
		leftDrawTilesCount -= 13 - 3*len(player.melds)
	}
	if d.playerNumber == 3 {
		leftDrawTilesCount += 13
	}

	melds := []model.Meld{}
	for _, m := range d.players[0].melds {
		melds = append(melds, *m)
	}

	const self = 0
	selfPlayer := d.players[self]

	return &model.PlayerInfo{
		HandTiles34: d.counts,
		Melds:       melds,
		DoraTiles:   d.doraList(),
		NumRedFives: d.numRedFives,

		RoundWindTile: d.roundWindTile,
		SelfWindTile:  selfPlayer.selfWindTile,
		IsParent:      d.dealer == self,
		// FIXME: 설명 보완
		IsRiichi: selfPlayer.isReached,

		DiscardTiles: normalDiscardTiles(selfPlayer.discardTiles),
		LeftTiles34:  d.leftCounts,

		LeftDrawTilesCount: leftDrawTilesCount,

		NukiDoraNum: selfPlayer.nukiDoraNum,
	}
}

func (d *roundData) analysis() error {
	if !debugMode {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("내부 오류:", err)
			}
		}()
	}

	if debugMode {
		if msg := d.parser.GetMessage(); len(msg) > 0 {
			const printLimit = 500
			if len(msg) > printLimit {
				msg = msg[:printLimit]
			}
			fmt.Println("수신", msg)
		}
	}

	// 설명
	if d.parser.IsLogin() {
		d.parser.HandleLogin()
	}

	if d.parser.SkipMessage() {
		return nil
	}

	// 설명
	// TODO: 설명 보완
	if !d.parser.IsInit() && !d.parser.IsRoundWin() && !d.parser.IsRyuukyoku() && d.players[0].isReached {
		return nil
	}

	if debugMode {
		fmt.Println("현재 좌석은", d.parser.GetSelfSeat())
	}

	var currentRoundCache *roundAnalysisCache
	if analysisCache := getAnalysisCache(d.parser.GetSelfSeat()); analysisCache != nil {
		currentRoundCache = analysisCache.wholeGameCache[d.roundNumber][d.benNumber]
	}

	switch {
	case d.parser.IsInit():
		// 설명
		if !debugMode && !d.skipOutput {
			clearConsole()
		}

		roundNumber, benNumber, dealer, doraIndicators, hands, numRedFives := d.parser.ParseInit()
		switch d.parser.GetDataSourceType() {
		case dataSourceTypeTenhou:
			d.reset(roundNumber, benNumber, dealer)
			d.gameMode = gameModeMatch // TODO: 설명 보완
		case dataSourceTypeMajsoul:
			if dealer != -1 { // 설명
				// 설명
				d.reset(0, 0, dealer)
				d.gameMode = gameModeMatch
				fmt.Printf("게임이 곧 시작됩니다. 배정된 좌석: ")
				color.HiGreen(util.MahjongZH[d.players[0].selfWindTile])
				return nil
			} else {
				// 설명
				newDealer := (4 - d.parser.GetSelfSeat() + roundNumber) % 4
				// 설명
				d.reset(roundNumber, benNumber, newDealer)
			}
		default:
			panic("not impl!")
		}
		d.updateRoundStateFromParser()

		// 설명
		if analysisCache := getAnalysisCache(d.parser.GetSelfSeat()); analysisCache != nil {
			currentRoundCache = analysisCache.wholeGameCache[d.roundNumber][d.benNumber]
		}

		d.doraIndicators = doraIndicators
		for _, dora := range doraIndicators {
			d.descLeftCounts(dora)
		}
		for _, tile := range hands {
			d.counts[tile]++
			d.descLeftCounts(tile)
		}
		d.numRedFives = numRedFives

		playerInfo := d.newModelPlayerInfo()

		// 설명
		if d.gameMode == gameModeRecordCache && len(hands) == 14 {
			currentRoundCache.addAIDiscardTileWhenDrawTile(simpleBestDiscardTile(playerInfo), -1, 0, 0)
		}

		if d.skipOutput {
			return nil
		}

		// 설명
		if d.gameMode == gameModeRecord {
			currentRoundCache.print()
		}

		color.New(color.FgHiGreen).Printf("%s", util.MahjongZH[d.roundWindTile])
		fmt.Printf("%d국 시작, 자풍은 ", roundNumber%4+1)
		color.New(color.FgHiGreen).Printf("%s", util.MahjongZH[d.players[0].selfWindTile])
		fmt.Println()
		info := fmt.Sprintln(util.TilesToMahjongZHInterface(d.doraIndicators)...)
		info = info[:len(info)-1]
		color.HiYellow("도라 표시패는 " + info)
		fmt.Println()
		// TODO: 설명 보완
		return analysisPlayerWithRisk(playerInfo, nil)
	case d.parser.IsOpen():
		// 설명
		who, meld, kanDoraIndicator := d.parser.ParseOpen()
		meldType := meld.MeldType
		meldTiles := meld.Tiles
		calledTile := meld.CalledTile

		// 설명
		for _, player := range d.players {
			player.canIppatsu = false
		}

		// 설명
		if kanDoraIndicator != -1 {
			d.newDora(kanDoraIndicator)
		}

		player := d.players[who]

		// 설명
		if meldType != meldTypeAnkan {
			player.isNaki = true
		}

		// 설명
		if meldType == meldTypeKakan {
			if who != 0 {
				// 설명
				d.descLeftCounts(calledTile)
			} else {
				// 설명
				d.counts[calledTile]--
				// 설명

				// 설명
				if d.gameMode == gameModeRecordCache {
					currentRoundCache.addKan(meldType)
				}
			}
			// 설명
			for _, _meld := range player.melds {
				// 설명
				if _meld.Tiles[0] == calledTile {
					_meld.MeldType = meldTypeKakan
					_meld.Tiles = append(_meld.Tiles, calledTile)
					_meld.ContainRedFive = meld.ContainRedFive
					break
				}
			}

			if debugMode {
				if who == 0 {
					if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 1 {
						return fmt.Errorf("손패 오류: %d장 %v", handsCount, d.counts)
					}
				}
			}

			break
		}

		// 설명
		d.players[who].melds = append(d.players[who].melds, meld)

		if who != 0 {
			// 설명
			// 설명
			if meldType != meldTypeAnkan {
				d.leftCounts[calledTile]++
			}
			for _, tile := range meldTiles {
				d.descLeftCounts(tile)
			}
		} else {
			// 설명
			if meldType == meldTypeAnkan {
				d.counts[meldTiles[0]] = 0

				// 설명
				if d.gameMode == gameModeRecordCache {
					currentRoundCache.addKan(meldType)
				}
			} else {
				d.counts[calledTile]++
				for _, tile := range meldTiles {
					d.counts[tile]--
				}
				if meld.RedFiveFromOthers {
					tileType := meldTiles[0] / 9
					d.numRedFives[tileType]++
				}

				// 설명
				if d.gameMode == gameModeRecordCache {
					currentRoundCache.addChiPonKan(meldType)
				}
			}

			if debugMode {
				if meldType == meldTypeMinkan || meldType == meldTypeAnkan {
					if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 1 {
						return fmt.Errorf("손패 오류: %d장 %v", handsCount, d.counts)
					}
				} else {
					if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 2 {
						return fmt.Errorf("손패 오류: %d장 %v", handsCount, d.counts)
					}
				}
			}
		}
	case d.parser.IsReach():
		// 설명
		// 설명
		who := d.parser.ParseReach()
		d.players[who].isReached = true
		d.players[who].canIppatsu = true
		//case "AGARI", "RYUUKYOKU":
		// 설명
		//case "PROF":
		// 설명
		//case "BYE":
		// 설명
		//case "REJOIN", "GO":
		// 설명
	case d.parser.IsFuriten():
		// 설명
		if d.skipOutput {
			return nil
		}
		color.HiYellow("후리텐")
		//case "U", "V", "W":
		// 설명
		//case "HELO", "RANKING", "TAIKYOKU", "UN", "LN", "SAIKAI":
		// 설명
	case d.parser.IsSelfDraw():
		if !debugMode && !d.skipOutput {
			clearConsole()
		}
		// 설명
		tile, isRedFive, kanDoraIndicator := d.parser.ParseSelfDraw()
		if d.pendingSelfNukiDoraReplacement || d.pendingSelfDrawnNorth {
			if d.counts[30] > 0 {
				d.counts[30]--
				if d.pendingSelfDrawnNorth {
					d.players[0].nukiDoraNum++
					if !d.skipOutput {
						color.HiYellow("%s 북빼기", d.players[0].name)
					}
				}
			}
			d.pendingSelfNukiDoraReplacement = false
			d.pendingSelfDrawnNorth = false
		}
		d.descLeftCounts(tile)
		d.counts[tile]++
		if d.playerNumber == 3 && tile == 30 {
			d.pendingSelfDrawnNorth = true
		}
		if isRedFive {
			d.numRedFives[tile/9]++
		}
		if kanDoraIndicator != -1 {
			d.newDora(kanDoraIndicator)
		}

		playerInfo := d.newModelPlayerInfo()

		// 설명
		riskTables := d.analysisTilesRisk()
		mixedRiskTable := riskTables.mixedRiskTable()

		// 설명
		if d.gameMode == gameModeRecordCache {
			bestAttackDiscardTile := simpleBestDiscardTile(playerInfo)
			bestDefenceDiscardTile := mixedRiskTable.getBestDefenceTile(playerInfo.HandTiles34)
			bestAttackDiscardTileRisk, bestDefenceDiscardTileRisk := 0.0, 0.0
			if bestDefenceDiscardTile >= 0 {
				bestAttackDiscardTileRisk = mixedRiskTable[bestAttackDiscardTile]
				bestDefenceDiscardTileRisk = mixedRiskTable[bestDefenceDiscardTile]
			}
			currentRoundCache.addAIDiscardTileWhenDrawTile(bestAttackDiscardTile, bestDefenceDiscardTile, bestAttackDiscardTileRisk, bestDefenceDiscardTileRisk)
		}

		if d.skipOutput {
			return nil
		}

		// 설명
		if d.gameMode == gameModeRecord {
			currentRoundCache.print()
		}

		// 설명
		d.printDiscards()
		fmt.Println()

		// 설명
		riskTables.printWithHands(d.counts, d.leftCounts)

		// 설명
		// TODO: 설명 보완
		return analysisPlayerWithRisk(playerInfo, mixedRiskTable)
	case d.parser.IsDiscard():
		who, discardTile, isRedFive, isTsumogiri, isReach, canBeMeld, kanDoraIndicator := d.parser.ParseDiscard()

		if kanDoraIndicator != -1 {
			d.newDora(kanDoraIndicator)
		}

		player := d.players[who]
		if isReach {
			player.isReached = true
			player.canIppatsu = true
		}

		if who == 0 {
			// 설명
			riskTables := d.analysisTilesRisk()
			mixedRiskTable := riskTables.mixedRiskTable()

			if d.pendingSelfNukiDoraReplacement && util.CountOfTiles34(d.counts)%3 == 1 {
				d.counts[discardTile]++
				d.pendingSelfNukiDoraReplacement = false
			}

			// 설명
			d.counts[discardTile]--

			d.globalDiscardTiles = append(d.globalDiscardTiles, discardTile)
			player.discardTiles = append(player.discardTiles, discardTile)
			player.latestDiscardAtGlobal = len(d.globalDiscardTiles) - 1

			if isRedFive {
				d.numRedFives[discardTile/9]--
			}

			// 설명
			if d.gameMode == gameModeRecordCache {
				currentRoundCache.addSelfDiscardTile(discardTile, mixedRiskTable[discardTile], isReach)
			}

			if debugMode {
				if handsCount := util.CountOfTiles34(d.counts); handsCount%3 != 1 {
					return fmt.Errorf("손패 오류: %d장 %v", handsCount, d.counts)
				}
			}

			return nil
		}

		// 설명
		d.descLeftCounts(discardTile)

		_disTile := discardTile
		if isTsumogiri {
			_disTile = ^_disTile
		}
		d.globalDiscardTiles = append(d.globalDiscardTiles, _disTile)
		player.discardTiles = append(player.discardTiles, _disTile)
		player.latestDiscardAtGlobal = len(d.globalDiscardTiles) - 1

		// 설명
		if !player.isReached && len(player.discardTiles) <= 5 {
			player.earlyOutsideTiles = append(player.earlyOutsideTiles, util.OutsideTiles(discardTile)...)
		}

		if player.isReached && player.reachTileAtGlobal == -1 {
			// 설명
			player.reachTileAtGlobal = len(d.globalDiscardTiles) - 1
			player.reachTileAt = len(player.discardTiles) - 1
			// 설명
			if !d.skipOutput {
				if isTsumogiri {
					color.HiYellow("%s 쯔모기리 리치!", player.name)
				} else {
					color.HiYellow("%s 리치!", player.name)
				}
			}
		} else if len(player.meldDiscardsAt) != len(player.melds) {
			// 설명
			// 설명
			// 설명
			player.meldDiscardsAt = append(player.meldDiscardsAt, len(player.discardTiles)-1)
			player.meldDiscardsAtGlobal = append(player.meldDiscardsAtGlobal, len(d.globalDiscardTiles)-1)
		}

		// 설명
		if player.reachTileAt < len(player.discardTiles)-1 {
			player.canIppatsu = false
		}

		playerInfo := d.newModelPlayerInfo()

		// 설명
		riskTables := d.analysisTilesRisk()
		mixedRiskTable := riskTables.mixedRiskTable()

		// 설명
		if d.gameMode == gameModeRecordCache {
			allowChi := who == 3
			_, results14, incShantenResults14 := util.CalculateMeld(playerInfo, discardTile, isRedFive, allowChi)
			bestAttackDiscardTile := -1
			if len(results14) > 0 {
				bestAttackDiscardTile = results14[0].DiscardTile
			} else if len(incShantenResults14) > 0 {
				bestAttackDiscardTile = incShantenResults14[0].DiscardTile
			}
			if bestAttackDiscardTile != -1 {
				bestDefenceDiscardTile := mixedRiskTable.getBestDefenceTile(playerInfo.HandTiles34)
				bestAttackDiscardTileRisk := 0.0
				if bestDefenceDiscardTile >= 0 {
					bestAttackDiscardTileRisk = mixedRiskTable[bestAttackDiscardTile]
				}
				currentRoundCache.addPossibleChiPonKan(bestAttackDiscardTile, bestAttackDiscardTileRisk)
			}
		}

		if d.skipOutput {
			return nil
		}

		// 설명
		//if d.gameMode == gameModeMatch && who == 3 && !canBeMeld {
		//	return nil
		//}

		if !debugMode {
			clearConsole()
		}

		// 설명
		if d.gameMode == gameModeRecord {
			currentRoundCache.print()
		}

		// 설명
		d.printDiscards()
		fmt.Println()
		riskTables.printWithHands(d.counts, d.leftCounts)

		if d.gameMode == gameModeMatch && !canBeMeld {
			return nil
		}

		// 설명
		// TODO: 설명 보완
		allowChi := d.playerNumber != 3 && who == 3 && playerInfo.LeftDrawTilesCount > 0
		return analysisMeld(playerInfo, discardTile, isRedFive, allowChi, mixedRiskTable)
	case d.parser.IsRoundWin():
		// TODO: 설명 보완

		if !debugMode {
			clearConsole()
		}
		fmt.Println("화료, 이번 국 종료")
		whos, points := d.parser.ParseRoundWin()
		if len(whos) == 3 {
			color.HiYellow("봉황급 방총 회피")
			if d.parser.GetDataSourceType() == dataSourceTypeMajsoul {
				color.HiYellow("(정신 차려요, 이건 작혼입니다)")
			}
		}
		for i, who := range whos {
			fmt.Println(d.players[who].name, points[i])
		}
	case d.parser.IsRyuukyoku():
		if !debugMode {
			clearConsole()
		}
		d.printDiscards()
		type_, _, _ := d.parser.ParseRyuukyoku()
		if type_ > 0 {
			color.HiYellow("유국, 이번 국 종료 [%d]", type_)
		} else {
			color.HiYellow("유국, 이번 국 종료")
		}
	case d.parser.IsNukiDora():
		who, isTsumogiri := d.parser.ParseNukiDora()
		player := d.players[who]
		player.nukiDoraNum++
		if !d.skipOutput {
			if !debugMode {
				clearConsole()
			}
			d.printDiscards()
			if isTsumogiri {
				color.HiYellow("%s 북빼기(쯔모기리)", player.name)
			} else {
				color.HiYellow("%s 북빼기", player.name)
			}
		}
		if who != 0 {
			// 설명
			d.descLeftCounts(30)
			// TODO
			_ = isTsumogiri
		} else {
			// 설명
			if d.counts[30] > 0 {
				d.counts[30]--
			}
			d.pendingSelfDrawnNorth = false
			d.pendingSelfNukiDoraReplacement = true
		}
		// 설명
		for _, player := range d.players {
			player.canIppatsu = false
		}
	case d.parser.IsNewDora():
		// 설명
		// 설명
		// 설명
		kanDoraIndicator := d.parser.ParseNewDora()
		d.newDora(kanDoraIndicator)
	default:
	}

	return nil
}
