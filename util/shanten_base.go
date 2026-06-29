package util

import (
	"fmt"
)

const (
	shantenStateAgari  = -1
	shantenStateTenpai = 0
)

// 설명
// 설명
func CalculateShantenOfChiitoi(tiles34 []int) int {
	shanten := 6
	numKind := 0
	for _, c := range tiles34 {
		if c == 0 {
			continue
		}
		if c >= 2 {
			shanten--
		}
		numKind++
	}
	shanten += MaxInt(0, 7-numKind)
	return shanten
}

type shanten struct {
	tiles         []int
	numberMelds   int
	numberTatsu   int
	numberPairs   int
	numberJidahai int // 설명
	ankanTiles    int // 설명
	isolatedTiles int // 설명
	minShanten    int
}

func (st *shanten) scanCharacterTiles(countOfTiles int) {
	ankanTiles := 0    // 설명
	isolatedTiles := 0 // 설명

	for i, c := range st.tiles[27:] {
		if c == 0 {
			continue
		}
		switch c {
		case 1:
			isolatedTiles |= 1 << uint(i)
		case 2:
			st.numberPairs++
		case 3:
			st.numberMelds++
		case 4:
			st.numberMelds++
			st.numberJidahai++
			ankanTiles |= 1 << uint(i)
			isolatedTiles |= 1 << uint(i)
		}
	}

	if st.numberJidahai > 0 && countOfTiles%3 == 2 {
		st.numberJidahai--
	}

	if isolatedTiles > 0 {
		st.isolatedTiles |= 1 << 27
		if ankanTiles|isolatedTiles == ankanTiles {
			// 설명
			st.ankanTiles |= 1 << 27
		}
	}
}

// 설명
// 설명
func (st *shanten) calcNormalShanten() int {
	_shanten := 8 - 2*st.numberMelds - st.numberTatsu - st.numberPairs
	numMentsuKouho := st.numberMelds + st.numberTatsu
	if st.numberPairs > 0 {
		numMentsuKouho += st.numberPairs - 1 // 설명
	} else if st.ankanTiles > 0 && st.isolatedTiles > 0 {
		if st.ankanTiles|st.isolatedTiles == st.ankanTiles { // 설명
			// 설명
			_shanten++
		}
	}
	if numMentsuKouho > 4 { // 설명
		_shanten += numMentsuKouho - 4
	}
	if _shanten != shantenStateAgari && _shanten < st.numberJidahai {
		return st.numberJidahai
	}
	return _shanten
}

// 설명
func (st *shanten) increaseSet(k int) {
	st.tiles[k] -= 3
	st.numberMelds++
}

func (st *shanten) decreaseSet(k int) {
	st.tiles[k] += 3
	st.numberMelds--
}

// 설명
func (st *shanten) increasePair(k int) {
	st.tiles[k] -= 2
	st.numberPairs++
}

func (st *shanten) decreasePair(k int) {
	st.tiles[k] += 2
	st.numberPairs--
}

// 설명
func (st *shanten) increaseSyuntsu(k int) {
	st.tiles[k]--
	st.tiles[k+1]--
	st.tiles[k+2]--
	st.numberMelds++
}
func (st *shanten) decreaseSyuntsu(k int) {
	st.tiles[k]++
	st.tiles[k+1]++
	st.tiles[k+2]++
	st.numberMelds--
}

// 설명
func (st *shanten) increaseTatsuFirst(k int) {
	st.tiles[k]--
	st.tiles[k+1]--
	st.numberTatsu++
}
func (st *shanten) decreaseTatsuFirst(k int) {
	st.tiles[k]++
	st.tiles[k+1]++
	st.numberTatsu--
}

// 설명
func (st *shanten) increaseTatsuSecond(k int) {
	st.tiles[k]--
	st.tiles[k+2]--
	st.numberTatsu++
}
func (st *shanten) decreaseTatsuSecond(k int) {
	st.tiles[k]++
	st.tiles[k+2]++
	st.numberTatsu--
}

// 설명
func (st *shanten) increaseIsolatedTile(k int) {
	st.tiles[k]--
	st.isolatedTiles |= 1 << uint(k)
}
func (st *shanten) decreaseIsolatedTile(k int) {
	st.tiles[k]++
	st.isolatedTiles &^= 1 << uint(k)
}

