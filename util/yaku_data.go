package util

import (
	"fmt"
	"sort"
)

var considerOldYaku bool

func SetConsiderOldYaku(b bool) {
	considerOldYaku = b
}

//

const (
	// https://en.wikipedia.org/wiki/Japanese_Mahjong_yaku
	// Special criteria
	YakuRiichi int = iota
	YakuChiitoi

	// Yaku based on luck
	YakuTsumo
	//YakuIppatsu
	//YakuHaitei
	//YakuHoutei
	//YakuRinshan
	//YakuChankan
	YakuDaburii

	// Yaku based on sequences
	YakuPinfu
	YakuRyanpeikou
	YakuIipeikou
	YakuSanshokuDoujun // *
	YakuIttsuu         // *

	// Yaku based on triplets and/or quads
	YakuToitoi
	YakuSanAnkou
	YakuSanshokuDoukou
	YakuSanKantsu

	// Yaku based on terminal or honor tiles
	YakuTanyao
	YakuYakuhai
	YakuChanta    // * 슌쯔가 필요
	YakuJunchan   // * 슌쯔가 필요
	YakuHonroutou // 치또이츠도 포함
	YakuShousangen

	// Yaku based on suits
	YakuHonitsu  // *
	YakuChinitsu // *

	// Yakuman
	//YakuKokushi
	//YakuKokushi13
	YakuSuuAnkou
	YakuSuuAnkouTanki
	YakuDaisangen
	YakuShousuushii
	YakuDaisuushii
	YakuTsuuiisou
	YakuChinroutou
	YakuRyuuiisou
	YakuChuuren
	YakuChuuren9
	YakuSuuKantsu
	//YakuTenhou
	//YakuChiihou

	// 고역
	YakuShiiaruraotai
	YakuUumensai
	YakuSanrenkou
	YakuIsshokusanjun

	// 고역 역만
	YakuDaisuurin
	YakuDaisharin
	YakuDaichikurin
	YakuDaichisei

	//_endYakuType  // enum 끝 표시. YakuType 개수 계산용
)

//const maxYakuType = _endYakuType

var YakuNameMap = map[int]string{
	// Special criteria
	YakuRiichi:  "리치",
	YakuChiitoi: "치또이츠",

	// Yaku based on luck
	YakuTsumo: "쯔모",
	//YakuIppatsu: "일발",
	//YakuHaitei:  "해저",
	//YakuHoutei:  "하저",
	//YakuRinshan: "영상",
	//YakuChankan: "창깡",
	YakuDaburii: "더블리치",

	// Yaku based on sequences
	YakuPinfu:          "핑후",
	YakuRyanpeikou:     "량페코",
	YakuIipeikou:       "이페코",
	YakuSanshokuDoujun: "삼색동순",
	YakuIttsuu:         "일기통관",

	// Yaku based on triplets and/or quads
	YakuToitoi:         "또이또이",
	YakuSanAnkou:       "산안커",
	YakuSanshokuDoukou: "삼색동각",
	YakuSanKantsu:      "산깡쯔",

	// Yaku based on terminal or honor tiles
	YakuTanyao:     "탕야오",
	YakuYakuhai:    "역패",
	YakuChanta:     "찬타",
	YakuJunchan:    "준찬타",
	YakuHonroutou:  "혼노두", // 치또이츠도 포함
	YakuShousangen: "소삼원",

	// Yaku based on suits
	YakuHonitsu:  "혼일색",
	YakuChinitsu: "청일색",

	// Yakuman
	//YakuKokushi:       "국사",
	//YakuKokushi13:     "국사 13면",
	YakuSuuAnkou:      "스안커",
	YakuSuuAnkouTanki: "스안커 단기",
	YakuDaisangen:     "대삼원",
	YakuShousuushii:   "소사희",
	YakuDaisuushii:    "대사희",
	YakuTsuuiisou:     "자일색",
	YakuChinroutou:    "청노두",
	YakuRyuuiisou:     "녹일색",
	YakuChuuren:       "구련보등",
	YakuChuuren9:      "순정구련보등",
	YakuSuuKantsu:     "스깡쯔",
	//YakuTenhou:        "천화",
	//YakuChiihou:       "지화",
}

var OldYakuNameMap = map[int]string{
	YakuShiiaruraotai: "십이낙태",
	YakuUumensai:      "오문제",
	YakuSanrenkou:     "삼연각",
	YakuIsshokusanjun: "일색삼순",

	YakuDaisuurin:   "대수린",
	YakuDaisharin:   "대차륜",
	YakuDaichikurin: "대죽림",
	YakuDaichisei:   "대칠성",
}

func YakuTypesToStr(yakuTypes []int) string {
	if len(yakuTypes) == 0 {
		return "[역 없음]"
	}
	names := []string{}
	for _, t := range yakuTypes {
		if name, ok := YakuNameMap[t]; ok {
			names = append(names, name)
		}
	}

	if considerOldYaku {
		for _, t := range yakuTypes {
			if name, ok := OldYakuNameMap[t]; ok {
				names = append(names, name)
			}
		}
	}

	return fmt.Sprint(names)
}

