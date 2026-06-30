package util

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"math"
	"sort"
)

// map[개량패]유효패. 유효패 수가 가장 큰 경우를 선택한다.
type Improves map[int]Waits

// 3k+1장 손패 분석 결과
type Hand13AnalysisResult struct {
	// 원래 손패
	Tiles34 []int

	// 남은 패
	LeftTiles34 []int

	// 이미 후로했는지 여부(멘젠이 아닌 상태)
	// 역 없음 등을 판단하는 데 사용한다.
	IsNaki bool

	// 샨텐 수
	Shanten int

	// 유효패
	// 남은 매수를 고려한다.
	// 어떤 유효패 4장이 모두 보이면 그 유효패의 value는 0이다.
	Waits Waits

	// 다마텐일 때의 유효패
	DamaWaits Waits

	// TODO: 후로 유효패: 다른 사람이 이 패를 버렸을 때 후로할 수 있고 샨텐을 전진시키는 패
	//MeldWaits Waits

	// map[유효패]샨텐 전진 후의 최대 유효패 수
	NextShantenWaitsCountMap map[int]int

	// 샨텐 전진 후 최대 유효패 수의 가중 평균
	AvgNextShantenWaitsCount float64

	// 현재 유효패와 샨텐 전진 후 유효패를 종합한 점수
	MixedWaitsScore float64

	// 개량: 이 패를 쯔모해도 샨텐은 전진하지 않지만 유효패가 늘어난다.
	// len(Improves)는 개량패 종류 수다.
	Improves Improves

	// 개량 경우 수. 유효패를 늘리는 쯔모-타패 방식이 몇 가지인지 계산한다.
	ImproveWayCount int

	// 유효패가 아닌 패를 쯔모했을 때의 유효패 수 가중 평균(비개량+개량).
	// 비개량패의 유효패 수는 Waits.AllCount()로 본다.
	// 여기서는 한 순의 개량 평균만 고려한다.
	// TODO: 개량을 고려할 때 샨텐 전진까지 필요한 쯔모 횟수의 기대값을 어떻게 계산할지. 몬테카를로?
	AvgImproveWaitsCount float64

	// 텐파이 시 손패 화료율
	// TODO: 텐파이가 아닐 때의 화료율?
	AvgAgariRate float64

	// 후리텐 가능률(1샨텐과 텐파이 시)
	FuritenRate float64

	// 역
	YakuTypes map[int]struct{}

	// 후로 시 한쪽 대기인지 여부
	IsPartWait bool

	// 도라 개수(손패+후로)
	DoraCount int

	// 리치하지 않은 상태의 기대 타점(후로 또는 다마텐)
	DamaPoint float64

	// 리치 상태의 기대 타점
	RiichiPoint float64

	// 국 수지
	MixedRoundPoint float64

	// TODO: 적도라 개량 안내
}

// 설명
// 설명
func (r *Hand13AnalysisResult) speedScore() float64 {
	if r.Waits.AllCount() == 0 || r.AvgNextShantenWaitsCount == 0 {
		return 0
	}
	leftCount := float64(CountOfTiles34(r.LeftTiles34))
	p2 := float64(r.Waits.AllCount()) / leftCount
	//p2 := r.AvgImproveWaitsCount / leftCount
	p1 := r.AvgNextShantenWaitsCount / leftCount
	// TODO: 설명 보완
	//	p1 = r.AvgAgariRate / 100
	//}
	p2_, p1_ := 1-p2, 1-p1
	const leftTurns = 10.0 // math.Max(5.0, leftCount/4)
	sumP2 := p2_ * (1 - math.Pow(p2_, leftTurns)) / p2
	sumP1 := p1_ * (1 - math.Pow(p1_, leftTurns)) / p1
	result := p2 * p1 * (sumP2 - sumP1) / (p2_ - p1_)
	return result * 100
}

func (r *Hand13AnalysisResult) mixedRoundPoint() float64 {
	const weight = -1500
	if r.RiichiPoint > 0 {
		return r.AvgAgariRate/100*(r.RiichiPoint+1500) + weight
	}
	return r.AvgAgariRate/100*(r.DamaPoint+1500) + weight
}

