package net

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"

	"xianxia/common/dbengine"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/game/global/db"
	"xianxia/servers/game/msg"
	"xianxia/servers/game/conf"
)

const LOGIN_ACCESS_TOKEN_LEN = 32
const (
	DB_RET_TYPE_ACCESSTOKEN = 1 + iota
	DB_RET_TYPE_PLAYERDATA
	DB_RET_TYPE_GET_DEVICE_ACCOUNT
	DB_RET_TYPE_SET_DEVICE_ACCOUNT
)

const (
	Login_Type_Device = 1
)

type CConnection struct {
	closed  bool
	agent     gate.Agent
	master    global.Player
	dbRequest bool
	dbId      int64
	deviceType string
	openId     string
	accessToken []byte
	loginType int
}

//发一条握手消息
func (conn *CConnection) OnAccept() {
	m := &msg.GSCL_Hi{}
	conn.Send(m)
}

func (conn *CConnection) IsClosed() bool {
	return conn.closed
}

func (conn *CConnection) OnRecv(data []byte) {
	if data == nil {
		return
	}

	//判断server是否成功启动
	if global.ServerG.GetState() != global.Server_State_Started {
		conn.Close()
		return
	}

	if conn.master == nil { //登录校验
		if conn.dbRequest {
			return
		}

		if len(data) < 4 {
			return
		}

		deviceTypeLen := conf.RdWrEndian.Uint16(data)
		openIdLen := conf.RdWrEndian.Uint16(data[2:])
		if len(data) != 4 + int(deviceTypeLen) + int(openIdLen) + LOGIN_ACCESS_TOKEN_LEN {
			log.Error("Connection AccessToken Lenght error len:%d", len(data))
			return
		}

		conn.deviceType = string(data[4:4+deviceTypeLen])
		conn.openId = string(data[4+deviceTypeLen:4+deviceTypeLen + openIdLen])
		conn.accessToken = data[4+deviceTypeLen + openIdLen:]

		conn.dbRequest = true
		global.ServerG.GetDBEngine().Request(conn, DB_RET_TYPE_ACCESSTOKEN, int64(0),"get", fmt.Sprintf("loginToken:%s", conn.openId))
	} else {
		if len(data) < 4 {
			return
		}
		moduleId := msg.RdWrEndian.Uint32(data)
		global.ServerG.GetMsgDispatcher().Dispatch(moduleId, conn.master, data[4:])
	}
}

func (conn *CConnection) Send(msg global.NetMessage) {
	conn.agent.WriteMsg(msg.MakeBuffer())
}

func (conn *CConnection) OnClose() {
	conn.closed = true
	if conn.master != nil {

		conn.master.Offline()
		conn.master = nil
	}
}

func (conn *CConnection) Close() {
	conn.agent.Close()
}

func (conn *CConnection) Kick() {
	conn.agent.Close()

	conn.closed = true
	if conn.master != nil {

		conn.master.Offline()
		conn.master = nil
	}
}

//db返回
func (conn *CConnection) OnRet(ret *dbengine.CDBRet) {
	if ret == nil || ret.Err != nil {
		conn.dbRequest = false
		log.Error("OnRet:%v", ret.Err)
		return
	}

	//連接已關閉
	if conn.IsClosed() {
		return
	}

	switch ret.OpType {
	case DB_RET_TYPE_ACCESSTOKEN: //获取deviceID
		if ret.Content == nil {
			m := &msg.GSCL_PlayerLoginTokenExpired{}

			conn.Send(m)
			conn.dbRequest = false
			return
		}
		values, err := redis.String(ret.Content, nil)
		if err == redis.ErrNil {
			log.Error("CConnection redis.String error:%v", err)
			conn.dbRequest = false
			return
		}

		abData := db.DB_AccountBase_Data{}
		err = json.Unmarshal([]byte(values), &abData)
		if err != nil {
			log.Error("CConnection json.Unmarshal DB_AccountBase_Data error:%v", err)
			conn.dbRequest = false
			return
		}

		conn.loginType = abData.LoginType
		if abData.LoginType == Login_Type_Device { //设备登录
			daData := db.DB_DeviceToken_Data{}
			err = json.Unmarshal([]byte(values), &daData)
			if err != nil {
				log.Error("CConnection json.Unmarshal DB_DeviceToken_Data error:%v", err)
				conn.dbRequest = false
				return
			}

			if daData.Token != string(conn.accessToken) {
				log.Error("CConnectiondaData.Token != string(conn.accessToken)")
				return
			}

			//设备登录的openid 就是deviceId
			conn.openId = fmt.Sprintf("%d", daData.DeviceId)

			global.ServerG.GetDBEngine().Request(conn, DB_RET_TYPE_GET_DEVICE_ACCOUNT, int64(0),"get", fmt.Sprintf("account:%d", daData.DeviceId))
		}
	case DB_RET_TYPE_GET_DEVICE_ACCOUNT: //获取playerDBID
		if ret.Content == nil { //无角色就创建
			conn.dbRequest = false
			player, err := global.ServerG.GetObjectMgr().CreatePlayer(conn)
			if err != nil {
				log.Error("CreatePlayer error:%v", err)
			} else {
				global.ServerG.GetDBEngine().Request(conn, DB_RET_TYPE_SET_DEVICE_ACCOUNT, int64(0),"set", fmt.Sprintf("account:%s", conn.openId), player.GetDBId())
				//绑定角色
				conn.master = player
				player.SetOpenId(conn.openId)
				player.SetDeviceType(conn.deviceType)
				player.SetLoginType(conn.loginType)

				player.Online(true)
			}
		} else {
			dbId, err := redis.Int64(ret.Content, nil)
			if err != nil {
				m := &msg.GSCL_Error{
					Desc: []byte("获取账号错误"),
				}

				conn.Send(m)
				return
			}

			//判断是否是换设备
			player := global.ServerG.GetObjectMgr().GetPlayer(dbId)
			if player != nil {
				if conn.loginType == Login_Type_Device {
					log.Error("CConnection Login_Type_Device player:%d 重复登录", dbId)

					m := &msg.GSCL_Error{
						Desc: []byte("重复登录"),
					}

					conn.Send(m)
					return
				} else {
					//todo kickPlayer
				}
			}

			global.ServerG.GetDBEngine().Request(conn, DB_RET_TYPE_PLAYERDATA, int64(0),"hgetall", fmt.Sprintf("player:%d", dbId))
		}

	case DB_RET_TYPE_PLAYERDATA:
		conn.dbRequest = false
		values, err := redis.Values(ret.Content, nil)
		if err == nil {
			dst := new(db.DB_PLayer_Props)
			err = redis.ScanStruct(values, dst)
			if err != nil || dst.DBId == 0 { //解析错误
				log.Error("struct DB_PLayer_Props error:%v", err)
			} else {
				if dst.Lock {
					m := &msg.GSCL_Error{
						Desc: []byte("已被封号"),
					}
					conn.Send(m)
					return
				}

				player, err := global.ServerG.GetObjectMgr().CreatePlayerFromDB(dst.DBId, dst, conn)
				if err != nil {
					log.Error("CreatePlayerFromDB error:%v", err)
				} else {
					//绑定角色
					conn.master = player
					player.SetOpenId(conn.openId)
					player.SetDeviceType(conn.deviceType)
					player.SetLoginType(conn.loginType)
				}
			}
		} else {
			log.Error("CConnection DB_RET_TYPE_PLAYERDATA Error:%v", err)
		}
	}
}

func CreateConnection(agent gate.Agent) global.Connection {
	return &CConnection{
		agent: agent,
	}
}
