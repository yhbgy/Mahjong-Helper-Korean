package util

import "sort"

const (
	WallSafeTypeDoubleNoChance = iota // 설명
	WallSafeTypeNoChance              // 설명
	WallSafeTypeDoubleOneChance
	WallSafeTypeMixedOneChance // 설명
	WallSafeTypeOneChance
)

type WallSafeTile struct {
	Tile34   int
	SafeType int
}

type WallSafeTileList []WallSafeTile

func (l WallSafeTileList) String() string {
	tiles := []int{}
	for _, t := range l {
		tiles = append(tiles, t.Tile34)
	}
	return TilesToStr(tiles)
}

func (l WallSafeTileList) sort() {
	normalIndex := func(tile34 int) int {
		idx := tile34 % 9
		if idx >= 5 {
			// 5678 -> 3210
			idx = 8 - idx
		}
		return idx
	}

	sort.Slice(l, func(i, j int) bool {
		li, lj := l[i], l[j]

		liIndex := normalIndex(li.Tile34)
		ljIndex := normalIndex(lj.Tile34)
		// 설명
		if liIndex > 2 && ljIndex <= 2 || liIndex <= 2 && ljIndex > 2 {
			return liIndex < ljIndex
		}

		if li.SafeType != lj.SafeType {
			return li.SafeType < lj.SafeType
		}

		return liIndex < ljIndex
	})
}

func (l WallSafeTileList) FilterWithHands(handsTiles34 []int) WallSafeTileList {
	newSafeTiles34 := WallSafeTileList{}
	for _, safeTile := range l {
		if handsTiles34[safeTile.Tile34] > 0 {
			newSafeTiles34 = append(newSafeTiles34, safeTile)
		}
	}
	newSafeTiles34.sort()
	return newSafeTiles34
}

// 설명
func CalcDNCSafeTiles(leftTiles34 []int) (dncSafeTiles WallSafeTileList) {
	nc := func(idx int) bool {
		return leftTiles34[idx] == 0
	}
	or := func(idx ...int) bool {
		for _, i := range idx {
			if nc(i) {
				return true
			}
		}
		return false
	}
	and := func(idx ...int) bool {
		for _, i := range idx {
			if !nc(i) {
				return false
			}
		}
		return true
	}

	const safeType = WallSafeTypeDoubleNoChance
	for i := 0; i < 3; i++ {
		// 설명
		if or(9*i+1, 9*i+2) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9 * i, safeType})
		}
		// 설명
		if nc(9*i+2) || and(9*i, 9*i+3) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9*i + 1, safeType})
		}
		// 설명
		for j := 2; j <= 6; j++ {
			idx := 9*i + j
			if and(idx-2, idx+1) || and(idx-1, idx+1) || and(idx-1, idx+2) {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		// 설명
		if nc(9*i+6) || and(9*i+5, 9*i+8) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9*i + 7, safeType})
		}
		// 설명
		if or(9*i+6, 9*i+7) {
			dncSafeTiles = append(dncSafeTiles, WallSafeTile{9*i + 8, safeType})
		}
	}
	dncSafeTiles.sort()
	return
}

// 설명
// 설명
// 설명
func CalcDNCSafeTilesWithDiscards(leftTiles34 []int, safeTiles34 []bool) (dncSafeTiles WallSafeTileList) {
	nc := func(idx int) bool {
		return leftTiles34[idx] == 0
	}

	const safeType = WallSafeTypeDoubleNoChance

	dncSafeTiles = CalcDNCSafeTiles(leftTiles34)

	// 설명
	// 설명
	for i := 0; i < 3; i++ {
		for j := 1; j < 3; j++ {
			idx := 9*i + j
			if nc(idx-1) && safeTiles34[idx+3] {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			if nc(idx-1) && safeTiles34[idx+3] || nc(idx+1) && safeTiles34[idx-3] {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 6; j < 8; j++ {
			idx := 9*i + j
			if nc(idx+1) && safeTiles34[idx-3] {
				dncSafeTiles = append(dncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
	}

	dncSafeTiles.sort()
	return
}

// 설명
func CalcNCSafeTiles(leftTiles34 []int) (ncSafeTiles WallSafeTileList) {
	nc := func(idx int) bool {
		return leftTiles34[idx] == 0
	}
	or := func(idx ...int) bool {
		for _, i := range idx {
			if nc(i) {
				return true
			}
		}
		return false
	}

	const safeType = WallSafeTypeNoChance
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			if or(idx+1, idx+2) {
				ncSafeTiles = append(ncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			if or(idx-2, idx-1) && or(idx+1, idx+2) {
				ncSafeTiles = append(ncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			if or(idx-2, idx-1) {
				ncSafeTiles = append(ncSafeTiles, WallSafeTile{idx, safeType})
			}
		}
	}
	ncSafeTiles.sort()
	return
}

// 설명
func CalcOCSafeTiles(leftTiles34 []int) (ocSafeTiles WallSafeTileList) {
	oc := func(idx int) bool {
		return leftTiles34[idx] == 1
	}
	or := func(idx ...int) bool {
		for _, i := range idx {
			if oc(i) {
				return true
			}
		}
		return false
	}
	and := func(idx ...int) bool {
		for _, i := range idx {
			if !oc(i) {
				return false
			}
		}
		return true
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			idx := 9*i + j
			if and(idx+1, idx+2) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeDoubleOneChance})
			} else if or(idx+1, idx+2) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeOneChance})
			}
		}
		for j := 3; j < 6; j++ {
			idx := 9*i + j
			if or(idx-2, idx-1) && or(idx+1, idx+2) {
				if and(idx-2, idx-1, idx+1, idx+2) {
					// 설명
					ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeDoubleOneChance})
				} else if and(idx-2, idx-1) || and(idx+1, idx+2) {
					// 설명
					ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeMixedOneChance})
				} else {
					ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeOneChance})
				}
			}
		}
		for j := 6; j < 9; j++ {
			idx := 9*i + j
			if and(idx-2, idx-1) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeDoubleOneChance})
			} else if or(idx-2, idx-1) {
				ocSafeTiles = append(ocSafeTiles, WallSafeTile{idx, WallSafeTypeOneChance})
			}
		}
	}
	ocSafeTiles.sort()
	return
}

func CalcWallTiles(leftTiles34 []int) (safeTiles WallSafeTileList) {
	safeTiles = append(safeTiles, CalcNCSafeTiles(leftTiles34)...)
	safeTiles = append(safeTiles, CalcOCSafeTiles(leftTiles34)...)
	safeTiles.sort()
	return
}
