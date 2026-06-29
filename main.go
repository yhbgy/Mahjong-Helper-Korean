package main

import (
	"flag"
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"github.com/fatih/color"
	"math/rand"
	"strings"
	"time"
)

var (
	considerOldYaku bool

	isMajsoul     bool
	isTenhou      bool
	isAnalysis    bool
	isInteractive bool

	showImproveDetail      bool
	showAgariAboveShanten1 bool
	showScore              bool
	showAllYakuTypes       bool

	humanDoraTiles string

	port int
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.BoolVar(&considerOldYaku, "old", false, "고역 허용")
	flag.BoolVar(&isMajsoul, "majsoul", false, "작혼 도우미")
	flag.BoolVar(&isTenhou, "tenhou", false, "천봉 도우미")
	flag.BoolVar(&isAnalysis, "analysis", false, "분석 모드")
	flag.BoolVar(&isInteractive, "interactive", false, "대화형 모드")
	flag.BoolVar(&isInteractive, "i", false, "-interactive와 같음")
	flag.BoolVar(&showImproveDetail, "detail", false, "개량 상세 표시")
	flag.BoolVar(&showAgariAboveShanten1, "agari", false, "텐파이 전 예상 화료율 표시")
	flag.BoolVar(&showAgariAboveShanten1, "a", false, "-agari와 같음")
	flag.BoolVar(&showScore, "score", false, "국 수지 표시")
	flag.BoolVar(&showScore, "s", false, "-score와 같음")
	flag.BoolVar(&showAllYakuTypes, "yaku", false, "모든 역 표시")
	flag.BoolVar(&showAllYakuTypes, "y", false, "-yaku와 같음")
	flag.StringVar(&humanDoraTiles, "dora", "", "도라로 볼 패 지정")
	flag.StringVar(&humanDoraTiles, "d", "", "-dora와 같음")
	flag.IntVar(&port, "port", 12121, "서버 포트 지정")
	flag.IntVar(&port, "p", 12121, "-port와 같음")
}

const (
	platformTenhou  = 0
	platformMajsoul = 1

	defaultPlatform = platformMajsoul
)

var platforms = map[int][]string{
	platformTenhou: {
		"천봉",
		"Web",
		"4K",
	},
	platformMajsoul: {
		"작혼",
		"국제 중국어 서버",
		"일본 서버",
		"국제 서버",
	},
}

const readmeURL = "https://github.com/EndlessCheng/mahjong-helper/blob/master/README.md"
const issueURL = "https://github.com/EndlessCheng/mahjong-helper/issues"
const issueCommonQuestions = "https://github.com/EndlessCheng/mahjong-helper/issues/104"
const qqGroupNum = "375865038"

func welcome() int {
	fmt.Println("사용 설명: " + readmeURL)
	fmt.Println("문제 제보: " + issueURL)
	fmt.Println("피드백 그룹: " + qqGroupNum)
	fmt.Println()

	fmt.Println("숫자를 입력해 사용할 사이트를 선택하세요:")
	for i, cnt := 0, 0; cnt < len(platforms); i++ {
		if platformInfo, ok := platforms[i]; ok {
			info := platformInfo[0] + " [" + strings.Join(platformInfo[1:], ",") + "]"
			fmt.Printf("%d - %s\n", i, info)
			cnt++
		}
	}

	choose := defaultPlatform
	fmt.Scanln(&choose) // 설명
	platformInfo, ok := platforms[choose]
	var platformName string
	if ok {
		platformName = platformInfo[0]
	}
	if !ok {
		choose = defaultPlatform
		platformName = platforms[choose][0]
	}

	clearConsole()
	color.HiGreen("선택됨 - %s", platformName)

	if choose == platformMajsoul {
		if len(gameConf.MajsoulAccountIDs) == 0 {
			color.HiYellow(`
알림: 처음 사용할 때는 CPU전을 한 판 시작하거나 게임에 다시 로그인해 주세요.
이 단계는 계정 ID와 게임 시작 시 자풍을 확인하기 위한 것이며, 완료되지 않으면 이후 데이터를 분석할 수 없습니다.

도우미가 반응하지 않으면 설치 단계를 완료했는지 확인해 주세요.
관련 링크 ` + issueCommonQuestions)
		}
	}

	return choose
}

func main() {
	flag.Parse()

	color.HiGreen("리치마작 도우미 %s (by EndlessCheng)", version)
	if version != versionDev {
		go checkNewVersion(version)
	}

	util.SetConsiderOldYaku(considerOldYaku)

	humanTiles := strings.Join(flag.Args(), " ")
	humanTilesInfo := &model.HumanTilesInfo{
		HumanTiles:     humanTiles,
		HumanDoraTiles: humanDoraTiles,
	}

	var err error
	switch {
	case isMajsoul:
		err = runServer(true, port)
	case isTenhou || isAnalysis:
		err = runServer(true, port)
	case isInteractive: // 설명
		err = interact(humanTilesInfo)
	case len(flag.Args()) > 0: // 설명
		_, err = analysisHumanTiles(humanTilesInfo)
	default: // 설명
		choose := welcome()
		isHTTPS := choose == platformMajsoul
		err = runServer(isHTTPS, port)
	}
	if err != nil {
		errorExit(err)
	}
}