func (st *shanten) run(depth int) {
	if st.minShanten == shantenStateAgari {
		return
	}

	// skip
	for ; depth < 27 && st.tiles[depth] == 0; depth++ {
	}

	if depth >= 27 {
		_shanten := st.calcNormalShanten()
		st.minShanten = MinInt(st.minShanten, _shanten)
		return
	}

	// i := depth % 9
	// 설명
	i := depth
	if i > 8 {
		i -= 9
	}
	if i > 8 {
		i -= 9
	}

	// 설명
	switch st.tiles[depth] {
	case 1:
		// 설명
		// 설명
		if i < 6 && st.tiles[depth+1] == 1 && st.tiles[depth+2] > 0 && st.tiles[depth+3] < 4 {
			// 설명
			// 설명
			st.increaseSyuntsu(depth)
			st.run(depth + 2)
			st.decreaseSyuntsu(depth)
		} else {
			// 설명
			st.increaseIsolatedTile(depth)
			st.run(depth + 1)
			st.decreaseIsolatedTile(depth)

			if i < 7 && st.tiles[depth+2] > 0 {
				if st.tiles[depth+1] != 0 {
					// 설명
					st.increaseSyuntsu(depth)
					st.run(depth + 1)
					st.decreaseSyuntsu(depth)
				}
				// 설명
				st.increaseTatsuSecond(depth)
				st.run(depth + 1)
				st.decreaseTatsuSecond(depth)
			}
			if i < 8 && st.tiles[depth+1] > 0 {
				// 설명
				st.increaseTatsuFirst(depth)
				st.run(depth + 1)
				st.decreaseTatsuFirst(depth)
			}
		}
	case 2:
		// 설명
		st.increasePair(depth)
		st.run(depth + 1)
		st.decreasePair(depth)

		if i < 7 && st.tiles[depth+1] > 0 && st.tiles[depth+2] > 0 {
			// 설명
			st.increaseSyuntsu(depth)
			st.run(depth)
			st.decreaseSyuntsu(depth)
		}
	case 3:
		// 설명
		st.increaseSet(depth)
		st.run(depth + 1)
		st.decreaseSet(depth)

		st.increasePair(depth)
		if i < 7 && st.tiles[depth+1] > 0 && st.tiles[depth+2] > 0 {
			// 설명
			st.increaseSyuntsu(depth)
			st.run(depth + 1)
			st.decreaseSyuntsu(depth)
		} else {
			if i < 7 && st.tiles[depth+2] > 0 {
				// 설명
				st.increaseTatsuSecond(depth)
				st.run(depth + 1)
				st.decreaseTatsuSecond(depth)
			}
			if i < 8 && st.tiles[depth+1] > 0 {
				// 설명
				st.increaseTatsuFirst(depth)
				st.run(depth + 1)
				st.decreaseTatsuFirst(depth)
			}
		}
		st.decreasePair(depth)

		if i < 7 && st.tiles[depth+1] >= 2 && st.tiles[depth+2] >= 2 {
			// 설명
			st.increaseSyuntsu(depth)
			st.increaseSyuntsu(depth)
			st.run(depth)
			st.decreaseSyuntsu(depth)
			st.decreaseSyuntsu(depth)
		}
	case 4:
		st.increaseSet(depth)
		if i < 7 && st.tiles[depth+2] > 0 {
			if st.tiles[depth+1] > 0 {
				// 설명
				st.increaseSyuntsu(depth)
				st.run(depth + 1)
				st.decreaseSyuntsu(depth)
			}
			// 설명
			st.increaseTatsuSecond(depth)
			st.run(depth + 1)
			st.decreaseTatsuSecond(depth)
		}
		if i < 8 && st.tiles[depth+1] > 0 {
			// 설명
			st.increaseTatsuFirst(depth)
			st.run(depth + 1)
			st.decreaseTatsuFirst(depth)
		}
		// 설명
		st.increaseIsolatedTile(depth)
		st.run(depth + 1)
		st.decreaseIsolatedTile(depth)
		st.decreaseSet(depth)

		st.increasePair(depth)
		if i < 7 && st.tiles[depth+2] > 0 {
			if st.tiles[depth+1] > 0 {
				// 설명
				st.increaseSyuntsu(depth)
				st.run(depth)
				st.decreaseSyuntsu(depth)
			}
			// 설명
			st.increaseTatsuSecond(depth)
			st.run(depth + 1)
			st.decreaseTatsuSecond(depth)
		}
		if i < 8 && st.tiles[depth+1] > 0 {
			// 설명
			st.increaseTatsuFirst(depth)
			st.run(depth + 1)
			st.decreaseTatsuFirst(depth)
		}
		st.decreasePair(depth)
	}
}

// 설명
// 설명
func CalculateShantenOfNormal(tiles34 []int, countOfTiles int) int {
	st := shanten{
		numberMelds: (14 - countOfTiles) / 3,
		minShanten:  8, // 설명
		tiles:       tiles34,
	}

	st.scanCharacterTiles(countOfTiles)

	for i, c := range st.tiles[:27] {
		if c == 4 {
			st.ankanTiles |= 1 << uint(i)
		}
	}

	st.run(0)

	return st.minShanten
}

// 설명
// 설명
func CalculateShanten(tiles34 []int) int {
	countOfTiles := CountOfTiles34(tiles34) // 설명
	if countOfTiles > 14 {
		panic(fmt.Sprintln("[CalculateShanten] 인자 오류 >14", tiles34, countOfTiles))
	}
	minShanten := CalculateShantenOfNormal(tiles34, countOfTiles)
	if countOfTiles >= 13 { // 설명
		minShanten = MinInt(minShanten, CalculateShantenOfChiitoi(tiles34))
	}
	return minShanten
}
