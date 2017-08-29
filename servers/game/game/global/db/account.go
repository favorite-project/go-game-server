package db

type DB_AccountBase_Data struct {
	LoginType int `redis:"loginType"`
}

type DB_DeviceToken_Data struct {
	LoginType int `redis:"loginType"`
	Token string `redis:"token"`
	DeviceId int `redis:"deviceId"`

}