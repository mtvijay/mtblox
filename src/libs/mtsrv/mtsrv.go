package mtsrv

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/mtbox/mtlog"
)

type ServiceDispatchFunc func(*NetServices, int) error

func showVersion(name, ver string) {
	log.Printf("%s service version build %v", name, ver)
}

func ParseConfig(configFile string, srvCfg interface{}) error {
	if body, err := ioutil.ReadFile(configFile); err != nil {
		log.Printf("Could not read configuration [%v]: %v", configFile, err)
		return err
	} else {
		if err := json.Unmarshal(body, srvCfg); err != nil {
			log.Printf("Failed to parse for external configuration, error: %v", err)
			return err
		}
	}

	return nil
}

func StoreConfig(configFile string, srvCfg interface{}) error {
	configJson, err := json.Marshal(srvCfg)
	if err != nil {
		log.Printf("Failed to parse for external configuration, error: %v", err)
		return err
	}
	err = ioutil.WriteFile(configFile, configJson, 0644)
	if err != nil {
		log.Printf("Failed to write configuration to file: error was: %v", err)
		return err
	}
	return nil
}

type RetentionPolicy struct {
	Duration      string `json:"Duration"`
	ShardDuration string `json:"ShardDuration"`
}

type NetServices struct {
	Name       string          `json:"Name"`
	Host       string          `json:"Server"`
	Port       uint32          `json:"Port"`
	User       string          `json:"User"`
	Password   string          `json:"Password"`
	Topics     []string        `json:"Topics"`
	Frequency  uint64          `json:"Frequency"`
	LoadFactor int             `json:"LoadFactor"`
	DbSerialNo int             `json:"DbSerialNo"`
	Retention  RetentionPolicy `json:"Retention"`
	TLSDisable bool            `json:"TLSDisable"`
}

func (cp *NetServices) ServiceAddr() string {
	if cp == nil {
		return ""
	}
	hostPoint := cp.Host
	if cp.Port != 0 {
		hostPoint += ":" + strconv.Itoa(int(cp.Port))
	}

	return hostPoint
}

type ServiceCommonConfig struct {
	PidFile        string          `json:"PidFile"`
	ServiceName    string          `json:"ServiceName"`
	ServiceInst    uint8           `json:"ServiceInst"`
	DebugMode      bool            `json:"DebugMode"`
	Threads        int             `json:"Threads"`
	RootPath       string          `json:"RootPath"`
	LogCfg         mtlog.LogConfig `json:"LogCfg"`
	SystemPeriodic bool            `json:"SystemPeriodic"`
	Services       []NetServices   `json:"Services"`
}

type Server struct {
	sync.Mutex
	metList map[string]MtMetric
}

// Initialize the common flags
func InitFlags(name string, ver string) string {
	flVersion := flag.Bool("v", false, "Print version information and quit")
	// TO BE DEPRECATED : Check if config file name has been passed as command-line argument
	flConfig := flag.String("cfg", "", "Json - Configuration File")
	// Check if config File name has been passed as environment variable
	envConfig := os.Getenv("cfgFile")

	flag.Parse()

	if *flVersion {
		showVersion(name, ver)
		return ""
	}

	if flag.NArg() != 0 {
		flag.Usage()
		return ""
	}

	if envConfig == "" && *flConfig == "" {
		log.Printf("Critical: No configuration file - cfg")
		flag.Usage()
		return ""
	} else if envConfig != "" {
		return envConfig
	} else {
		return *flConfig
	}
}

func (s *Server) RunCommonLoop(scCfg *ServiceCommonConfig, disp ServiceDispatchFunc) {
	runtime.GOMAXPROCS(scCfg.Threads)
	wg := sync.WaitGroup{}

	if scCfg.SystemPeriodic {
		log.Printf("Starting periodic")
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.systemServicePeriodic()
		}()
	}

	for _, v := range scCfg.Services {
		wg.Add(1)
		// start the individual service
		go func(cp NetServices) {
			defer wg.Done()
			disp(&cp, int(scCfg.LogCfg.Level))
		}(v)

		// FIXME: we need to add dependancy model
		// add a delay here to actually so that there is time
		// between other services to start
		time.Sleep(time.Second)
	}
	wg.Wait()
}

func (s *Server) RegisterMetric(name string, met MtMetric) {
	s.Lock()
	defer s.Unlock()
	s.metList[name] = met
	return
}

func (s *Server) UnregisterMetric(name string) {
	s.Lock()
	defer s.Unlock()
	delete(s.metList, name)
	return
}

//
func NewServer(scCfg *ServiceCommonConfig) *Server {
	s := Server{}
	s.metList = make(map[string]MtMetric)
	return &s
}
