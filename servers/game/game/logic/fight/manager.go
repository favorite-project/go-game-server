package fight

import (
	_ "fmt"
	"github.com/name5566/leaf/log"
	"sort"
	"xianxia/servers/game/game/global"
)

//根据出手速度排序从大到小排序
type FighterSlice []global.IFightObject

func (s FighterSlice) Len() int {
	return len(s)
}

func (s FighterSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s FighterSlice) Less(i, j int) bool {
	return s[i].GetFightProp(global.Creature_Prop_Two_AttackSpeed) > s[j].GetFightProp(global.Creature_Prop_Two_AttackSpeed)
}

type CFightMgr struct {
}

var FightMgr *CFightMgr

func init() {
	FightMgr = &CFightMgr{}
}

func (mgr *CFightMgr) DoFight(master global.Player, attackers []global.Creature, defencers []global.Creature, fightMode uint32, bBoss bool) *global.FightResultData {

	if attackers == nil || defencers == nil {
		return nil
	}

	resultData := &global.FightResultData{
		Attackers: make(FighterSlice, len(attackers)),
		Defenders: make(FighterSlice, len(defencers)),
		AttackWin: false,
		BBoss:     bBoss,
		Items:     []global.IFightItemData{},
	}

	for i, c := range attackers {
		fo := mgr.createFightObjectFromCreature(c, master)
		if fo == nil {
			log.Error("CFightMgr::DoFight attacker empty")
			return nil
		}

		fo.pos = int8(i)
		resultData.Attackers[i] = fo
	}

	for i, c := range defencers {
		fo := mgr.createFightObjectFromCreature(c, master)
		if fo == nil {
			log.Error("CFightMgr::DoFight defencer empty")
			return nil
		}

		fo.pos = int8(i + 100)
		resultData.Defenders[i] = fo
	}

	attackObjs := FighterSlice(resultData.Attackers)
	defendObjs := FighterSlice(resultData.Defenders)
	fightObjs := FighterSlice(append(attackObjs, defendObjs...))

	//按照出手速度排序
	sort.Sort(attackObjs)
	sort.Sort(defendObjs)
	sort.Sort(fightObjs)

	skillMgr := global.ServerG.GetSkillMgr()
	fightRound := 0
	for curRound := int8(1); curRound < global.MAX_FIGHT_ROUNT; curRound++ {
		for _, attacker := range fightObjs {
			//技能释放
			if !attacker.CanAttack() {
				continue
			}

			skillId := int32(0)
			if attacker.CanUseSkill() {
				skillId = attacker.GetNowSkill()
			}

			sItemArr := skillMgr.Logic(skillId, attacker, attackObjs, defendObjs)
			if (sItemArr == nil || len(sItemArr) == 0) && skillId != 0 { //普通攻击
				skillId = 0
				sItemArr = skillMgr.Logic(0, attacker, attackObjs, defendObjs)
			}

			if sItemArr == nil || len(sItemArr) == 0 { //回合没输出，出现bug
				//log.Error("FightMgr::DoFight round:%d empty logicResult:%p", curRound, sItemArr)
				//return nil

				//一方已经死完了
				break
			}

			attacker.ResetNowSkill(skillId)

			fiItem := &global.FightItemData_Attack{
				FightItemData_Base: global.FightItemData_Base{
					CurRound:   curRound,
					IType:      global.FIGHT_ITEM_ATTACK,
					EventDatas: sItemArr,
				},
				AttackerPos: attacker.GetPos(),
				SkillId:     skillId,
			}

			resultData.Items = append(resultData.Items, fiItem)
			fightRound++
		}

		//round Update
		allFiBArr := []global.IFightEventData{}
		for _, fighter := range fightObjs {
			rFiBItems := fighter.Update(curRound)
			if rFiBItems != nil {
				allFiBArr = append(allFiBArr, rFiBItems...)
			}
		}

		if len(allFiBArr) > 0 {
			fiItem := &global.FightItemData_Buff{
				FightItemData_Base: global.FightItemData_Base{
					CurRound:   curRound,
					IType:      global.FIGHT_ITEM_BUFF,
					EventDatas: allFiBArr,
				},
			}
			resultData.Items = append(resultData.Items, fiItem)
		}

		//检查战斗是否结束
		allDead := true
		for _, attacker := range attackObjs {
			if !attacker.IsDead() {
				allDead = false
				break
			}
		}

		//todo可以根据先手速度判断谁先死
		if !allDead {
			allDead = true
			for _, defender := range defendObjs {
				if !defender.IsDead() {
					allDead = false
					break
				}
			}

			if allDead {
				resultData.AttackWin = true
				break
			}
		} else {
			break
		}
	}

	fightOverEventInfo := &global.Fight_Event_Info{
		Mode:       fightMode,
		Master:     master.GetDBId(),
		Attackers:  attackers,
		Defencers:  defencers,
		FightRound: fightRound,
		Win:        resultData.AttackWin,
		BBoss:      bBoss,
	}

	//抛出战斗事件
	global.ServerG.GetEventRouter().DoEvent(global.Event_Type_FightOver, master, fightOverEventInfo)

	return resultData
}

