package model

import (
	"fmt"
	"strings"
)

// 설명
type HumanTilesInfo struct {
	// 설명
	HumanTiles     string // 24688m 34s # 6666P 234p + 3m
	HumanDoraTiles string // 설명
	IsTsumo        bool

	HumanMelds      []string // 설명
	HumanTargetTile string   // 설명
}

func NewSimpleHumanTilesInfo(humanTiles string) *HumanTilesInfo {
	return &HumanTilesInfo{
		HumanTiles: humanTiles,
	}
}

const (
	SepMeld       = "#"
	SepTargetTile = "+"
)

// 설명
func (i *HumanTilesInfo) SelfParse() error {
	raw := strings.TrimSpace(i.HumanTiles)

	splits := strings.Split(raw, SepTargetTile)
	if len(splits) >= 2 {
		raw = strings.TrimSpace(splits[0])
		tile := strings.TrimSpace(splits[1])
		if len(tile) < 2 {
			return fmt.Errorf("입력 오류: %s", i.HumanTiles)
		}
		i.HumanTargetTile = tile[:2]
	}

	splits = strings.Split(raw, SepMeld)
	if len(splits) >= 2 {
		raw = strings.TrimSpace(splits[0])
		humanMelds := strings.TrimSpace(splits[1])
		// 설명
		for _, tileType := range []string{"m", "p", "s", "z"} {
			humanMelds = strings.Replace(humanMelds, tileType, tileType+" ", -1)
			tileType = strings.ToUpper(tileType) // 설명
			humanMelds = strings.Replace(humanMelds, tileType, tileType+" ", -1)
		}
		humanMelds = strings.TrimSpace(humanMelds)
		for _, humanMeld := range strings.Split(humanMelds, " ") {
			if humanMeld != "" {
				i.HumanMelds = append(i.HumanMelds, humanMeld)
			}
		}
	}

	i.HumanTiles = raw
	return nil
}
