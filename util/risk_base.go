package util

import "fmt"

// 설명
// TODO: 설명 보완
func calcLowRiskTiles27(safeTiles34 []bool, leftTiles34 []int) []int {
	lowRiskTiles27 := make([]int, 27)
	const _true = 1
	for i, safe := range safeTiles34[:27] {
		if safe {
			lowRiskTiles27[i] = _true
		}
	}
	for i := 0; i < 3; i++ {
		// 설명
		if leftTiles34[9*i+1] == 0 {
			lowRiskTiles27[9*i] = _true
		}
		// 설명
		if leftTiles34[9*i+2] == 0 {
			lowRiskTiles27[9*i] = _true
			lowRiskTiles27[9*i+1] = _true
		}
		// 설명
		if leftTiles34[9*i+3] == 0 {
			lowRiskTiles27[9*i+1] = _true
			lowRiskTiles27[9*i+2] = _true
		}
		// 설명
		if leftTiles34[9*i+5] == 0 {
			lowRiskTiles27[9*i+6] = _true
			lowRiskTiles27[9*i+7] = _true
		}
		// 설명
		if leftTiles34[9*i+6] == 0 {
			lowRiskTiles27[9*i+7] = _true
			lowRiskTiles27[9*i+8] = _true
		}
		// 설명
		if leftTiles34[9*i+7] == 0 {
			lowRiskTiles27[9*i+8] = _true
		}
	}
	return lowRiskTiles27
}

// 설명
func calcTileType27(discardTiles []int) []tileType {
	sujiType27 := make([]tileType, 27)

	safeTiles34 := make([]int, 34)
	// 설명
	for _, tile := range discardTiles {
		safeTiles34[tile] = 1
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			sujiType27[idx] = TileTypeTable[j][safeTiles34[idx+3]]
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			mixSafeTile := safeTiles34[idx-3]<<1 | safeTiles34[idx+3]
			sujiType27[idx] = TileTypeTable[j][mixSafeTile]
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			sujiType27[idx] = TileTypeTable[j][safeTiles34[idx-3]]
		}
	}

	return sujiType27
}

type RiskTiles34 []float64

