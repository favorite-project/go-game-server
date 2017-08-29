package msg

import (
	"encoding/binary"
	"xianxia/servers/game/conf"
)

type MsgHandler func([]interface{})

var Processor *BProcessor = NewBProcessor()
var RdWrEndian binary.ByteOrder

func init() {
	RdWrEndian = binary.BigEndian

	if conf.LittleEndian {
		RdWrEndian = binary.LittleEndian
	}
}
