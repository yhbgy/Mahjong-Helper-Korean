package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"os"
)

func interact(humanTilesInfo *model.HumanTilesInfo) error {
	if !debugMode {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("내부 오류:", err)
			}
		}()
	}

	playerInfo, err := analysisHumanTiles(humanTilesInfo)
	if err != nil {
		return err
	}
	tiles34 := playerInfo.HandTiles34
	leftTiles34 := playerInfo.LeftTiles34
	var tile string
	for {
		count := util.CountOfTiles34(tiles34)
		switch count % 3 {
		case 0:
			return fmt.Errorf("인자 오류: %d장", count)
		case 1:
			fmt.Print("> 쯔모 ")
			fmt.Scanf("%s\n", &tile)
			tile, isRedFive, err := util.StrToTile34(tile)
			if err != nil {
				// 설명
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if tiles34[tile] == 4 {
				// 설명
				fmt.Fprintln(os.Stderr, "더 이상 쯔모할 수 없습니다")
				continue
			}
			if isRedFive {
				playerInfo.NumRedFives[tile/9]++
			}
			leftTiles34[tile]--
			tiles34[tile]++
		case 2:
			fmt.Print("> 타 ")
			fmt.Scanf("%s\n", &tile)
			tile, isRedFive, err := util.StrToTile34(tile)
			if err != nil {
				// 설명
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if tiles34[tile] == 0 {
				// 설명
				fmt.Fprintln(os.Stderr, "버린 패가 존재하지 않습니다")
				continue
			}
			if isRedFive {
				playerInfo.NumRedFives[tile/9]--
			}
			tiles34[tile]--
			playerInfo.DiscardTiles = append(playerInfo.DiscardTiles, tile) // 설명
		}
		if err := analysisPlayerWithRisk(playerInfo, nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
