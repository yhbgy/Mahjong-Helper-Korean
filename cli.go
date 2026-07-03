package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/fatih/color"
	"math"
	"sort"
	"strings"
)

func printAccountInfo(accountID int) {
	fmt.Printf("계정 ID는 ")
	color.New(color.FgHiGreen).Printf("%d", accountID)
	fmt.Printf("이며, 이 숫자는 작혼 서버 계정 데이터베이스의 ID입니다. 값이 작을수록 가입 시점이 빠릅니다.\n")
}

//

func (d *roundData) printRoundState() {
	names := []string{"자신", "하가", "대면", "상가"}
	nukiCount := 0
	for i, player := range d.players {
		if i >= len(names) || (d.playerNumber == 3 && player.selfWindTile == 30) {
			continue
		}
		nukiCount += player.nukiDoraNum
	}
	if len(d.scores) == 0 && d.benNumber == 0 && d.liqibang == 0 && nukiCount == 0 && len(d.doraIndicators) == 0 {
		return
	}

	if d.roundWindTile >= 27 && d.roundWindTile < len(util.MahjongZH) {
		fmt.Printf("%s %d국", util.MahjongZH[d.roundWindTile], d.roundNumber%4+1)
	} else {
		fmt.Printf("%d국", d.roundNumber%4+1)
	}
	if d.benNumber > 0 {
		fmt.Printf(" %d본장", d.benNumber)
	}
	if d.liqibang > 0 {
		fmt.Printf(" 리치봉 %d개", d.liqibang)
	}
	if len(d.scores) > 0 {
		fmt.Print(" | 점수:")
		for i, score := range d.scores {
			if i >= len(names) || (d.playerNumber == 3 && i == 3) {
				continue
			}
			fmt.Printf(" %s %d", names[i], score)
		}
	}
	if len(d.doraIndicators) > 0 {
		fmt.Print(" | 도라 표시패:")
		for _, tile := range d.doraIndicators {
			if tile >= 0 && tile < len(util.MahjongZH) {
				fmt.Printf(" %s", util.MahjongZH[tile])
			}
		}
	}
	if nukiCount > 0 {
		fmt.Print(" | 북빼기:")
		for i, player := range d.players {
			if i >= len(names) || (d.playerNumber == 3 && player.selfWindTile == 30) || player.nukiDoraNum == 0 {
				continue
			}
			fmt.Printf(" %s %d", names[i], player.nukiDoraNum)
		}
	}
	fmt.Println()
}

func (p *playerInfo) printDiscards() {
	// TODO: 부자연스러운 버림패나 위험패를 강조한다. 예:
	// - 초반부터 중장패를 버림
	// - 중장패를 버리기 시작한 뒤 요구패를 테다시함
	//   (누군가가 퐁해서 생긴 현상일 수도 있다. 예: 133m에서 누군가 2m을 퐁)
	// - 도라를 버림
	// - 적도라를 버림
	// - 누군가 리치한 상황에서 위험도가 높은 패를 여러 번 버림
	//   (상대가 읽고 버렸거나, 상대 손패와 강 정보상 안전패가 생겼을 가능성도 있다)
	// - 그 밖의 예시는 《마신의 눈》 번역 참고: https://tieba.baidu.com/p/3311909701
	//      단순한 예로, 테다시로 또이츠 하나를 깨면 치또이츠 가능성은 거의 사라진다.
	//      상대가 초반에 양면 타쯔를 테다시했다면 혼일/청일 계열 또는 또이츠형을 추정할 수 있고,
	//      그 뒤 리치나 후로가 나오면 손패를 더 읽기 쉬워진다.
	// https://tieba.baidu.com/p/3311909701
	//      후로 이후와 종반의 테다시는 가능한 한 기억한다. 다른 사람이 테다시하기 전의 안전패는 먼저 버리는 편이 좋다.
	// https://tieba.baidu.com/p/3372239806
	//      치할 때 나온 패의 색은 위험하다. 퐁 이후에는 모든 패가 위험하다.

	fmt.Print(p.name)
	if p.isReached {
		color.New(color.FgHiYellow).Print("(리치)")
	}
	fmt.Print(":")
	for i, disTile := range p.discardTiles {
		fmt.Printf(" ")
		// TODO: 도라와 적도라 표시
		bgColor := color.BgBlack
		fgColor := color.FgWhite
		var tile string
		if disTile >= 0 { // 테다시
			tile = util.MahjongZH[disTile]
			if p.isNaki { // 후로
				fgColor = getOtherDiscardAlertColor(disTile) // 중장패 테다시를 강조
				if util.InInts(i, p.meldDiscardsAt) {
					bgColor = color.BgWhite // 후로 직후 버린 패는 배경을 강조
					fgColor = color.FgBlack
				}
			}
		} else { // 쯔모기리
			disTile = ^disTile
			tile = util.MahjongZH[disTile] + "*"
			fgColor = color.FgHiBlack // 어둡게 표시
		}
		color.New(bgColor, fgColor).Print(tile)
	}
	fmt.Println()
}

