package mail
import (
	"xianxia/common/event"
	"xianxia/servers/game/game/global"
	"xianxia/common/dbengine"
	"github.com/name5566/leaf/log"
	"fmt"
	"time"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"xianxia/servers/game/msg"
)

var MailMgr *CMailMgr
func init() {
	MailMgr = &CMailMgr{
		mPlayerMails:make(map[int64]*global.PlayerMailsInfo),
	}
}

const (
	Mail_GlobalSys_DB_Get = 1
	Mail_GlobalSys_DB_Set = 2
	Mail_Player_DB_Get = 3
	Mail_Player_DB_Set = 4
)

type CMailMgr struct {
	mPlayerMails map[int64]*global.PlayerMailsInfo
	*global.GlobalSysMailsInfo
}

func (mgr *CMailMgr) Create() bool {
	//注册事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOnline, mgr) //上线事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOffline, mgr)//下线事件

	global.ServerG.GetDBEngine().Request(mgr, Mail_GlobalSys_DB_Get, 0,"get", "sysmails")
	conn := global.ServerG.GetDBEngine().Redis.Get()
	data, err := redis.String(conn.Do("get", "sysmails"))
	if err != nil && err != redis.ErrNil {
		conn.Close()
		return false
	}

	mgr.initGlobalSysMailInfo()
	if  err != redis.ErrNil {
		value, err := redis.String(data, nil)
		if err != nil {
			conn.Close()
			log.Error("CMailMgr::Create Mail_GlobalSys_DB_Get content:%s redis.String error:%s", value, err)
			return false
		}

		err = json.Unmarshal([]byte(value), mgr.GlobalSysMailsInfo)
		if err != nil {
			conn.Close()
			log.Error("CMailMgr::Create Mail_GlobalSys_DB_Get content:%s json.Unmarshal error:%s", value, err)
			return false
		}
	}

	conn.Close()

	return true
}

func (mgr *CMailMgr) Stop() bool {
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOnline, mgr)
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOffline, mgr)

	return true
}



func (mgr *CMailMgr) Update(now time.Time, elspNanoSecond int64) {
}

func (mgr *CMailMgr) OnEvent(event *event.CEvent) {
	if event == nil {
		return
	}

	if event.Obj == nil {
		return
	}

	player, ok := event.Obj.(global.Player)
	if !ok {
		return
	}

	switch(event.Type) {
	case global.Event_Type_PlayerOnline:
		global.ServerG.GetDBEngine().Request(mgr, Mail_Player_DB_Get, player.GetDBId(),"get", fmt.Sprintf("mail:%d", player.GetDBId()))
	case global.Event_Type_PlayerOffline:
		dbId := player.GetDBId()
		if _, ok := mgr.mPlayerMails[dbId];ok {
			if mgr.checkMailsExpire(dbId) {
				mgr.savePlayerDB(dbId)
			}

			delete(mgr.mPlayerMails, dbId)
		}
	}
}

func (mgr *CMailMgr) OnRet(ret *dbengine.CDBRet) {
	if ret.Err != nil {
		log.Error("CMailMgr::OnRet error:%s", ret.Err)
		return
	}

	switch(ret.OpType) {
	case Mail_Player_DB_Set:
	case Mail_Player_DB_Get:
		player := global.ServerG.GetObjectMgr().GetPlayer(ret.DBId)
		if player == nil || !player.IsOnline() {
			return
		}

		//下发到客户端
		var pMails *global.PlayerMailsInfo
		if nil == ret.Content {
			pMails = mgr.initPlayerMailInfo()
		} else {
			value, err := redis.String(ret.Content, nil)
			if err != nil {
				log.Error("CMailMgr::OnRet Mail_Player_DB_Get content:%s redis.String error:%s", ret.Content, err)
				return
			}

			pMails = &global.PlayerMailsInfo{}
			err = json.Unmarshal([]byte(value), pMails)
			if err != nil {
				log.Error("CMailMgr::OnRet Mail_Player_DB_Get content:%s json.Unmarshal error:%s", ret.Content, err)
				return
			}
		}

		mgr.mPlayerMails[ret.DBId] = pMails

		needUpdate := mgr.checkMailsExpire(ret.DBId)
		now := int32(time.Now().Unix())

		//系统邮件检查
		for id, sysmail := range mgr.GlobalSysMailsInfo.Mails {
			if now >= sysmail.ExpireTime {
				continue
			}

			found := false
			for _, pmail := range pMails.Mails {
				if pmail.SysId == id {
					found = true
					break
				}
			}

			if !found {
				newid := pMails.Id
				pMails.Id += 1
				newMail := &global.PlayerMailInfo{
					Id:newid,
					SysId:id,
					Type:global.MAIL_TYPE_GLOBAL_SYSTEM,
					Items:sysmail.Items,
					Title:sysmail.Title,
					Content:sysmail.Content,
					ExpireTime:sysmail.ExpireTime,
					State:global.MAIL_STATE_UNREAD,
					CreateTime:now,
				}
				pMails.Mails[newid] = newMail
				needUpdate = true
			}
		}

		if needUpdate {
			mgr.savePlayerDB(ret.DBId)
		}

		//下发客户端
		m := &msg.GSCL_PlayerMailsInfo {
			Mails: make(map[int32]*global.PlayerMailInfo),
		}

		for id, pmail := range pMails.Mails {
			if pmail.State == global.MAIL_STATE_READ { //领取过的系统奖励都不下发
				continue
			}

			m.Mails[id] = pmail
		}

		player.GetConnection().Send(m)
	}
}