// 설명
func (r *Hand13AnalysisResult) String() string {
	s := fmt.Sprintf("%d 유효패 %s\n%.2f 개량 유효패 [%d(%d)종]",
		r.Waits.AllCount(),
		//r.Waits.AllCount()+r.MeldWaits.AllCount(),
		TilesToKoreanStrWithBracket(r.Waits.indexes()),
		r.AvgImproveWaitsCount,
		len(r.Improves),
		r.ImproveWayCount,
	)
	if len(r.DamaWaits) > 0 {
		s += fmt.Sprintf("(다마텐 유효패 %s)", TilesToKoreanStrWithBracket(r.DamaWaits.indexes()))
	}
	if r.Shanten >= 1 {
		mixedScore := r.MixedWaitsScore
		//for i := 2; i <= r.Shanten; i++ {
		//	mixedScore /= 4
		//}
		s += fmt.Sprintf(" %.2f %s유효패(%.2f 종합 점수)",
			r.AvgNextShantenWaitsCount,
			NumberToChineseShanten(r.Shanten-1),
			mixedScore,
		)
	}
	if r.AvgAgariRate > 0 {
		s += fmt.Sprintf("[%.2f%% 화료율] ", r.AvgAgariRate)
	}
	if r.MixedRoundPoint > 0 {
		s += fmt.Sprintf(" [국 수지%d]", int(math.Round(r.MixedRoundPoint)))
	}
	if r.DamaPoint > 0 {
		s += fmt.Sprintf("[다마텐%d]", int(math.Round(r.DamaPoint)))
	}
	if r.RiichiPoint > 0 {
		s += fmt.Sprintf("[리치%d]", int(math.Round(r.RiichiPoint)))
	}
	if r.Shanten >= 0 && r.Shanten <= 1 {
		if r.FuritenRate > 0 {
			if r.FuritenRate < 1 {
				s += "[후리텐 가능성]"
			} else {
				s += "[후리텐]"
			}
		}
	}
	if len(r.YakuTypes) > 0 {
		s += YakuTypesWithDoraToStr(r.YakuTypes, r.DoraCount)
	}
	return s
}

