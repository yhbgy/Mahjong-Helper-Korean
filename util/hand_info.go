package util

import "github.com/EndlessCheng/mahjong-helper/util/model"

type _handInfo struct {
	*model.PlayerInfo
	divideResult *DivideResult // 설명

	// 설명
	allShuntsuFirstTiles []int
	allKotsuTiles        []int
}

// 설명
func (hi *_handInfo) getAllShuntsuFirstTiles() []int {
	shuntsuFirstTiles := append([]int{}, hi.divideResult.ShuntsuFirstTiles...)
	for _, meld := range hi.Melds {
		if meld.MeldType == model.MeldTypeChi {
			shuntsuFirstTiles = append(shuntsuFirstTiles, meld.Tiles[0])
		}
	}
	return shuntsuFirstTiles
}

// 설명
func (hi *_handInfo) getAllKotsuTiles() []int {
	kotsuTiles := append([]int{}, hi.divideResult.KotsuTiles...)
	for _, meld := range hi.Melds {
		if meld.MeldType != model.MeldTypeChi {
			kotsuTiles = append(kotsuTiles, meld.Tiles[0])
		}
	}
	return kotsuTiles
}

// 설명
func (hi *_handInfo) containHonor() bool {
	// 설명
	if hi.divideResult.IsChiitoi {
		for _, c := range hi.HandTiles34[27:] {
			if c > 0 {
				return true
			}
		}
		return false
	}

	if hi.divideResult.PairTile >= 27 {
		return true
	}
	for _, tile := range hi.allKotsuTiles {
		if tile >= 27 {
			return true
		}
	}
	return false
}

// 설명
func (hi *_handInfo) isYakuTile(tile int) bool {
	return tile >= 31 || tile == hi.RoundWindTile || tile == hi.SelfWindTile
}

// 설명
func (hi *_handInfo) isDoubleWindTile(tile int) bool {
	return hi.RoundWindTile == hi.SelfWindTile && tile == hi.RoundWindTile
}

// 설명
func (hi *_handInfo) numAnkan() (cnt int) {
	for _, meld := range hi.Melds {
		if meld.MeldType == model.MeldTypeAnkan {
			cnt++
		}
	}
	return
}

// 설명
func (hi *_handInfo) numKantsu() (cnt int) {
	for _, meld := range hi.Melds {
		if meld.IsKan() {
			cnt++
		}
	}
	return
}

// 설명
// 설명
func (hi *_handInfo) numAnkou() (cnt int, isMinkou bool) {
	num := len(hi.divideResult.KotsuTiles) + hi.numAnkan()
	// 설명
	if hi.IsTsumo {
		return num, false
	}
	// 설명
	if hi.WinTile == hi.divideResult.PairTile {
		return num, false
	}
	// 설명
	for _, tile := range hi.divideResult.ShuntsuFirstTiles {
		if hi.WinTile >= tile && hi.WinTile <= tile+2 {
			return num, false
		}
	}
	// 설명
	return num - 1, true
}

// 설명
func (hi *_handInfo) _countSpecialKotsu(specialTilesL, specialTilesLR int) (cnt int) {
	for _, tile := range hi.allKotsuTiles {
		if tile >= specialTilesL && tile <= specialTilesLR {
			cnt++
		}
	}
	return
}
