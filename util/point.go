package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

// TODO: 설명 보완

func roundUpPoint(point int) int {
	if point == 0 {
		return 0
	}
	return ((point-1)/100 + 1) * 100
}

func calcBasicPoint(han int, fu int, yakumanTimes int) (basicPoint int) {
	switch {
	case yakumanTimes > 0: // 설명
		basicPoint = 8000 * yakumanTimes
	case han >= 13: // 설명
		basicPoint = 8000
	case han >= 11: // 설명
		basicPoint = 6000
	case han >= 8: // 설명
		basicPoint = 4000
	case han >= 6: // 설명
		basicPoint = 3000
	default:
		basicPoint = fu * (1 << uint(2+han))
		if basicPoint > 2000 { // 설명
			basicPoint = 2000
		}
	}
	return
}

// 설명
// 설명
func CalcPointRon(han int, fu int, yakumanTimes int, isParent bool) (point int) {
	basicPoint := calcBasicPoint(han, fu, yakumanTimes)
	if isParent {
		point = 6 * basicPoint
	} else {
		point = 4 * basicPoint
	}
	return roundUpPoint(point)
}

// 설명
// 설명
func CalcPointTsumo(han int, fu int, yakumanTimes int, isParent bool) (childPoint int, parentPoint int) {
	basicPoint := calcBasicPoint(han, fu, yakumanTimes)
	if isParent {
		childPoint = 2 * basicPoint
	} else {
		childPoint = basicPoint
		parentPoint = 2 * basicPoint
	}
	return roundUpPoint(childPoint), roundUpPoint(parentPoint)
}

// 설명
// 설명
func CalcPointTsumoSum(han int, fu int, yakumanTimes int, isParent bool) int {
	childPoint, parentPoint := CalcPointTsumo(han, fu, yakumanTimes, isParent)
	if isParent {
		return 3 * childPoint
	}
	return 2*childPoint + parentPoint
}

//

type PointResult struct {
	Point      int
	FixedPoint float64 // 설명

	han          int
	fu           int
	yakumanTimes int
	isParent     bool

	divideResult *DivideResult
	winTile      int
	yakuTypes    []int
	agariRate    float64 // 설명
}

// 설명
// 설명
// 설명
func CalcPoint(playerInfo *model.PlayerInfo) (result *PointResult) {
	result = &PointResult{}
	isNaki := playerInfo.IsNaki()
	var han, fu int
	numDora := playerInfo.CountDora()
	for _, divideResult := range DivideTiles34(playerInfo.HandTiles34) {
		_hi := &_handInfo{
			PlayerInfo:   playerInfo,
			divideResult: divideResult,
		}
		yakuTypes := findYakuTypes(_hi, isNaki)
		if len(yakuTypes) == 0 {
			// 설명
			continue
		}
		yakumanTimes := CalcYakumanTimes(yakuTypes, isNaki)
		if yakumanTimes == 0 {
			han = CalcYakuHan(yakuTypes, isNaki)
			han += numDora
			fu = _hi.calcFu(isNaki)
		}
		var pt int
		if _hi.IsTsumo {
			pt = CalcPointTsumoSum(han, fu, yakumanTimes, _hi.IsParent)
		} else {
			pt = CalcPointRon(han, fu, yakumanTimes, _hi.IsParent)
		}
		_result := &PointResult{
			pt,
			float64(pt),
			han,
			fu,
			yakumanTimes,
			_hi.IsParent,
			divideResult,
			_hi.WinTile,
			yakuTypes,
			0.0, // 설명
		}
		// 설명
		if pt > result.Point {
			result = _result
		} else if pt == result.Point {
			if han > result.han {
				result = _result
			}
		}
	}
	return
}

// 설명
// 설명
// 설명
func CalcAvgPoint(playerInfo model.PlayerInfo, waits Waits) (avgPoint float64, pointResults []*PointResult) {
	isFuriten := playerInfo.IsFuriten(waits)
	if isFuriten {
		// 설명
		if !playerInfo.IsRiichi {
			playerInfo.IsTsumo = true
		}
	}

	tileAgariRate := CalculateAgariRateOfEachTile(waits, &playerInfo)
	sum := 0.0
	weight := 0.0
	for tile, left := range waits {
		if left == 0 {
			continue
		}
		playerInfo.HandTiles34[tile]++
		playerInfo.WinTile = tile
		result := CalcPoint(&playerInfo) // 설명
		playerInfo.HandTiles34[tile]--
		if result.Point == 0 {
			// 설명
			continue
		}
		pt := float64(result.Point)
		if playerInfo.IsRiichi {
			// 설명
			pt = result.fixedRiichiPoint(isFuriten)
			result.FixedPoint = pt
		}
		w := tileAgariRate[tile]
		sum += pt * w
		weight += w
		result.agariRate = w
		pointResults = append(pointResults, result)
	}
	if weight > 0 {
		avgPoint = sum / weight
	}
	return
}

// 설명
// 설명
// TODO: 설명 보완
// TODO: 설명 보완
func CalcAvgRiichiPoint(playerInfo model.PlayerInfo, waits Waits) (avgRiichiPoint float64, pointResults []*PointResult) {
	if playerInfo.IsNaki() {
		return 0, nil
	}
	playerInfo.IsRiichi = true
	return CalcAvgPoint(playerInfo, waits)
}
