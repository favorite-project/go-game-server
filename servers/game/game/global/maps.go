package global

type Map interface {
	Singleton
	GetPlayerList() []int64
	GetRegionMonsters(int32, bool) []Monster
	IsInstance() bool
	GetInstanceMonsters(fightIndex int) (mons []Monster, bBoss bool)
}

type MapMgr interface {
	Singleton

	Create() bool
	Stop() bool

	GetMap(int32) Map

	NextMap(Player) bool
	ChangeMap(Player, int32, int32) error
}
