package main

var debugMode = false

type gameMode int

const (
	// TODO: 설명 보완
	gameModeMatch       gameMode = iota // 설명
	gameModeRecord                      // 설명
	gameModeRecordCache                 // 설명
	gameModeLive                        // 설명
)

const (
	dataSourceTypeTenhou = iota
	dataSourceTypeMajsoul
)

const (
	meldTypeChi    = iota // 설명
	meldTypePon           // 설명
	meldTypeAnkan         // 설명
	meldTypeMinkan        // 설명
	meldTypeKakan         // 설명
)

// 설명
func normalDiscardTiles(discardTiles []int) []int {
	newD := make([]int, len(discardTiles))
	copy(newD, discardTiles)
	for i, discardTile := range newD {
		if discardTile < 0 {
			newD[i] = ^discardTile
		}
	}
	return newD
}