func (n *shantenSearchNode13) analysis(playerInfo *model.PlayerInfo, considerImprove bool) (result13 *Hand13AnalysisResult) {
	tiles34 := playerInfo.HandTiles34
	leftTiles34 := playerInfo.LeftTiles34
	shanten13 := n.shanten
	waits := n.waits
	waitsCount := waits.AllCount()

	nextShantenWaitsCountMap := map[int]int{} // 설명
	improves := Improves{}
	improveWayCount := 0
	// 설명
	maxImproveWaitsCount34 := make([]int, 34)
	for i := 0; i < 34; i++ {
		maxImproveWaitsCount34[i] = waitsCount // 설명
	}
	avgRoundPoint := 0.0
	roundPointWeight := 0
	yakuTypes := map[int]struct{}{}

	for i := 0; i < 34; i++ {
		// 설명
		if leftTiles34[i] == 0 {
			continue
		}
		leftTiles34[i]--
		tiles34[i]++

		if node14, ok := n.children[i]; ok && node14 != nil { // 설명
			// 설명
			maxNextShantenWaitsCount := 0
			for _, node13 := range node14.children {
				maxNextShantenWaitsCount = MaxInt(maxNextShantenWaitsCount, node13.waits.AllCount())
			}
			nextShantenWaitsCountMap[i] = maxNextShantenWaitsCount

			//const minRoundPoint = -1e10
			//maxRoundPoint := minRoundPoint

			if results14 := node14.analysis(playerInfo, false); len(results14) > 0 {
				bestResult14 := results14[0]

				// 설명
				w := leftTiles34[i] + 1
				avgRoundPoint += float64(w) * bestResult14.Result13.MixedRoundPoint
				roundPointWeight += w

				// 설명
				for t := range bestResult14.Result13.YakuTypes {
					yakuTypes[t] = struct{}{}
				}
			}

			//for discardTile, node13 := range node14.children {
			//
			//
			// 설명
			// TODO: 설명 보완
			//	_isRedFive := playerInfo.IsOnlyRedFive(discardTile)
			//	playerInfo.DiscardTile(discardTile, _isRedFive)
			//
			// 설명
			//	if newShanten13 == 0 {
			// 설명
			//		_avgAgariRate := CalculateAvgAgariRate(newWaits, playerInfo) / 100
			//		var _roundPoint float64
			//		if isNaki {
			// FIXME: 설명 보완
			//			_avgPoint, _ := CalcAvgPoint(*playerInfo, newWaits)
			// 설명
			//				_avgAgariRate = 0
			//			}
			//			_roundPoint = _avgAgariRate*(_avgPoint+1500) - 1500
			//		} else {
			//			_avgRiichiPoint, _ := CalcAvgRiichiPoint(*playerInfo, newWaits)
			//			_roundPoint = _avgAgariRate*(_avgRiichiPoint+1500) - 1500
			//		}
			//		maxRoundPoint = math.Max(maxRoundPoint, _roundPoint)
			// 설명
			//		//fillYakuTypes(newShanten13, newWaits)
			//	}
			//
			//	playerInfo.UndoDiscardTile(discardTile, _isRedFive)
			//}
			// 설명
			//w := leftTiles34[i] + 1
			////avgAgariRate += maxAgariRate * float64(w)
			//if maxRoundPoint > minRoundPoint {
			//	avgRoundPoint += float64(w) * maxRoundPoint
			//	roundPointWeight += w
			//}
			//fmt.Println(i, maxAvgRiichiRonPoint)
			//avgRiichiPoint += maxAvgRiichiRonPoint * float64(w)
		} else if considerImprove { // 설명
			for j := 0; j < 34; j++ {
				if tiles34[j] == 0 || j == i {
					continue
				}
				// 설명
				// TODO: 설명 보완
				_isRedFive := playerInfo.IsOnlyRedFive(j)
				playerInfo.DiscardTile(j, _isRedFive)
				// 설명
				if newShanten13, improveWaits := CalculateShantenAndWaits13(tiles34, leftTiles34); newShanten13 == shanten13 {
					// 설명
					// TODO: 설명 보완
					if improveWaitsCount := improveWaits.AllCount(); improveWaitsCount > waitsCount {
						improveWayCount++
						if improveWaitsCount > maxImproveWaitsCount34[i] {
							maxImproveWaitsCount34[i] = improveWaitsCount
							// 설명
							improves[i] = improveWaits
						}
						// 설명
					}
				}
				playerInfo.UndoDiscardTile(j, _isRedFive)
			}
		}

		tiles34[i]--
		leftTiles34[i]++
	}

	_tiles34 := make([]int, 34)
	copy(_tiles34, tiles34)
	result13 = &Hand13AnalysisResult{
		Tiles34:                  _tiles34,
		LeftTiles34:              leftTiles34,
		IsNaki:                   playerInfo.IsNaki(),
		Shanten:                  shanten13,
		Waits:                    waits,
		DamaWaits:                Waits{},
		NextShantenWaitsCountMap: nextShantenWaitsCountMap,
		Improves:                 improves,
		ImproveWayCount:          improveWayCount,
		AvgImproveWaitsCount:     float64(waitsCount),
		YakuTypes:                yakuTypes,
		DoraCount:                playerInfo.CountDora(),
	}

	// 설명
	if waitsCount > 0 {
		//avgAgariRate /= float64(waitsCount)
		if roundPointWeight > 0 {
			avgRoundPoint /= float64(roundPointWeight)
			//if shanten13 == 1 {
			// TODO: 설명 보완
			//} else if shanten13 == 2 {
			// TODO: 설명 보완
			//}
		}
		//avgRiichiPoint /= float64(waitsCount)
		if shanten13 == shantenStateTenpai {
			// TODO: 설명 보완
			avgRonPoint, pointResults := CalcAvgPoint(*playerInfo, waits)
			result13.DamaPoint = avgRonPoint
			// 설명
			for _, pr := range pointResults {
				result13.DamaWaits[pr.winTile] = leftTiles34[pr.winTile]
			}

			if !result13.IsNaki {
				avgRiichiPoint, riichiPointResults := CalcAvgRiichiPoint(*playerInfo, waits)
				result13.RiichiPoint = avgRiichiPoint
				result13.AvgAgariRate = CalculateAvgAgariRate(waits, playerInfo)
				for _, pr := range riichiPointResults {
					for _, yakuType := range pr.yakuTypes {
						result13.YakuTypes[yakuType] = struct{}{}
					}
				}
			} else {
				// 설명
				agariRate := 0.0
				for _, pr := range pointResults { // 설명
					agariRate = agariRate + pr.agariRate - agariRate*pr.agariRate/100
					for _, yakuType := range pr.yakuTypes {
						result13.YakuTypes[yakuType] = struct{}{}
					}
				}
				result13.AvgAgariRate = agariRate

				// 설명
				result13.IsPartWait = len(pointResults) < len(waits.AvailableTiles())
			}
		}
	}

	// 설명
	if len(playerInfo.Melds) == 0 && shanten13 == 3 && CountPairsOfTiles34(tiles34)+shanten13 == 6 {
		// 설명
		if waitsCount <= 21 {
			result13.YakuTypes[YakuChiitoi] = struct{}{}
		}
	}

	// 설명
	if shanten13 <= 1 {
		for _, discardTile := range playerInfo.DiscardTiles {
			if _, ok := waits[discardTile]; ok {
				result13.FuritenRate = 0.5 // TODO: 설명 보완
				if shanten13 == shantenStateTenpai {
					result13.FuritenRate = 1
				}
			}
		}
	}

	// 설명
	//if shanten13 <= 1 {
	//result13.DamaPoint = avgRonPoint
	//if !result13.IsNaki {
	//	result13.RiichiPoint = avgRiichiPoint
	//}
	// 설명
	//if result13.FuritenRate == 1 && result13.RiichiPoint > 0 {
	//	result13.DamaPoint = 0
	//}
	if shanten13 == shantenStateTenpai {
		result13.MixedRoundPoint = result13.mixedRoundPoint()
	} else {
		result13.MixedRoundPoint = avgRoundPoint
	}
	//}

	// 설명
	if len(nextShantenWaitsCountMap) > 0 {
		nextShantenWaitsSum := 0
		weight := 0
		for tile, c := range nextShantenWaitsCountMap {
			w := leftTiles34[tile]
			nextShantenWaitsSum += w * c
			weight += w
		}
		result13.AvgNextShantenWaitsCount = float64(nextShantenWaitsSum) / float64(weight)
	}
	if len(improves) > 0 {
		improveWaitsSum := 0
		weight := 0
		for i := 0; i < 34; i++ {
			w := leftTiles34[i]
			improveWaitsSum += w * maxImproveWaitsCount34[i]
			weight += w
		}
		result13.AvgImproveWaitsCount = float64(improveWaitsSum) / float64(weight)
	}
	result13.MixedWaitsScore = result13.speedScore()

	// 설명
	if shanten13 == 2 {
		result13.MixedWaitsScore /= 4 // TODO: 설명 보완
	}

	return
}

