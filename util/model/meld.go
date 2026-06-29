package model

const (
	MeldTypeChi    = iota // 설명
	MeldTypePon           // 설명
	MeldTypeAnkan         // 설명
	MeldTypeMinkan        // 설명
	MeldTypeKakan         // 설명
)

type Meld struct {
	MeldType int // 설명

	// Tiles == sort(SelfTiles + CalledTile)
	Tiles      []int // 설명
	SelfTiles  []int // 설명
	CalledTile int   // 설명

	// TODO: 설명 보완
	ContainRedFive    bool // 설명
	RedFiveFromOthers bool // 설명
}

// 설명
func (m *Meld) IsKan() bool {
	return m.MeldType == MeldTypeAnkan || m.MeldType == MeldTypeMinkan || m.MeldType == MeldTypeKakan
}
