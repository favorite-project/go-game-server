package db

type DB_PLayer_Props struct {

	Level    int32 `redis:"level"`
	Money    int32 `redis:"money"`
	Diamond  int32 `redis:"diamond"`
	VipLevel int32 `redis:"viplevel"`
	RechargeNum int32 `redis:"rechargeNum"` //累计充值的金额，可用来折算成vip
	FightVal int32 `redis:"fightVal"`
	Name           []byte `redis:"name"` //global.NAME_MAX_LEN
	CreateTime     int32  `redis:"createTime"`
	DBId           int64  `redis:"dbid"`
	MapId          int32  `redis:"mapid"`
	RegionId       int32  `redis:"regionid"`
	Occupation     int32  `redis:"occupation"`
	Sex            int32  `redis:"sex"`
	Exp            int32  `redis:"exp"`
	Pic            []byte `redis:"pic"`
	FreeFightCount int32  `redis:"freefightcount"`
	MaxMapId       int32  `redis:"maxmapid"`
	MaxRegionId    int32  `redis:"maxregionid"`
	EquipReslove   int32  `redis:"equipreslove"`
	SkillPoint     int32  `redis:"skillpoint"`
	LastOnlineTime int32 `redis:"lastonlinetime"`
	LastOffTimeTime    int32 `redis:"lastofflinetime"`
	OpenId 			string `redis:"openid"`
	DeviceType 		string `redis:"devicetype"`
	LoginType 		int `redis:"logintype"`
	QuickFightCount int32 `redis:"quickFightCount"`
	QuickFightTime  int32 `redis:"quickFightTime"`
	VipExp          int32 `redis:"vipExp"`
	Lock 			bool `redis:"lock"`
	AdvanceLevel     int32 `redis:"advanceLevel"`
	AdvanceExp     int32 `redis:"advanceExp"`
	VipRewardTime     int32 `redis:"vipRewardTime"`
}
