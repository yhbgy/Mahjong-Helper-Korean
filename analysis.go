package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
	"strings"
)

func simpleBestDiscardTile(playerInfo *model.PlayerInfo) int {
	shanten, results14, incShantenResults14 := util.CalculateShantenWithImproves14(playerInfo)
	bestAttackDiscardTile := -1
	if len(results14) > 0 {
		bestAttackDiscardTile = results14[0].DiscardTile
	} else if len(incShantenResults14) > 0 {
		bestAttackDiscardTile = incShantenResults14[0].DiscardTile
	} else {
		return -1
	}
	if shanten == 1 && len(playerInfo.DiscardTiles) < 9 && len(results14) > 0 && len(incShantenResults14) > 0 && !playerInfo.IsNaki() { // 울었을 때의 샹텐 후퇴는 아직 고려하지 않음
		if results14[0].Result13.Waits.AllCount() < 9 && results14[0].Result13.MixedWaitsScore < incShantenResults14[0].Result13.MixedWaitsScore {
			bestAttackDiscardTile = incShantenResults14[0].DiscardTile
		}
	}
	return bestAttackDiscardTile
}

// TODO: model 패키지로 옮기기
func humanMeld(meld model.Meld) string {
	humanMeld := util.TilesToStr(meld.Tiles)
	if meld.MeldType == model.MeldTypeAnkan {
		return strings.ToUpper(humanMeld)
	}
	return humanMeld
}
func humanHands(playerInfo *model.PlayerInfo) string {
	humanHands := util.Tiles34ToStr(playerInfo.HandTiles34)
	if len(playerInfo.Melds) > 0 {
		humanHands += " " + model.SepMeld
		for i := len(playerInfo.Melds) - 1; i >= 0; i-- {
			humanHands += " " + humanMeld(playerInfo.Melds[i])
		}
	}
	return humanHands
}

func analysisPlayerWithRisk(playerInfo *model.PlayerInfo, mixedRiskTable riskTable) error {
	// 손패
	humanTiles := humanHands(playerInfo)
	fmt.Println(humanTiles)
	fmt.Println(strings.Repeat("=", len(humanTiles)))

	countOfTiles := util.CountOfTiles34(playerInfo.HandTiles34)
	switch countOfTiles % 3 {
	case 1:
		result := util.CalculateShantenWithImproves13(playerInfo)
		fmt.Println("현재 " + util.NumberToChineseShanten(result.Shanten) + ":")
		r := &analysisResult{
			discardTile34:  -1,
			result13:       result,
			mixedRiskTable: mixedRiskTable,
		}
		r.printWaitsWithImproves13_oneRow()
	case 2:
		// 손패 분석
		shanten, results14, incShantenResults14 := util.CalculateShantenWithImproves14(playerInfo)

		// 안내 메시지
		if shanten == -1 {
			color.HiRed("[이미 화료]")
		} else if shanten == 0 {
			if len(results14) > 0 {
				r13 := results14[0].Result13
				if r13.RiichiPoint > 0 && r13.FuritenRate == 0 && r13.DamaPoint >= 5200 && r13.DamaWaits.AllCount() == r13.Waits.AllCount() {
					color.HiGreen("다마텐 타점 충분: 화료율 중시면 다마텐, 타점 중시면 리치")
				}
				// 국 수지가 비슷할 때: 화료율 중시면 xx, 타점 중시면 xx를 안내
			}
		} else if shanten == 1 {
			// 초중반 멘젠일 때 샨텐 후퇴를 안내
			if len(playerInfo.DiscardTiles) < 9 && !playerInfo.IsNaki() {
				alertBackwardToShanten2(results14, incShantenResults14)
			}
		}

		// TODO: 유국이 가까울 때 하이테이 차례가 누구인지 안내

		// 무엇을 버릴지에 대한 분석 결과
		printResults14WithRisk(results14, mixedRiskTable)
		printResults14WithRisk(incShantenResults14, mixedRiskTable)
	default:
		err := fmt.Errorf("인자 오류: %d장", countOfTiles)
		if debugMode {
			panic(err)
		}
		return err
	}

	fmt.Println()
	return nil
}

