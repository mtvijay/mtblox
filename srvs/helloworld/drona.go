// NewDronaCtx
package main

import (
	"context"
	"sync"
	"time"
)//


// Operation types supported
const (
	DefaultNumberOfHandlers     = 11

	StatsUpdateTicker = 5 * time.Second // timer for updating client for stats
	FailPostTimeout   = 2 * time.Minute
)

type DronaRequest struct {
	sync.RWMutex

	// If cancelContext is set it can be used to cancel some operations
	cancelContext context.Context
	cancelFunc    context.CancelFunc

	// Object that needs to be downloaded
	name      string
	// Status of Download, we convert here to string because this
	// field is going to be json marshalled
	status string

	// result to be sent out
	result chan *DronaRequest

	startTime, endTime time.Time

	// generated SasURI TTL
	Duration time.Duration
}

type DronaCtx struct {
	reqChan  chan *DronaRequest
	respChan chan *DronaRequest

	// Number of handlers
	noHandlers int

	// add waitGroups here
	wg *sync.WaitGroup

	// Also open the quit channel so that we can bail
	quitChan chan bool
}

//Keep working till we are told otherwise
func (ctx *DronaCtx) ListenAndServe() {
	for {
		select {
		case req := <-ctx.reqChan:
			_ = ctx.handleRequest(req)

		case <-ctx.quitChan:
			_ = ctx.handleQuit()
			return
		}
	}
}

func (ctx *DronaCtx) handleRequest(req *DronaRequest) error {
	var err error

	return err
}

func (ctx *DronaCtx) handleQuit() error {
	return nil
}


func NewDronaCtx(name string, noHandlers int) (*DronaCtx, error) {
	dSync := DronaCtx{}

	// Setup the load value
	dSync.noHandlers = noHandlers
	if noHandlers == 0 {
		dSync.noHandlers = DefaultNumberOfHandlers
	}

	wg := new(sync.WaitGroup)
	dSync.wg = wg

	// Finally make channels
	dSync.reqChan = make(chan *DronaRequest, dSync.noHandlers)
	dSync.respChan = make(chan *DronaRequest, dSync.noHandlers)
	dSync.quitChan = make(chan bool)

	// Initialize syncer handlers and start listening
	for i := 0; i < dSync.noHandlers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dSync.ListenAndServe()
		}()
	}

	return &dSync, nil
}
