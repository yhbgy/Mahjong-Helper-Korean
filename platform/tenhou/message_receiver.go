package tenhou

import (
	"encoding/json"
	"regexp"
	"time"
)

type MessageReceiver struct {
	originMessageQueue  chan []byte
	orderedMessageQueue chan []byte
}

func NewMessageReceiver() *MessageReceiver {
	const maxQueueSize = 100
	mr := &MessageReceiver{
		originMessageQueue:  make(chan []byte, maxQueueSize),
		orderedMessageQueue: make(chan []byte, maxQueueSize),
	}
	go mr.run()
	return mr
}

var isSelfDraw = regexp.MustCompile("^T[0-9]{1,3}$").MatchString

// TODO: 설명 보완
func (mr *MessageReceiver) isSelfDraw(data []byte) bool {
	d := struct {
		Tag string `json:"tag"`
	}{}
	if err := json.Unmarshal(data, &d); err != nil {
		return false
	}
	return isSelfDraw(d.Tag)
}

func (mr *MessageReceiver) run() {
	for data := range mr.originMessageQueue {
		if !mr.isSelfDraw(data) {
			mr.orderedMessageQueue <- data
			continue
		}

		// 설명
		time.Sleep(75 * time.Millisecond) // 설명

		// 설명
		if len(mr.originMessageQueue) == 0 {
			mr.orderedMessageQueue <- data
			continue
		}

		// 설명
		// 설명
		// 설명
		mr.originMessageQueue <- data
	}
}

func (mr *MessageReceiver) Put(data []byte) {
	mr.originMessageQueue <- data
}

func (mr *MessageReceiver) Get() []byte {
	return <-mr.orderedMessageQueue
}

func (mr *MessageReceiver) IsEmpty() bool {
	return len(mr.originMessageQueue) == 0 && len(mr.orderedMessageQueue) == 0
}