func (mgr *CMailMgr) initGlobalSysMailInfo() {
	mgr.GlobalSysMailsInfo =  &global.GlobalSysMailsInfo{
		Mails:make(map[int32]*global.SysMailInfo),
	}
}

func (mgr *CMailMgr) initPlayerMailInfo() *global.PlayerMailsInfo {
	return &global.PlayerMailsInfo{
		Id: int32(1),
		Mails:make(map[int32]*global.PlayerMailInfo),
	}
}

func (mgr *CMailMgr) checkMailsExpire(dbId int64)  bool {
	if _, ok := mgr.mPlayerMails[dbId];!ok {
		return false
	}

	now := int32(time.Now().Unix())
	has := false
	for {
		foundId := int32(-1)
		for id, pmail := range mgr.mPlayerMails[dbId].Mails {
			//是否过期
			if now >= pmail.ExpireTime {
				foundId = id
				has = true
				break
			} else {
				if pmail.Items != nil && len(pmail.Items) != 0 { //带奖励的非全员邮件领取后一律删除
					if  pmail.SysId == 0 && pmail.State == global.MAIL_STATE_READ {
						foundId = id
						has = true
						break
					}
				}
			}
		}

		if foundId < 0 {
			break
		} else {
			delete(mgr.mPlayerMails[dbId].Mails, foundId)
		}
	}

	return has
}

func (mgr *CMailMgr) SendMail(sysId int32, mType byte, title string, content string, Items map[int32]int32, receiverDBId int64, expireTime int32) bool {
	//所有邮件都过期
	now := int32(time.Now().Unix())
	if expireTime == 0 {
		expireTime = now + global.MAIL_UNREWARD_MAX_SEC
	}

	//是否是系统邮件
	if mType == global.MAIL_TYPE_GLOBAL_SYSTEM {
		if _, ok := mgr.GlobalSysMailsInfo.Mails[sysId]; ok {
			log.Error("Mail::SendMail sysId:%d  error", sysId)
			return false
		}

		newMail := &global.SysMailInfo{
			Id:sysId,
			Items:Items,
			Title: title,
			Content: content,
			ExpireTime: expireTime,
		}

		mgr.GlobalSysMailsInfo.Mails[sysId] = newMail

		j, err := json.Marshal(mgr.GlobalSysMailsInfo)
		if err != nil {
			log.Error("Mail::SendMail sysId:%d  json.Marshal error:%s", sysId, err)
			return false
		}

		global.ServerG.GetDBEngine().Request(mgr, Mail_GlobalSys_DB_Set, 0, "set", "sysmails", j)

		//在线玩家通知
		onlinePlayes := global.ServerG.GetObjectMgr().GetOnlinePlayer()
		for _, dbid := range onlinePlayes {
			player :=  global.ServerG.GetObjectMgr().GetPlayer(dbid)
			if player == nil {
				continue
			}

			if data, ok := mgr.mPlayerMails[dbid]; ok {
				 mail := &global.PlayerMailInfo{
					Id:data.Id,
					 SysId:sysId,
					Type:mType,
					Items:Items,
					Title:title,
					Content:content,
					ExpireTime:expireTime,
					State:global.MAIL_STATE_UNREAD,
					CreateTime:now,
				}

				data.Mails[data.Id] = mail
				data.Id += 1
				mgr.savePlayerDB(dbid)

				m := &msg.GSCL_PlayerMailsInfo{
					Mails:make(map[int32]*global.PlayerMailInfo),
				}

				m.Mails[mail.Id] = mail

				player.GetConnection().Send(m)
			}
		}

		return true
	}

	newMail := &global.PlayerMailInfo{
		Id:int32(0),
		Type:mType,
		Items:Items,
		Title:title,
		Content:content,
		ExpireTime:expireTime,
		State:global.MAIL_STATE_UNREAD,
		CreateTime:int32(time.Now().Unix()),
	}

	if pMails, ok := mgr.mPlayerMails[receiverDBId];ok { //在线
		newMail.Id = int32(pMails.Id)
		pMails.Id += 1
		pMails.Mails[newMail.Id] = newMail
		mgr.savePlayerDB(receiverDBId)

		if player := global.ServerG.GetObjectMgr().GetPlayer(receiverDBId); player != nil {
			m := &msg.GSCL_PlayerMailsInfo{
				Mails:make(map[int32]*global.PlayerMailInfo),
			}
			m.Mails[newMail.Id] = newMail
			player.GetConnection().Send(m)
		}

	} else {
		conn := global.ServerG.GetDBEngine().Redis.Get()
		data, err := redis.String(conn.Do("get", fmt.Sprintf("mail:%d", receiverDBId)))
		if err != nil && err != redis.ErrNil {
			conn.Close()
			return false
		}

		pMails := mgr.initPlayerMailInfo()
		if err != redis.ErrNil {
			err = json.Unmarshal([]byte(data), pMails)
			if err != nil {
				conn.Close()
				return false
			}
		}

		newMail.Id = int32(pMails.Id)
		pMails.Id += 1

		pMails.Mails[newMail.Id] = newMail
		j, err := json.Marshal(pMails)
		if err != nil {
			conn.Close()
			return false
		}

		_, err = conn.Do("set", fmt.Sprintf("mail:%d", receiverDBId), j)
		if err != nil {
			conn.Close()
			return false
		}

		conn.Close()
	}

	return true
}

