package majsoul

import (
	"encoding/binary"
	"fmt"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/api"
	"github.com/EndlessCheng/mahjong-helper/platform/majsoul/proto/lq"
	"github.com/golang/protobuf/proto"
	"os"
	"reflect"
	"strings"
)

// 설명
// 설명
type Message struct {
	Name            string        `json:"name"`
	RequestMessage  proto.Message `json:"request_message,omitempty"`
	ResponseMessage proto.Message `json:"response_message,omitempty"`
	NotifyMessage   proto.Message `json:"notify_message,omitempty"`
}

type MessageReceiver struct {
	originMessageQueue  chan []byte   // 설명
	orderedMessageQueue chan *Message // 설명

	indexToMessageMap map[uint16]*Message
}

func NewMessageReceiver() *MessageReceiver {
	const maxQueueSize = 100
	mr := &MessageReceiver{
		originMessageQueue:  make(chan []byte, maxQueueSize),
		orderedMessageQueue: make(chan *Message, maxQueueSize),
		indexToMessageMap:   map[uint16]*Message{},
	}
	go mr.run()
	return mr
}

func (mr *MessageReceiver) run() {
	for data := range mr.originMessageQueue {
		messageType := data[0]
		switch messageType {
		case api.MessageTypeNotify:
			notifyName, data, err := api.UnwrapData(data[1:])
			if err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.api.UnwrapData.NOTIFY", err)
				continue
			}
			notifyName = notifyName[1:] // 설명

			mt := proto.MessageType(notifyName)
			if mt == nil {
				fmt.Fprintf(os.Stderr, "MessageReceiver.run %s 를 찾을 수 없습니다. 확인하세요!\n", notifyName)
				continue
			}
			messagePtr := reflect.New(mt.Elem())
			if err := proto.Unmarshal(data, messagePtr.Interface().(proto.Message)); err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.proto.Unmarshal.NOTIFY", notifyName, err)
				continue
			}

			mr.orderedMessageQueue <- &Message{
				Name:          notifyName,
				NotifyMessage: messagePtr.Interface().(proto.Message),
			}
		case api.MessageTypeRequest:
			messageIndex := binary.LittleEndian.Uint16(data[1:3])

			rawMethodName, data, err := api.UnwrapData(data[3:])
			if err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.api.UnwrapData.REQUEST", err)
				continue
			}
			rawMethodName = rawMethodName[1:] // 설명

			// 설명
			splits := strings.Split(rawMethodName, ".")
			clientName, methodName := splits[1], splits[2]
			methodType := lq.FindMethod(clientName, methodName)
			reqType := methodType.In(1)
			respType := methodType.Out(0)

			messagePtr := reflect.New(reqType.Elem())
			if err := proto.Unmarshal(data, messagePtr.Interface().(proto.Message)); err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.proto.Unmarshal.REQUEST", rawMethodName, err)
				continue
			}
			reqMessage := messagePtr.Interface().(proto.Message)

			messagePtr = reflect.New(respType.Elem())
			respMessage := messagePtr.Interface().(proto.Message)

			mr.indexToMessageMap[messageIndex] = &Message{
				Name:            rawMethodName,
				RequestMessage:  reqMessage,
				ResponseMessage: respMessage,
			}
		case api.MessageTypeResponse:
			// 설명
			messageIndex := binary.LittleEndian.Uint16(data[1:3])
			message, ok := mr.indexToMessageMap[messageIndex]
			if !ok {
				// 설명
				continue
			}
			delete(mr.indexToMessageMap, messageIndex)
			if err := api.UnwrapMessage(data[3:], message.ResponseMessage); err != nil {
				fmt.Fprintln(os.Stderr, "MessageReceiver.run.proto.Unmarshal.RESPONSE", message.Name, err)
				continue
			}
			mr.orderedMessageQueue <- message
		default:
			panic(fmt.Sprintln("[MessageReceiver] 데이터 오류", messageType))
		}
	}
}

func (mr *MessageReceiver) Put(data []byte) {
	mr.originMessageQueue <- data
}

func (mr *MessageReceiver) Get() *Message {
	return <-mr.orderedMessageQueue
}
