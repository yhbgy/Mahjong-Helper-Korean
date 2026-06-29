package main

const (
	majsoulGameConfigCategoryFriends = 1 // 설명
	majsoulGameConfigCategoryMatch   = 2 // 설명
)

// 설명
// 설명
type majsoulGameConfig struct {
	Category int `json:"category"`
	Mode     *struct {
		Mode       int `json:"mode"`
		DetailRule *struct {
			GuyiMode int `json:"guyi_mode"`
		} `json:"detail_rule"`
	} `json:"mode"`
}

func (c *majsoulGameConfig) isGuyiMode() bool {
	return c != nil && c.Mode != nil && c.Mode.DetailRule != nil && c.Mode.DetailRule.GuyiMode == 1
}
