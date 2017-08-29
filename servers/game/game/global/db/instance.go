package db

type Player_InstanceDB_Data struct {
	DbId int64
	MFreeCount map[int32]int32 //免费次数
}