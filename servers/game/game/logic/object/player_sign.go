package object

import (
	"github.com/name5566/leaf/log"
	"xianxia/servers/game/msg"
	"time"
	"xianxia/servers/game/utils"
	"xianxia/servers/game/game/global"
	"xianxia/servers/game/conf"
	"fmt"
)


func(p *player) sendSignInfo() {
	info, err := p.playerInfo.GetPlayerSignInfo(p.dbId)
	if err != nil {
		log.Error("player::sendSignInfo %d GetSignInfo nil", p.dbId)
		return
	}

	m := &msg.GSCL_SignInfo{
		PlayerSignInfo:info,
	}

	p.conn.Send(m)
}

func(p *player) msg_sign(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	info, err := p.playerInfo.GetPlayerSignInfo(p.dbId)
	if err != nil {
		log.Error("player::msg_sign %d GetSignInfo nil", p.dbId)
		return
	}

	day := int32(conf.RdWrEndian.Uint32(recvData))
	autoSell := true
	var items *global.RewardData

	now := int32(time.Now().Unix())
	if day == 0 {
		//签到
		if utils.CheckIsSameDayBySec(int64(now), int64(info.SignTime), 0) {
			return
		}

	} else {
		//累计签到奖励
		for _, s := range info.RewardStates {
			if day == s {
				return
			}
		}

	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("SignReward", day)
	if cfg == nil {
		fmt.Println(11111111111)
		return
	}

	items = getRewardFromStr(cfg.(*global.SignRewardCfg).Items)
	if items == nil {
		fmt.Println(222222222)
		return
	}

	if p.IsBagFullMulti(items, autoSell) {
		fmt.Println(33333333333)
		return
	}

	if day==0 {
		info.SignTime = now
		info.SignCnt++
	} else {
		info.RewardStates = append(info.RewardStates, day)
	}

	p.playerInfo.SetPlayerSignInfo(p.dbId, info)
	fmt.Println("sign_suc")

	_, _, sellInfo, _ := p.AddItems(items, true, autoSell)
	m := &msg.GSCL_SignReward{
		Day:day,
		Reward:&msg.CommonMsg_Reward {
			Items:items,
			Sells:sellInfo,
		},
	}

	p.conn.Send(m)
}

func(p *player) sendLoginInfo() {
	info, err := p.playerInfo.GetPlayerLoginActInfo(p.dbId)
	if err != nil {
		log.Error("player::sendLoginInfo %d GetPlayerLoginActInfo nil", p.dbId)
		return
	}

	now := int32(time.Now().Unix())
	if !utils.CheckIsSameDayBySec(int64(now), int64(info.LoginTime), 0) {
		info.TodayReward = false
		info.LoginTime = now
		info.LoginCnt++
		p.playerInfo.SetPlayerLoginActInfo(p.dbId, info)
	}

	m := &msg.GSCL_LoginInfo{
		PlayerLoginInfo:info,
	}

	p.conn.Send(m)
}

func(p *player) msg_login(recvData []byte) {
	if len(recvData) < 4 {
		return
	}

	info, err := p.playerInfo.GetPlayerLoginActInfo(p.dbId)
	if err != nil {
		log.Error("player::msg_login %d GetPlayerLoginActInfo nil", p.dbId)
		return
	}

	day := int32(conf.RdWrEndian.Uint32(recvData))
	autoSell := true
	var items *global.RewardData

	now := int32(time.Now().Unix())
	if !utils.CheckIsSameDayBySec(int64(now), int64(info.LoginTime), 0) {
		info.TodayReward = false
		info.LoginTime = now
		info.LoginCnt++
	}

	if day == 0 {
		//签到
		if info.TodayReward {
			return
		}

	} else {
		//累计签到奖励
		for _, s := range info.RewardStates {
			if day == s {
				return
			}
		}

	}

	cfg := global.ServerG.GetConfigMgr().GetCfg("LoginReward", day)
	if cfg == nil {
		return
	}

	items = getRewardFromStr(cfg.(*global.LoginRewardCfg).Items)
	if items == nil {
		return
	}

	if p.IsBagFullMulti(items, autoSell) {
		return
	}

	if day==0 {
		info.TodayReward = true
	} else {
		info.RewardStates = append(info.RewardStates, day)
	}

	p.playerInfo.SetPlayerLoginActInfo(p.dbId, info)
	_, _, sellInfo, _ := p.AddItems(items, true, autoSell)
	m := &msg.GSCL_LoginReward{
		Day:day,
		Reward:&msg.CommonMsg_Reward {
			Items:items,
			Sells:sellInfo,
		},
	}

	p.conn.Send(m)
}