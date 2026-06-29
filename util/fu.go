package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

func roundUpFu(fu int) int {
	return ((fu-1)/10 + 1) * 10
}

// 설명
func (hi *_handInfo) calcFu(isNaki bool) int {
	divideResult := hi.divideResult

	// 설명
	if divideResult.IsChiitoi {
		return 25
	}

	const baseFu = 20

	// 설명
	fu := baseFu

	// 설명
	_, ronKotsu := hi.numAnkou()
	for _, tile := range divideResult.KotsuTiles {
		var _fu int
		// 설명
		if ronKotsu && tile == hi.WinTile {
			_fu = 2
		} else {
			_fu = 4
		}
		if isYaochupai(tile) {
			_fu *= 2
		}
		fu += _fu
	}

	// 설명
	for _, meld := range hi.Melds {
		_fu := 0
		switch meld.MeldType {
		case model.MeldTypePon:
			_fu = 2
		case model.MeldTypeMinkan, model.MeldTypeKakan:
			_fu = 8
		case model.MeldTypeAnkan:
			_fu = 16
		}
		if _fu > 0 {
			if isYaochupai(meld.Tiles[0]) {
				_fu *= 2
			}
			fu += _fu
		}
	}

	// 설명
	if hi.isYakuTile(divideResult.PairTile) {
		fu += 2
		if hi.isDoubleWindTile(divideResult.PairTile) {
			fu += 2
		}
	}

	if fu == baseFu {
		// 설명
		if isNaki {
			// 설명
			return 30
		}
		// 설명
		// 설명
		isPinfu := false
		for _, tile := range divideResult.ShuntsuFirstTiles {
			t9 := tile % 9
			if t9 < 6 && tile == hi.WinTile || t9 > 0 && tile+2 == hi.WinTile {
				isPinfu = true
				break
			}
		}
		if hi.IsTsumo {
			if isPinfu {
				// 설명
				return 20
			}
			// 설명
			return 30
		} else {
			// 설명
			if isPinfu {
				// 설명
				return 30
			}
			// 설명
			return 40
		}
	}

	// 설명
	if !isNaki && !hi.IsTsumo {
		fu += 10
	}

	// 설명
	if hi.IsTsumo {
		fu += 2
	}

	// 설명
	// 설명
	if divideResult.PairTile == hi.WinTile {
		fu += 2 // 설명
	} else {
		for _, tile := range divideResult.ShuntsuFirstTiles {
			if tile+1 == hi.WinTile {
				fu += 2 // 설명
				break
			}
			if tile%9 == 0 && tile+2 == hi.WinTile || tile%9 == 6 && tile == hi.WinTile {
				fu += 2 // 설명
				break
			}
		}
	}

	// 설명
	return roundUpFu(fu)
}
