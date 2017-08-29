package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"xianxia/servers/gate/conf"
	"net"
	"io"
	"encoding/binary"
)

var conns map[gate.Agent]*gsAgent

func init() {
	conns = make(map[gate.Agent]*gsAgent)
}

type gsAgent struct {
	gsconn   net.Conn
	aconn  gate.Agent
	writeChan chan []byte
}

func (agent *gsAgent) WriteMsg(data []byte) {
	if len(agent.writeChan) == 10 {
		log.Error("WriteMsg writechan full")
		return
	}

	var msgLen uint32 = uint32(len(data))
	msg := make([]byte, uint32(conf.LenMsgLen)+msgLen)

	// write len
	switch conf.LenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		if conf.LittleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if conf.LittleEndian {
			binary.LittleEndian.PutUint32(msg, msgLen)
		} else {
			binary.BigEndian.PutUint32(msg, msgLen)
		}
	}

	copy(msg[conf.LenMsgLen:], data)
	agent.writeChan <- msg
}

func (agent *gsAgent) OnClose() {
	agent.aconn = nil
	agent.writeChan <- nil
}

func rpcNewGSAgent(args []interface{}) {
	gsagent := args[0].(*gsAgent)
	if _, ok := conns[gsagent.aconn]; ok {
		log.Error("rpcNewGSAgent gsconn existed!!!")
		return
	}
	conns[gsagent.aconn] = gsagent

	//开启写协程
	gsagent.writeChan = make(chan []byte, 10)
	go func() {
		for b := range gsagent.writeChan {
			if b == nil {
				close(gsagent.writeChan)
				break
			}

			gsagent.gsconn.Write(b)
		}
	}()
}

func rpcCloseGSAgent(args []interface{}) {
	gsagent := args[0].(*gsAgent)
	if gsagent.aconn != nil {
		gsagent.aconn.Close()
		gsagent.aconn.Destroy()
	}

	if _, ok := conns[gsagent.aconn]; !ok {
		return
	}

	delete(conns, gsagent.aconn)
	gsagent.OnClose()
}

func rpcGSRecv(args []interface{}) {
	gsagent := args[0].(*gsAgent)
	if _, ok := conns[gsagent.aconn]; !ok {
		log.Error("rpcGSRecv gsconn empty!!!")
		return
	}

	gsagent.aconn.WriteMsg(args[1])
}

func newGSConn(aconn gate.Agent) {
	agent := &gsAgent{
		aconn:aconn,
	}

	go func() {
		conn, err := net.Dial("tcp", conf.Server.GameServerAddr)
		if err != nil {
			log.Error("connect %s err:%s", conf.Server.GameServerAddr, err)
			ChanRPC.Go("CloseGSAgent", agent)
			return
		}

		agent.gsconn = conn
		ChanRPC.Go("NewGSAgent", agent)
		for {
			var b [4]byte
			bufMsgLen := b[:conf.LenMsgLen]

			// read len
			if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
				log.Error("newGSConn io.ReadFull error1:%s", err)
				break
			}

			// parse len
			var msgLen uint32
			switch conf.LenMsgLen {
			case 1:
				msgLen = uint32(bufMsgLen[0])
			case 2:
				if conf.LittleEndian {
					msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
				} else {
					msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
				}
			case 4:
				if conf.LittleEndian {
					msgLen = binary.LittleEndian.Uint32(bufMsgLen)
				} else {
					msgLen = binary.BigEndian.Uint32(bufMsgLen)
				}
			}

			// data
			msgData := make([]byte, msgLen)
			if _, err := io.ReadFull(conn, msgData); err != nil {
				log.Error("newGSConn io.ReadFull error2:%s", err)
				break
			}

			ChanRPC.Go("GSRecv", agent, msgData)
		}

		ChanRPC.Go("CloseGSAgent", agent)
	}()
}


func closeGSConn(aconn gate.Agent) {
	gsagent, ok := conns[aconn]
	if !ok {
		return
	}

	delete(conns, aconn)
	gsagent.gsconn.Close()
	gsagent.OnClose()
}

func writeGSConn(aconn gate.Agent, msg []byte) {

	if _, ok := conns[aconn]; !ok {
		return
	}

	conns[aconn].WriteMsg(msg)
}
