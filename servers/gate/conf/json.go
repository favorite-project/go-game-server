package conf

import (
	"encoding/json"
	"github.com/name5566/leaf/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var Server struct {
	LogLevel       string
	LogPath        string
	WSAddr         string
	TCPAddr        string
	MaxConnNum     int
	ConsolePort    int
	ProfilePath    string
	GameServerAddr string
}

func init() {
	path, _ := filepath.Abs(os.Args[0])
	path = strings.Replace(path, "\\", "/", -1)
	path = string([]byte(path)[:strings.LastIndex(path, "/")+1])

	data, err := ioutil.ReadFile(path + "conf/server.json")
	if err != nil {
		log.Fatal("%v", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("%v", err)
	}
}
