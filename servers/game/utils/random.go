package utils

import (
	"os"
	"math/rand"
	"encoding/binary"
	"strings"
	"strconv"
	"time"
	"fmt"
)

func randInit() {
	file, err := os.Open("/dev/urandom")
	defer file.Close()
	var seed int64
	if err != nil {
		seed = int64(time.Now().Nanosecond())
	} else {
		buf := make([]byte, 8)
		file.Read(buf)
		seed = int64(binary.LittleEndian.Uint64(buf))
	}

	rand.Seed(seed)
}


// 一百以内的概率
func HundredRandomTrue(maxVal int32) bool{
	if maxVal == 100 {
		return true
	}
	if maxVal > 100 || maxVal < 1{
		return false
	}
	randInit()
	r := int(rand.Int31n(int32(100)))
	if  r <= int(maxVal){
		return true
	}

	return false

}
// randstr 结构:   itemid+概率;itemid+概率;itemid+概率
func GetResult(total int,randstr string) (itemid int32, err error){
	randInit()
	r,m := int(rand.Int31n(int32(total))),0
	randArr := strings.Split(randstr,";")
	for _,v := range randArr {
		rewardArr := strings.Split(v, "+")
		rate,err := strconv.Atoi(rewardArr[1])
		if err != nil {
			return 0,err
		}

		if r <= (rate+m) {
			rewardItemId,err := strconv.Atoi(rewardArr[0])
			if err != nil {
				return 0,err
			}
			return int32(rewardItemId),nil
		} else {
			m += rate
		}
	}

	return 0,nil
}

func combineRandStr(randStr string,total int) (ret string) {
	randArr := strings.Split(randStr,";")
	sum := 0
	for _,v := range randArr {
		rewardArr := strings.Split(v, "+")
		rate,err := strconv.Atoi(rewardArr[2])
		if err != nil {
			return
		}
		sum += rate
	}
	// 不中的概率
	blank_num := total - sum
	randStr = fmt.Sprintf("%s;0+0+%d" ,randStr, blank_num)
	return randStr
}

// randstr 结构:   itemid+数量+概率;itemid+数量+概率;itemid+数量+概率
func GetResultNew(total int,randstr string) (itemid int32, num int32, err error){
	randInit()
	r,m := int(rand.Int31n(int32(total))),0
	randstr = combineRandStr(randstr,total)
	if randstr == "" {
		fmt.Println("合并错误")
		return 0,0,err
	}
	//如果没有";"号，表示是必中的:
	if !strings.Contains(randstr,";") {
		rewardArr := strings.Split(randstr, "+")
		rewardItemId,err := strconv.Atoi(rewardArr[0])
		if err != nil {
			return 0,0,err
		}
		conver_num,_ := strconv.Atoi(rewardArr[1])
		return int32(rewardItemId),int32(conver_num),nil
	}

	randArr := strings.Split(randstr,";")
	for _,v := range randArr {
		rewardArr := strings.Split(v, "+")
		rate,err := strconv.Atoi(rewardArr[2])
		if err != nil {
			return 0,0,err
		}

		if r <= (rate+m) {
			rewardItemId,err := strconv.Atoi(rewardArr[0])
			if err != nil {
				return 0,0,err
			}

			conver_num,_ := strconv.Atoi(rewardArr[1])
			return int32(rewardItemId),int32(conver_num),nil
		} else {
			m += rate
		}
	}

	return 0,0,nil
}