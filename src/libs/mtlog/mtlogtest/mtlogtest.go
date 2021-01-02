package main

import (
	"encoding/json"
	"io/ioutil"

	"libs/mtlog"
)

const (
	MTLOG_TEST_CFG_FILE = "src/zc/libs/mtlog/mtlogtest/mtlog_test.cfg"
)

type Marksheet struct {
	English uint16 `json:"English"`
	Maths   uint16 `json:"Maths"`
	Science uint16 `json:"Science"`
	Arts    uint16 `json:"Arts"`
}

type StudentRecord struct {
	Name  string      `json:"Name"`
	Class int32       `json:"Class"`
	Marks []Marksheet `json:"Marks"`
}

var (
	configFile string
	stdRecord  [3]StudentRecord
)

func main() {
	configFile = MTLOG_TEST_CFG_FILE
	mtlog.SetLevel(mtlog.DebugLevel)
	if body, err := ioutil.ReadFile(configFile); err != nil {
		mtlog.Errorf("Could not read configuration %v Error %v", configFile, err)
		return
	} else {
		mtlog.Debug("Config Details : ", string(body))
		if err := json.Unmarshal(body, &stdRecord); err != nil {
			mtlog.Errorf("Configuration parse error : %v", err)
			return
		} else {
			mtlog.Debug(mtlog.JsonStringify(stdRecord))
		}
	}
}