//

type handsRisk struct {
	tile int
	risk float64
}

func tilesToKoreanList(tiles []int) string {
	if len(tiles) == 0 {
		return ""
	}

	names := make([]string, len(tiles))
	for i, tile := range tiles {
		names[i] = util.MahjongZH[tile]
	}
	return strings.Join(names, " ")
}

// 34종 패의 위험도
type riskTable util.RiskTiles34

func (t riskTable) printWithHands(hands []int, fixedRiskMulti float64) (containLine bool) {
	// 방총률이 0인 패 출력(현물, 또는 노찬스이면서 남은 수가 0인 패)
	safeCount := 0
	for i, c := range hands {
		if c > 0 && t[i] == 0 {
			fmt.Printf(" " + util.MahjongZH[i])
			safeCount++
		}
	}

	// 위험패를 방총률순으로 출력하고 강조
	handsRisks := []handsRisk{}
	for i, c := range hands {
		if c > 0 && t[i] > 0 {
			handsRisks = append(handsRisks, handsRisk{i, t[i]})
		}
	}
	sort.Slice(handsRisks, func(i, j int) bool {
		return handsRisks[i].risk < handsRisks[j].risk
	})
	if len(handsRisks) > 0 {
		if safeCount > 0 {
			fmt.Print(" |")
			containLine = true
		}
		for _, hr := range handsRisks {
			// 색상에는 텐파이율도 반영한다.
			color.New(getNumRiskColor(hr.risk * fixedRiskMulti)).Printf(" " + util.MahjongZH[hr.tile])
		}
	}

	return
}

func (t riskTable) getBestDefenceTile(tiles34 []int) (result int) {
	minRisk := 100.0
	maxRisk := 0.0
	for tile, c := range tiles34 {
		if c == 0 {
			continue
		}
		risk := t[tile]
		if risk < minRisk {
			minRisk = risk
			result = tile
		}
		if risk > maxRisk {
			maxRisk = risk
		}
	}
	if maxRisk == 0 {
		return -1
	}
	return result
}

//

type riskInfo struct {
	// 3인 마작이면 3, 4인 마작이면 4
	playerNumber int

	// 해당 플레이어의 텐파이율(리치 시 100.0)
	tenpaiRate float64

	// 해당 플레이어 기준 안전패
	// 깡한 패도 안전패로 본다. 스지/벽 위험도 판단에 도움이 된다.
	safeTiles34 []bool

	// 패별 방총률 표
	riskTable riskTable

	// 남은 무스지 123789
	// 총 18종. 남은 무스지 수가 적을수록 해당 무스지가 더 위험하다.
	leftNoSujiTiles []int

	// 쯔모기리 리치 여부
	isTsumogiriRiichi bool

	// 론 점수
	// 디버그 전용
	_ronPoint float64
}

type riskInfoList []*riskInfo

// 텐파이율을 반영한 종합 위험도
func (l riskInfoList) mixedRiskTable() riskTable {
	mixedRiskTable := make(riskTable, 34)
	for i := range mixedRiskTable {
		mixedRisk := 0.0
		for _, ri := range l[1:] {
			if ri.tenpaiRate <= 15 {
				continue
			}
			_risk := ri.riskTable[i] * ri.tenpaiRate / 100
			mixedRisk = mixedRisk + _risk - mixedRisk*_risk/100
		}
		mixedRiskTable[i] = mixedRisk
	}
	return mixedRiskTable
}

