package util

import (
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/stretchr/testify/assert"
	"sort"
	"strings"
	"testing"
)

func calcStrYaku(humanTiles string, humanWinTile string, isTsumo bool, melds ...model.Meld) string {
	output := ""
	pi := &model.PlayerInfo{
		HandTiles34:   MustStrToTiles34(humanTiles),
		Melds:         melds,
		IsTsumo:       isTsumo,
		WinTile:       MustStrToTile34(humanWinTile),
		RoundWindTile: 27,
		SelfWindTile:  27,
	}
	isNaki := pi.IsNaki()
	for _, result := range DivideTiles34(pi.HandTiles34) {
		yakuTypes := findYakuTypes(&_handInfo{
			PlayerInfo:   pi,
			divideResult: result,
		}, isNaki)
		sort.Ints(yakuTypes)
		output += YakuTypesToStr(yakuTypes) + " "
	}
	return strings.TrimSpace(output)
}

func Test_findYakuTypes(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("[치또이츠 혼노두 혼일색]", calcStrYaku("99s 112233445566z", "9s", false))
	assert.Equal("[치또이츠 혼일색]", calcStrYaku("22m 112233445566z", "2m", false))
	assert.Equal("[핑후 이페코 삼색동순]", calcStrYaku("345m 345s 334455p 44z", "3m", false))
	assert.Equal("[삼색동각]", calcStrYaku("333m 333s 333345p 11z", "3m", false))
	assert.Equal("[핑후 이페코 탕야오] [이페코 삼색동순 탕야오]", calcStrYaku("22334455m 234s 234p", "3m", false))
	assert.Equal("[산안커 역패 역패 소삼원]", calcStrYaku("234m 333p 55666777z", "3m", false))
	assert.Equal("[이페코 일기통관 혼일색]", calcStrYaku("123445566789m 11z", "3m", false))
	assert.Equal("[또이또이 산안커 혼일색] [이페코 혼일색]", calcStrYaku("111222333444m 11z", "3m", false))
	assert.Equal("[스안커] [쭔모 이페코 혼일색]", calcStrYaku("111222333444m 11z", "3m", true))
	assert.Equal("[역패 역패 찬타]", calcStrYaku("123m 123999s 11155z", "3m", false))
	assert.Equal("[량페코]", calcStrYaku("334455m 667788s 77z", "3m", false))
	assert.Equal("[핑후 량페코]", calcStrYaku("334455m 667788s 44z", "3m", false))
	assert.Equal("[준찬타]", calcStrYaku("123m 123999s 11789p", "3m", false))

	// 설명
	assert.Equal("[구련보등]", calcStrYaku("11122345678999m", "3m", false))
	assert.Equal("[순정구련보등]", calcStrYaku("11123345678999m", "3m", false))
	assert.Equal("[녹일색]", calcStrYaku("22334466688s 666z", "6z", false))
	assert.Equal("[스안커]", calcStrYaku("111999m 111p 11122z", "1z", true))
	assert.Equal("[소사희 자일색]", calcStrYaku("11122233344555z", "1z", false))
	assert.Equal("[자일색]", calcStrYaku("11223344556677z", "1z", false))
	assert.Equal("[스안커 단기 대사희 자일색]", calcStrYaku("11122233344455z", "5z", false))
	assert.Equal("[대삼원]", calcStrYaku("12333m 555666777z", "1m", false))
	assert.Equal("[청노두]", calcStrYaku("111999m 111999s 11p", "1m", false))

	// 설명
	assert.Equal("[삼색동각]", calcStrYaku("333m 333p 333567s 11z", "3m", false))
	assert.Equal("[산안커 삼색동각]", calcStrYaku("333345m 333p 333s 11z", "3m", false))

	// 설명
	assert.Equal("[일기통관 역패 역패 혼일색]", calcStrYaku("123p 11177z", "3p", false,
		model.Meld{MeldType: model.MeldTypeChi, Tiles: MustStrToTiles("456p")},
		model.Meld{MeldType: model.MeldTypeChi, Tiles: MustStrToTiles("789p")},
	))
	assert.Equal("[또이또이 역패 역패 혼노두]", calcStrYaku("111p 11177z", "1p", false,
		model.Meld{MeldType: model.MeldTypePon, Tiles: MustStrToTiles("999p")},
		model.Meld{MeldType: model.MeldTypePon, Tiles: MustStrToTiles("111s")},
	))
	assert.Equal("[또이또이 산깡쯔 혼일색]", calcStrYaku("333m 77z", "3m", false,
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("4444z")},
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("2222z")},
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("3333z")},
	))
	assert.Equal("[또이또이 산깡쯔 탕야오]", calcStrYaku("333m 77s", "3m", false,
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("4444s")},
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("2222s")},
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("3333s")},
	))
	assert.Equal("[스깡쯔]", calcStrYaku("77z", "7z", false,
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("1111z")},
		model.Meld{MeldType: model.MeldTypeAnkan, Tiles: MustStrToTiles("1111p")},
		model.Meld{MeldType: model.MeldTypeKakan, Tiles: MustStrToTiles("2222z")},
		model.Meld{MeldType: model.MeldTypeMinkan, Tiles: MustStrToTiles("3333z")},
	))
	assert.Equal("[스안커 단기 대사희 자일색 스깡쯔]", calcStrYaku("77z", "7z", false,
		model.Meld{MeldType: model.MeldTypeAnkan, Tiles: MustStrToTiles("1111z")},
		model.Meld{MeldType: model.MeldTypeAnkan, Tiles: MustStrToTiles("2222z")},
		model.Meld{MeldType: model.MeldTypeAnkan, Tiles: MustStrToTiles("3333z")},
		model.Meld{MeldType: model.MeldTypeAnkan, Tiles: MustStrToTiles("4444z")},
	))

	// 역 없음
	assert.Equal("[역 없음]", calcStrYaku("333m 123s 123p 77z", "3m", false,
		model.Meld{MeldType: model.MeldTypeChi, Tiles: MustStrToTiles("789p")},
	))
}

