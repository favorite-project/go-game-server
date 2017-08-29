package msg

import (
	"bytes"
	"encoding/binary"
	_ "errors"
	"fmt"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"reflect"
	"xianxia/servers/game/conf"
)

type CusProcessor struct {
	msgInfo map[uint32]*MsgInfo
}

type MsgInfo struct {
	msgType    reflect.Type
	msgRouter  *chanrpc.Server
	msgHandler MsgHandler
}

type MsgHead struct {
	MsgID uint32
}

func NewProcessor() *CusProcessor {
	p := new(CusProcessor)
	p.msgInfo = make(map[uint32]*MsgInfo)
	return p
}

func (p *CusProcessor) checkMsg(msg interface{}) (msgID uint32) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Struct || msgType.NumField() < 1 {
		log.Fatal("struct message type and Field Num >= 1  required")
	}

	msgHead := reflect.ValueOf(msg).Field(0)
	if msgHead.Type() != reflect.TypeOf(MsgHead{}) {
		log.Fatal("message field 0 must be MsgHead")
	}

	msgID = uint32(msgHead.Field(0).Uint())
	if msgID == 0 {
		log.Fatal("message ID empty")
	}

	return msgID
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *CusProcessor) Register(msg interface{}) uint32 {
	msgID := p.checkMsg(msg)
	if _, ok := p.msgInfo[msgID]; ok {
		log.Fatal("message %v is already registered", msgID)
	}

	i := new(MsgInfo)
	i.msgType = reflect.TypeOf(msg)
	p.msgInfo[msgID] = i
	return msgID
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *CusProcessor) SetRouter(msg interface{}, msgRouter *chanrpc.Server) {
	msgID := p.checkMsg(msg)
	i, ok := p.msgInfo[msgID]
	if !ok {
		log.Fatal("message %v not registered", msgID)
	}

	i.msgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *CusProcessor) SetHandler(msg interface{}, msgHandler MsgHandler) {
	msgID := p.checkMsg(msg)
	i, ok := p.msgInfo[msgID]
	if !ok {
		log.Fatal("message %v not registered", msgID)
	}

	i.msgHandler = msgHandler
}

// goroutine safe
func (p *CusProcessor) Route(msg interface{}, userData interface{}) error {
	msgID := p.checkMsg(msg)
	i, ok := p.msgInfo[msgID]
	if !ok {
		return fmt.Errorf("message %v not registered", msgID)
	}
	if i.msgHandler != nil {
		i.msgHandler([]interface{}{msg, userData})
	}
	if i.msgRouter != nil {
		i.msgRouter.Go(i.msgType, msg, userData)
	}
	return nil
}

// goroutine safe\=-
func (p *CusProcessor) Unmarshal(data []byte) (interface{}, error) {
	var msgID uint32
	if conf.LittleEndian {
		msgID = binary.LittleEndian.Uint32(data[:4])
	} else {
		msgID = binary.BigEndian.Uint32(data[:4])
	}

	i, ok := p.msgInfo[msgID]
	if !ok {
		return nil, fmt.Errorf("message %v not registered", msgID)
	}

	msg := reflect.New(i.msgType)
	pmsg := msg.Interface()
	reader := bytes.NewReader(data)
	var err error
	if conf.LittleEndian {
		err = binary.Read(reader, binary.LittleEndian, pmsg)
	} else {
		err = binary.Read(reader, binary.BigEndian, pmsg)
	}

	return reflect.ValueOf(pmsg).Elem().Interface(), err
}

// goroutine safe
func (p *CusProcessor) Marshal(msg interface{}) ([][]byte, error) {

	/*
		msgID := p.checkMsg(msg)
		if _, ok := p.msgInfo[msgID]; !ok {
			return nil, fmt.Errorf("message %v not registered", msgID)
		}
	*/

	// data
	buf := new(bytes.Buffer)
	var err error
	if conf.LittleEndian {
		err = binary.Write(buf, binary.LittleEndian, msg)
	} else {
		err = binary.Write(buf, binary.BigEndian, msg)
	}

	return [][]byte{buf.Bytes()}, err
}
