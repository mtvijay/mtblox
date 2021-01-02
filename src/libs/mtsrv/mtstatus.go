package mtsrv

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"libs/metrics"
	"libs/mtlog"
	"runtime"
	"time"
)

type SrvHealthStatusType int

const (
	ServiceHealthStatus_HEALTH_UNK SrvHealthStatusType = iota
	ServiceHealthStatus_HEALTH_RED
	ServiceHealthStatus_HEALTH_YELLOW
	ServiceHealthStatus_HEALTH_GREEN
)

type PeripheralStatusType int

const (
	FAILED PeripheralStatusType = iota
	CONNECTED
	CONNECTING
)

const (
	FAILED_UPTIME        = "service is not up"
	HEALTH_SEND_INTERVAL = "per min" //XXX
)

type SrvsComponentComposition struct {
	Databases     []*PersistenceStatusDetail
	Transports    []*TransportBlockStatusDetail
	SecurityStack []*SecurityStackDetail
	WebHooks      *WebHooksList
}

type Health struct {
	name       string
	startTime  time.Time
	processCtx *process.Process
}

//top level str for checking microservices health.
type MtsrvStatus struct {
	BriefStatus    *BriefStatusSummary
	DetailedStatus *DetailedStatusSummary
}

//brief summary of micoservice health.
//XXX FIXME what do we need to report in
//brief summary?
type BriefStatusSummary struct {
	ServiceName string
	Status      SrvHealthStatusType
}

type DetailedStatusSummary struct {
	ConnectedPeripheral *PeripheralList
	ProcessStatus       *ProcessDetail
	Security            []*SecurityStackDetail
	WebHooks            *WebHooksList
}

//list and status of peripherals connected
//with the microservices.
type PeripheralList struct {
	Databases  []*PersistenceStatusDetail    //can be cassandra, influx, redis etc.
	Transports []*TransportBlockStatusDetail //can be kafka, http etc.
}

type TransportBlockStatusDetail struct {
	Name                                   string               //name of transport srvs
	Status                                 PeripheralStatusType // Connectivity status.
	UpTime                                 string               // Time since the transport service is up.
	TotalRequests                          uint64               // Total requests received
	TotalResponse                          uint64               // Total response sent
	TotalRequestsInOneInterval             uint64               // Total requests received in one measurement interval
	TotalResponseInOneInterval             uint64               // Total response sent in one measurement interval
	RequestRate                            string               // Request per min.
	ResponseRate                           string
	AvgTimeToServeRequests                 int64
	AvgTimeToServeResponse                 int64
	TotalSuccessfulRequests                uint64
	TotalUnsuccessfulRequests              uint64
	TotalSuccessfulResponse                uint64
	TotalUnsuccessfulResponse              uint64
	TotalSuccessfulRequestsInOneInterval   uint64
	TotalUnsuccessfulRequestsInOneInterval uint64
	TotalSuccessfulResponseInOneInterval   uint64
	TotalUnsuccessfulResponseInOneInterval uint64
	Error                                  string
}

//status and operation detail of databases
//connected with the microservices.
type PersistenceStatusDetail struct {
	Name                              string               //name of database srvs
	Status                            PeripheralStatusType // Connectivity status.
	NonLocalIp                        string
	UpTime                            string // Time since the db is up.
	TotalReadOperations               uint64 // Total read operation on DB.
	TotalWriteOperations              uint64 // Total write operation on DB.
	RequestRate                       string // Query per min?
	AvgExecutionTimeOfReadOperations  int64
	AvgExecutionTimeOfWriteOperations int64
	TotalUnsuccessfulReadOperations   uint64 // Total unsuccessful read operation on DB.
	TotalUnsuccessfulWriteOperations  uint64 // Total unsuccessful write operation on DB.
	Error                             string
}

