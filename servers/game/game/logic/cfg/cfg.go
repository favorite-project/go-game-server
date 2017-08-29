package cfg

import (
	"errors"
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/recordfile"
	"reflect"
	"xianxia/servers/game/conf"
	"xianxia/servers/game/game/global"
)

type Csv struct {
	Rf   *recordfile.RecordFile
	Type reflect.Type
}

type CConfigMgr struct {
	allCsvs map[string]*Csv
}

var ConfigMgr *CConfigMgr

func init() {
	ConfigMgr = &CConfigMgr{
		allCsvs: make(map[string]*Csv),
	}

	ConfigMgr.allCsvs["Monster"] = &Csv{Type: reflect.TypeOf(global.MonsterCfg{})}
	ConfigMgr.allCsvs["Region"] = &Csv{Type: reflect.TypeOf(global.RegionCfg{})}
	ConfigMgr.allCsvs["Map"] = &Csv{Type: reflect.TypeOf(global.MapCfg{})}
	ConfigMgr.allCsvs["PropsCaculation"] = &Csv{Type: reflect.TypeOf(global.PropsCaculationCfg{})}
	ConfigMgr.allCsvs["Equipment"] = &Csv{Type: reflect.TypeOf(global.EquipmentCfg{})}
	ConfigMgr.allCsvs["Item"] = &Csv{Type: reflect.TypeOf(global.ItemCfg{})}
	ConfigMgr.allCsvs["PlayerLevel"] = &Csv{Type: reflect.TypeOf(global.PlayerLevelCfg{})}
	ConfigMgr.allCsvs["DropBox"] = &Csv{Type: reflect.TypeOf(global.DropBoxCfg{})}
	ConfigMgr.allCsvs["EquipmentProps"] = &Csv{Type: reflect.TypeOf(global.EquipmentPropsCfg{})}
	ConfigMgr.allCsvs["Skill"] = &Csv{Type: reflect.TypeOf(global.SkillCfg{})}
	ConfigMgr.allCsvs["Buff"] = &Csv{Type: reflect.TypeOf(global.BuffCfg{})}
	ConfigMgr.allCsvs["RandomStone"] = &Csv{Type: reflect.TypeOf(global.RandomStone{})}
	ConfigMgr.allCsvs["EquipCreate"] = &Csv{Type: reflect.TypeOf(global.EquipCreateCfg{})}
	ConfigMgr.allCsvs["EquipUpdate"] = &Csv{Type: reflect.TypeOf(global.EquipUpdate{})}
	ConfigMgr.allCsvs["Instance"] = &Csv{Type: reflect.TypeOf(global.InstanceCfg{})}
	ConfigMgr.allCsvs["Vip"] = &Csv{Type: reflect.TypeOf(global.VipCfg{})}
	ConfigMgr.allCsvs["Challenge"] = &Csv{Type: reflect.TypeOf(global.Challenge{})}
	ConfigMgr.allCsvs["ChallengeMonster"] = &Csv{Type: reflect.TypeOf(global.ChallengeMonster{})}
	ConfigMgr.allCsvs["Mine"] = &Csv{Type: reflect.TypeOf(global.MineCfg{})}
	ConfigMgr.allCsvs["MineWork"] = &Csv{Type: reflect.TypeOf(global.MineWorkCfg{})}
	ConfigMgr.allCsvs["Suit"] = &Csv{Type: reflect.TypeOf(global.SuitCfg{})}
	ConfigMgr.allCsvs["SignReward"] = &Csv{Type: reflect.TypeOf(global.SignRewardCfg{})}
	ConfigMgr.allCsvs["LoginReward"] = &Csv{Type: reflect.TypeOf(global.LoginRewardCfg{})}
	ConfigMgr.allCsvs["Advance"] = &Csv{Type: reflect.TypeOf(global.AdvanceCfg{})}
}

func (mgr *CConfigMgr) Start() bool {
	err := mgr.reload("Start")
	if err != nil {
		log.Error("%v", err)
		return false
	}

	return true
}

func (mgr *CConfigMgr) GetCsv(csvName string) *recordfile.RecordFile {
	csv, ok := mgr.allCsvs[csvName]
	if ok {
		return csv.Rf
	}

	return nil
}

func (mgr *CConfigMgr) GetCfg(csvName string, id interface{}) interface{} {
	csv, ok := mgr.allCsvs[csvName]
	if !ok || csv.Rf == nil {
		return nil
	}

	return csv.Rf.Index(id)
}

func (mgr *CConfigMgr) reload(opType string) error {
	for csvName, csv := range mgr.allCsvs {
		csvName = fmt.Sprintf("%s%s.csv", conf.Server.CsvDataPath, csvName)
		rf, err := recordfile.New(reflect.New(csv.Type).Elem().Interface())
		if err != nil {
			return errors.New(fmt.Sprintf("ConfigMgr %s Load %s New Error:%v", opType, csvName, err))
		}

		rf.Comma = ','
		err = rf.Read(csvName, 3)
		if err != nil {
			return errors.New(fmt.Sprintf("ConfigMgr %s Load %s Read Error:%v", opType, csvName, err))
		}

		csv.Rf = rf
	}

	return nil
}

func (mgr *CConfigMgr) Reload() {
	mgr.reload("Reload")
}