func YakuTypesWithDoraToStr(yakuTypes map[int]struct{}, numDora int) string {
	if len(yakuTypes) == 0 {
		return "[역 없음]"
	}
	yt := []int{}
	for t := range yakuTypes {
		yt = append(yt, t)
	}
	sort.Ints(yt)
	names := []string{}
	for _, t := range yt {
		names = append(names, YakuNameMap[t])
	}
	// TODO: old yaku
	if numDora > 0 {
		names = append(names, fmt.Sprintf("도라%d", numDora))
	}
	return fmt.Sprint(names)
}

//

type _yakuHanMap map[int]int
type _yakumanTimesMap map[int]int

var YakuHanMap = _yakuHanMap{
	YakuRiichi:  1,
	YakuChiitoi: 2,

	YakuTsumo: 1,
	//YakuIppatsu: 1,
	//YakuHaitei:  1,
	//YakuHoutei:  1,
	//YakuRinshan: 1,
	//YakuChankan: 1,
	YakuDaburii: 2,

	YakuPinfu:          1,
	YakuRyanpeikou:     3,
	YakuIipeikou:       1,
	YakuSanshokuDoujun: 2,
	YakuIttsuu:         2,

	YakuToitoi:         2,
	YakuSanAnkou:       2,
	YakuSanshokuDoukou: 2,
	YakuSanKantsu:      2,

	YakuTanyao:     1,
	YakuYakuhai:    1,
	YakuChanta:     2,
	YakuJunchan:    3,
	YakuHonroutou:  2,
	YakuShousangen: 2,

	YakuHonitsu:  3,
	YakuChinitsu: 6,
}

var NakiYakuHanMap = _yakuHanMap{
	//YakuHaitei:  1,
	//YakuHoutei:  1,
	//YakuRinshan: 1,
	//YakuChankan: 1,

	YakuSanshokuDoujun: 1,
	YakuIttsuu:         1,

	YakuToitoi:         2,
	YakuSanAnkou:       2,
	YakuSanshokuDoukou: 2,
	YakuSanKantsu:      2,

	YakuTanyao:     1,
	YakuYakuhai:    1,
	YakuChanta:     1,
	YakuJunchan:    2,
	YakuHonroutou:  2,
	YakuShousangen: 2,

	YakuHonitsu:  2,
	YakuChinitsu: 5,
}

var OldYakuHanMap = _yakuHanMap{
	YakuUumensai:      2,
	YakuSanrenkou:     2,
	YakuIsshokusanjun: 3,
}

var OldNakiYakuHanMap = _yakuHanMap{
	YakuShiiaruraotai: 1, // 4후로 대조차
	YakuUumensai:      2,
	YakuSanrenkou:     2,
	YakuIsshokusanjun: 2,
}

// yakuTypes(비역만)의 누적 판수를 계산한다.
func CalcYakuHan(yakuTypes []int, isNaki bool) (cntHan int) {
	var yakuHanMap _yakuHanMap
	if !isNaki {
		yakuHanMap = YakuHanMap
	} else {
		yakuHanMap = NakiYakuHanMap
	}

	for _, yakuType := range yakuTypes {
		if han, ok := yakuHanMap[yakuType]; ok {
			cntHan += han
		}
	}

	if considerOldYaku {
		if !isNaki {
			yakuHanMap = OldYakuHanMap
		} else {
			yakuHanMap = OldNakiYakuHanMap
		}

		for _, yakuType := range yakuTypes {
			if han, ok := yakuHanMap[yakuType]; ok {
				cntHan += han
			}
		}
	}

	return
}

//

var YakumanTimesMap = map[int]int{
	//YakuKokushi:       1,
	//YakuKokushi13:     2,
	YakuSuuAnkou:      1,
	YakuSuuAnkouTanki: 2,
	YakuDaisangen:     1,
	YakuShousuushii:   1,
	YakuDaisuushii:    2,
	YakuTsuuiisou:     1,
	YakuChinroutou:    1,
	YakuRyuuiisou:     1,
	YakuChuuren:       1,
	YakuChuuren9:      2,
	YakuSuuKantsu:     1,
	//YakuTenhou:        1,
	//YakuChiihou:       1,
}

var NakiYakumanTimesMap = map[int]int{
	YakuDaisangen:   1,
	YakuShousuushii: 1,
	YakuDaisuushii:  2,
	YakuTsuuiisou:   1,
	YakuChinroutou:  1,
	YakuRyuuiisou:   1,
	YakuSuuKantsu:   1,
}

var OldYakumanTimesMap = map[int]int{
	YakuDaisuurin:   1,
	YakuDaisharin:   1,
	YakuDaichikurin: 1,
	YakuDaichisei:   1, // 자일색과 복합되어 실제로는 더블 역만
}

// 역만 배수를 계산한다.
func CalcYakumanTimes(yakuTypes []int, isNaki bool) (times int) {
	var yakumanTimesMap _yakumanTimesMap
	if !isNaki {
		yakumanTimesMap = YakumanTimesMap
	} else {
		yakumanTimesMap = NakiYakumanTimesMap
	}

	for _, yakuman := range yakuTypes {
		if t, ok := yakumanTimesMap[yakuman]; ok {
			times += t
		}
	}

	if considerOldYaku && !isNaki {
		for _, yakuman := range yakuTypes {
			if t, ok := OldYakumanTimesMap[yakuman]; ok {
				times += t
			}
		}
	}

	return
}
