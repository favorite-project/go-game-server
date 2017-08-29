package object

import (
	"errors"
	"time"
	"xianxia/common/event"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/game/global/db"
	"github.com/name5566/leaf/log"
	"github.com/garyburd/redigo/redis"
	"fmt"
	"encoding/json"
)

//生物管理器
type MapCreature map[int64]global.Creature
type CObjectManager struct {
	Players       MapCreature   //所有玩家，来自登录
	OnlinePlayers map[int64]int //在线
	curUUid       int64
	mOfflinePlayes  map[int64]global.OffLinePlayer //在线
}

var ObjectMgr *CObjectManager

//初始化
func init() {
	ObjectMgr = &CObjectManager{
		Players:       make(MapCreature),
		OnlinePlayers: make(map[int64]int),
		mOfflinePlayes:make(map[int64]global.OffLinePlayer),
	}

}

func (mgr *CObjectManager) Create() bool {
	//注册消息
	global.ServerG.GetMsgDispatcher().Register(global.Message_RootKey_Player, mgr)

	//注册事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOnline, mgr)
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOffline, mgr)
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_FightOver, mgr)

	return true
}

func (mgr *CObjectManager) Stop() bool {
	global.ServerG.GetMsgDispatcher().UnRegister(global.Message_RootKey_Player)

	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOnline, mgr)
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOffline, mgr)
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_FightOver, mgr)

	return true
}

func (mgr *CObjectManager) GenerateUUid() int64 {
	return int64(conf.Server.PlatformId*1000000000+conf.Server.ServerID*100000000) + mgr.curUUid
}

//统一循环调度
func (mgr *CObjectManager) Update(now time.Time, elspNanoSecond int64) {
	for _, c := range mgr.Players {
		if player, ok := c.(global.Player); ok {
			player.Update(now, elspNanoSecond)
		}
	}
}

func (mgr *CObjectManager) GetPlayer(id int64) global.Player {
	player, ok := mgr.Players[id]
	if !ok {
		return nil
	}

	p2, ok := player.(global.Player) //转换不会改变地址
	if !ok {
		return nil
	}

	return p2
}

func (mgr *CObjectManager) CreatePlayerFromDB(id int64, dbData *db.DB_PLayer_Props, conn global.Connection) (global.Player, error) {
	if dbData == nil {
		return nil, errors.New("dbData nil")
	}

	p, err := mgr.Players[id]
	if err {
		return p.(global.Player), nil
	}

	p2 := &player{
		dbId: id,
		conn: conn,
	}

	p2.initFromDB(dbData)

	mgr.Players[id] = p2

	return p2, nil
}

func (mgr *CObjectManager) CreatePlayer(conn global.Connection) (global.Player, error) {
	p2 := &player{
		conn: conn,
	}

	err := p2.create()
	if err != nil {
		return nil, err
	}

	mgr.Players[p2.GetDBId()] = p2

	return p2, nil
}

func (mgr *CObjectManager) OnRecv(obj interface{}, recvData []byte) {
	player, ok := obj.(global.Player)
	if !ok {
		return
	}

	player.OnRecv(recvData)
}

func (mgr *CObjectManager) OnEvent(ev *event.CEvent) {
	if ev == nil {
		return
	}

	switch ev.Type {
	case global.Event_Type_PlayerOnline:
		if ev.Obj == nil {
			return
		}

		p, ok := ev.Obj.(global.Player)
		if !ok {
			return
		}

		mgr.OnlinePlayers[p.GetDBId()] = 1

		//上线了删除离线数据
		if _, ok := mgr.mOfflinePlayes[p.GetDBId()]; ok {
			delete(mgr.mOfflinePlayes, p.GetDBId())
		}

	case global.Event_Type_PlayerOffline:

		if ev.Obj == nil {
			return
		}

		p, ok := ev.Obj.(global.Player)
		if !ok {
			return
		}

		dbId := p.GetDBId()
		if _, ok := mgr.OnlinePlayers[dbId]; ok {
			delete(mgr.OnlinePlayers,dbId)
		}

		if _, ok := mgr.Players[dbId]; ok {
			delete(mgr.Players, dbId)
		}

	case global.Event_Type_FightOver:
		if ev.Content == nil {
			return
		}

		fightOverInfo, ok := ev.Content.(*global.Fight_Event_Info)
		if !ok {
			return
		}

		for _, c := range fightOverInfo.Attackers {
			if c == nil {
				continue
			}

			c.OnFightEvent(true, fightOverInfo)
		}

		for _, c := range fightOverInfo.Defencers {
			if c == nil {
				continue
			}

			c.OnFightEvent(false, fightOverInfo)
		}
	}
}

