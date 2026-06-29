package util

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

type shantenSearchNode13 struct {
	shanten  int
	waits    Waits
	children map[int]*shantenSearchNode14 // 설명
}

func (n *shantenSearchNode13) printWithPrefix(prefix string) string {
	output := ""
	for drawTile, node14 := range n.children {
		output += prefix + fmt.Sprintln("쯔모", Mahjong[drawTile]) + node14.printWithPrefix(prefix+"  ")
	}
	return output
}

func (n *shantenSearchNode13) String() string {
	return n.printWithPrefix("")
}

type shantenSearchNode14 struct {
	shanten  int
	children map[int]*shantenSearchNode13 // 설명
}

func (n *shantenSearchNode14) printWithPrefix(prefix string) string {
	if n == nil || n.shanten == shantenStateAgari {
		return prefix + "end\n"
	}
	output := ""
	for discardTile, node13 := range n.children {
		output += prefix + fmt.Sprintln("버림", Mahjong[discardTile]) + node13.printWithPrefix(prefix+"  ")
	}
	return output
}

func (n *shantenSearchNode14) String() string {
	return n.printWithPrefix("")
}

func _search13(currentShanten int, playerInfo *model.PlayerInfo, stopAtShanten int) *shantenSearchNode13 {
	waits := Waits{}
	children := map[int]*shantenSearchNode14{}
	tiles34 := playerInfo.HandTiles34
	leftTiles34 := playerInfo.LeftTiles34

	isTenpai := currentShanten == shantenStateTenpai

	// 설명
	//if !isTenpai {
	//	needCheck34 := make([]bool, 34)
	//	idx := -1
	//	for i := 0; i < 3; i++ {
	//		for j := 0; j < 9; j++ {
	//			idx++
	//			if tiles34[idx] == 0 {
	//				continue
	//			}
	//			if j == 0 {
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//				needCheck34[idx+2] = true
	//			} else if j == 1 {
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//				needCheck34[idx+2] = true
	//			} else if j < 7 {
	//				needCheck34[idx-2] = true
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//				needCheck34[idx+2] = true
	//			} else if j == 7 {
	//				needCheck34[idx-2] = true
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//				needCheck34[idx+1] = true
	//			} else {
	//				needCheck34[idx-2] = true
	//				needCheck34[idx-1] = true
	//				needCheck34[idx] = true
	//			}
	//		}
	//	}
	//	for i := 27; i < 34; i++ {
	//		if tiles34[i] > 0 {
	//			needCheck34[i] = true
	//		}
	//	}
	//}

	for i := 0; i < 34; i++ {
		//if !needCheck34[i] {
		//	continue
		//}
		if tiles34[i] == 4 {
			continue
		}
		tiles34[i]++
		if isTenpai {
			// 설명
			if IsAgari(tiles34) {
				waits[i] = leftTiles34[i]
				children[i] = nil
			}
		} else {
			if CalculateShanten(tiles34) < currentShanten {
				// 설명
				// 설명
				waits[i] = leftTiles34[i]
				if leftTiles34[i] > 0 && currentShanten-1 >= stopAtShanten {
					leftTiles34[i]--
					children[i] = _search14(currentShanten-1, playerInfo, stopAtShanten)
					leftTiles34[i]++
				} else {
					children[i] = nil
				}
			}
		}
		tiles34[i]--
	}

	return &shantenSearchNode13{
		shanten:  currentShanten,
		waits:    waits,
		children: children,
	}
}

// 설명
func _search14(targetShanten int, playerInfo *model.PlayerInfo, stopAtShanten int) *shantenSearchNode14 {
	// 설명
	children := map[int]*shantenSearchNode13{}
	tiles34 := playerInfo.HandTiles34
	for i := 0; i < 34; i++ {
		if tiles34[i] == 0 {
			continue
		}
		tiles34[i]--
		if CalculateShanten(tiles34) == targetShanten {
			// 설명
			children[i] = _search13(targetShanten, playerInfo, stopAtShanten)
		}
		tiles34[i]++
	}

	return &shantenSearchNode14{
		shanten:  targetShanten,
		children: children,
	}
}

// 설명
func CalculateShantenAndWaits13(tiles34 []int, leftTiles34 []int) (shanten int, waits Waits) {
	if len(leftTiles34) == 0 {
		leftTiles34 = InitLeftTiles34WithTiles34(tiles34)
	}

	shanten = CalculateShanten(tiles34)
	pi := &model.PlayerInfo{HandTiles34: tiles34, LeftTiles34: leftTiles34}
	node13 := _search13(shanten, pi, shanten) // 설명
	waits = node13.waits
	return
}

// 설명
func searchShanten14(shanten int, playerInfo *model.PlayerInfo, stopAtShanten int) *shantenSearchNode14 {
	if shanten == shantenStateAgari {
		return &shantenSearchNode14{
			shanten:  shanten,
			children: map[int]*shantenSearchNode13{},
		}
	}
	return _search14(shanten, playerInfo, stopAtShanten)
}