func _stopShanten(shanten int) int {
	if shanten >= 3 {
		return shanten - 1
	}
	return shanten - 2
}

// 3k+1장 패에서 샹텐 수, 유효패, 개량 등을 계산한다(남은 매수 고려).
func CalculateShantenWithImproves13(playerInfo *model.PlayerInfo) (r *Hand13AnalysisResult) {
	if len(playerInfo.LeftTiles34) == 0 {
		playerInfo.FillLeftTiles34()
	}

	shanten := CalculateShanten(playerInfo.HandTiles34)
	shantenSearchRoot := _search13(shanten, playerInfo, _stopShanten(shanten))
	return shantenSearchRoot.analysis(playerInfo, true)
}

//

const (
	honorRiskRoundWind = 4
	honorRiskYaku      = 3
	honorRiskOtakaze   = 2
	honorRiskSelfWind  = 1
)

type tileValue float64

const (
	doraValue                tileValue = 10000
	doraFirstNeighbourValue  tileValue = 1000
	doraSecondNeighbourValue tileValue = 100
	honoredValue             tileValue = 15
)

func calculateIsolatedTileValue(tile int, playerInfo *model.PlayerInfo) tileValue {
	value := tileValue(100)

	// 설명
	for _, doraTile := range playerInfo.DoraTiles {
		if tile == doraTile {
			value += doraValue
			//} else if doraTile < 27 {
			//	if tile/3 != doraTile/3 {
			//		continue
			//	}
			//	t9 := tile % 9
			//	dt9 := doraTile % 9
			//	if t9+1 == dt9 || t9-1 == dt9 {
			//		value += doraFirstNeighbourValue
			//	} else if t9+2 == dt9 || t9-2 == dt9 {
			//		value += doraSecondNeighbourValue
			//	}
		}
	}

	if tile >= 27 {
		if tile == playerInfo.SelfWindTile || tile == playerInfo.RoundWindTile || tile >= 31 {
			// 설명
			value += honoredValue
			if playerInfo.SelfWindTile == playerInfo.RoundWindTile && tile == playerInfo.SelfWindTile {
				value += honoredValue // 설명
			} else if tile == playerInfo.SelfWindTile {
				value++ // 설명
			} else if tile == playerInfo.RoundWindTile {
				value-- // 설명
			}
			if tile == 31 {
				value -= 0.1
			}
			if tile == 32 {
				value -= 0.2
			}
		} else {
			// 설명
			for i := 1; i <= 3; i++ {
				otakazeTile := playerInfo.SelfWindTile + i
				if otakazeTile > 30 {
					otakazeTile -= 4
				}
				if tile == otakazeTile {
					// 설명
					value -= tileValue(4 - i)
					break
				}
			}
		}
		left := playerInfo.LeftTiles34[tile]
		if left == 2 {
			value *= 0.9
		} else if left == 1 {
			value *= 0.2
		} else if left == 0 {
			value = 0
		}
	}

	return value
}

