package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalcPointRon(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(3600, CalcPointRon(1, 110, 0, false)) // saki
	assert.Equal(5200, CalcPointRon(3, 40, 0, false))
	assert.Equal(6400, CalcPointRon(3, 50, 0, false))
	assert.Equal(7700, CalcPointRon(3, 60, 0, false))
	assert.Equal(8000, CalcPointRon(3, 70, 0, false))

	assert.Equal(7700, CalcPointRon(3, 40, 0, true))
	assert.Equal(11600, CalcPointRon(4, 30, 0, true))
	assert.Equal(12000, CalcPointRon(4, 40, 0, true))

	assert.Equal(32000, CalcPointRon(0, 0, 1, false))
	assert.Equal(64000, CalcPointRon(0, 0, 2, false))
	assert.Equal(96000, CalcPointRon(0, 0, 3, false))
	assert.Equal(128000, CalcPointRon(0, 0, 4, false))
	assert.Equal(160000, CalcPointRon(0, 0, 5, false))
	assert.Equal(192000, CalcPointRon(0, 0, 6, false))

	assert.Equal(48000, CalcPointRon(0, 0, 1, true))
	assert.Equal(96000, CalcPointRon(0, 0, 2, true))
	assert.Equal(144000, CalcPointRon(0, 0, 3, true))
	assert.Equal(192000, CalcPointRon(0, 0, 4, true))
	assert.Equal(240000, CalcPointRon(0, 0, 5, true))
	assert.Equal(288000, CalcPointRon(0, 0, 6, true))
}

func TestCalcPointTsumoSum(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(3600, CalcPointTsumoSum(1, 110, 0, false)) // saki
	assert.Equal(5200, CalcPointTsumoSum(3, 40, 0, false))
	assert.Equal(6400, CalcPointTsumoSum(3, 50, 0, false))
	assert.Equal(7900, CalcPointTsumoSum(3, 60, 0, false))
	assert.Equal(8000, CalcPointTsumoSum(3, 70, 0, false))

	assert.Equal(7800, CalcPointTsumoSum(3, 40, 0, true))
	assert.Equal(12000, CalcPointTsumoSum(4, 40, 0, true))
}