func Test_findOldYakuTypes(t *testing.T) {
	considerOldYaku = true

	assert := assert.New(t)

	assert.Equal("[산안커 삼연각] [핑후 이페코 일색삼순]", calcStrYaku("222333444p 11m 789s", "9s", false))
	assert.Equal("[역패 찬타 오문제]", calcStrYaku("123p 111m 789s 11777z", "9s", false))
	assert.Equal("[준찬타 십이낙태]", calcStrYaku("99p", "9p", true,
		model.Meld{MeldType: model.MeldTypeChi, Tiles: MustStrToTiles("123m")},
		model.Meld{MeldType: model.MeldTypeChi, Tiles: MustStrToTiles("789p")},
		model.Meld{MeldType: model.MeldTypeChi, Tiles: MustStrToTiles("789s")},
		model.Meld{MeldType: model.MeldTypePon, Tiles: MustStrToTiles("999m")},
	))
	assert.Equal("[대수린] [대수린] [대수린]", calcStrYaku("22334455667788m", "2m", false))
	assert.Equal("[대차륜] [대차륜] [대차륜]", calcStrYaku("22334455667788p", "2p", false))
	assert.Equal("[대죽림] [대죽림] [대죽림]", calcStrYaku("22334455667788s", "2s", false))
	assert.Equal("[자일색 대칠성]", calcStrYaku("11223344556677z", "2z", false))
}

func Benchmark_findYakuTypes(b *testing.B) {
	pi := &model.PlayerInfo{
		HandTiles34:   MustStrToTiles34("345m 345789p 34555s"),
		IsTsumo:       false,
		WinTile:       MustStrToTile34("5s"),
		RoundWindTile: 27,
		SelfWindTile:  27,
	}
	for i := 0; i < b.N; i++ {
		// 1750 ns/op
		for _, result := range DivideTiles34(pi.HandTiles34) {
			findYakuTypes(&_handInfo{
				PlayerInfo:   pi,
				divideResult: result,
			}, false)
		}
	}
}
