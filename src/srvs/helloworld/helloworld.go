package main

import (
	"sync"
	"time"

	"github.com/mtbox/mtlog"
	"github.com/mtbox/mtsrv"
)

//this structure will keep health status
//of services in key value pair.
type HealthReport struct {
	SrvsHealthReport map[string]ReportStatusAndCounter
	FullHealthReport map[string]ReportStatusAndCounter
	sync.Mutex
}

type ReportStatusAndCounter struct {
	Status          string
	HostName        string
	EnvName         string
	ServiceName     string
	ServiceInstance string
	Counter         int
}

const (
	VERSION                 = "heloooo1.0"
	serviceName             = "helloworld"
	REPORT_PREPARE_INTERVAL = 61 //Keeping this interval more than health report interval.
	DECLARE_DEAD_AFTER      = 3  //declare dead after 3 cycles.
)

var (
	healthRep              *HealthReport
	cdConfig               CfgLocalServices
	healthCtx              *mtsrv.Health
	vaultClientCreated     bool = false
	startVaultTokenRenewal bool = false
)

// this is collective all the basic setup is done and we
// are ready for receiving the kakfa messages
func systemReady() bool {
	return true
}

type VaultConfig struct {
	RoleId   string
	SecretId string
}

type CfgLocalServices struct {
	ScCfg mtsrv.ServiceCommonConfig `json:"ServiceCommonConfig"`
	SpCfg ServiceSpecificConfig     `json:"ServiceSpecificConfig"`
}

type ServiceSpecificConfig struct {
	Simulate    bool
	VaultDetail VaultConfig
}

func serviceMain(spCfg ServiceSpecificConfig) {
	simulate := "no"
	if spCfg.Simulate {
		simulate = "yes"
	}
	mtlog.Tracef("Simulate mode = %s", simulate)
}

func main() {
	cfgFile := mtsrv.InitFlags(serviceName, VERSION)

	err := mtsrv.ParseConfig(cfgFile, &cdConfig)
	if err != nil {
		mtlog.Errorf("Critical: Configuration file parsing failed")
		return
	}
	mtlog.InitLogging(cdConfig.ScCfg.LogCfg)

	s := mtsrv.NewServer(&cdConfig.ScCfg)

	serviceMain(cdConfig.SpCfg)
	mtlog.Tracef("Setting Log level to ", mtlog.LogLevel(cdConfig.ScCfg.LogCfg.Level).String())
	mtlog.SetLevel(mtlog.LogLevel(cdConfig.ScCfg.LogCfg.Level))

	//Initialize health report.
	healthRep = &HealthReport{}
	healthRep.SrvsHealthReport = make(map[string]ReportStatusAndCounter)
	healthRep.FullHealthReport = make(map[string]ReportStatusAndCounter)

	s.RunCommonLoop(&cdConfig.ScCfg, ListenAndServe)
}

func Report() {
	healthRep.Lock()
	defer healthRep.Unlock()
}

func PrepareHealthReport() {
	ticker := time.NewTicker(time.Duration(REPORT_PREPARE_INTERVAL) * time.Second)
	for range ticker.C {
		Report()
	}
}

func fillhealthReport() {
}

func sendMicroserviceHealthReport(healtStatusSendInterval uint64) {
	ticker := time.NewTicker(time.Duration(healtStatusSendInterval) * time.Second)
	for range ticker.C {
		fillhealthReport()
	}
}

func KafkaMessageReceiver(kafkaMessgaeChan chan []byte) {
	mtlog.Info("Starting KafkaMessageReceiver for helloworld...")
	for {
		select {
		case msg := <-kafkaMessgaeChan:
			mtlog.Info("Received...%v", msg)
		}
	}
	mtlog.Info("Exiting KafkaMessageReceiver for helloworld...")
}

//create a vault client for helloworld service.
//we will use this client for vault api call.
func vaultMain(cp *mtsrv.NetServices) {
	for {
		time.Sleep(time.Minute)
	}
}

func ListenAndServe(cp *mtsrv.NetServices, logging int) error {
	broker := cp.ServiceAddr()

	messageChan := make(chan []byte)
	mtlog.Tracef("Starting network service endpoint %v at %s", cp.Name, broker)
	switch cp.Name {
	case "kafka":
		// wait till system is ready before we start kafka subscribers
		for !systemReady() {
			time.Sleep(time.Minute)
		}
		go KafkaMessageReceiver(messageChan)

	case "cassandra":
	case "health":
		var hErr error
		healthCtx, hErr = mtsrv.NewHealthSrvs(serviceName)
		if hErr == nil {
			mtlog.Info("Starting health monitoring service")
			go sendMicroserviceHealthReport(cp.Frequency)
			go PrepareHealthReport()
		}
	case "zvault":
		vaultMain(cp)

	case "tls":
		hctx, err := NewConnect(serviceName, false, nil)
		if err != nil {
			mtlog.Errorf("Critical: web service create failed %v", err)
			return err
		}
		_ = hctx.AppendRouter(initProjectHandlers())
	}
	return nil
}