//status and operation detail of security service
//connected with the microservices.(eg vault)
type SecurityStackDetail struct {
	Name                        string               //name of security service
	Status                      PeripheralStatusType // Connectivity status.
	UpTime                      string               // Time since the security srvs is up.
	TotalSuccessfulEncryption   uint64               //Total successful encryption.
	TotalSuccessfulDecryption   uint64               //Total successful decryption.
	TotalUnsuccessfulEncryption uint64               //Total unsuccessful encryption.
	TotalUnsuccessfulDecryption uint64               //Total unsuccessful decryption.
	Error                       string
}

//process detail.
type ProcessDetail struct {
	Id         int32          // PId of process.
	Name       string         // Name of process.
	UpTime     string         // Time since the process is in up or exit state.
	State      string         // State of the process.
	NumThreads int32          //NumThreads returns the number of threads used by the process.
	Stats      *ProcessStats  // Realtime stats for container.
	Runtime    *RuntimeDetail // Runtime stats, like no of active go routines.
}

//aggregating all types of stats of one process.
type ProcessStats struct {
	PageFaultStat  *PageFaultDetail
	IOCounterStats *IOCounterStatDetail
	CpuStats       *CpuStatDetail
	MemoryStat     *MemoryStatDetail
	Network        []*NetworkStatDetail
}

//page fault detail will help us to improve
//the performance of our process.
type PageFaultDetail struct {
	MinorFaults      uint64 //satisfied by accessing the disk
	MajorFaults      uint64 //satisfied by sharing pages that are already in memory.
	ChildMinorFaults uint64
	ChildMajorFaults uint64
}

//Then number of bytes process has read and
//write from the storage and their counter.
type IOCounterStatDetail struct {
	ReadCount  uint64 // Attempt to count the number of bytes which this process really did cause to be fetched from the storage layer.
	WriteCount uint64 // Attempt to count the number of bytes which this process caused to be sent to the storage layer
	ReadBytes  uint64 // The number of bytes which this task has caused to be read from storage.
	WriteBytes uint64 // The number of bytes which this task has caused, or shall cause to be written to disk
}

//aggregates and wraps all CPU related info of container.
type CpuStatDetail struct {
	CpuPercent float64 // Percent of the CPU time this process uses.
	User       float64 // Time spent by tasks of the cgroup in user mode.
	System     float64 // Time spent by tasks of the cgroup in kernel mode.
}

//aggregates all memory stats.
type MemoryStatDetail struct {
	MemPercent float32 // Percent of the total RAM this process uses
	Alloc      uint64  // Alloc is bytes of allocated heap objects.
	TotalAlloc uint64  // TotalAlloc is cumulative bytes allocated for heap objects.
	Sys        uint64  // Sys is the total bytes of memory obtained from the OS.
	Mallocs    uint64  // Mallocs is the cumulative count of heap objects allocated.
	Frees      uint64  // Frees is the cumulative count of heap objects freed.
	HeapSys    uint64  // HeapSys estimates the largest size the heap has had.
	// HeapIdle minus HeapReleased estimates the amount of memory
	// that could be returned to the OS, but is being retained by
	HeapIdle     uint64
	NextGC       uint64 // NextGC is the target heap size of the next GC cycle.
	LastGC       uint64 // LastGC is the time the last garbage collection finished.(nanoseconds since 1970)
	PauseTotalNs uint64 // PauseTotalNs is the cumulative nanoseconds in GC.
	NumGC        uint32 // NumGC is the number of completed GC cycles.
}

// aggregates the network stats of one container.
type NetworkStatDetail struct {
	InterfaceName string

	RxPackets uint64 // Number of packets received.
	RxBytes   uint64 // Number of bytes received.
	TxPackets uint64 // Number of packets sent.
	TxBytes   uint64 // Number of bytes sent.

	RxErrors  uint64 // Total number of errors while receiving.(Errin)
	TxErrors  uint64 // Total number of errors while sending.(Errout)
	RxDropped uint64 // Total number of incoming packets which were dropped.
	TxDropped uint64 // Total number of outgoing packets which were dropped.

	EndpointID string // Endpoint ID.
	InstanceID string // Instance ID.
}

