package internal

import (
	"github.com/name5566/leaf/gate"
	"reflect"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC(reflect.TypeOf([]byte{}), rpcAgentRecv)

	skeleton.RegisterChanRPC("NewGSAgent", rpcNewGSAgent)
	skeleton.RegisterChanRPC("CloseGSAgent", rpcCloseGSAgent)
	skeleton.RegisterChanRPC("GSRecv", rpcGSRecv)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	newGSConn(a)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	closeGSConn(a)
}

func rpcAgentRecv(args []interface{}) {
	a := args[1].(gate.Agent)
	msg := args[0].([]byte)
	writeGSConn(a, msg)
}
