package msg

import (
	"errors"
	"fmt"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"reflect"
	"xianxia/common/encrypt"
)

type BProcessor struct {
	MsgRouter *chanrpc.Server
	MsgType   reflect.Type
	Encrypter encrypt.Encrypter
}

func NewBProcessor() *BProcessor {
	p := new(BProcessor)
	p.MsgType = reflect.TypeOf([]byte{})

	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *BProcessor) Register(msg interface{}) error {
	return nil
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *BProcessor) SetRouter(msg interface{}, msgRouter *chanrpc.Server) {
	if reflect.TypeOf(msg) != p.MsgType {
		log.Fatal("msg:%v type must be %v", reflect.TypeOf(msg), p.MsgType)
		return
	}

	p.MsgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *BProcessor) SetHandler(msg interface{}, msgHandler MsgHandler) {

}

// goroutine safe
func (p *BProcessor) Route(msg interface{}, userData interface{}) error {
	if p.MsgRouter != nil {
		p.MsgRouter.Go(p.MsgType, msg, userData)
	}

	return nil
}

// goroutine safe\=-
func (p *BProcessor) Unmarshal(data []byte) (interface{}, error) {
	if p.Encrypter != nil {
		var err error
		data, err = p.Encrypter.Decrypt(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// goroutine safe
func (p *BProcessor) Marshal(msg interface{}) ([][]byte, error) {
	if reflect.TypeOf(msg) != p.MsgType {
		return nil, errors.New(fmt.Sprintf("msg:%v Marshal must be []byte", msg))
	}

	data := msg.([]byte)
	if p.Encrypter != nil {
		var err error
		data, err = p.Encrypter.Encrypt(data)
		if err != nil {
			return nil, err
		}
	}

	return [][]byte{data}, nil
}