func calculateTileValue(tile int, playerInfo *model.PlayerInfo) (value tileValue) {
	// 설명
	for _, doraTile := range playerInfo.DoraTiles {
		if tile == doraTile {
			value += doraValue
		} else if doraTile < 27 {
			if tile/3 != doraTile/3 {
				continue
			}
			t9 := tile % 9
			dt9 := doraTile % 9
			if t9+1 == dt9 || t9-1 == dt9 {
				value += doraFirstNeighbourValue
			} else if t9+2 == dt9 || t9-2 == dt9 {
				value += doraSecondNeighbourValue
			}
		}
	}
	return
}

type Hand14AnalysisResult struct {
	// 설명
	DiscardTile int

	// 설명
	IsDiscardDoraTile bool

	// 설명
	DiscardTileValue tileValue

	// 설명
	isIsolatedYaochuDiscardTile bool

	// 설명
	Result13 *Hand13AnalysisResult

	DiscardHonorTileRisk int

	// 설명
	LeftDrawTilesCount int

	// 설명
	// 설명
	OpenTiles []int
}

func (r *Hand14AnalysisResult) String() string {
	meldInfo := ""
	if len(r.OpenTiles) > 0 {
		meldType := "치"
		if r.OpenTiles[0] == r.OpenTiles[1] {
			meldType = "퐁"
		}
		meldInfo = fmt.Sprintf("%s%s %s, ", string([]rune(MahjongZH[r.OpenTiles[0]])[:1]), MahjongZH[r.OpenTiles[1]], meldType)
	}
	return meldInfo + fmt.Sprintf("%s 타: %s", MahjongZH[r.DiscardTile], r.Result13.String())
}

type Hand14AnalysisResultList []*Hand14AnalysisResult