// 후로를 분석한다.
// playerInfo: 자신의 정보
// targetTile34: 다른 사람이 버린 패
// isRedFive: 해당 버림패가 적5인지 여부
// allowChi: 치가 가능한지 여부
// mixedRiskTable: 위험도 표
func analysisMeld(playerInfo *model.PlayerInfo, targetTile34 int, isRedFive bool, allowChi bool, mixedRiskTable riskTable) error {
	if handsCount := util.CountOfTiles34(playerInfo.HandTiles34); handsCount%3 != 1 {
		return fmt.Errorf("손패 오류: %d장 %v", handsCount, playerInfo.HandTiles34)
	}
	// 원래 손패 분석
	result := util.CalculateShantenWithImproves13(playerInfo)
	// 후로 분석
	shanten, results14, incShantenResults14 := util.CalculateMeld(playerInfo, targetTile34, isRedFive, allowChi)
	if len(results14) == 0 && len(incShantenResults14) == 0 {
		return nil // fmt.Errorf("입력 오류: 이 패로는 울 수 없습니다")
	}

	// 후로
	humanTiles := humanHands(playerInfo)
	handsTobeNaki := humanTiles + " " + model.SepTargetTile + " " + util.Tile34ToStr(targetTile34) + "?"
	fmt.Println(handsTobeNaki)
	fmt.Println(strings.Repeat("=", len(handsTobeNaki)))

	// 원래 손패 분석 결과
	fmt.Println("현재 " + util.NumberToChineseShanten(result.Shanten) + ":")
	r := &analysisResult{
		discardTile34:  -1,
		result13:       result,
		mixedRiskTable: mixedRiskTable,
	}
	r.printWaitsWithImproves13_oneRow()

	// 안내 메시지
	// TODO: 국 수지가 비슷할 때 화료율 중시면 xx, 타점 중시면 xx를 안내
	if shanten == -1 {
		color.HiRed("[이미 화료]")
	} else if shanten <= 1 {
		// 후로 후 텐파이 또는 1샨텐이면 형식 텐파이를 안내
		if len(results14) > 0 && results14[0].LeftDrawTilesCount > 0 && results14[0].LeftDrawTilesCount <= 16 {
			color.HiGreen("형식 텐파이를 고려?")
		}
	}

	// TODO: 유국이 가까울 때 하이테이 차례가 누구인지 안내

	// 후로 후 무엇을 버릴지에 대한 분석 결과
	printResults14WithRisk(results14, mixedRiskTable)
	printResults14WithRisk(incShantenResults14, mixedRiskTable)
	return nil
}

func analysisHumanTiles(humanTilesInfo *model.HumanTilesInfo) (playerInfo *model.PlayerInfo, err error) {
	defer func() {
		if er := recover(); er != nil {
			err = er.(error)
		}
	}()

	if err = humanTilesInfo.SelfParse(); err != nil {
		return
	}

	tiles34, numRedFives, err := util.StrToTiles34(humanTilesInfo.HumanTiles)
	if err != nil {
		return
	}

	tileCount := util.CountOfTiles34(tiles34)
	if tileCount > 14 {
		return nil, fmt.Errorf("입력 오류: %d장", tileCount)
	}

	if tileCount%3 == 0 {
		color.HiYellow("%s 는 %d장입니다\n도우미가 무작위로 한 장을 보충했습니다", humanTilesInfo.HumanTiles, tileCount)
		util.RandomAddTile(tiles34)
	}

	melds := []model.Meld{}
	for _, humanMeld := range humanTilesInfo.HumanMelds {
		tiles, _numRedFives, er := util.StrToTiles(humanMeld)
		if er != nil {
			return nil, er
		}
		isUpper := humanMeld[len(humanMeld)-1] <= 'Z'
		var meldType int
		switch {
		case len(tiles) == 3 && tiles[0] != tiles[1]:
			meldType = model.MeldTypeChi
		case len(tiles) == 3 && tiles[0] == tiles[1]:
			meldType = model.MeldTypePon
		case len(tiles) == 4 && isUpper:
			meldType = model.MeldTypeAnkan
		case len(tiles) == 4 && !isUpper:
			meldType = model.MeldTypeMinkan
		default:
			return nil, fmt.Errorf("입력 오류: %s", humanMeld)
		}
		containRedFive := false
		for i, c := range _numRedFives {
			if c > 0 {
				containRedFive = true
				numRedFives[i] += c
			}
		}
		melds = append(melds, model.Meld{
			MeldType:       meldType,
			Tiles:          tiles,
			ContainRedFive: containRedFive,
		})
	}

	playerInfo = model.NewSimplePlayerInfo(tiles34, melds)
	playerInfo.NumRedFives = numRedFives

	if humanTilesInfo.HumanDoraTiles != "" {
		playerInfo.DoraTiles, _, err = util.StrToTiles(humanTilesInfo.HumanDoraTiles)
		if err != nil {
			return
		}
	}

	if humanTilesInfo.HumanTargetTile != "" {
		if tileCount%3 == 2 {
			return nil, fmt.Errorf("입력 오류: %s 는 %d장입니다", humanTilesInfo.HumanTiles, tileCount)
		}
		targetTile34, isRedFive, er := util.StrToTile34(humanTilesInfo.HumanTargetTile)
		if er != nil {
			return nil, er
		}
		if er := analysisMeld(playerInfo, targetTile34, isRedFive, true, nil); er != nil {
			return nil, er
		}
		return
	}

	playerInfo.IsTsumo = humanTilesInfo.IsTsumo
	err = analysisPlayerWithRisk(playerInfo, nil)
	return
}