func (l riskInfoList) printWithHands(hands []int, leftCounts []int) {
	// 텐파이율이 일정 값을 넘으면 방총률을 출력한다.
	const (
		minShownTenpaiRate4 = 50.0
		minShownTenpaiRate3 = 20.0
	)

	minShownTenpaiRate := minShownTenpaiRate4
	if l[0].playerNumber == 3 {
		minShownTenpaiRate = minShownTenpaiRate3
	}

	dangerousPlayerCount := 0
	// 안전패와 위험패 출력
	names := []string{"", "하가", "대면", "상가"}
	for i := len(l) - 1; i >= 1; i-- {
		tenpaiRate := l[i].tenpaiRate
		if len(l[i].riskTable) > 0 && (debugMode || tenpaiRate > minShownTenpaiRate) {
			dangerousPlayerCount++
			fmt.Print(names[i] + " 안전패:")
			//if debugMode {
			//fmt.Printf("(%d*%2.2f%% 텐파이율)", int(l[i]._ronPoint), l[i].tenpaiRate)
			//}
			containLine := l[i].riskTable.printWithHands(hands, tenpaiRate/100)

			// 텐파이율 출력
			fmt.Print(" ")
			if !containLine {
				fmt.Print("  ")
			}
			fmt.Print("[")
			if tenpaiRate == 100 {
				fmt.Print("100.%")
			} else {
				fmt.Printf("%4.1f%%", tenpaiRate)
			}
			fmt.Print("텐파이율]")

			// 무스지 개수 출력
			fmt.Print(" ")
			const badMachiLimit = 3
			noSujiInfo := ""
			if l[i].isTsumogiriRiichi {
				noSujiInfo = "쯔모기리 리치"
			} else if len(l[i].leftNoSujiTiles) == 0 {
				noSujiInfo = "나쁜 대기 텐파이/후리텐"
			} else if len(l[i].leftNoSujiTiles) <= badMachiLimit {
				noSujiInfo = "나쁜 대기 텐파이 가능성/후리텐"
			}
			if noSujiInfo != "" {
				fmt.Printf("[%d무스지: ", len(l[i].leftNoSujiTiles))
				color.New(color.FgHiYellow).Printf("%s", noSujiInfo)
				fmt.Print("]")
			} else {
				fmt.Printf("[%d무스지]", len(l[i].leftNoSujiTiles))
			}

			fmt.Println()
		}
	}

	// 둘 이상의 플레이어가 리치/후로했다면 텐파이율을 반영한 가중 종합 방총률을 출력한다.
	mixedPlayers := 0
	for _, ri := range l[1:] {
		if ri.tenpaiRate > 0 {
			mixedPlayers++
		}
	}
	if dangerousPlayerCount > 0 && mixedPlayers > 1 {
		fmt.Print("종합 안전패:")
		mixedRiskTable := l.mixedRiskTable()
		mixedRiskTable.printWithHands(hands, 1)
		fmt.Println()
	}

	// 노찬스/원찬스로 생긴 안전패 출력
	// TODO: 다른 함수로 분리
	if dangerousPlayerCount > 0 {
		ncSafeTileList := util.CalcNCSafeTiles(leftCounts).FilterWithHands(hands)
		ocSafeTileList := util.CalcOCSafeTiles(leftCounts).FilterWithHands(hands)
		if len(ncSafeTileList) > 0 {
			fmt.Printf("NC:")
			for _, safeTile := range ncSafeTileList {
				fmt.Printf(" " + util.MahjongZH[safeTile.Tile34])
			}
			fmt.Println()
		}
		if len(ocSafeTileList) > 0 {
			fmt.Printf("OC:")
			for _, safeTile := range ocSafeTileList {
				fmt.Printf(" " + util.MahjongZH[safeTile.Tile34])
			}
			fmt.Println()
		}

		// 다른 표시 방식: 벽패 표시
		//printedNC := false
		//for i, c := range leftCounts[:27] {
		//	if c != 0 || i%9 == 0 || i%9 == 8 {
		//		continue
		//	}
		//	if !printedNC {
		//		printedNC = true
		//		fmt.Printf("NC:")
		//	}
		//	fmt.Printf(" " + util.MahjongZH[i])
		//}
		//if printedNC {
		//	fmt.Println()
		//}
		//printedOC := false
		//for i, c := range leftCounts[:27] {
		//	if c != 1 || i%9 == 0 || i%9 == 8 {
		//		continue
		//	}
		//	if !printedOC {
		//		printedOC = true
		//		fmt.Printf("OC:")
		//	}
		//	fmt.Printf(" " + util.MahjongZH[i])
		//}
		//if printedOC {
		//	fmt.Println()
		//}
		fmt.Println()
	}
}

//

