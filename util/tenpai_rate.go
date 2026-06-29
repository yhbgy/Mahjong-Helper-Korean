package util

import "github.com/EndlessCheng/mahjong-helper/util/model"

// 설명
// TODO: 설명 보완
func CalcTenpaiRate(melds []*model.Meld, discardTiles []int, meldDiscardsAt []int) float64 {
	isNaki := false
	for _, meld := range melds {
		if meld.MeldType != model.MeldTypeAnkan {
			isNaki = true
		}
	}

	if !isNaki {
		// 설명
		turn := len(discardTiles)
		return float64(turn)
	}

	if len(melds) == 4 {
		return 100
	}

	_tenpaiRate := tenpaiRate[len(melds)]

	turn := MinInt(len(discardTiles), len(_tenpaiRate)-1)
	_tenpaiRateWithTurn := _tenpaiRate[turn]

	// 설명
	// 설명
	countTedashi := 0
	if len(meldDiscardsAt) > 0 {
		latestDiscardAt := meldDiscardsAt[len(meldDiscardsAt)-1]
		if len(discardTiles) > latestDiscardAt {
			for _, disTile := range discardTiles[latestDiscardAt+1:] {
				if disTile >= 0 {
					countTedashi++
				}
			}
		}
	}
	countTedashi = MinInt(countTedashi, len(_tenpaiRateWithTurn)-1)

	return _tenpaiRateWithTurn[countTedashi]
}
