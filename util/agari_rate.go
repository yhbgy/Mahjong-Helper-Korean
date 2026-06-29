package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"sort"
)

// 설명
// 설명
func CalculateAgariRateOfEachTile(waits Waits, playerInfo *model.PlayerInfo) map[int]float64 {
	if playerInfo == nil {
		playerInfo = &model.PlayerInfo{}
	}

	tileAgariRate := map[int]float64{}

	// 설명
	if playerInfo.IsFuriten(waits) {
		for tile, left := range waits {
			rate := 0.0
			for i := 0; i < left; i++ {
				rate = rate + furitenBaseAgariRate - rate*furitenBaseAgariRate/100
			}
			tileAgariRate[tile] = rate
		}
		return tileAgariRate
	}

	// 설명
	if len(waits) == 1 {
		for tile, left := range waits {
			if tile >= 27 {
				rate := honorTileDankiAgariTable[left]
				if InInts(tile, playerInfo.DoraTiles) {
					// 설명
					// 설명
					rate *= honorDoraAgariMulti
				}
				tileAgariRate[tile] = rate
				return tileAgariRate
			}
		}
	}

	// 설명
	tileType27 := calcTileType27(playerInfo.DiscardTiles)
	for tile, left := range waits {
		var rate float64
		if tile < 27 { // 설명
			rate = agariMap[tileType27[tile]][left]
		} else { // 설명
			rate = honorTileNonDankiAgariTable[left]
		}
		if InInts(tile, playerInfo.DoraTiles) {
			// 설명
			// 설명
			if tile >= 27 {
				rate *= honorDoraAgariMulti
			} else {
				rate *= numberDoraAgariMulti
			}
		}
		tileAgariRate[tile] = rate
	}

	return tileAgariRate
}

// 설명
func CalculateAvgAgariRate(waits Waits, playerInfo *model.PlayerInfo) float64 {
	if playerInfo == nil {
		playerInfo = &model.PlayerInfo{}
	}

	// 설명
	if playerInfo.IsFuriten(waits) {
		rate := 0.0
		for i := 0; i < waits.AllCount(); i++ {
			rate = rate + furitenBaseAgariRate - rate*furitenBaseAgariRate/100
		}
		return rate
	}

	tileAgariRate := CalculateAgariRateOfEachTile(waits, playerInfo)
	agariRate := 0.0
	for _, rate := range tileAgariRate {
		agariRate = agariRate + rate - agariRate*rate/100
	}

	// 설명
	// 설명
	waitTiles := []int{}
	for tile, left := range waits {
		if left > 0 {
			if tile >= 27 {
				return agariRate
			}
			waitTiles = append(waitTiles, tile)
		}
	}
	if len(waitTiles) > 1 {
		suitType := waitTiles[0] / 9
		for _, tile := range waitTiles[1:] {
			if tile/9 != suitType {
				return agariRate
			}
		}
		sort.Ints(waitTiles)
		if len(waitTiles) == 2 && waitTiles[0]+3 == waitTiles[1] ||
			len(waitTiles) == 3 && waitTiles[0]+3 == waitTiles[1] && waitTiles[1]+3 == waitTiles[2] {
			agariRate *= ryanmenAgariMulti
		}
	}

	return agariRate
}
