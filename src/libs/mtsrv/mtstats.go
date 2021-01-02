package mtsrv

import (
	"libs/mtlog"
	"time"
)

func (s *Server) statusUpdate() int {
	s.Lock()
	defer s.Unlock()
	for _, met := range s.metList {
		metStr, _ := met.Marshal()
		mtlog.Tracef("%v %v", time.Now(), metStr)
	}
	return 1
}

type MetricType int

const (
	metCounter MetricType = 1
	metGauge   MetricType = 2
	metState   MetricType = 3
)

type MetricEntry struct {
	mtype  MetricType
	mkey   string
	mvalue interface{}
}

// genric collection structure for metrics..
type MetricCollect struct {
	CollectionName string
	metrics        []MetricEntry
}

// Add counter type
func (mc *MetricCollect) AddCounter(key string, val interface{}) {
	mEntry := MetricEntry{mtype: metCounter, mkey: key, mvalue: val}
	mc.metrics = append(mc.metrics, mEntry)
}

func (mc *MetricCollect) AddGauge(key string, val interface{}) {
	mEntry := MetricEntry{mtype: metGauge, mkey: key, mvalue: val}
	mc.metrics = append(mc.metrics, mEntry)
}

// The stats structure implemented by the client
type MtMetric interface {
	// readLock/lock or just return nothing, it depends on the
	// how the locking is implemented. Before calling Marshal
	// server code will call the lock. Its client responsbilites
	// to do safe implementation.
	Lock()
	Unlock()

	// Client can implement marshaller to get the collection
	Marshal() (*MetricCollect, error)
}

//
// systemServicePeridoic
//
func (s *Server) systemServicePeriodic() {
	c := time.Tick(1 * time.Minute)
	for now := range c {
		mtlog.Tracef("%v %v", now, s.statusUpdate())
	}
}
