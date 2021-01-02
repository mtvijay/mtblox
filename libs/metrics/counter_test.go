package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zededa/zedcloud/libs/zlog"
)

type testMetrics struct {
	Name  string
	Zctr1 Zcounter
	Zctr2 Zcounter
}

var itm1 testMetrics

func (tc testMetrics) GetName() string {
	return tc.Name
}

func (tc testMetrics) Print() {
	DefaultPrint(tc)
}

func (tc testMetrics) Log() {
	DefaultLogger(tc)
}

func TestGetCounter(t *testing.T) {
	cfg := zlog.LogConfig{
		Level:      0,
		Path:       "/tmp/",
		File:       "metrics.log",
		MaxBackups: 3,
		MaxSize:    100,
		MaxAge:     7,
		Compress:   true,
	}
	zlog.InitLogging(cfg)
	zlog.Tracef("Starting the test")
	itm1 = testMetrics{Name: "counter1"}
	LogCounters(&itm1, time.Second)
	itm1.Zctr1.Label("ctr1")
	itm1.Zctr1.Set(222)
}

func TestMetricSet(t *testing.T) {
	itm1.Zctr2.Label("ctr2")
	itm1.Zctr2.Set(10000)
	itm1.Zctr1.Add(2)
	assert.Equal(t, itm1.Zctr1.Value(), uint64(224))
}

func TestMetricCount(t *testing.T) {
	t.Skip("Has a data race. skipping until we can fix it")
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		itm1.Zctr2.Add(10)
	}
	assert.Equal(t, itm1.Zctr2.Value(), uint64(10600))
}

func TestMetricName(t *testing.T) {
	assert.Equal(t, itm1.Zctr2.Name(), "ctr2")
	assert.Equal(t, itm1.Zctr1.Name(), "ctr1")
}
