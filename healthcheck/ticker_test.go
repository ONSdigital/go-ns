package healthcheck

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	shortGap = time.Duration(600 * time.Millisecond)
	longGap  = time.Duration(2 * time.Second)
)

type elfCheck struct {
	svc          string
	pause        time.Duration
	failingCalls int
	calls        []time.Time
	mutex        sync.Mutex
}

func (e *elfCheck) Healthcheck() (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.calls = append(e.calls, time.Now())
	logData := log.Data{"svc": e.svc, "calls": e.calls, "allow_fails": e.failingCalls}

	if e.pause != 0 {
		log.Info("HCk SLEEPING", logData)
		time.Sleep(e.pause)
	}

	if len(e.calls) > e.failingCalls {
		log.Info("HCk CALLED ok", logData)
		return e.svc, nil
	}
	log.Info("HCk CALLED bad", logData)
	return e.svc, errors.New(e.svc + " failed you")
}

func TestUnitTickerSuccess(t *testing.T) {

	stateChangeChan := make(chan bool, 1)
	requestCheckChan := make(chan bool, 1)
	startTime := time.Now()
	stateChanges := 0

	mcli := elfCheck{svc: "ok-svc", failingCalls: 0}

	ticker := NewTickerWithAlerts(longGap, shortGap, stateChangeChan, requestCheckChan, &mcli)
	defer ticker.Close()

	Convey("given a healthcheck that always succeeds", t, func() {

		Convey("when the ticker has been created", func() {

			Convey("then we should see success and it happened soon after shortGap", func() {

				ctx, cancel := context.WithTimeout(context.Background(), longGap)
				defer cancel()
				var newState bool

				// block until timeout, cancel or state-change
				select {
				case <-ctx.Done():
				case newState = <-stateChangeChan:
					stateChanges++
				}

				mcli.mutex.Lock()
				defer mcli.mutex.Unlock()

				So(ctx.Err(), ShouldBeNil)
				So(len(mcli.calls), ShouldEqual, 1)
				So(newState, ShouldBeTrue)
				So(stateChanges, ShouldEqual, 1)
				So(time.Now(), ShouldHappenBetween, startTime.Add(shortGap), startTime.Add(shortGap*2))
			})

			Convey("then we should be able to wait for another healthcheck after longGap", func() {

				ctx, cancel := context.WithTimeout(context.Background(), longGap)
				var ctxErr error

				// background check for length of mcli.calls
				// - cancel context (caught in `select` below) when mcli.calls is (expected) 2
				go func() {
					for _ = range time.Tick(50 * time.Millisecond) {
						mcli.mutex.Lock()
						numCalls := len(mcli.calls)
						mcli.mutex.Unlock()
						if numCalls > 1 {
							cancel()
							return
						}
					}
				}()

				// block until timeout, cancel or state-change
				select {
				case <-ctx.Done():
					ctxErr = ctx.Err()
				case <-stateChangeChan:
					// should not get here, so stateChanges should remain at 1
					stateChanges++
				}

				mcli.mutex.Lock()
				defer mcli.mutex.Unlock()

				So(ctxErr, ShouldEqual, context.Canceled)
				So(len(mcli.calls), ShouldEqual, 2)
				So(stateChanges, ShouldEqual, 1)
				So(mcli.calls[1], ShouldHappenBetween, startTime.Add(longGap), mcli.calls[0].Add(longGap+shortGap))

			})
		})
	})
}

func TestUnitTickerFailure(t *testing.T) {

	stateChangeChan := make(chan bool, 1)
	requestCheckChan := make(chan bool, 1)
	startTime := time.Now()

	Convey("given a healthcheck that always fails", t, func() {

		mcli := elfCheck{svc: "failing-svc", failingCalls: 1000}

		ticker := NewTickerWithAlerts(longGap, shortGap, stateChangeChan, requestCheckChan, &mcli)
		defer ticker.Close()

		Convey("when the ticker is created", func() {

			ctx, cancel := context.WithTimeout(context.Background(), longGap+shortGap)
			defer cancel()
			var stateChanged = false

			// block until timeout, cancel or state-change
			select {
			case <-ctx.Done():
			case <-stateChangeChan:
				// should not get here
				stateChanged = true
			}

			Convey("then we have timed out waiting for success, and see several calls to Healthcheck", func() {

				mcli.mutex.Lock()
				defer mcli.mutex.Unlock()

				So(stateChanged, ShouldBeFalse)
				So(ctx.Err(), ShouldNotBeNil)
				So(len(mcli.calls), ShouldBeGreaterThan, 1)
				So(mcli.calls[0], ShouldHappenBetween, startTime.Add(shortGap), startTime.Add(shortGap*2))
				So(mcli.calls[1], ShouldHappenBetween, mcli.calls[0].Add(shortGap), mcli.calls[0].Add(shortGap*3))
				So(time.Now(), ShouldHappenAfter, startTime.Add(shortGap*2))
			})
		})
	})

	Convey("given a healthcheck that fails at first, then recovers", t, func() {

		mcli := elfCheck{svc: "ok-eventually-svc", failingCalls: 2}

		Convey("when the ticker is created", func() {

			ticker := NewTickerWithAlerts(longGap, shortGap, stateChangeChan, requestCheckChan, &mcli)
			defer ticker.Close()

			ctx, cancel := context.WithTimeout(context.Background(), longGap+shortGap+shortGap)
			defer cancel()
			var newState bool
			var stateChanges int

			// block until timeout, cancel or state-change
			select {
			case <-ctx.Done():
			case newState = <-stateChangeChan:
				stateChanges++
			}

			Convey("then we eventually got success, and see several calls to Healthcheck", func() {

				mcli.mutex.Lock()
				defer mcli.mutex.Unlock()

				So(ctx.Err(), ShouldBeNil)
				So(newState, ShouldBeTrue)
				So(stateChanges, ShouldEqual, 1)
				So(len(mcli.calls), ShouldBeGreaterThan, 1)
			})
		})
	})

	Convey("given a healthcheck that times out", t, func() {
		mcli := elfCheck{
			svc:          "timeout-svc",
			pause:        3 * longGap,
			failingCalls: 1000,
		}

		ticker := NewTickerWithAlerts(longGap, shortGap, stateChangeChan, requestCheckChan, &mcli)
		defer ticker.Close()

		ctx, cancel := context.WithTimeout(context.Background(), mcli.pause-1*time.Second)
		defer cancel()
		var stateChanges int

		// block until timeout, cancel or state-change
		select {
		case <-ctx.Done():
		case <-stateChangeChan:
			// should not get here
			stateChanges++
		}

		Convey("when the ticker is created", func() {

			Convey("then we eventually time out, and see a single call to Healthcheck", func() {

				mcli.mutex.Lock()
				defer mcli.mutex.Unlock()

				So(ctx.Err(), ShouldNotBeNil)
				So(stateChanges, ShouldEqual, 0)
				So(len(mcli.calls), ShouldEqual, 1)
			})
		})
	})

}
