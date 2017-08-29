package msg

import (
	"errors"
	"fmt"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"reflect"
	"xianxia/common/encrypt"
	"xianxia/servers/gate/conf"
)

type GateProcessor struct {
	MsgRouter *chanrpc.Server
	MsgType   reflect.Type
	Encrypter encrypt.Encrypter
}

type MsgHandler func([]interface{})

func NewProcessor() *GateProcessor {
	p := new(GateProcessor)
	p.MsgType = reflect.TypeOf([]byte{})

	p.Encrypter = &encrypt.Cycle{
		Key:        conf.CycleEncrypterKey,
		RandKeyLen: 4,
	}

	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *GateProcessor) Register(msg interface{}) error {
	return nil
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *GateProcessor) SetRouter(msg interface{}, msgRouter *chanrpc.Server) {
	if reflect.TypeOf(msg) != p.MsgType {
		log.Fatal("msg:%v type must be %v", reflect.TypeOf(msg), p.MsgType)
		return
	}

	p.MsgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *GateProcessor) SetHandler(msg interface{}, msgHandler MsgHandler) {

}

// goroutine safe
func (p *GateProcessor) Route(msg interface{}, userData interface{}) error {
	if p.MsgRouter != nil {
		p.MsgRouter.Go(p.MsgType, msg, userData)
	}

	return nil
}

// goroutine safe\=-
func (p *GateProcessor) Unmarshal(data []byte) (interface{}, error) {
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
func (p *GateProcessor) Marshal(msg interface{}) ([][]byte, error) {
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
