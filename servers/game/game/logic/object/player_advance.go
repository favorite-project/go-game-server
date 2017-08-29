package object

import (
	"xianxia/servers/game/game/global"
)

func(p *player) msg_advance_levelUp(recvData []byte) {
	advanceLv, _ := p.GetProp(global.Player_Prop_Advance_Level)
	advanceExp, _ := p.GetProp(global.Player_Prop_Advance_Exp)

	icfg := global.ServerG.GetConfigMgr().GetCfg("Advance", advanceLv + int32(1))
	if icfg == nil {
		return
	}

	icfg = global.ServerG.GetConfigMgr().GetCfg("Advance", advanceLv)
	if icfg == nil {
		return
	}

	cfg := icfg.(*global.AdvanceCfg)
	if advanceExp  < cfg.NeedExp {
		return
	}

	p.SetProp(global.Player_Prop_Advance_Level, int32(1), true)
	p.SetProp(global.Player_Prop_Advance_Exp, -cfg.NeedExp, true)

}

func (p *player) getAdvanceBoss() []global.Creature {
	advanceLv, _ := p.GetProp(global.Player_Prop_Advance_Level)
	icfg := global.ServerG.GetConfigMgr().GetCfg("Advance", advanceLv + int32(1))
	if icfg != nil {
		return nil
	}

	icfg = global.ServerG.GetConfigMgr().GetCfg("Advance", advanceLv)
	if icfg == nil {
		return nil
	}

	cfg := icfg.(*global.AdvanceCfg)
	playerLv, _ := p.GetProp(global.Player_Prop_Level)
	if playerLv < cfg.PlayerLevel {
		return nil
	}

	items := &global.RewardData{
		Items:make(map[int32]*global.RewardItem),
	}

	if cfg.ItemId1 != 0 {
		if p.GetBagItemNum(cfg.ItemId1) < cfg.ItemNum1 {
			return nil
		}

		items.Items[cfg.ItemId1] = &global.RewardItem{
			Id: cfg.ItemId1,
			Num:cfg.ItemNum1,
		}
	}

	if cfg.ItemId2 != 0 {
		if p.GetBagItemNum(cfg.ItemId2) < cfg.ItemNum2 {
			return nil
		}

		items.Items[cfg.ItemId2] = &global.RewardItem{
			Id: cfg.ItemId2,
			Num:cfg.ItemNum2,
		}
	}

	if cfg.ItemId3 != 0 {
		if p.GetBagItemNum(cfg.ItemId3) < cfg.ItemNum3 {
			return nil
		}

		items.Items[cfg.ItemId3] = &global.RewardItem{
			Id: cfg.ItemId3,
			Num:cfg.ItemNum3,
		}
	}

	if cfg.BossId == 0 {
		return nil
	}

	micfg := global.ServerG.GetConfigMgr().GetCfg("Monster", cfg.BossId)
	if micfg == nil {
		return nil
	}

	mon := global.ServerG.GetObjectMgr().CreateMonster(micfg.(*global.MonsterCfg))
	if mon == nil {
		return nil
	}

	mons := []global.Creature{mon}

	return mons
}