// 설명
// 설명
func (l Hand14AnalysisResultList) Sort(improveFirst bool) {
	if len(l) <= 1 {
		return
	}

	shanten := l[0].Result13.Shanten

	sort.Slice(l, func(i, j int) bool {
		ri, rj := l[i].Result13, l[j].Result13
		riWaitsCount, rjWaitsCount := ri.Waits.AllCount(), rj.Waits.AllCount()

		// 설명
		// 설명
		if riWaitsCount == 0 || rjWaitsCount == 0 {
			if riWaitsCount == 0 && rjWaitsCount == 0 {
				return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
			}
			return riWaitsCount > rjWaitsCount
		}

		switch shanten {
		case 0:
			// 설명
			// 설명
			if !InDelta(ri.MixedRoundPoint, rj.MixedRoundPoint, 100) {
				return ri.MixedRoundPoint > rj.MixedRoundPoint
			}
			// 설명
			if !Equal(ri.AvgAgariRate, rj.AvgAgariRate) {
				return ri.AvgAgariRate > rj.AvgAgariRate
			}
		case 1:
			// 설명
			var riScore, rjScore float64
			if shanten >= 2 && improveFirst {
				// 설명
				//riScore = float64(ri.AvgImproveWaitsCount) * ri.MixedRoundPoint
				//rjScore = float64(rj.AvgImproveWaitsCount) * rj.MixedRoundPoint
				break
			} else {
				// 설명
				wi := float64(riWaitsCount)
				if ri.MixedRoundPoint < 0 {
					wi = 1 / wi
				}
				wj := float64(rjWaitsCount)
				if rj.MixedRoundPoint < 0 {
					wj = 1 / wj
				}
				riScore = wi * ri.MixedRoundPoint
				rjScore = wj * rj.MixedRoundPoint
			}
			if !Equal(riScore, rjScore) {
				return riScore > rjScore
			}
		}

		if shanten >= 2 {
			// 설명
			if l[i].isIsolatedYaochuDiscardTile && l[j].isIsolatedYaochuDiscardTile {
				// 설명
				if l[i].DiscardTileValue != l[j].DiscardTileValue {
					return l[i].DiscardTileValue < l[j].DiscardTileValue
				}
			} else if l[i].isIsolatedYaochuDiscardTile && l[i].DiscardTileValue < 500 {
				return true
			} else if l[j].isIsolatedYaochuDiscardTile && l[j].DiscardTileValue < 500 {
				return false
			}
		}

		//if improveFirst {
		// 설명
		//	if !Equal(ri.AvgImproveWaitsCount, rj.AvgImproveWaitsCount) {
		//		return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
		//	}
		//}

		// 설명
		// 설명
		// 설명
		// 설명

		if !Equal(ri.MixedWaitsScore, rj.MixedWaitsScore) {
			return ri.MixedWaitsScore > rj.MixedWaitsScore
		}

		if riWaitsCount != rjWaitsCount {
			return riWaitsCount > rjWaitsCount
		}

		if !Equal(ri.AvgNextShantenWaitsCount, rj.AvgNextShantenWaitsCount) {
			return ri.AvgNextShantenWaitsCount > rj.AvgNextShantenWaitsCount
		}

		// shanten == 1
		if !Equal(ri.AvgAgariRate, rj.AvgAgariRate) {
			return ri.AvgAgariRate > rj.AvgAgariRate
		}

		if !Equal(ri.AvgImproveWaitsCount, rj.AvgImproveWaitsCount) {
			return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
		}

		if l[i].DiscardTileValue != l[j].DiscardTileValue {
			// 설명
			return l[i].DiscardTileValue < l[j].DiscardTileValue
		}

		// 설명
		idxI, idxJ := l[i].DiscardTile, l[j].DiscardTile
		if idxI < 27 && idxJ < 27 {
			idxI %= 9
			if idxI > 4 {
				idxI = 8 - idxI
			}
			idxJ %= 9
			if idxJ > 4 {
				idxJ = 8 - idxJ
			}
			return idxI > idxJ
		}
		if idxI < 27 || idxJ < 27 {
			// 설명
			return idxI < idxJ
		}
		// 설명
		return l[i].DiscardHonorTileRisk > l[j].DiscardHonorTileRisk

		// 설명
		//if len(ri.Improves) != len(rj.Improves) {
		//	return len(ri.Improves) > len(rj.Improves)
		//}
		//if ri.ImproveWayCount != rj.ImproveWayCount {
		//	return ri.ImproveWayCount > rj.ImproveWayCount
		//}
	})
}

func (l *Hand14AnalysisResultList) filterOutDiscard(cantDiscardTile int) {
	newResults := Hand14AnalysisResultList{}
	for _, r := range *l {
		if r.DiscardTile != cantDiscardTile {
			newResults = append(newResults, r)
		}
	}
	*l = newResults
}

