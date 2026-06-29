package util

import "fmt"

func _calcKey(tiles34 []int) (key int) {
	bitPos := -1

	// 설명
	idx := -1
	for i := 0; i < 3; i++ {
		prevInHand := false // 설명
		for j := 0; j < 9; j++ {
			idx++
			if c := tiles34[idx]; c > 0 {
				prevInHand = true
				bitPos++
				switch c {
				case 2:
					key |= 0x3 << uint(bitPos)
					bitPos += 2
				case 3:
					key |= 0xF << uint(bitPos)
					bitPos += 4
				case 4:
					key |= 0x3F << uint(bitPos)
					bitPos += 6
				}
			} else {
				if prevInHand {
					prevInHand = false
					key |= 0x1 << uint(bitPos)
					bitPos++
				}
			}
		}
		if prevInHand {
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	// 설명
	for i := 27; i < 34; i++ {
		if c := tiles34[i]; c > 0 {
			bitPos++
			switch c {
			case 2:
				key |= 0x3 << uint(bitPos)
				bitPos += 2
			case 3:
				key |= 0xF << uint(bitPos)
				bitPos += 4
			case 4:
				key |= 0x3F << uint(bitPos)
				bitPos += 6
			}
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	return
}

// 설명
func IsAgari(tiles34 []int) bool {
	key := _calcKey(tiles34)
	_, isAgari := winTable[key]
	return isAgari
}

//

// 설명
type DivideResult struct {
	PairTile          int   // 설명
	KotsuTiles        []int // 설명
	ShuntsuFirstTiles []int // 설명

	// 설명
	// 설명
	// 설명
	IsChiitoi       bool // 설명
	IsChuurenPoutou bool // 설명
	IsIttsuu        bool // 설명
	IsRyanpeikou    bool // 설명
	IsIipeikou      bool // 설명
}

// 설명
func (d *DivideResult) String() string {
	if d.IsChiitoi {
		return "[치또이츠]"
	}

	output := ""

	humanTilesList := []string{TilesToStr([]int{d.PairTile, d.PairTile})}
	for _, kotsuTile := range d.KotsuTiles {
		humanTilesList = append(humanTilesList, TilesToStr([]int{kotsuTile, kotsuTile, kotsuTile}))
	}
	for _, shuntsuFirstTile := range d.ShuntsuFirstTiles {
		humanTilesList = append(humanTilesList, TilesToStr([]int{shuntsuFirstTile, shuntsuFirstTile + 1, shuntsuFirstTile + 2}))
	}
	output += fmt.Sprint(humanTilesList)

	if d.IsChuurenPoutou {
		output += "[구련보등]"
	}
	if d.IsIttsuu {
		output += "[일기통관]"
	}
	if d.IsRyanpeikou {
		output += "[량페코]"
	}
	if d.IsIipeikou {
		output += "[이페코]"
	}

	return output
}

// 설명
// http://hp.vector.co.jp/authors/VA046927/mjscore/mjalgorism.html
// http://hp.vector.co.jp/authors/VA046927/mjscore/AgariIndex.java
func DivideTiles34(tiles34 []int) (divideResults []*DivideResult) {
	tiles14 := make([]int, 14)
	tiles14TailIndex := 0

	key := 0
	bitPos := -1

	// 설명
	idx := -1
	for i := 0; i < 3; i++ {
		prevInHand := false // 설명
		for j := 0; j < 9; j++ {
			idx++
			if c := tiles34[idx]; c > 0 {
				tiles14[tiles14TailIndex] = idx
				tiles14TailIndex++

				prevInHand = true
				bitPos++
				switch c {
				case 2:
					key |= 0x3 << uint(bitPos)
					bitPos += 2
				case 3:
					key |= 0xF << uint(bitPos)
					bitPos += 4
				case 4:
					key |= 0x3F << uint(bitPos)
					bitPos += 6
				}
			} else {
				if prevInHand {
					prevInHand = false
					key |= 0x1 << uint(bitPos)
					bitPos++
				}
			}
		}
		if prevInHand {
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	// 설명
	for i := 27; i < 34; i++ {
		if c := tiles34[i]; c > 0 {
			tiles14[tiles14TailIndex] = i
			tiles14TailIndex++

			bitPos++
			switch c {
			case 2:
				key |= 0x3 << uint(bitPos)
				bitPos += 2
			case 3:
				key |= 0xF << uint(bitPos)
				bitPos += 4
			case 4:
				key |= 0x3F << uint(bitPos)
				bitPos += 6
			}
			key |= 0x1 << uint(bitPos)
			bitPos++
		}
	}

	results, ok := winTable[key]
	if !ok {
		return
	}

	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	for _, r := range results {
		// 설명
		pairTile := tiles14[(r>>6)&0xF]

		// 설명
		numKotsu := r & 0x7
		kotsuTiles := make([]int, numKotsu)
		for i := range kotsuTiles {
			kotsuTiles[i] = tiles14[(r>>uint(10+i*4))&0xF]
		}

		// 설명
		numShuntsu := (r >> 3) & 0x7
		shuntsuFirstTiles := make([]int, numShuntsu)
		for i := range shuntsuFirstTiles {
			shuntsuFirstTiles[i] = tiles14[(r>>uint(10+(numKotsu+i)*4))&0xF]
		}

		divideResults = append(divideResults, &DivideResult{
			PairTile:          pairTile,
			KotsuTiles:        kotsuTiles,
			ShuntsuFirstTiles: shuntsuFirstTiles,
			IsChiitoi:         r&(1<<26) != 0,
			IsChuurenPoutou:   r&(1<<27) != 0,
			IsIttsuu:          r&(1<<28) != 0,
			IsRyanpeikou:      r&(1<<29) != 0,
			IsIipeikou:        r&(1<<30) != 0,
		})
	}

	return
}
