package msg

import (
	"xianxia/servers/game/game/global"
	"bytes"
	"encoding/binary"
	"xianxia/servers/game/game/global/db"
)

//开启副本
type GSCL_EnterInstance struct {
	global.RootMessage
	Suc bool
}

func (msg *GSCL_EnterInstance) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Instance
	msg.RootMessage.RootKeySub = global.Message_RootKey_Instance_Enter

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)
	binary.Write(buf, RdWrEndian, msg.Suc)

	return buf.Bytes()
}

//副本信息
type GSCL_PlayerInstanceInfo struct {
	global.RootMessage
	*db.Player_InstanceDB_Data
}

func (msg *GSCL_PlayerInstanceInfo) MakeBuffer() []byte {
	msg.RootMessage.RootKey = global.Message_RootKey_Instance
	msg.RootMessage.RootKeySub = global.Message_RootKey_Instance_Info

	buf := new(bytes.Buffer)
	binary.Write(buf, RdWrEndian, msg.RootMessage)

	if msg.Player_InstanceDB_Data != nil {
		binary.Write(buf, RdWrEndian, uint16(len(msg.Player_InstanceDB_Data.MFreeCount)))
		for instanceId, count := range msg.Player_InstanceDB_Data.MFreeCount {
			binary.Write(buf, RdWrEndian, instanceId)
			binary.Write(buf, RdWrEndian, count)
		}
	} else {
		binary.Write(buf, RdWrEndian, uint16(0))
	}

	return buf.Bytes()
}