func (mgr *CFightMgr) DoRoundAttack(attacker global.IFightObject, defender global.IFightObject) global.IFightEventData {
	if attacker == nil || defender == nil {
		return nil
	}

	attackBlood := int32(0)
	//计算命中闪避
	hitValue := attacker.GetFightProp(global.Creature_Prop_Two_Hit)
	missValue := defender.GetFightProp(global.Creature_Prop_Two_Miss)
	hitValue += 10000
	hitValue -= missValue

	/*
	   FIGHT_EVENT_SKILL_NORMAL = int8(1) + iota //普通伤害
	   	FIGHT_EVENT_SKILL_NOHIT                   //未命中
	   	FIGHT_EVENT_SKILL_CRIT                    //暴击
	*/
	feItem := &global.FightEventData_Skill{
		FightEventData_Base: global.FightEventData_Base{
			EType: global.FIGHT_EVENT_SKILL_NORMAL,
			Pos:   defender.GetPos(),
		},
		ChangeProps: make(map[int32]int32),
	}

	randAttack := global.ServerG.GetRandSrc()
	if randAttack.Int31n(10000)+1 > hitValue { //未命中
		feItem.EType = global.FIGHT_EVENT_SKILL_NOHIT
	} else {
		//计算攻击防御
		attackValue := attacker.GetFightProp(global.Creature_Prop_Two_Attack)
		defenceValue := defender.GetFightProp(global.Creature_Prop_Two_Defence)

		//计算暴击抗暴
		critValue := attacker.GetFightProp(global.Creature_Prop_Two_Crit)
		tenacityValue := defender.GetFightProp(global.Creature_Prop_Two_Tenacity)

		//最终伤害
		attackBlood = int32(attackValue * attackValue / (attackValue + defenceValue))

		critValue -= tenacityValue
		crited := true
		if randAttack.Int31n(10000)+1 > critValue {
			crited = false
		}

		//做伤害偏移
		randAValue := randAttack.Int31n(int32(attackBlood/20) + 1)
		randAdd := randAttack.Intn(2)
		if randAdd == 1 {
			attackBlood += randAValue
		} else {
			attackBlood -= randAValue
		}

		//暴击
		if crited {
			attackBlood = int32(attackBlood * global.Crit_Per / 1000)
			feItem.EType = global.FIGHT_EVENT_SKILL_CRIT
		}

		//计算单次伤害百分比
		attackBlood = int32(attackBlood * attacker.GetFightProp(global.Creature_Prop_Two_BaseGain) / 1000)

		//额外伤害
		attackBlood += attacker.GetFightProp(global.Creature_Prop_Two_FAAdd)

		//最终伤害增益
		attackBlood = int32(attackBlood * attacker.GetFightProp(global.Creature_Prop_Two_FAGain) / 1000)

		defender.SetBlood(defender.GetBlood() - attackBlood)

		feItem.ChangeProps[int32(global.Creature_Prop_Two_Blood)] = -attackBlood
		//fmt.Println("attackInfo:", attackBlood, attackValue, defenceValue, attacker.GetFightProp(global.Creature_Prop_Two_FAAdd))
	}

	return feItem
}

func (mgr *CFightMgr) createFightObjectFromCreature(c global.Creature, master global.Player) *fightObject {
	if c == nil {
		return nil
	}

	fo := &fightObject{
		object: c,
		master: master,
	}

	fo.initFightProps() //战斗属性
	fo.initSkillData()  //技能数据

	return fo
}