//runtime stats.
type RuntimeDetail struct {
	NumOfGoRoutines        int // Number of go-routines running inside a container.
	NumOfCrashedGoRoutines int // Number go-routines crashed inside a container.
	NumOfCpu               int // Number of logical CPUs usable by the current process.
}

//XXX his will contain aws connectivity, email send etc related info.
type WebHooksList struct {
}

//Initialize a new health service.
func NewHealthSrvs(serviceName string) (*Health, error) {
	processes, pErr := process.Processes()
	if pErr != nil {
		mtlog.Errorf("Unable to fetch process detail:%v", pErr)
		return nil, pErr
	}
	healthCtx := &Health{}

	for _, process := range processes {
		processName, nErr := process.Name()
		if nErr != nil {
			mtlog.Errorf("Unable to fetch process name:%v", nErr)
			return nil, nErr
		}
		if serviceName == processName {
			healthCtx.processCtx = process
			healthCtx.name = processName
			healthCtx.startTime = time.Now()
			break
		}
	}
	return healthCtx, nil
}

//Complete microservice health status.
//1) Detail of microservice health status.
//2) And brief health status.
//Fill Peripheral detail and process detail.
func (hCtx *Health) CompleteSrvsStatus(srvsComponent *SrvsComponentComposition) (*MtsrvStatus, error) {
	srvsStatus := &MtsrvStatus{}
	detailedSumm := &DetailedStatusSummary{}
	periphDetail := &PeripheralList{}

	//fill database detail.
	if len(srvsComponent.Databases) != 0 {
		for _, db := range srvsComponent.Databases {
			periphDetail.Databases = append(periphDetail.Databases, db)
		}
	}

	//fill transport detail.
	if len(srvsComponent.Transports) != 0 {
		for _, trans := range srvsComponent.Transports {
			periphDetail.Transports = append(periphDetail.Transports, trans)
		}
	}
	detailedSumm.ConnectedPeripheral = periphDetail

	//fill security stack detail.
	if len(srvsComponent.SecurityStack) != 0 {
		detailedSumm.Security = srvsComponent.SecurityStack
	}

	//fill process detail.
	processDetail, pErr := hCtx.ProcessDetail()
	if pErr == nil {
		detailedSumm.ProcessStatus = processDetail
	}
	srvsStatus.DetailedStatus = detailedSumm

	briefSrvsStatus, bErr := hCtx.BriefSrvsStatus(srvsStatus)
	if bErr != nil {
		return nil, bErr
	}

	if briefSrvsStatus.Status != ServiceHealthStatus_HEALTH_GREEN {
		mtlog.Errorf("Health of microservice %v is %v", briefSrvsStatus.ServiceName, briefSrvsStatus.Status)
	}

	srvsStatus.BriefStatus = briefSrvsStatus

	return srvsStatus, nil
}

