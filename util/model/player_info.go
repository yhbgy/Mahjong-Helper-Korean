package model

import "fmt"

type PlayerInfo struct {
	HandTiles34 []int  // 설명
	Melds       []Meld // 설명
	DoraTiles   []int  // 설명
	NumRedFives []int  // 설명

	IsTsumo       bool // 설명
	WinTile       int  // 설명
	RoundWindTile int  // 설명
	SelfWindTile  int  // 설명
	IsParent      bool // 설명
	IsDaburii     bool // 설명
	IsRiichi      bool // 설명

	DiscardTiles []int // 설명
	LeftTiles34  []int // 설명

	LeftDrawTilesCount int // 설명

	// 설명
	// 설명

	NukiDoraNum int // 설명
}

func NewSimplePlayerInfo(tiles34 []int, melds []Meld) *PlayerInfo {
	leftTiles34 := InitLeftTiles34WithTiles34(tiles34)
	for _, meld := range melds {
		for _, tile := range meld.Tiles {
			leftTiles34[tile]--
			if leftTiles34[tile] < 0 {
				panic(fmt.Sprint("후로 데이터가 올바르지 않습니다", melds))
			}
		}
	}
	return &PlayerInfo{
		HandTiles34:   tiles34,
		Melds:         melds,
		NumRedFives:   make([]int, 3),
		RoundWindTile: 27,
		SelfWindTile:  27,
		LeftTiles34:   leftTiles34,
	}
}

// 설명
func (pi *PlayerInfo) CountDora() (count int) {
	for _, doraTile := range pi.DoraTiles {
		count += pi.HandTiles34[doraTile]
		for _, m := range pi.Melds {
			for _, tile := range m.Tiles {
				if tile == doraTile {
					count++
				}
			}
		}
	}
	// 설명
	for _, num := range pi.NumRedFives {
		count += num
	}
	// 설명
	if pi.NukiDoraNum > 0 {
		count += pi.NukiDoraNum
		// 설명
		for _, doraTile := range pi.DoraTiles {
			if doraTile == 30 {
				count += pi.NukiDoraNum
			}
		}
	}
	return
}

// 설명
// TODO: 설명 보완
//func (pi *PlayerInfo) CountUraDora() (count float64) {
//	if !pi.IsRiichi || pi.IsNaki() {
//		return 0
//	}
//	uraDoraTileLeft := make([]int, len(pi.LeftTiles34))
//	for tile, left := range pi.LeftTiles34 {
//		uraDoraTileLeft[DoraTile(tile)] = left
//	}
//	sum := 0
//	weight := 0
//	for tile, c := range pi.HandTiles34 {
//		w := uraDoraTileLeft[tile]
//		sum += w * c
//		weight += w
//	}
//	for _, meld := range pi.Melds {
//		for tile, c := range meld.Tiles {
//			w := uraDoraTileLeft[tile]
//			sum += w * c
//			weight += w
//		}
//	}
// 설명
//	return float64(len(pi.DoraTiles)*sum) / float64(weight)
//}

// 설명
// 설명
func (pi *PlayerInfo) IsNaki() bool {
	for _, meld := range pi.Melds {
		if meld.MeldType != MeldTypeAnkan {
			return true
		}
	}
	return false
}

// 설명
// 설명
// TODO: 설명 보완
func (pi *PlayerInfo) IsFuriten(waits map[int]int) bool {
	for _, discardTile := range pi.DiscardTiles {
		if _, ok := waits[discardTile]; ok {
			return true
		}
	}
	return false
}

/************* 아래 인터페이스는 임시 내부 호출용 ************/

func (pi *PlayerInfo) FillLeftTiles34() {
	pi.LeftTiles34 = InitLeftTiles34WithTiles34(pi.HandTiles34)
}

// 설명
func (pi *PlayerInfo) IsOnlyRedFive(tile int) bool {
	return tile < 27 && tile%9 == 4 && pi.HandTiles34[tile] > 0 && pi.HandTiles34[tile] == pi.NumRedFives[tile/9]
}

func (pi *PlayerInfo) DiscardTile(tile int, isRedFive bool) {
	// 설명
	pi.HandTiles34[tile]--
	if isRedFive {
		pi.NumRedFives[tile/9]--
	}
	pi.DiscardTiles = append(pi.DiscardTiles, tile)
}

func (pi *PlayerInfo) UndoDiscardTile(tile int, isRedFive bool) {
	// 설명
	pi.DiscardTiles = pi.DiscardTiles[:len(pi.DiscardTiles)-1]
	pi.HandTiles34[tile]++
	if isRedFive {
		pi.NumRedFives[tile/9]++
	}
}

//func (pi *PlayerInfo) DrawTile(tile int) {
// 설명
//}
//
//func (pi *PlayerInfo) UndoDrawTile(tile int) {
// 설명
//}

func (pi *PlayerInfo) AddMeld(meld Meld) {
	// 설명
	// 설명
	for _, tile := range meld.SelfTiles {
		pi.HandTiles34[tile]--
	}
	pi.Melds = append(pi.Melds, meld)
	if meld.RedFiveFromOthers {
		tile := meld.Tiles[0]
		pi.NumRedFives[tile/9]++
	}
}

func (pi *PlayerInfo) UndoAddMeld() {
	// 설명
	latestMeld := pi.Melds[len(pi.Melds)-1]
	for _, tile := range latestMeld.SelfTiles {
		pi.HandTiles34[tile]++
	}
	pi.Melds = pi.Melds[:len(pi.Melds)-1]
	if latestMeld.RedFiveFromOthers {
		tile := latestMeld.Tiles[0]
		pi.NumRedFives[tile/9]--
	}
}