func alertBackwardToShanten2(results util.Hand14AnalysisResultList, incShantenResults util.Hand14AnalysisResultList) {
	if len(results) == 0 || len(incShantenResults) == 0 {
		return
	}

	if results[0].Result13.Waits.AllCount() < 9 {
		if results[0].Result13.MixedWaitsScore < incShantenResults[0].Result13.MixedWaitsScore {
			color.HiGreen("샨텐 후퇴?")
		}
	}
}

// 안내할 역
var yakuTypesToAlert = []int{
	//util.YakuKokushi,
	//util.YakuKokushi13,
	util.YakuSuuAnkou,
	util.YakuSuuAnkouTanki,
	util.YakuDaisangen,
	util.YakuShousuushii,
	util.YakuDaisuushii,
	util.YakuTsuuiisou,
	util.YakuChinroutou,
	util.YakuRyuuiisou,
	util.YakuChuuren,
	util.YakuChuuren9,
	util.YakuSuuKantsu,
	//util.YakuTenhou,
	//util.YakuChiihou,

	util.YakuChiitoi,
	util.YakuPinfu,
	util.YakuRyanpeikou,
	util.YakuIipeikou,
	util.YakuSanshokuDoujun,
	util.YakuIttsuu,
	util.YakuToitoi,
	util.YakuSanAnkou,
	util.YakuSanshokuDoukou,
	util.YakuSanKantsu,
	util.YakuTanyao,
	util.YakuChanta,
	util.YakuJunchan,
	util.YakuHonroutou,
	util.YakuShousangen,
	util.YakuHonitsu,
	util.YakuChinitsu,

	util.YakuShiiaruraotai,
	util.YakuUumensai,
	util.YakuSanrenkou,
	util.YakuIsshokusanjun,
}

/*

8     3삭 타 텐파이[2만, 7만]
9.20  [20 개량]  4.00 텐파이수

4     텐파이 [2만, 7만]
4.50  [ 4 개량]  55.36% 참고 화료율

8     45만 치, 4만 타 텐파이[2만, 7만]
9.20  [20 개량]  4.00 텐파이수

*/
// 무엇을 버릴지에 대한 분석 결과 출력(두 줄)
func printWaitsWithImproves13_twoRows(result13 *util.Hand13AnalysisResult, discardTile34 int, openTiles34 []int) {
	shanten := result13.Shanten
	waits := result13.Waits

	waitsCount, waitTiles := waits.ParseIndex()
	c := getWaitsCountColor(shanten, float64(waitsCount))
	color.New(c).Printf("%-6d", waitsCount)
	if discardTile34 != -1 {
		if len(openTiles34) > 0 {
			meldType := "치"
			if openTiles34[0] == openTiles34[1] {
				meldType = "퐁"
			}
			color.New(color.FgHiWhite).Printf("%s%s", string([]rune(util.MahjongZH[openTiles34[0]])[:1]), util.MahjongZH[openTiles34[1]])
			fmt.Printf("%s，", meldType)
		}
		fmt.Print("타 ")
		fmt.Print(util.MahjongZH[discardTile34])
		fmt.Print(" ")
	}
	//fmt.Print("대기")
	//if shanten <= 1 {
	//	fmt.Print("[")
	//	if len(waitTiles) > 0 {
	//		fmt.Print(util.MahjongZH[waitTiles[0]])
	//		for _, idx := range waitTiles[1:] {
	//			fmt.Print(", " + util.MahjongZH[idx])
	//		}
	//	}
	//	fmt.Println("]")
	//} else {
	fmt.Println(util.TilesToKoreanStrWithBracket(waitTiles))
	//}

	if len(result13.Improves) > 0 {
		fmt.Printf("%-6.2f[%2d 개량]", result13.AvgImproveWaitsCount, len(result13.Improves))
	} else {
		fmt.Print(strings.Repeat(" ", 15))
	}

	fmt.Print(" ")

	if shanten >= 1 {
		c := getWaitsCountColor(shanten-1, result13.AvgNextShantenWaitsCount)
		color.New(c).Printf("%5.2f", result13.AvgNextShantenWaitsCount)
		fmt.Printf(" %s", util.NumberToChineseShanten(shanten-1))
		if shanten >= 2 {
			fmt.Printf("유효패")
		} else { // shanten == 1
			fmt.Printf("수")
			if showAgariAboveShanten1 {
				fmt.Printf("(%.2f%% 참고 화료율)", result13.AvgAgariRate)
			}
		}
		if showScore {
			mixedScore := result13.MixedWaitsScore
			//for i := 2; i <= shanten; i++ {
			//	mixedScore /= 4
			//}
			fmt.Printf("(%.2f 종합 점수)", mixedScore)
		}
	} else { // shanten == 0
		fmt.Printf("%5.2f%% 참고 화료율", result13.AvgAgariRate)
	}

	fmt.Println()
}