//Fill transport related stats.
//For kafka we are assuming producer is receiving requests of sending msg and
//subscriber is receiving response by getting the msg sent by producer.
func (hCtx *Health) TransportHealthDetail(transportStat *metrics.TransportHealthStat) (*TransportBlockStatusDetail, error) {
	if hCtx == nil {
		return nil, fmt.Errorf("Health context is nil")
	}
	transDetail := &TransportBlockStatusDetail{}

	//check connected transport status of microservice.
	//For kafka we will do this in two steps-
	//1)check if producer and consumer exists.
	//2)check for Tx and Rx errors.
	time_init := time.Time{}
	if transportStat.IsConnected {
		if transportStat.TotalTxErrInOneInterval.Value().(uint64) == 0 && transportStat.TotalRxErrInOneInterval.Value().(uint64) == 0 {
			transDetail.Status = CONNECTED
			transDetail.UpTime = time.Since(transportStat.InitializationTime).String()
		} else {
			transDetail.Status = CONNECTING
			transDetail.UpTime = time_init.String()
			errStr := fmt.Sprintf("TotalTx Err In One Interval Is %v, TotalRx Err In One Interval Is %v ",
				transportStat.TotalTxErrInOneInterval.Value().(uint64), transportStat.TotalRxErrInOneInterval.Value().(uint64))
			transDetail.Error = errStr
		}
	} else {
		transDetail.Status = FAILED
		transDetail.UpTime = time_init.String()
		if transportStat.Error != nil {
			transDetail.Error = transportStat.Error.Error()
		}
	}

	transDetail.Name = transportStat.TransportName

	transDetail.TotalSuccessfulRequests = transportStat.TotalTx.Value().(uint64)
	transDetail.TotalUnsuccessfulRequests = transportStat.TotalTxErr.Value().(uint64)
	transDetail.TotalSuccessfulResponse = transportStat.TotalRx.Value().(uint64)
	transDetail.TotalUnsuccessfulResponse = transportStat.TotalRxErr.Value().(uint64)
	transDetail.TotalRequests = transDetail.TotalSuccessfulRequests + transDetail.TotalUnsuccessfulRequests
	transDetail.TotalResponse = transDetail.TotalSuccessfulResponse + transDetail.TotalUnsuccessfulResponse

	transDetail.TotalSuccessfulRequestsInOneInterval = transportStat.TotalTxInOneInterval.Value().(uint64)
	transDetail.TotalUnsuccessfulRequestsInOneInterval = transportStat.TotalTxErrInOneInterval.Value().(uint64)
	transDetail.TotalSuccessfulResponseInOneInterval = transportStat.TotalRxInOneInterval.Value().(uint64)
	transDetail.TotalUnsuccessfulResponseInOneInterval = transportStat.TotalRxErrInOneInterval.Value().(uint64)
	transDetail.TotalRequestsInOneInterval = transDetail.TotalSuccessfulRequestsInOneInterval + transDetail.TotalUnsuccessfulRequestsInOneInterval
	transDetail.TotalResponseInOneInterval = transDetail.TotalSuccessfulResponseInOneInterval + transDetail.TotalUnsuccessfulResponseInOneInterval

	transDetail.RequestRate = fmt.Sprintf("%v %v", transDetail.TotalRequestsInOneInterval, HEALTH_SEND_INTERVAL)
	transDetail.ResponseRate = fmt.Sprintf("%v %v", transDetail.TotalResponseInOneInterval, HEALTH_SEND_INTERVAL)
	if transDetail.TotalSuccessfulRequestsInOneInterval != 0 {
		transDetail.AvgTimeToServeRequests = transportStat.TotalTimeToSendMsgsInOneInterval / int64(transDetail.TotalSuccessfulRequestsInOneInterval)
	}
	if transDetail.TotalSuccessfulResponseInOneInterval != 0 {
		transDetail.AvgTimeToServeResponse = transportStat.TotalTimeToReceiveMsgsInOneInterval / int64(transDetail.TotalSuccessfulResponseInOneInterval)
	}

	return transDetail, nil
}

