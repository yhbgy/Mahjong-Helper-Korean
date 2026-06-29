package main

import (
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/util"
	"sort"
	"strconv"
	"time"
)

type _majsoulRecordAccount struct {
	AccountID int `json:"account_id"`
	// 설명
	Seat     int    `json:"seat"` // 설명
	Nickname string `json:"nickname"`
}

// 설명
type majsoulRecordBaseInfo struct {
	UUID      string `json:"uuid"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`

	Config *majsoulGameConfig `json:"config"`

	Accounts []_majsoulRecordAccount `json:"accounts"`
}

func (i *majsoulRecordBaseInfo) sort() {
	sort.Slice(i.Accounts, func(i_, j int) bool {
		return i.Accounts[i_].Seat < i.Accounts[j].Seat
	})
}

var seatNameZH = []string{"동", "남", "서", "북"}

func (i *majsoulRecordBaseInfo) String() string {
	i.sort()

	const timeFormat = "2006-01-02 15:04:05"
	output := fmt.Sprintf("%s\n%s부터\n%s까지\n\n", i.UUID, time.Unix(i.StartTime, 0).Format(timeFormat), time.Unix(i.EndTime, 0).Format(timeFormat))

	maxAccountID := 0
	for _, account := range i.Accounts {
		maxAccountID = util.MaxInt(maxAccountID, account.AccountID)
	}
	accountShownWidth := len(strconv.Itoa(maxAccountID))
	for _, account := range i.Accounts {
		output += fmt.Sprintf("%s %*d %s\n", seatNameZH[account.Seat], accountShownWidth, account.AccountID, account.Nickname)
	}
	return output
}

func (i *majsoulRecordBaseInfo) getSelfSeat(accountID int) (int, error) {
	if len(i.Accounts) == 0 {
		return -1, fmt.Errorf("패보 기본 정보가 비어 있습니다")
	}
	for _, account := range i.Accounts {
		if account.AccountID == accountID {
			return account.Seat, nil
		}
	}
	// 설명
	return 0, nil
}

//

// 설명
type majsoulRecordAction struct {
	Name   string          `json:"name"`
	Action *majsoulMessage `json:"data"`
}

type majsoulRoundActions []*majsoulRecordAction

func (l majsoulRoundActions) append(action *majsoulRecordAction) (majsoulRoundActions, error) {
	if action == nil {
		return nil, fmt.Errorf("데이터 이상: 가져온 동작 내용이 비어 있습니다")
	}
	newL := l

	if action.Name == "RecordNewRound" {
		newL = majsoulRoundActions{action}
	} else {
		if len(newL) == 0 {
			return nil, fmt.Errorf("데이터 이상: RecordNewRound를 받지 못했습니다")
		}
		newL = append(newL, action)
	}

	return newL, nil
}

func parseMajsoulRecordAction(actions []*majsoulRecordAction) (roundActionsList []majsoulRoundActions, err error) {
	if len(actions) == 0 {
		return nil, fmt.Errorf("데이터 이상: 가져온 패보 내용이 비어 있습니다")
	}

	var currentRoundActions majsoulRoundActions
	for _, action := range actions {
		if action.Name == "RecordNewRound" {
			if len(currentRoundActions) > 0 {
				roundActionsList = append(roundActionsList, currentRoundActions)
			}
			currentRoundActions = []*majsoulRecordAction{action}
		} else {
			if len(currentRoundActions) == 0 {
				return nil, fmt.Errorf("데이터 이상: RecordNewRound를 받지 못했습니다")
			}
			currentRoundActions = append(currentRoundActions, action)
		}
	}
	if len(currentRoundActions) > 0 {
		roundActionsList = append(roundActionsList, currentRoundActions)
	}
	return
}
