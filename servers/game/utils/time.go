package utils

import (
	"time"
	//"fmt"
	"errors"
)

func GetTodayEndUnix() int64 {
	now := time.Now()
	year, month, day:= now.Date()
	//tomorrow := now.Add(24*time.Hour)
	//tomorrow_str := fmt.Sprintf("%d-%d-%d 00:00:00", year, month, tomorrow.Day())
	today_end := time.Date(year, month, day, 23, 59, 59, 0, time.Local).Unix()
	return today_end
}


func GetTodayEndUnixInHour(now *time.Time, hour int) (int64, error) {
	if hour < 0 || hour >= 24 {
		return int64(0), errors.New("hour error")
	}

	if now == nil {
		*now = time.Now()
	}

	sec := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local).Unix()

	nowHour := now.Hour()
	if nowHour >= hour {
		sec += 24 * 3600
	}

	return sec, nil
}

func CheckIsSameDay(time1 *time.Time, time2 *time.Time, hour int) bool {
	if time1 == nil || time2 == nil {
		return false
	}

	t1EndSec, _ := GetTodayEndUnixInHour(time1, hour)
	t2EndSec, _ := GetTodayEndUnixInHour(time2, hour)

	return t1EndSec == t2EndSec
}

func CheckIsSameDayBySec(time1 int64, time2 int64, hour int) bool {
	t1 := time.Unix(time1, int64(0))
	t2 := time.Unix(time2, int64(0))
	return CheckIsSameDay(&t1, &t2, hour)
}
