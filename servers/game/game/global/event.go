package global

const (
	Event_Type_PlayerOnline = 1 + iota
	Event_Type_PlayerOffline
	Event_Type_ChangeMap
	Event_Type_FightOver
)

type ChangeMap_Event_Info struct {
	Player    int64
	OMapId    int32
	ORegionId int32
	MapId     int32
	RegionId  int32
}

type Fight_Event_Info struct {
	Mode       uint32
	Master     int64
	Attackers  []Creature
	Defencers  []Creature
	FightRound int
	Win        bool
	BBoss      bool
}
