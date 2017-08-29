package object

import (
	"github.com/name5566/leaf/log"
	"strconv"
	"strings"
	"xianxia/servers/game/game/global"
)

type monster struct {
	creature
	props [global.Monster_Prop_Max - global.Monster_Prop_Level]int32
}

func (m *monster) initProps(cfgData *global.MonsterCfg) bool {
	if cfgData == nil {
		return false
	}

	m.bitflags = make(map[int]byte)

	m.creature.cfgId = cfgData.Id

	m.initBase()

	//设置属性
	m.creature.props[global.Creature_Prop_Two_Defence] = cfgData.Defence
	m.creature.props[global.Creature_Prop_Two_Attack] = cfgData.Attack
	m.creature.props[global.Creature_Prop_Two_Blood] = cfgData.Blood
	m.creature.props[global.Creature_Prop_Two_AttackSpeed] = cfgData.Attackspeed
	m.creature.props[global.Creature_Prop_Two_Miss] = cfgData.Miss
	m.creature.props[global.Creature_Prop_Two_Hit] = cfgData.Hit
	m.creature.props[global.Creature_Prop_Two_Crit] = cfgData.Crit
	m.creature.props[global.Creature_Prop_Two_Tenacity] = cfgData.Tenacity

	//基本数据
	m.creature.cType = global.Creature_Type_Monster
	copy(m.creature.name[:len(m.creature.name)], []byte(cfgData.Name))

	//怪物属性
	m.props[global.Monster_Prop_Level-global.Monster_Prop_Level] = cfgData.Level
	m.props[global.Monster_Prop_Quality-global.Monster_Prop_Level] = cfgData.Quality

	//技能
	if len(cfgData.Skills) > 0 {
		m.SkillDBData = &global.SkillDBData{
			Equips:make(map[int32]*global.SkillDBItem),
		}

		skillStrArr := strings.Split(cfgData.Skills, "#")
		for i, skillIdStr := range skillStrArr {
			skillId, err := strconv.Atoi(skillIdStr)
			if err != nil {
				log.Error("monster::initProps monsterid:%d skill strconvAtoi(%s) Error:%s", cfgData.Id, skillIdStr, err)
				continue
			}

			m.SkillDBData.Equips[int32(skillId)] = &global.SkillDBItem{
				SkillId: int32(skillId),
				Pos:int32(i),
			}
		}
	}

	//todo 计算出二级属性
	//m.caculateProps(true)

	return true
}

func (m *monster) GetProp(index int) (int32, bool) {
	value, suc := m.creature.GetProp(index)
	if suc {
		return value, suc
	}

	if index >= global.Monster_Prop_Max || index < global.Monster_Prop_Level {
		return 0, false
	}

	return m.props[index-global.Monster_Prop_Level], true
}

func (m *monster) SetProp(index int, value int32, bAdd bool) (int32, bool) {
	v, suc := m.creature.SetProp(index, value, bAdd)
	if suc {
		return v, suc
	}

	if index >= global.Monster_Prop_Max || index < global.Monster_Prop_Level {
		return 0, false
	}

	if bAdd {
		m.props[index-global.Monster_Prop_Level] += value
	} else {
		m.props[index-global.Monster_Prop_Level] = value
	}

	return m.props[index-global.Monster_Prop_Level], true
}

func (m *monster) GetSkillData() *global.SkillDBData {
	return m.SkillDBData
}
