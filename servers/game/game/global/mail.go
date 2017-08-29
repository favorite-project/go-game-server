package global

import (
	"xianxia/common/event"
	"xianxia/common/dbengine"
)

const(
	MAIL_TYPE_GLOBAL_SYSTEM = byte(1) + iota //全员系统邮件
	MAIL_TYPE_PLAYER_SYSTEM //系统单发邮件，包括系统通知
)

const(
	MAIL_STATE_UNREAD = byte(1) + iota //全员系统邮件
	MAIL_STATE_READ
)

const MAIL_UNREWARD_MAX_SEC = int32(7 * 24 * 3600) //没有领奖信息的，只保存7天的邮件

type SysMailInfo struct {
	Id int32 `json:"id"` //全员系统邮件
	Items map[int32]int32 `json:"items"`
	Title string `json:"title"`
	Content string `json:"content"`
	ExpireTime int32 `json:"expireTime"`
}

type PlayerMailInfo struct {
	Id int32 `json:"id"`
	SysId int32 `json:"sysId"` //全员系统邮件
	Type byte `json:"type"`
	Items map[int32]int32 `json:"items"`
	Title string `json:"title"`
	Content string `json:"content"`
	ExpireTime int32 `json:"expireTime"`
	CreateTime int32 `json:"createTime"`
	State byte `json:"state"`
}

type GlobalSysMailsInfo struct {
	Mails map[int32]*SysMailInfo
}

type PlayerMailsInfo struct {
	Id int32
	Mails map[int32]*PlayerMailInfo
}

const (
	MAIL_REWARD_RET_SUC = int32(0) + iota //成功
	MAIL_REWARD_RET_BAGFULL //背包满
	MAIL_REWARD_RET_EXPIRED //过期了
	MAIL_REWARD_RET_REWARDED //已领取过
)

type MailMgr interface {
	Singleton
	Create() bool
	Stop() bool
	OnEvent(event *event.CEvent)
	OnRet(ret *dbengine.CDBRet)
	SendMail(sysId int32, mType byte, title string, content string, Items map[int32]int32, receiverDBId int64, expireTime int32) bool
	Reward(player Player, id int32)
}