func TestCalcRonPointWithHands(t *testing.T) {
	assert := assert.New(t)

	// 설명
	newPIWithWinTile := func(humanTiles string, winHumanTile string) *model.PlayerInfo {
		return &model.PlayerInfo{
			HandTiles34:   MustStrToTiles34(humanTiles),
			WinTile:       MustStrToTile34(winHumanTile),
			RoundWindTile: MustStrToTile34("2z"),
			SelfWindTile:  MustStrToTile34("2z"),
		}
	}
	assert.Equal(12000, CalcPoint(newPIWithWinTile("11m 112233445566z", "1m")).Point)    // [치또이츠 혼노두 혼일색]
	assert.Equal(7700, CalcPoint(newPIWithWinTile("345m 345s 334455p 44z", "3m")).Point) // [핑후 이페코 삼색동순]
	assert.Equal(2600, CalcPoint(newPIWithWinTile("333m 333s 333345p 11z", "3m")).Point) // [삼색동각]
	assert.Equal(8000, CalcPoint(newPIWithWinTile("22334455m 234s 234p", "3m")).Point)   // 설명
	assert.Equal(12000, CalcPoint(newPIWithWinTile("234m 333p 55666777z", "3m")).Point)  // [산안커 역패 역패 소삼원]
	assert.Equal(12000, CalcPoint(newPIWithWinTile("123445566789m 11z", "3m")).Point)    // 설명
	assert.Equal(3200, CalcPoint(newPIWithWinTile("123m 123999s 11155z", "3m")).Point)   // [찬타]
	assert.Equal(5200, CalcPoint(newPIWithWinTile("334455m 667788s 77z", "3m")).Point)   // [량페코]
	assert.Equal(7700, CalcPoint(newPIWithWinTile("334455m 667788s 44z", "3m")).Point)   // [핑후 량페코]
	assert.Equal(5200, CalcPoint(newPIWithWinTile("123m 123999s 11789p", "3m")).Point)   // [준찬타]
	assert.Equal(2600, CalcPoint(newPIWithWinTile("345m 12355789s 222z", "3m")).Point)   // [역패 역패]
	// 설명
	assert.Equal(32000, CalcPoint(newPIWithWinTile("11122345678999m", "3m")).Point)
	assert.Equal(64000, CalcPoint(newPIWithWinTile("11122345678999m", "2m")).Point)
	assert.Equal(160000, CalcPoint(newPIWithWinTile("11122233344455z", "5z")).Point)

	// 설명
	newPIWithRiichi := func(humanTiles string, winHumanTile string) *model.PlayerInfo {
		return &model.PlayerInfo{
			HandTiles34:   MustStrToTiles34(humanTiles),
			WinTile:       MustStrToTile34(winHumanTile),
			RoundWindTile: MustStrToTile34("2z"),
			SelfWindTile:  MustStrToTile34("3z"),
			IsRiichi:      true,
		}
	}
	assert.Equal(1300, CalcPoint(newPIWithRiichi("345m 222789p 333s 66z", "3m")).Point) // [리치]

	// 설명
	ronPoints := []int{}
	for doraCount := 0; doraCount < 13; doraCount++ {
		ronPoint := CalcPoint(&model.PlayerInfo{
			NumRedFives:   []int{doraCount, 0, 0}, // 설명
			HandTiles34:   MustStrToTiles34("345m 222789p 333s 66z"),
			WinTile:       MustStrToTile34("3m"),
			RoundWindTile: MustStrToTile34("2z"),
			SelfWindTile:  MustStrToTile34("3z"),
			IsRiichi:      true,
		}).Point
		ronPoints = append(ronPoints, ronPoint)
	}
	assert.Equal(ronPoints, []int{1300, 2600, 5200, 8000, 8000, 12000, 12000, 16000, 16000, 16000, 24000, 24000, 32000})

	// 설명
	ronPoints = []int{}
	for doraCount := 0; doraCount < 13; doraCount++ {
		ronPoint := CalcPoint(&model.PlayerInfo{
			NumRedFives:   []int{doraCount, 0, 0}, // 설명
			HandTiles34:   MustStrToTiles34("345m 222789p 333s 66z"),
			WinTile:       MustStrToTile34("3m"),
			RoundWindTile: MustStrToTile34("2z"),
			SelfWindTile:  MustStrToTile34("3z"),
			IsParent:      true,
			IsRiichi:      true,
		}).Point
		ronPoints = append(ronPoints, ronPoint)
	}
	assert.Equal(ronPoints, []int{2000, 3900, 7700, 12000, 12000, 18000, 18000, 24000, 24000, 24000, 36000, 36000, 48000})

	const eps = 1

	first := func(a float64, _ ...interface{}) float64 { return a }

	// 설명
	newPIWithWaits := func(humanTiles string) (model.PlayerInfo, Waits) {
		tiles34 := MustStrToTiles34(humanTiles)
		_, waits := CalculateShantenAndWaits13(tiles34, nil)
		return model.PlayerInfo{
			HandTiles34:   tiles34,
			RoundWindTile: MustStrToTile34("2z"),
			SelfWindTile:  MustStrToTile34("2z"),
		}, waits
	}
	assert.InDelta(3700, first(CalcAvgRiichiPoint(newPIWithWaits("34m 123567p 12355s"))), eps)   // 리치핑후
	assert.InDelta(7500, first(CalcAvgRiichiPoint(newPIWithWaits("13m 123567p 12355s"))), eps)   // 리치삼색동순
	assert.InDelta(4291, first(CalcAvgRiichiPoint(newPIWithWaits("12366m 234p 345s 55z"))), eps) // 설명

	// 설명
	newFuritenPIWithWaits := func(humanTiles string, humanDiscardTiles string) (model.PlayerInfo, Waits) {
		tiles34 := MustStrToTiles34(humanTiles)
		_, waits := CalculateShantenAndWaits13(tiles34, nil)
		return model.PlayerInfo{
			HandTiles34:   tiles34,
			RoundWindTile: MustStrToTile34("2z"),
			SelfWindTile:  MustStrToTile34("2z"),
			DiscardTiles:  MustStrToTiles(humanDiscardTiles),
		}, waits
	}
	assert.InDelta(4070, first(CalcAvgRiichiPoint(newFuritenPIWithWaits("45678m 123p 56799s", "9m"))), eps) // 리치핑후(쭔모)
}

func BenchmarkCalcAvgRiichiPoint(b *testing.B) {
	humanTiles := "11123678m 11122z" // 설명
	tiles34 := MustStrToTiles34(humanTiles)
	playerInfo := model.PlayerInfo{
		HandTiles34:   tiles34,
		Melds:         nil,
		IsTsumo:       true,
		RoundWindTile: MustStrToTile34("2z"),
		SelfWindTile:  MustStrToTile34("2z"),
		LeftTiles34:   InitLeftTiles34WithTiles34(tiles34),
	}
	_, waits := CalculateShantenAndWaits13(playerInfo.HandTiles34, playerInfo.LeftTiles34)
	for i := 0; i < b.N; i++ {
		// 7277 ns/op
		CalcAvgRiichiPoint(playerInfo, waits)
	}
}