//Fill database related stats.
func (hCtx *Health) DatabaseHealthDetail(dbStat *metrics.DatabaseHealthStat) (*PersistenceStatusDetail, error) {
	if hCtx == nil {
		return nil, fmt.Errorf("Health context is nil")
	}
	dbDetail := &PersistenceStatusDetail{}

	//check DB status of microservice.
	//we will do this in two steps-
	//1)check if DB client is connected.
	//2)check unsuccessful R/W operation.
	time_init := time.Time{}
	if dbStat.IsConnected {
		if dbStat.TotalUnsuccessfulReadOp == 0 && dbStat.TotalUnsuccessfulWriteOp == 0 {
			dbDetail.UpTime = time.Since(dbStat.InitializationTime).String()
			dbDetail.Status = CONNECTED
		} else {
			dbDetail.UpTime = time_init.String()
			dbDetail.Status = CONNECTING
			errStr := fmt.Sprintf("Total Unsuccessful Read Operation Is %v, Total Unsuccessful Write Operation Is %v ",
				dbStat.TotalUnsuccessfulReadOp, dbStat.TotalUnsuccessfulWriteOp)
			dbDetail.Error = errStr
		}
	} else {
		dbDetail.UpTime = time_init.String()
		dbDetail.Status = FAILED
		if dbStat.Error != nil {
			dbDetail.Error = dbStat.Error.Error()
		}
	}

	dbDetail.Name = dbStat.DatabaseName
	dbDetail.TotalReadOperations = dbStat.TotalReadOp
	dbDetail.TotalWriteOperations = dbStat.TotalWriteOp

	dbDetail.TotalUnsuccessfulReadOperations = dbStat.TotalUnsuccessfulReadOp
	dbDetail.TotalUnsuccessfulWriteOperations = dbStat.TotalUnsuccessfulWriteOp

	totalOp := dbStat.TotalReadOp + dbStat.TotalWriteOp
	dbDetail.RequestRate = fmt.Sprintf("%v %v", totalOp, HEALTH_SEND_INTERVAL)
	if dbStat.TotalReadOp != 0 {
		dbDetail.AvgExecutionTimeOfReadOperations = dbStat.TotalLatencyReadOp / int64(dbStat.TotalReadOp)
	}
	if dbStat.TotalWriteOp != 0 {
		dbDetail.AvgExecutionTimeOfWriteOperations = dbStat.TotalLatencyWriteOp / int64(dbStat.TotalWriteOp)
	}

	return dbDetail, nil
}

//Fill security stack related details.
func (hCtx *Health) SecurityStackHealthDetail(securityStat *metrics.SecurityStackHealthStat) (*SecurityStackDetail, error) {
	if hCtx == nil {
		return nil, fmt.Errorf("Health context is nil")
	}
	securityDetail := &SecurityStackDetail{}

	//check security connectivity status of
	//microservice.
	//we will do this in two steps-
	//1)check if security srvs client is connected?
	//2)check unsuccessful Encryption/Decryption operation.
	time_init := time.Time{}
	if securityStat.IsConnected {
		if securityStat.TotalUnsuccessfulEncryption.Value().(uint64) == 0 && securityStat.TotalUnsuccessfulDecryption.Value().(uint64) == 0 {
			securityDetail.UpTime = time.Since(securityStat.InitializationTime).String()
			securityDetail.Status = CONNECTED
		} else {
			securityDetail.UpTime = time_init.String()
			securityDetail.Status = CONNECTING
			errStr := fmt.Sprintf("Total Unsuccessful Encryption Operation Is %v, Total Unsuccessful Decryption Operation Is %v ",
				securityStat.TotalUnsuccessfulEncryption.Value().(uint64), securityStat.TotalUnsuccessfulDecryption.Value().(uint64))
			securityDetail.Error = errStr
		}
	} else {
		securityDetail.UpTime = time_init.String()
		securityDetail.Status = FAILED
		if securityStat.Error != nil {
			securityDetail.Error = securityStat.Error.Error()
		}
	}
	securityDetail.Name = securityStat.SecuritySrvsName
	securityDetail.TotalSuccessfulEncryption = securityStat.TotalSuccessfulEncryption.Value().(uint64)
	securityDetail.TotalSuccessfulDecryption = securityStat.TotalSuccessfulDecryption.Value().(uint64)
	securityDetail.TotalUnsuccessfulEncryption = securityStat.TotalUnsuccessfulEncryption.Value().(uint64)
	securityDetail.TotalUnsuccessfulDecryption = securityStat.TotalUnsuccessfulDecryption.Value().(uint64)
	return securityDetail, nil
}