func (mgr *CObjectManager) PlayerIsOnline(dbId int64) bool {
	_, ok := mgr.OnlinePlayers[dbId]
	return ok
}

func (mgr *CObjectManager) SetPlayerOnline(dbId int64) {
	_, ok := mgr.OnlinePlayers[dbId]
	if !ok {
		mgr.OnlinePlayers[dbId] = 1
	}
}

func (mgr *CObjectManager) CreateMonster(cfgData *global.MonsterCfg) global.Monster {
	if cfgData == nil {
		return nil
	}

	mon := &monster{}

	mon.initProps(cfgData)

	return mon
}

func(mgr *CObjectManager) GetOnlinePlayer() []int64 {
	onlineDbIds := make([]int64, len(mgr.OnlinePlayers))
	i:= 0
	for dbid, _:= range mgr.OnlinePlayers {
		onlineDbIds[i] = dbid
		i++
	}

	return onlineDbIds
}

func(mgr *CObjectManager) GetOfflinePlayer(dbid int64) global.OffLinePlayer {

	if op, ok := mgr.mOfflinePlayes[dbid]; ok {
		return op
	}

	//同步从db里拉取
	conn := global.ServerG.GetDBEngine().Redis.Get()
	if conn == nil {
		log.Error("redis GetOfflinePlayer %d redis.Get Conn nil", dbid)
		return nil
	}

	now :=  int32(time.Now().Unix())
	//角色基本属性信息
	values, err := redis.Values(conn.Do("hgetall", fmt.Sprintf("player:%d", dbid)))
	if err != nil {
		log.Error("GetOfflinePlayer get redis player %d error:%v", dbid, err)
		conn.Close()
		return nil
	}

	dst := new(db.DB_PLayer_Props)
	err = redis.ScanStruct(values, dst)
	if err != nil { //解析错误
		log.Error("GetOfflinePlayer struct DB_PLayer_Props error:%v", err)
		conn.Close()
		return nil
	}

	op := &offlineplayer{}
	op.setProps(dst)

	//角色装备信息
	equips, err := redis.String(conn.Do("get", fmt.Sprintf("equip:%d", dbid)))
	if err != nil && err != redis.ErrNil {
		log.Error("GetOfflinePlayer get redis player %d equip error:%v", dbid, err)
		conn.Close()
		return nil
	}

	var equipData *global.EquipDBData = nil
	if err != redis.ErrNil {
		equipData = &global.EquipDBData {
			EquipData:make(map[int16]*global.ItemDBData),
		}

		err = json.Unmarshal([]byte(equips), equipData)
		if err != nil {
			log.Error("GetOfflinePlayer get redis player %d equip json error:%v", dbid, err)
			conn.Close()
			return nil
		}
	}
	op.setEquips(equipData)

	//角色技能信息
	skills, err := redis.String(conn.Do("get", fmt.Sprintf("skill:%d", dbid)))
	if err != nil && err != redis.ErrNil {
		log.Error("GetOfflinePlayer get redis player %d skill error:%v", dbid, err)
		conn.Close()
		return nil
	}

	var skillData *global.SkillDBData = nil
	if err != redis.ErrNil {
		skillData = &global.SkillDBData {
			Equips:make(map[int32]*global.SkillDBItem),
			Bags:make(map[int32]*global.SkillDBItem),
		}

		err = json.Unmarshal([]byte(skills), skillData)
		if err != nil {
			log.Error("GetOfflinePlayer get redis player %d skill json error:%v", dbid, err)
			conn.Close()
			return nil
		}
	}
	op.setSkills(skillData)

	op.loadTime = now

	mgr.mOfflinePlayes[dbid] = op

	conn.Close()
	return op

}