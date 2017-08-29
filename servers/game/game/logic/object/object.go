package object

import (
	"strconv"
	"strings"
	"time"
	"xianxia/servers/game/game/global"
)

//属性定义
type CreatueProps [global.Creature_Prop_Two_Max]int32

//生物类基类
type creature struct {
	cfgId    int32
	cType    byte
	mapId    int32
	regionId int32
	name     [global.NAME_MAX_LEN]byte
	pic      [global.PIC_MAX_LEN]byte
	props    CreatueProps
	bitflags map[int]byte
	*global.SkillDBData

	//副本战斗相关
	instanceMapId int32
	instanceFightIndex int
}

func (c *creature) initBase() {
	c.props[global.Creature_Prop_Two_FAGain] = 1000   //最终伤害增益
	c.props[global.Creature_Prop_Two_BaseGain] = 1000 //单次伤害百分比
}

func (c *creature) GetProp(index int) (int32, bool) {
	if index >= global.Creature_Prop_Two_Max || index < global.Creature_Prop_One_Power {
		return 0, false
	}

	return c.props[index], true
}

func (c *creature) SetProp(index int, value int32, bAdd bool) (int32, bool) {
	if index >= global.Creature_Prop_Two_Max || index < global.Creature_Prop_One_Power {
		return 0, false
	}

	ov := c.props[index]
	if bAdd {
		c.props[index] += value
	} else {
		c.props[index] = value
	}

	if c.bitflags != nil && ov != c.props[index] {
		c.bitflags[index] = byte(1)
	}

	if index >= global.Creature_Prop_One_Power && index < global.Creature_Prop_One_Max {
		c.caculateProps(index, c.props[index]-ov)
	}

	return c.props[index], true
}

func (c *creature) GetType() byte {
	return c.cType
}

func (c *creature) GetName() []byte {
	return c.name[:len(c.name)]
}
func (c *creature) GetPic() []byte {
	return c.pic[:len(c.pic)]
}

func (c *creature) GetMapId() (int32, int32) {
	return c.mapId, c.regionId
}

func (c *creature) SetMapId(mapId int32, regionId int32) {
	c.mapId = mapId
	c.regionId = regionId
}

func (c *creature) GetCfgId() int32 {
	return c.cfgId
}

func (c *creature) Update(time.Time, int64) {

}

func (c *creature) OnFightEvent(bAttacker bool, fev *global.Fight_Event_Info) {

}

//计算二级属性
func (c *creature) caculateProps(propOneId int, changeValue int32) {
	if propOneId < global.Creature_Prop_One_Power || propOneId >= global.Creature_Prop_One_Max || changeValue == 0 {
		return
	}

	nowValue, _ := c.GetProp(propOneId)
	oldValue := nowValue - changeValue
	if oldValue < 0 || nowValue < 0 {
		return
	}

	csv := global.ServerG.GetConfigMgr().GetCsv("PropsCaculation")
	if csv == nil {
		return
	}

	cfgArr := []*global.PropsCaculationCfg{}
	for i := 0; i < csv.NumRecord(); i++ {
		cfg, ok := csv.Record(i).(*global.PropsCaculationCfg)
		if !ok {
			continue
		}

		if cfg.PropId != propOneId {
			continue
		}

		if (oldValue < cfg.ValueMin && nowValue < cfg.ValueMin) || (oldValue > cfg.ValueMax && nowValue > cfg.ValueMax) {
			continue
		}

		cfgArr = append(cfgArr, cfg)
	}

	mapValues1 := make(map[int]float64)
	mapValues2 := make(map[int]float64)
	for _, cfg := range cfgArr {
		leftValue1 := int32(0)
		if oldValue >= cfg.ValueMin {
			leftValue1 = oldValue - cfg.ValueMin + 1
			if oldValue > cfg.ValueMax {
				leftValue1 = cfg.ValueMax - cfg.ValueMin + 1
			}
		}

		leftValue2 := int32(0)
		if nowValue >= cfg.ValueMin {
			leftValue2 = nowValue - cfg.ValueMin + 1
			if nowValue > cfg.ValueMax {
				leftValue2 = cfg.ValueMax - cfg.ValueMin + 1
			}
		}

		epropstrArr := strings.Split(cfg.EffectProps, ";")
		if epropstrArr == nil {
			continue
		}

		for _, epropstr := range epropstrArr {
			propvalue := strings.Split(epropstr, "+")
			if propvalue == nil || len(propvalue) != 2 {
				continue
			}

			pid, err := strconv.Atoi(propvalue[0])
			if err != nil {
				continue
			}

			if pid < global.Creature_Prop_Two_Attack || pid >= global.Creature_Prop_Two_Max {
				continue
			}

			v, err := strconv.ParseFloat(propvalue[1], 32)
			if err != nil {
				continue
			}

			if _, ok := mapValues1[pid]; !ok {
				mapValues1[pid] = 0
			}
			if _, ok := mapValues2[pid]; !ok {
				mapValues2[pid] = 0
			}

			mapValues1[pid] += v * float64(leftValue1) //math.Ceil(v * float64(cValue))
			mapValues2[pid] += v * float64(leftValue2) //math.Ceil(v * float64(cValue))
		}
	}

	for id, value := range mapValues1 {
		v := int32(mapValues2[id]) - int32(value)
		ov := c.props[id]
		c.props[id] += v
		if c.bitflags != nil && c.props[id] != ov {
			c.bitflags[id] = byte(1)
		}
	}
}

func (c *creature) GetSkillData() *global.SkillDBData {
	return nil
}

func (c *creature) GetInstanceMapId() int32 {
	return c.instanceMapId
}

func (c *creature)  SetInstanceMapId(instanceMapId int32) {
	c.instanceMapId = instanceMapId
}

func (c *creature) GetInstanceFightIndex() int {
	return c.instanceFightIndex
}

func (c *creature) SetInstanceFightIndex(instanceFightIndex int) {
	c.instanceFightIndex = instanceFightIndex
}