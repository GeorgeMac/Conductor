package conductor

import (
	"github.com/e-dard/signalman"
	"io"
	"sync"
	"syscall"
)

// Conductor
//
// Nano-Framework to orchestrate a clean shutdown of services, ensure
// routines have finished and close registered closers.
type Conductor struct {
	wg   *sync.WaitGroup
	mu   *sync.RWMutex
	cl   []io.Closer
	errf func(error)
}

// Constructor a new Conductor with a function for
// handling errors on calls to Close (errf).
func NewConductor(errf func(error)) (cond *Conductor) {
	return &Conductor{
		wg:   &sync.WaitGroup{},
		mu:   &sync.RWMutex{},
		errf: errf,
	}
}

// Go sends a function in a go routine and register it
// against the Conductor. This will ensure that the
// routine has finished before exiting the program.
func (c *Conductor) Go(f func()) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		f()
	}()
}

// RegisterClosers safely ensures that the io.Closer is
// closed before the program exits and that any errors
// returned on Close are delegated to the handler function.
func (c *Conductor) RegisterClosers(cl ...io.Closer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.cl = append(c.cl, cl...)
}

// FinishAsap begins the finish process. This will ensure
// no new closers can be registered, then that all recorded
// routines have finished and finally that all registered
// io.Closers have been closed.
func (c *Conductor) FinishAsap() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.wg.Wait()

	for _, cc := range c.cl {
		if err := cc.Close(); err != nil {
			c.errf(err)
		}
	}
}

// FinishOnSignals performs the same graceful shutdown as FinishAsap.
// However, it registers the shutdown against syscall.Signals and
// blocks the routine until one of those signals are called.
func (c *Conductor) FinishOnSignals(sigs ...syscall.Signal) {
	st := make(chan struct{})
	for _, sig := range sigs {
		signalman.Register(sig, func() error {
			c.FinishAsap()
			close(st)
			return nil
		})
	}
	<-st
}
