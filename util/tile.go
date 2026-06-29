package util

import (
	"fmt"
	"math/rand"
	"sort"
)

var Mahjong = [...]string{
	"1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m",
	"1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p",
	"1s", "2s", "3s", "4s", "5s", "6s", "7s", "8s", "9s",
	"1z", "2z", "3z", "4z", "5z", "6z", "7z",
}

var MahjongU = [...]string{
	"1M", "2M", "3M", "4M", "5M", "6M", "7M", "8M", "9M",
	"1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p",
	"1S", "2S", "3S", "4S", "5S", "6S", "7S", "8S", "9S",
	"1Z", "2Z", "3Z", "4Z", "5Z", "6Z", "7Z",
}

var MahjongZH = [...]string{
	"1만", "2만", "3만", "4만", "5만", "6만", "7만", "8만", "9만",
	"1통", "2통", "3통", "4통", "5통", "6통", "7통", "8통", "9통",
	"1삭", "2삭", "3삭", "4삭", "5삭", "6삭", "7삭", "8삭", "9삭",
	"동", "남", "서", "북", "백", "발", "중",
}

var YaochuTiles = [...]int{0, 8, 9, 17, 18, 26, 27, 28, 29, 30, 31, 32, 33}

func TilesToMahjongZH(tiles []int) (words []string) {
	for _, tile := range tiles {
		words = append(words, MahjongZH[tile])
	}
	return
}

func TilesToMahjongZHInterface(tiles []int) (words []interface{}) {
	for _, tile := range tiles {
		words = append(words, MahjongZH[tile])
	}
	return
}

// 유효패
// map[유효패]남은 수
type Waits map[int]int

func (w Waits) AllCount() (count int) {
	for _, cnt := range w {
		count += cnt
	}
	return count
}

// 남은 수가 0이 아닌 유효패
func (w Waits) AvailableTiles() []int {
	if len(w) == 0 {
		return nil
	}

	tileIndexes := []int{}
	for idx, left := range w {
		if left > 0 {
			tileIndexes = append(tileIndexes, idx)
		}
	}
	sort.Ints(tileIndexes)

	return tileIndexes
}

func (w Waits) indexes() []int {
	if len(w) == 0 {
		return nil
	}

	tileIndexes := make([]int, 0, len(w))
	for idx := range w {
		tileIndexes = append(tileIndexes, idx)
	}
	sort.Ints(tileIndexes)

	return tileIndexes
}

func (w Waits) ParseIndex() (allCount int, indexes []int) {
	return w.AllCount(), w.indexes()
}

func (w Waits) _parse(template [34]string) (allCount int, tiles []string) {
	if len(w) == 0 {
		return 0, nil
	}

	tileIndexes := make([]int, 0, len(w))
	for idx, cnt := range w {
		tileIndexes = append(tileIndexes, idx)
		allCount += cnt
	}
	sort.Ints(tileIndexes)

	tiles = make([]string, len(tileIndexes))
	for i, idx := range tileIndexes {
		tiles[i] = template[idx]
	}

	return allCount, tiles
}

func (w Waits) parse() (allCount int, tiles []string) {
	return w._parse(Mahjong)
}

func (w Waits) parseZH() (allCount int, tilesZH []string) {
	return w._parse(MahjongZH)
}

func (w Waits) tilesZH() []string {
	_, tiles := w.parseZH()
	return tiles
}

func (w Waits) String() string {
	return fmt.Sprintf("%d 유효패 %s", w.AllCount(), TilesToStrWithBracket(w.indexes()))
}

func (w Waits) Equals(w1 Waits) bool {
	tiles0, tiles1 := w.AvailableTiles(), w1.AvailableTiles()
	if len(tiles0) != len(tiles1) {
		return false
	}
	for i := range tiles0 {
		if tiles0[i] != tiles1[i] {
			return false
		}
	}
	return true
}

func isMan(tile int) bool {
	return tile < 9
}

func isPin(tile int) bool {
	return tile >= 9 && tile < 18
}

func isSou(tile int) bool {
	return tile >= 18 && tile < 27
}

func isYaochupai(tile int) bool {
	if tile >= 27 {
		return true
	}
	t := tile % 9
	return t == 0 || t == 8
}

// tiles34가 13장일 때 tile을 넣었을 경우 고립패인지 판단한다.
func isIsolatedTile(tile int, tiles34 []int) bool {
	if tile >= 27 {
		return tiles34[tile] == 0
	}
	t := tile % 9
	l := tile - t + MaxInt(0, t-2)
	r := tile - t + MinInt(8, t+2)
	for i := l; i <= r; i++ {
		if tiles34[i] > 0 {
			return false
		}
	}
	return true
}

// 손패 매수 계산
func CountOfTiles34(tiles34 []int) (count int) {
	for _, c := range tiles34 {
		count += c
	}
	return
}

// 손패 또이츠 수 계산
func CountPairsOfTiles34(tiles34 []int) (count int) {
	for _, c := range tiles34 {
		if c >= 2 {
			count++
		}
	}
	return
}

func InitLeftTiles34() []int {
	leftTiles34 := make([]int, 34)
	for i := range leftTiles34 {
		leftTiles34[i] = 4
	}
	return leftTiles34
}

// 전달된 패를 제거한 뒤 남은 패를 반환한다.
func InitLeftTiles34WithTiles34(tiles34 []int) []int {
	leftTiles34 := make([]int, 34)
	for i, count := range tiles34 {
		leftTiles34[i] = 4 - count
	}
	return leftTiles34
}

// 바깥쪽 패 계산
func OutsideTiles(tile int) (outsideTiles []int) {
	if tile >= 27 {
		return
	}
	switch tile%9 + 1 {
	case 1, 9:
		return
	case 2, 3, 4:
		for i := tile - tile%9; i < tile; i++ {
			outsideTiles = append(outsideTiles, i)
		}
	case 5:
		// 초반 5 타패 후에는 37이 비교적 안전하다. TODO: 한쪽 스지 A 46도 고려
		outsideTiles = append(outsideTiles, tile-2, tile+2)
	case 6, 7, 8:
		for i := tile - tile%9 + 8; i > tile; i-- {
			outsideTiles = append(outsideTiles, i)
		}
	default:
		panic(fmt.Errorf("[OutsideTiles] 코드 오류: tile = %d", tile))
	}
	return
}

// 무작위로 패 한 장을 보충한다.
func RandomAddTile(tiles34 []int) {
	for {
		if tile := rand.Intn(34); tiles34[tile] < 4 {
			tiles34[tile]++
			break
		}
	}
}
