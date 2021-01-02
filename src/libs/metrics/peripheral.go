package metrics

import (
	"time"
)

const (
	RED    = 0
	GREEN  = 1
	YELLOW = 2
)

//database health stat.
type DatabaseHealthStat struct {
	DatabaseName             string
	MicroserviceName         string
	TotalReadOp              uint64
	TotalWriteOp             uint64
	TotalUnsuccessfulReadOp  uint64
	TotalUnsuccessfulWriteOp uint64
	TotalLatencyReadOp       int64
	TotalLatencyWriteOp      int64
	InitializationTime       time.Time
	Status                   int
	UpTime                   string
	IsConnected              bool
	Error                    error
}

//send cassandra health stat per microservice.
func (s *DatabaseHealthStat) GetStat() *DatabaseHealthStat {
	return s
}

//reset cassandra initialization time.
func (s *DatabaseHealthStat) ResetInitTime() {
	s.InitializationTime = time.Now()
}

//reset cassandra stats to zero.
func (s *DatabaseHealthStat) ResetStat() {
	s.TotalReadOp = 0
	s.TotalWriteOp = 0
	s.TotalUnsuccessfulReadOp = 0
	s.TotalUnsuccessfulWriteOp = 0
	s.TotalLatencyReadOp = 0
	s.TotalLatencyWriteOp = 0
}

//kafka health stat.
type TransportHealthStat struct {
	TransportName string

	TotalRx    Zcounter
	TotalRxErr Zcounter
	TotalTx    Zcounter
	TotalTxErr Zcounter

	TotalRxInOneInterval                Zcounter
	TotalRxErrInOneInterval             Zcounter
	TotalTxInOneInterval                Zcounter
	TotalTxErrInOneInterval             Zcounter
	TotalTimeToSendMsgsInOneInterval    int64
	TotalTimeToReceiveMsgsInOneInterval int64
	StartReceiveTime                    time.Time
	InitializationTime                  time.Time
	Status                              int
	UpTime                              string
	IsConnected                         bool
	Error                               error
}

type SecurityStackHealthStat struct {
	SecuritySrvsName            string
	TotalSuccessfulEncryption   Zcounter
	TotalSuccessfulDecryption   Zcounter
	TotalUnsuccessfulEncryption Zcounter
	TotalUnsuccessfulDecryption Zcounter
	InitializationTime          time.Time
	Status                      int
	UpTime                      string
	IsConnected                 bool
	Error                       error
}