//Brief of microservice health status
func (hCtx *Health) BriefSrvsStatus(srvsStat *MtsrvStatus) (*BriefStatusSummary, error) {
	if hCtx == nil {
		return nil, fmt.Errorf("health context is nil")
	}

	briefStatus := &BriefStatusSummary{}
	briefStatus.ServiceName = hCtx.name

	detailedStatus := srvsStat.DetailedStatus
	if detailedStatus == nil {
		return nil, fmt.Errorf("DetailedStatus is nil")
	}

	connPeripheral := detailedStatus.ConnectedPeripheral
	if connPeripheral != nil {

		if len(connPeripheral.Databases) != 0 {
			for _, db := range connPeripheral.Databases {
				if db != nil {
					if db.Status == CONNECTED {
						briefStatus.Status = ServiceHealthStatus_HEALTH_GREEN
					} else {
						if db.Status == CONNECTING {
							mtlog.Errorf("%v Database Detail: %v", db.Name, *db)
							briefStatus.Status = ServiceHealthStatus_HEALTH_YELLOW
							return briefStatus, nil
						}
						mtlog.Errorf("%v Database Detail: %v", db.Name, *db)
						briefStatus.Status = ServiceHealthStatus_HEALTH_RED
						return briefStatus, nil
					}
				}
			}
		}

		if len(connPeripheral.Transports) != 0 {
			for _, trans := range connPeripheral.Transports {
				if trans != nil {
					if trans.Status == CONNECTED {
						briefStatus.Status = ServiceHealthStatus_HEALTH_GREEN
					} else {
						if trans.Status == CONNECTING {
							mtlog.Errorf("%v Transport Detail: %v", trans.Name, *trans)
							briefStatus.Status = ServiceHealthStatus_HEALTH_YELLOW
							return briefStatus, nil
						}
						mtlog.Errorf("%v Transport Detail: %v", trans.Name, *trans)
						briefStatus.Status = ServiceHealthStatus_HEALTH_RED
						return briefStatus, nil
					}
				}
			}
		}
	}

	//check security service status.
	if len(detailedStatus.Security) != 0 {
		for _, security := range detailedStatus.Security {
			if security != nil {
				if security.Status == CONNECTED {
					briefStatus.Status = ServiceHealthStatus_HEALTH_GREEN
				} else {
					if security.Status == CONNECTING {
						mtlog.Errorf("%v Security Service Detail: %v", security.Name, *security)
						briefStatus.Status = ServiceHealthStatus_HEALTH_YELLOW
						return briefStatus, nil
					}
					mtlog.Errorf("%v Security Service Detail: %v", security.Name, *security)
					briefStatus.Status = ServiceHealthStatus_HEALTH_RED
					return briefStatus, nil
				}
			}
		}
	}

	//check for process status and mark health
	//accordingly.
	processStatus := detailedStatus.ProcessStatus
	if processStatus == nil {
		briefStatus.Status = ServiceHealthStatus_HEALTH_RED
		return briefStatus, nil
	}
	if processStatus.State == "T" || processStatus.State == "I" || processStatus.State == "W" ||
		processStatus.State == "Z" || processStatus.State == "L" {
		briefStatus.Status = ServiceHealthStatus_HEALTH_RED
		return briefStatus, nil
	}
	briefStatus.Status = ServiceHealthStatus_HEALTH_GREEN
	return briefStatus, nil
}