// 설명
// 설명
// 설명
// 설명
// 설명
// 설명
// 설명
func CalculateRiskTiles34(turns int, safeTiles34 []bool, leftTiles34 []int, doraTiles []int, roundWindTile int, playerWindTile int) (risk34 RiskTiles34) {
	risk34 = make(RiskTiles34, 34)

	// 설명
	// 설명
	doraMulti := func(tile int, tileType tileType) float64 {
		multi := 1.0
		for _, dora := range doraTiles {
			if tile == dora {
				multi *= FixedDoraRiskRateMulti[tileType]
			}
		}
		return multi
	}

	// 설명
	// 설명
	// 설명
	// 설명
	// 설명

	// 설명
	// 설명
	lowRiskTiles27 := calcLowRiskTiles27(safeTiles34, leftTiles34)
	// 설명
	// TODO: 설명 보완
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			t := TileTypeTable[j][lowRiskTiles27[idx+3]]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
			if j == 0 && safeTiles34[idx+3] && leftTiles34[idx] == 0 {
				// 설명
				risk34[idx] = 0
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			mixSafeTile := lowRiskTiles27[idx-3]<<1 | lowRiskTiles27[idx+3]
			t := TileTypeTable[j][mixSafeTile]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			t := TileTypeTable[j][lowRiskTiles27[idx-3]]
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
			if j == 8 && safeTiles34[idx-3] && leftTiles34[idx] == 0 {
				// 설명
				risk34[idx] = 0
			}
		}
		// 설명
		if leftTiles34[9*i+4] == 0 {
			t := tileTypeSuji37
			risk34[9*i+2] = RiskRate[turns][t] * doraMulti(9*i+2, t)
			risk34[9*i+6] = RiskRate[turns][t] * doraMulti(9*i+6, t)
		}
	}
	for i := 27; i < 34; i++ {
		if leftTiles34[i] > 0 {
			// 설명
			isYakuHai := i == roundWindTile || i == playerWindTile || i >= 31
			t := HonorTileType[boolToInt(isYakuHai)][leftTiles34[i]-1]
			risk34[i] = RiskRate[turns][t] * doraMulti(i, t)
		} else {
			// 설명
			risk34[i] = 0
		}
	}

	// TODO: 설명 보완
	// 설명

	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	ncSafeTile34 := CalcNCSafeTiles(leftTiles34)
	for _, ncSafeTile := range ncSafeTile34 {
		idx := ncSafeTile.Tile34
		switch idx%9 + 1 {
		case 1, 9:
			t := tileTypeSuji19
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 2, 8:
			t := tileTypeSuji19
			risk34[idx] = RiskRate[turns][t] * 1.1 * doraMulti(idx, t)
		case 3, 7:
			t := tileTypeSuji28
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 4, 6:
			t := tileTypeDoubleSuji46
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		case 5:
			t := tileTypeDoubleSuji5
			risk34[idx] = RiskRate[turns][t] * doraMulti(idx, t)
		default:
			panic(fmt.Errorf("[CalculateRiskTiles34] 코드 오류: ncSafeTile = %d", ncSafeTile.Tile34))
		}
	}

	// 설명
	// 설명
	dncSafeTiles := CalcDNCSafeTilesWithDiscards(leftTiles34, safeTiles34)
	for _, dncSafeTile := range dncSafeTiles {
		tile := dncSafeTile.Tile34
		if leftTiles34[tile] > 0 {
			t := tileTypeSuji19
			risk34[tile] = RiskRate[turns][t] * doraMulti(tile, t)
			// 설명
			if t9 := tile % 9; t9 > 0 && t9 < 8 {
				risk34[tile] *= 1.1
			}
		} else {
			risk34[tile] = 0
		}
	}

	// 설명
	for i, isSafe := range safeTiles34 {
		if isSafe {
			risk34[i] = 0
		}
	}

	return
}

// 설명
// 설명
func (l RiskTiles34) FixWithEarlyOutside(discardTiles []int) RiskTiles34 {
	for _, dTile := range discardTiles {
		l[dTile] *= 0.4
	}
	return l
}

func (l RiskTiles34) FixWithGlobalMulti(multi float64) RiskTiles34 {
	for i := range l {
		l[i] *= multi
	}
	return l
}

// 설명
func (l RiskTiles34) FixWithPoint(ronPoint float64) RiskTiles34 {
	return l.FixWithGlobalMulti(ronPoint / RonPointRiichiHiIppatsu)
}

// 설명
// 설명
func CalculateLeftNoSujiTiles(safeTiles34 []bool, leftTiles34 []int) (leftNoSujiTiles []int) {
	isNoSujiTiles27 := make([]bool, 27)

	for i := 0; i < 3; i++ {
		// 설명
		for j := 3; j < 6; j++ {
			if !safeTiles34[9*i+j] {
				isNoSujiTiles27[9*i+j-3] = true
				isNoSujiTiles27[9*i+j+3] = true
			}
		}
		// 설명
		if leftTiles34[9*i+4] == 0 {
			isNoSujiTiles27[9*i+2] = false
			isNoSujiTiles27[9*i+6] = false
		}
	}

	// 설명
	for i, c := range leftTiles34[:27] {
		if c == 0 {
			isNoSujiTiles27[i] = false
		}
	}

	// 설명
	lowRiskTiles27 := calcLowRiskTiles27(safeTiles34, leftTiles34)
	const _true = 1
	for i, isSafe := range lowRiskTiles27 {
		if isSafe == _true {
			isNoSujiTiles27[i] = false
		}
	}

	for i, isNoSujiTile := range isNoSujiTiles27 {
		if isNoSujiTile {
			leftNoSujiTiles = append(leftNoSujiTiles, i)
		}
	}

	return
}

// TODO: 설명 보완
// TODO: 설명 보완
// TODO: 설명 보완
