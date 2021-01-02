package metrics

import (
	"encoding/json"
	"sync/atomic"
)

type counter interface {
	Inc()
	Add(uint64)
	Set(uint64)
	Label(string)
	Name() string
	Value() interface{}
}

//
type Zcounter struct {
	label string
	x     uint64
}

type zcounterJson struct {
	Label string `json:",omitempty"`
	X     uint64
}

func (a Zcounter) MarshalJSON() ([]byte, error) {
	b := zcounterJson{}
	b.Label = a.label
	b.X = a.x
	return json.Marshal(b)
}

func (zc *Zcounter) Inc() {
	atomic.AddUint64(&zc.x, 1)
}

func (zc *Zcounter) Add(deltaX uint64) {
	atomic.AddUint64(&zc.x, deltaX)
}

func (zc *Zcounter) Set(baseV uint64) {
	atomic.StoreUint64(&zc.x, baseV)
}

func (zc *Zcounter) Reset() {
	atomic.StoreUint64(&zc.x, 0)
}

func (zc *Zcounter) Value() interface{} {
	return atomic.LoadUint64(&zc.x)
}

// fixme: Ideally would like to have write/read lock
func (zc *Zcounter) Label(l string) {
	zc.label = l
}

func (zc *Zcounter) Name() string {
	return zc.label
}

//
type ZcounterSmall struct {
	label string
	x     uint32
}

//
type ZcounterLarge struct {
	label string
	xh    uint64
	xl    uint64
}
