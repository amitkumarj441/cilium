package envoy

import (
	"math"
	"sync"

	"github.com/cilium/cilium/pkg/lock"

	log "github.com/sirupsen/logrus"
)

// StreamControlCtx holds the state of a gRPC stream server instance we need to know about.
type StreamControlCtx struct {
	wg       sync.WaitGroup
	handlers []*StreamControl
}

func (ctx *StreamControlCtx) stop() {
	for _, ctrl := range ctx.handlers {
		ctrl.stopHandling()
	}
	// Wait for everyone to be done.
	ctx.wg.Wait()
}

// StreamControl implements a generic Envoy streamed gRPC API. API
// servers should derive from this and add API-specific functionality
// of their own.
// A 'sync.Cond' is used to signal the stream server goroutine that the current version
// of the resoure should be streamed. Envoy stream requests embed a version field that
// is interpreted as a NAK if it repeats the same version as the previous request, and as
// an ACK when it contains the version field from the latest response we sent over.
// 64-bit version numbers are assumed to never wrap around.
type StreamControl struct {
	name           string
	cond           sync.Cond // L mutex used to protect members below as well.
	handled        bool
	ackedVersion   uint64
	sentVersion    uint64
	currentVersion uint64
}

func makeStreamControl(name string) StreamControl {
	return StreamControl{
		name:           name,
		cond:           sync.Cond{L: &lock.Mutex{}},
		ackedVersion:   math.MaxUint64,
		sentVersion:    0,
		currentVersion: 1,
	}
}

// update stream control based on received 'version'. Returns true if
// the current version should be sent.
// Lock must be held.
func (ctrl *StreamControl) updateVersionLocked(version uint64) bool {
	// Bump current version UP to sync with the version history of the caller, if we receive
	// evidence for a version we have not sent yet. This can happen when we restart.
	if ctrl.sentVersion < version {
		ctrl.currentVersion = version + 1
		log.Debug("Envoy: ", ctrl.name, " version bumped to ", ctrl.currentVersion)
	}

	// Roll back acked version if this is a NAK (== version is the same as the
	// previous acked version).
	if version == ctrl.ackedVersion {
		// NAK, send again
		ctrl.sentVersion = version
		log.Debug("Envoy: ", ctrl.name, " NAK received, sending again after version ", version)
	} else {
		log.Debug("Envoy: ", ctrl.name, " ACK received: ", version)
	}
	ctrl.ackedVersion = version // remember the last acked version

	return ctrl.currentVersion > ctrl.sentVersion
}

func (ctrl *StreamControl) updateVersion(version uint64) bool {
	ctrl.cond.L.Lock()
	defer ctrl.cond.L.Unlock()
	return ctrl.updateVersionLocked(version)
}

// 'handler()' is called to send the current version if it is later than 'version'
// Starts a handler gorouting on demand tracked by 'ctx'.
func (ctrl *StreamControl) startHandler(ctx *StreamControlCtx, handler func() error) {
	ctrl.handled = true
	ctx.handlers = append(ctx.handlers, ctrl)

	ctx.wg.Add(1)
	go func() {
		defer ctx.wg.Done()

		log.Debug("Envoy: Starting stream handler for: ", ctrl.name)

		ctrl.cond.L.Lock()
		for {
			// Quit waiting if we should stop or have something to send.
			// cond.Wait automatically unlocks and locks again.
			for ctrl.handled && ctrl.sentVersion == ctrl.currentVersion {
				ctrl.cond.Wait()
			}
			if !ctrl.handled {
				break // end handling
			}
			// Send the current version
			if handler() == nil {
				ctrl.sentVersion = ctrl.currentVersion
			}
		}
		ctrl.cond.L.Unlock()

		log.Debug("Envoy: Stream handler stopped: ", ctrl.name)
	}()
}

// 'handler()' is called to send the current version if it is later than 'version'
// Starts a handler gorouting on demand tracked by 'ctx'.
func (ctrl *StreamControl) handleVersion(ctx *StreamControlCtx, version uint64, handler func() error) {
	ctrl.cond.L.Lock()
	mustSend := ctrl.updateVersionLocked(version)

	// Start the handler if needed
	if mustSend && !ctrl.handled {
		ctrl.startHandler(ctx, handler)
	}
	ctrl.cond.L.Unlock()

	if mustSend {
		log.Debug("Envoy: ", ctrl.name, " version ", ctrl.currentVersion, " queued.")
		ctrl.cond.Signal()
	} else {
		log.Debug("Envoy: ", ctrl.name, " version ", ctrl.currentVersion, " acked.")
	}
}

func (ctrl *StreamControl) stopHandling() {
	ctrl.cond.L.Lock()
	ctrl.handled = false
	ctrl.cond.L.Unlock()
	ctrl.cond.Signal()
}

// f is called while the lock is held
func (ctrl *StreamControl) bumpVersionFunc(f func()) {
	ctrl.cond.L.Lock()
	f()
	ctrl.currentVersion++
	ctrl.cond.L.Unlock()
	ctrl.cond.Signal()
}

func (ctrl *StreamControl) bumpVersion() {
	ctrl.cond.L.Lock()
	ctrl.currentVersion++
	ctrl.cond.L.Unlock()
	ctrl.cond.Signal()
}