type analysisResult struct {
	discardTile34     int
	isDiscardTileDora bool
	openTiles34       []int
	result13          *util.Hand13AnalysisResult

	mixedRiskTable riskTable

	highlightAvgImproveWaitsCount bool
	highlightMixedScore           bool
}

/*
4[ 4.56] 8통 타 => 44.50% 참고 화료율[ 4 개량] [7p 7s] [다마텐2000] [삼색] [후리텐]

4[ 4.56] 8통 타 => 0.00% 참고 화료율[ 4 개량] [7p 7s] [무역]

31[33.58] 7삭 타 =>  5.23텐파이수 [19.21속도] [16개량] [6789p 56789s] [국수지3120] [후리텐 가능]

48[50.64] 5통 타 => 24.25이샹텐 [12개량] [123456789p 56789s]

31[33.62] 77삭 퐁, 5통 타 => 5.48텐파이수 [15 개량] [123456789p]

*/
// 무엇을 버릴지에 대한 분석 결과 출력(한 줄)
func (r *analysisResult) printWaitsWithImproves13_oneRow() {
	discardTile34 := r.discardTile34
	openTiles34 := r.openTiles34
	result13 := r.result13

	shanten := result13.Shanten

	// 유효패 수
	waitsCount := result13.Waits.AllCount()
	c := getWaitsCountColor(shanten, float64(waitsCount))
	color.New(c).Printf("%2d", waitsCount)
	// 개량 유효패 평균
	if len(result13.Improves) > 0 {
		if r.highlightAvgImproveWaitsCount {
			color.New(color.FgHiWhite).Printf("[%5.2f]", result13.AvgImproveWaitsCount)
		} else {
			fmt.Printf("[%5.2f]", result13.AvgImproveWaitsCount)
		}
	} else {
		fmt.Print(strings.Repeat(" ", 7))
	}

	fmt.Print(" ")

	// 3k+2장 손패의 타패 분석인지 여부
	if discardTile34 != -1 {
		// 후로 분석
		if len(openTiles34) > 0 {
			meldType := "치"
			if openTiles34[0] == openTiles34[1] {
				meldType = "퐁"
			}
			color.New(color.FgHiWhite).Printf("%s%s", string([]rune(util.MahjongZH[openTiles34[0]])[:1]), util.MahjongZH[openTiles34[1]])
			fmt.Printf("%s,", meldType)
		}
		// 버림패
		if r.isDiscardTileDora {
			color.New(color.FgHiWhite).Print("도라타")
		} else {
			fmt.Print("타")
		}
		tileZH := util.MahjongZH[discardTile34]
		if discardTile34 >= 27 {
			tileZH = " " + tileZH
		}
		if r.mixedRiskTable != nil {
			// 실제 위험도가 있으면 그 값으로 버림패 위험도를 표시
			risk := r.mixedRiskTable[discardTile34]
			if risk == 0 {
				fmt.Print(tileZH)
			} else {
				color.New(getNumRiskColor(risk)).Print(tileZH)
			}
			fmt.Printf(" [위험 %.1f]", risk)
		} else {
			fmt.Print(tileZH)
		}
	}

	fmt.Print(" => ")

	if shanten >= 1 {
		// 샨텐 전진 후 유효패 수 평균
		incShanten := shanten - 1
		c := getWaitsCountColor(incShanten, result13.AvgNextShantenWaitsCount)
		color.New(c).Printf("%5.2f", result13.AvgNextShantenWaitsCount)
		fmt.Printf("%s", util.NumberToChineseShanten(incShanten))
		if incShanten >= 1 {
			//fmt.Printf("유효패")
		} else { // incShanten == 0
			fmt.Printf("수")
			//if showAgariAboveShanten1 {
			//	fmt.Printf("(%.2f%% 참고 화료율)", result13.AvgAgariRate)
			//}
		}
	} else { // shanten == 0
		// 전진 후 화료율
		// 후리텐 또는 한쪽 대기면 빨간색으로 표시
		if result13.FuritenRate == 1 || result13.IsPartWait {
			color.New(color.FgHiRed).Printf("%5.2f%% 참고 화료율", result13.AvgAgariRate)
		} else {
			fmt.Printf("%5.2f%% 참고 화료율", result13.AvgAgariRate)
		}
	}

	// 손패 속도. 빠른 연장 판단에 사용
	if result13.MixedWaitsScore > 0 && shanten >= 1 && shanten <= 2 {
		fmt.Print(" ")
		if r.highlightMixedScore {
			color.New(color.FgHiWhite).Printf("[%5.2f속도]", result13.MixedWaitsScore)
		} else {
			fmt.Printf("[%5.2f속도]", result13.MixedWaitsScore)
		}
	}

	// 국 수지
	if showScore && result13.MixedRoundPoint != 0.0 {
		fmt.Print(" ")
		color.New(color.FgHiGreen).Printf("[국 수지%4d]", int(math.Round(result13.MixedRoundPoint)))
	}

	// (다마텐) 론 점수
	if result13.DamaPoint > 0 {
		fmt.Print(" ")
		ronType := "론"
		if !result13.IsNaki {
			ronType = "다마텐"
		}
		color.New(color.FgHiGreen).Printf("[%s%d]", ronType, int(math.Round(result13.DamaPoint)))
	}

	// 리치 점수. 쯔모, 일발, 우라도라를 고려
	if result13.RiichiPoint > 0 {
		fmt.Print(" ")
		color.New(color.FgHiGreen).Printf("[리치%d]", int(math.Round(result13.RiichiPoint)))
	}

	if len(result13.YakuTypes) > 0 {
		// 역(2샨텐 이내에서 표시)
		if result13.Shanten <= 2 {
			if !showAllYakuTypes && !debugMode {
				shownYakuTypes := []int{}
				for yakuType := range result13.YakuTypes {
					for _, yt := range yakuTypesToAlert {
						if yakuType == yt {
							shownYakuTypes = append(shownYakuTypes, yakuType)
						}
					}
				}
				if len(shownYakuTypes) > 0 {
					sort.Ints(shownYakuTypes)
					fmt.Print(" ")
					color.New(color.FgHiGreen).Printf(util.YakuTypesToStr(shownYakuTypes))
				}
			} else {
				// debug
				fmt.Print(" ")
				color.New(color.FgHiGreen).Printf(util.YakuTypesWithDoraToStr(result13.YakuTypes, result13.DoraCount))
			}
			// 한쪽 대기
			if result13.IsPartWait {
				fmt.Print(" ")
				color.New(color.FgHiRed).Printf("[한쪽 대기]")
			}
		}
	} else if result13.IsNaki && shanten >= 0 && shanten <= 2 {
		// 후로 시 역 없음 안내(텐파이부터 2샨텐까지)
		fmt.Print(" ")
		color.New(color.FgHiRed).Printf("[역 없음]")
	}

	// 후리텐 안내
	if result13.FuritenRate > 0 {
		fmt.Print(" ")
		if result13.FuritenRate < 1 {
			color.New(color.FgHiYellow).Printf("[후리텐 가능성]")
		} else {
			color.New(color.FgHiRed).Printf("[후리텐]")
		}
	}

	// 개량 수
	if showScore {
		fmt.Print(" ")
		if len(result13.Improves) > 0 {
			fmt.Printf("[%2d개량]", len(result13.Improves))
		} else {
			fmt.Print(strings.Repeat(" ", 4))
			fmt.Print(strings.Repeat("　", 2)) // 전각 공백
		}
	}

	// 유효패 종류
	fmt.Print(" ")
	waitTiles := result13.Waits.AvailableTiles()
	fmt.Print(util.TilesToKoreanStrWithBracket(waitTiles))
	if shanten == 1 {
		fmt.Printf(" [텐파이 진입: %s]", tilesToKoreanList(waitTiles))
	}

	//

	fmt.Println()

	if showImproveDetail {
		for tile, waits := range result13.Improves {
			fmt.Printf("쯔모 %s 개량 -> %s\n", util.Mahjong[tile], waits.String())
		}
	}
}