func (l Hand14AnalysisResultList) addOpenTile(openTiles []int) {
	for _, r := range l {
		r.OpenTiles = openTiles
	}
}

func (n *shantenSearchNode14) analysis(playerInfo *model.PlayerInfo, considerImprove bool) (results Hand14AnalysisResultList) {
	for discardTile, node13 := range n.children {
		isRedFive := playerInfo.IsOnlyRedFive(discardTile)

		// 패를 버린 뒤 3k+1장 상태의 손패를 분석한다.
		// 이 패가 5라면 적5만 있는 경우에만 적5를 버린다(TODO: 적5로 37을 속이는 경우 고려).
		playerInfo.DiscardTile(discardTile, isRedFive)
		result13 := node13.analysis(playerInfo, considerImprove)

		// 타패 후의 분석 결과를 기록한다.
		r14 := &Hand14AnalysisResult{
			DiscardTile:        discardTile,
			IsDiscardDoraTile:  InInts(discardTile, playerInfo.DoraTiles),
			Result13:           result13,
			LeftDrawTilesCount: playerInfo.LeftDrawTilesCount,
		}
		results = append(results, r14)

		if n.shanten >= 2 {
			if isYaochupai(discardTile) && isIsolatedTile(discardTile, playerInfo.HandTiles34) {
				r14.isIsolatedYaochuDiscardTile = true
				r14.DiscardTileValue = calculateIsolatedTileValue(discardTile, playerInfo)
			} else {
				r14.DiscardTileValue = calculateTileValue(discardTile, playerInfo)
			}
		}

		if discardTile >= 27 {
			switch discardTile {
			case playerInfo.RoundWindTile:
				r14.DiscardHonorTileRisk = honorRiskRoundWind
			case 31, 32, 33:
				r14.DiscardHonorTileRisk = honorRiskYaku
			case playerInfo.SelfWindTile:
				r14.DiscardHonorTileRisk = honorRiskSelfWind
			default:
				r14.DiscardHonorTileRisk = honorRiskOtakaze
			}
		}

		playerInfo.UndoDiscardTile(discardTile, isRedFive)
	}

	// 아래 로직은 "종합 속도"로 대체되었다.
	//improveFirst := func(l []*Hand14AnalysisResult) bool {
	//	if !considerImprove || len(l) <= 1 {
	//		return false
	//	}
	//
	//	shanten := l[0].Result13.Shanten
	//	// 이샹텐 이하는 유효패를 우선하고, 개량은 그 다음으로 본다.
	//	if shanten <= 1 {
	//		return false
	//	}
	//
	//	// 치또이와 일반형의 샹텐 수를 비교해 치또이가 더 작으면 개량을 우선한다.
	//	tiles34 := playerInfo.HandTiles34
	//	shantenChiitoi := CalculateShantenOfChiitoi(tiles34)
	//	shantenNormal := CalculateShantenOfNormal(tiles34, CountOfTiles34(tiles34))
	//	return shantenChiitoi < shantenNormal
	//}
	//improveFst := improveFirst(results)

	results.Sort(false)

	return
}

// 설명
func CalculateShantenWithImproves14(playerInfo *model.PlayerInfo) (shanten int, results Hand14AnalysisResultList, incShantenResults Hand14AnalysisResultList) {
	if len(playerInfo.LeftTiles34) == 0 {
		playerInfo.FillLeftTiles34()
	}

	shanten = CalculateShanten(playerInfo.HandTiles34)
	stopAtShanten := _stopShanten(shanten)
	shantenSearchRoot := searchShanten14(shanten, playerInfo, stopAtShanten)
	results = shantenSearchRoot.analysis(playerInfo, true)
	incShantenSearchRoot := searchShanten14(shanten+1, playerInfo, stopAtShanten+1)
	incShantenResults = incShantenSearchRoot.analysis(playerInfo, true)
	return
}