func (mgr *CMailMgr) Reward(player global.Player, id int32) {
	if player == nil {
		return
	}

	dbId := player.GetDBId()
	if _, ok := mgr.mPlayerMails[dbId];!ok {
		return
	}

	if _, ok := mgr.mPlayerMails[dbId].Mails[id];!ok {
		return
	}

	pmail :=  mgr.mPlayerMails[dbId].Mails[id]
	if pmail.Items == nil || len(pmail.Items) == 0 {
		return
	}

	m := &msg.GSCL_MailReward{
		Ret:global.MAIL_REWARD_RET_SUC,
		Id:int32(id),
	}

	if pmail.State == global.MAIL_STATE_READ {
		m.Ret = global.MAIL_REWARD_RET_REWARDED
		player.GetConnection().Send(m)
		return
	}

	now := int32(time.Now().Unix())
	//过期了
	if pmail.ExpireTime != 0 && now >= pmail.ExpireTime {
		m.Ret = global.MAIL_REWARD_RET_EXPIRED
		player.GetConnection().Send(m)
		return
	}

	rd := &global.RewardData{
		Items:make(map[int32]*global.RewardItem),
	}

	for itemId, num := range pmail.Items {
		if ri, ok := rd.Items[itemId]; ok {
			ri.Num += num
		} else {
			rd.Items[itemId] = &global.RewardItem{
				Id:itemId,
				Num:num,
			}
		}
	}

	pmail.State = global.MAIL_STATE_READ
	autoSell := false
	if player.IsBagFullMulti(rd, autoSell) {
		m.Ret = global.MAIL_REWARD_RET_BAGFULL
		player.GetConnection().Send(m)
		return
	}

	_, _, sellItems, err := player.AddItems(rd, true, autoSell)
	if err != nil {
		log.Error("邮件领取：addItems error")
		return
	}

	mSellItems := make(map[int32]int32)
	for _, sItem := range sellItems {
		mSellItems[sItem.CfgId] = sItem.Num
	}

	m.Items = pmail.Items
	m.SellItems = mSellItems
	player.GetConnection().Send(m)

	mgr.checkMailsExpire(dbId)
	mgr.savePlayerDB(dbId)
}

func (mgr *CMailMgr) savePlayerDB(dbId int64) {
	if _, ok := mgr.mPlayerMails[dbId];!ok {
		return
	}

	j, err := json.Marshal(mgr.mPlayerMails[dbId])
	if err != nil {
		log.Error("CMailMgr:::savePlayerDB %d json.Marshal error:%s", dbId, err)
		return
	}

	global.ServerG.GetDBEngine().Request(mgr, Mail_Player_DB_Set, dbId, "set", fmt.Sprintf("mail:%d", dbId), j)
}