func printResults14WithRisk(results14 util.Hand14AnalysisResultList, mixedRiskTable riskTable) {
	if len(results14) == 0 {
		return
	}
	results14 = sortResults14ForDisplay(results14, mixedRiskTable)

	maxMixedScore := -1.0
	maxAvgImproveWaitsCount := -1.0
	for _, result := range results14 {
		if result.Result13.MixedWaitsScore > maxMixedScore {
			maxMixedScore = result.Result13.MixedWaitsScore
		}
		if result.Result13.AvgImproveWaitsCount > maxAvgImproveWaitsCount {
			maxAvgImproveWaitsCount = result.Result13.AvgImproveWaitsCount
		}
	}

	if len(results14[0].OpenTiles) > 0 {
		fmt.Print("후로 후")
	}
	fmt.Println(util.NumberToChineseShanten(results14[0].Result13.Shanten) + "：")

	if results14[0].Result13.Shanten == 0 {
		// 대기는 같지만 타점이 다른지 확인
		isDiffPoint := false
		baseWaits := results14[0].Result13.Waits
		baseDamaPoint := results14[0].Result13.DamaPoint
		baseRiichiPoint := results14[0].Result13.RiichiPoint
		for _, result14 := range results14[1:] {
			if baseWaits.Equals(result14.Result13.Waits) && (baseDamaPoint != result14.Result13.DamaPoint || baseRiichiPoint != result14.Result13.RiichiPoint) {
				isDiffPoint = true
				break
			}
		}

		if isDiffPoint {
			color.HiGreen("타패 선택 주의: 타점")
		}
	}

	// FIXME: 선택지가 많을 때 타패 후보를 어떻게 줄일지
	//const maxShown = 10
	//if len(results14) > maxShown { // 출력 개수 제한
	//	results14 = results14[:maxShown]
	//}
	for _, result := range results14 {
		r := &analysisResult{
			result.DiscardTile,
			result.IsDiscardDoraTile,
			result.OpenTiles,
			result.Result13,
			mixedRiskTable,
			result.Result13.AvgImproveWaitsCount == maxAvgImproveWaitsCount,
			result.Result13.MixedWaitsScore == maxMixedScore,
		}
		r.printWaitsWithImproves13_oneRow()
	}
}