// 설명
func calculateMeldShanten(tiles34 []int, calledTile int, isRedFive bool, allowChi bool) (minShanten int, meldCombinations []model.Meld) {
	// 설명
	if tiles34[calledTile] >= 2 {
		meldCombinations = append(meldCombinations, model.Meld{
			MeldType:          model.MeldTypePon,
			Tiles:             []int{calledTile, calledTile, calledTile},
			SelfTiles:         []int{calledTile, calledTile},
			CalledTile:        calledTile,
			RedFiveFromOthers: isRedFive,
		})
	}
	// 설명
	if allowChi && calledTile < 27 {
		checkChi := func(tileA, tileB int) {
			if tiles34[tileA] > 0 && tiles34[tileB] > 0 {
				_tiles := []int{tileA, tileB, calledTile}
				sort.Ints(_tiles)
				meldCombinations = append(meldCombinations, model.Meld{
					MeldType:          model.MeldTypeChi,
					Tiles:             _tiles,
					SelfTiles:         []int{tileA, tileB},
					CalledTile:        calledTile,
					RedFiveFromOthers: isRedFive,
				})
			}
		}
		t9 := calledTile % 9
		if t9 >= 2 {
			checkChi(calledTile-2, calledTile-1)
		}
		if t9 >= 1 && t9 <= 7 {
			checkChi(calledTile-1, calledTile+1)
		}
		if t9 <= 6 {
			checkChi(calledTile+1, calledTile+2)
		}
	}

	// 설명
	minShanten = 99
	for _, c := range meldCombinations {
		tiles34[c.SelfTiles[0]]--
		tiles34[c.SelfTiles[1]]--
		minShanten = MinInt(minShanten, CalculateShanten(tiles34))
		tiles34[c.SelfTiles[0]]++
		tiles34[c.SelfTiles[1]]++
	}

	return
}

// TODO: 설명 보완
// 설명
//if isOpen {
//if newShanten, combinations, shantens := calculateMeldShanten(tiles34, i, true); newShanten < shanten {
// 설명
// 설명
//	meldWaits[i] = leftTile - tiles34[i]
//	for i, comb := range combinations {
//		if comb[0] == comb[1] && shantens[i] == newShanten {
//			meldWaits[i] *= 3
//			break
//		}
//	}
//}
//}

// 설명
// 설명
// 설명
// 설명
func CalculateMeld(playerInfo *model.PlayerInfo, calledTile int, isRedFive bool, allowChi bool) (minShanten int, results Hand14AnalysisResultList, incShantenResults Hand14AnalysisResultList) {
	if len(playerInfo.LeftTiles34) == 0 {
		playerInfo.FillLeftTiles34()
	}

	minShanten, meldCombinations := calculateMeldShanten(playerInfo.HandTiles34, calledTile, isRedFive, allowChi)

	for _, c := range meldCombinations {
		// 설명
		playerInfo.AddMeld(c)
		_shanten, _results, _incShantenResults := CalculateShantenWithImproves14(playerInfo)
		playerInfo.UndoAddMeld()

		// 설명
		_results.filterOutDiscard(calledTile)
		_incShantenResults.filterOutDiscard(calledTile)

		// 설명
		if c.MeldType == model.MeldTypeChi {
			cannotDiscardTile := -1
			if c.SelfTiles[0] < calledTile && c.SelfTiles[1] < calledTile && calledTile%9 >= 3 {
				cannotDiscardTile = calledTile - 3
			} else if c.SelfTiles[0] > calledTile && c.SelfTiles[1] > calledTile && calledTile%9 <= 5 {
				cannotDiscardTile = calledTile + 3
			}
			if cannotDiscardTile != -1 {
				_results.filterOutDiscard(cannotDiscardTile)
				_incShantenResults.filterOutDiscard(cannotDiscardTile)
			}
		}

		// 설명
		_results.addOpenTile(c.SelfTiles)
		_incShantenResults.addOpenTile(c.SelfTiles)

		// 설명
		if _shanten == minShanten {
			results = append(results, _results...)
			incShantenResults = append(incShantenResults, _incShantenResults...)
		} else if _shanten == minShanten+1 {
			incShantenResults = append(incShantenResults, _results...)
		}
	}

	results.Sort(false)
	incShantenResults.Sort(false)

	return
}