//Fill cpu, memory, network, IOCounters details
//of a process.
func (hCtx *Health) ProcessDetail() (*ProcessDetail, error) {
	pCtx := hCtx.processCtx
	if pCtx == nil {
		return nil, fmt.Errorf("Process context is nil")
	}
	var processDetail = &ProcessDetail{}
	processDetail.Id = pCtx.Pid
	processDetail.Name = hCtx.name
	processDetail.UpTime = time.Since(hCtx.startTime).String()
	status, sErr := pCtx.Status()
	if sErr != nil {
		mtlog.Errorf("Unable to fetch process state:%v", sErr)
	} else {
		processDetail.State = status
	}

	noThreads, tErr := pCtx.NumThreads()
	if tErr != nil {
		mtlog.Errorf("Unable to fetch number of threads:", tErr)
	} else {
		processDetail.NumThreads = noThreads
	}

	//fill process stats.
	pStats := &ProcessStats{}

	//stats of page fault.
	pFault := &PageFaultDetail{}
	faults, fErr := pCtx.PageFaults()
	if fErr != nil {
		mtlog.Errorf("Unable to fetch PageFaults detail :%v", fErr)
	} else {
		pFault.MinorFaults = faults.MinorFaults
		pFault.MajorFaults = faults.MajorFaults
		pFault.ChildMinorFaults = faults.ChildMinorFaults
		pFault.ChildMajorFaults = faults.ChildMajorFaults
	}

	//stats of process IOCounter.
	pioCounter := &IOCounterStatDetail{}
	iocounters, ioErr := pCtx.IOCounters()
	if ioErr != nil {
		mtlog.Errorf("Unable to fetch IOCounter detail :%v", ioErr)
	} else {
		pioCounter.ReadCount = iocounters.ReadCount
		pioCounter.WriteCount = iocounters.WriteCount
		pioCounter.ReadBytes = iocounters.ReadBytes
		pioCounter.WriteBytes = iocounters.WriteBytes
	}
	pStats.IOCounterStats = pioCounter

	//stats of process cpu.
	cpuStat := &CpuStatDetail{}
	cpuPercent, cpErr := pCtx.CPUPercent()
	if cpErr != nil {
		mtlog.Errorf("Unable to fetch CPU used percent:%v", cpErr)
	} else {
		cpuStat.CpuPercent = cpuPercent
	}
	cpuTime, ctErr := pCtx.Times()
	if ctErr != nil {
		mtlog.Errorf("Unable to fetch CPU time detail:%v", ctErr)
	} else {
		cpuStat.User = cpuTime.User
		cpuStat.System = cpuTime.System
	}
	pStats.CpuStats = cpuStat

	//stats of process memory.
	memStat := &MemoryStatDetail{}

	memPercent, mErr := pCtx.MemoryPercent()
	if mErr != nil {
		mtlog.Errorf("Unable to fetch memory percent:%v", mErr)
	} else {
		memStat.MemPercent = memPercent
	}

	//runtime memory detail.
	memRuntime := runtime.MemStats{}
	runtime.ReadMemStats(&memRuntime)

	memStat.Alloc = memRuntime.Alloc
	memStat.TotalAlloc = memRuntime.TotalAlloc
	memStat.Sys = memRuntime.Sys
	memStat.Mallocs = memRuntime.Mallocs
	memStat.Frees = memRuntime.Frees
	memStat.HeapSys = memRuntime.HeapSys
	memStat.HeapIdle = memRuntime.HeapIdle
	memStat.NextGC = memRuntime.NextGC
	memStat.LastGC = memRuntime.LastGC
	memStat.PauseTotalNs = memRuntime.PauseTotalNs
	memStat.NumGC = memRuntime.NumGC

	pStats.MemoryStat = memStat

	//stats of process network.
	netioStats, nErr := pCtx.NetIOCounters(true)
	if nErr != nil {
		mtlog.Errorf("Unable to fetch network detail:%v", nErr)
	}
	for _, netioStat := range netioStats {
		netDetail := &NetworkStatDetail{}
		netDetail.InterfaceName = netioStat.Name
		netDetail.RxPackets = netioStat.PacketsRecv
		netDetail.RxBytes = netioStat.BytesRecv
		netDetail.TxPackets = netioStat.PacketsSent
		netDetail.TxBytes = netioStat.BytesSent
		netDetail.RxErrors = netioStat.Errin
		netDetail.TxErrors = netioStat.Errout
		netDetail.RxDropped = netioStat.Dropin
		netDetail.TxDropped = netioStat.Dropout

		pStats.Network = append(pStats.Network, netDetail)
	}

	processDetail.Stats = pStats

	runtimeDetail := &RuntimeDetail{}
	runtimeDetail.NumOfGoRoutines = runtime.NumGoroutine()
	runtimeDetail.NumOfCpu = runtime.NumCPU()

	processDetail.Runtime = runtimeDetail

	return processDetail, nil
}