func sortResults14ForDisplay(results14 util.Hand14AnalysisResultList, mixedRiskTable riskTable) util.Hand14AnalysisResultList {
	sorted := append(util.Hand14AnalysisResultList{}, results14...)
	riskOf := func(result *util.Hand14AnalysisResult) float64 {
		if mixedRiskTable == nil || result.DiscardTile < 0 || result.DiscardTile >= len(mixedRiskTable) {
			return 0
		}
		return mixedRiskTable[result.DiscardTile]
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		ri, rj := sorted[i].Result13, sorted[j].Result13
		if ri.Shanten != rj.Shanten {
			return ri.Shanten < rj.Shanten
		}

		wi, wj := ri.Waits.AllCount(), rj.Waits.AllCount()
		switch ri.Shanten {
		case 0:
			if !util.InDelta(ri.MixedRoundPoint, rj.MixedRoundPoint, 100) {
				return ri.MixedRoundPoint > rj.MixedRoundPoint
			}
			if !util.Equal(ri.AvgAgariRate, rj.AvgAgariRate) {
				return ri.AvgAgariRate > rj.AvgAgariRate
			}
		case 1, 2:
			if !util.Equal(ri.MixedWaitsScore, rj.MixedWaitsScore) {
				return ri.MixedWaitsScore > rj.MixedWaitsScore
			}
			if wi != wj {
				return wi > wj
			}
			if !util.Equal(ri.AvgNextShantenWaitsCount, rj.AvgNextShantenWaitsCount) {
				return ri.AvgNextShantenWaitsCount > rj.AvgNextShantenWaitsCount
			}
		default:
			if !util.Equal(ri.AvgImproveWaitsCount, rj.AvgImproveWaitsCount) {
				return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
			}
			if wi != wj {
				return wi > wj
			}
		}

		riskI, riskJ := riskOf(sorted[i]), riskOf(sorted[j])
		if !util.Equal(riskI, riskJ) {
			return riskI < riskJ
		}
		if !util.Equal(ri.AvgImproveWaitsCount, rj.AvgImproveWaitsCount) {
			return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
		}
		if !util.Equal(ri.MixedRoundPoint, rj.MixedRoundPoint) {
			return ri.MixedRoundPoint > rj.MixedRoundPoint
		}
		return sorted[i].DiscardTile < sorted[j].DiscardTile
	})
	return sorted
}
