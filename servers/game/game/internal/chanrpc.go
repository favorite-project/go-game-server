package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/game/logic/net"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC(reflect.TypeOf([]byte("")), rpcMsg)
}

func rpcNewAgent(args []interface{}) {
	agent := args[0].(gate.Agent)
	connection := net.CreateConnection(agent)
	agent.SetUserData(connection)
	connection.OnAccept()
}

func rpcCloseAgent(args []interface{}) {
	agent := args[0].(gate.Agent)
	agent.UserData().(global.Connection).OnClose()
	agent.SetUserData(nil)
	agent.Destroy()
}

func rpcMsg(args []interface{}) {
	agent := args[1].(gate.Agent)
	data, ok := args[0].([]byte)
	if !ok {
		log.Error("msg type error:%v, %v", data, agent)
		return
	}

	agent.UserData().(global.Connection).OnRecv(data)
}
