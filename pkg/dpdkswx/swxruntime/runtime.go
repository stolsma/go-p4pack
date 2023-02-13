// Copyright 2022 - Sander Tolsma. All rights reserved
// SPDX-License-Identifier: Apache-2.0

package swxruntime

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"github.com/stolsma/go-p4pack/pkg/dpdkswx/eal"
	"github.com/stolsma/go-p4pack/pkg/logging"
)

var log logging.Logger

func init() {
	// keep the logger up to date, also after new log config
	logging.Register("dpdkswx/swxruntime", func(logger logging.Logger) {
		log = logger
	})
}

// MainCtx is the main lcore context and is supplied to the function running in the main lcore.
type MainCtx struct {
	Value interface{}   // Value is a user-specified context. Change it and it will persist across function invocations
	ch    chan *mainJob // channel to receive functions to execute.
	done  bool          // signal to kill current thread
}

type mainJob struct {
	fn  func(*MainCtx) error
	ret chan<- error
}

// Stores all the DPDK SWX thread info and gives control functions
type Runtime struct {
	mainCtx MainCtx
	running bool
}

func Create() (rt *Runtime) {
	rt = &Runtime{}
	return rt
}

func CreateAndStart(args []string) (rt *Runtime, nArgs int, err error) {
	rt = Create()
	nArgs, err = rt.Start(args)
	return
}

// ExecOnMainAsync asynchronously executes given function on main lcore.
func (rt *Runtime) ExecOnMainAsync(ret chan error, fn func(*MainCtx) error) <-chan error {
	if !rt.running {
		ret <- fmt.Errorf("swx runtime not initialized")
		return ret
	}

	log.Info("Entering 'execute on main' !")
	ctx := &rt.mainCtx
	ctx.ch <- &mainJob{fn, ret}
	return ret
}

// ExecOnMain executes function on main lcore.
func (rt *Runtime) ExecOnMain(fn func(*MainCtx) error) error {
	return <-rt.ExecOnMainAsync(make(chan error, 1), fn)
}

// ErrMainCorePanic is an error returned by ExecOnMain(Async) in case given function panics.
type ErrMainCorePanic struct {
	Pc []uintptr
	R  interface{} // value returned by recover()
}

// Error implements error interface.
func (e *ErrMainCorePanic) Error() string {
	return fmt.Sprintf("panic on main core %v", e.R)
}

// Unwrap returns error value if it was supplied to panic() as an argument.
func (e *ErrMainCorePanic) Unwrap() error {
	if err, ok := e.R.(error); ok {
		return err
	}
	return nil
}

func fPrintStack(w io.Writer, pc []uintptr) {
	frames := runtime.CallersFrames(pc)
	var frame runtime.Frame
	more := true
	for n := 0; more; n++ {
		frame, more = frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			continue
		}
		fmt.Fprintf(w, " -- (%2d): %s, %s:%d\n", n, frame.Function, frame.File, frame.Line)
	}
}

// FprintStack prints PCs into w.
func (e *ErrMainCorePanic) FprintStack(w io.Writer) {
	fPrintStack(w, e.Pc)
}

// panicCatcher launches function and returns possible panic as an error.
func panicCatcher(fn func(*MainCtx) error, ctx *MainCtx) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		pc := make([]uintptr, 64)
		n := 0

		for {
			if n = runtime.Callers(1, pc); n < len(pc) {
				break
			}
			pc = append(pc, make([]uintptr, len(pc))...)
		}
		// this function is called from runtime package, so to unwind the stack we may skip (1) runtime.Callers
		// function, (2) this caller function
		err = &ErrMainCorePanic{pc[:n], r}
	}()
	err = fn(ctx)
	return err
}

// to run as main core job listener
func (rt *Runtime) mainCoreJobListener() int {
	// get context
	ctx := &rt.mainCtx

	log.Info("Listening for jobs to execute on main core !")

	// run loop
	for job := range ctx.ch {
		log.Info("Got job to execute on main core !")
		err := panicCatcher(job.fn, ctx)
		if job.ret != nil {
			job.ret <- err
		}
		if ctx.done {
			break
		}
	}

	log.Info("Main core job listener is done!")
	return 0
}

// launch thread_main dataplane runners on all dpdk worker lcores, should be run at Start only!!!
func (rt *Runtime) launchWorkers() error {
	// init per-lcore dataplane contexts
	if err := ThreadsInit(); err != nil {
		return err
	}

	// launch every EAL worker lcore with dataplane runners
	return ThreadsStart()
}

// Initializes EAL as in rte_eal_init and starts the SWX pipeline workers.
// Options for EAL are specified in a parsed command line string. It also star a Main Lcore handler that waits for
// (Go lang) executable functions to be run in the main Lcore context.
// Returns number of parsed args and error.
func (rt *Runtime) Start(args []string) (n int, err error) {
	var wg sync.WaitGroup

	log.Infof("Starting swxruntime with dpdkargs: %v", args)

	wg.Add(1)
	go func() {
		// we should initialize EAL and run EAL threads in a separate goroutine because its thread is going to be acquired
		// by EAL and become main lcore thread!!!!!!
		runtime.LockOSThread()
		log.Info("swxruntime: lock this go thread to dpdk main core (lockOSThread)")

		// initialize EAL
		if n, err = eal.RteEalInit(args); err != nil {
			wg.Done()
			return
		}

		// threadLaunch initializes and runs thread_main on all worker lcores.
		if err = rt.launchWorkers(); err != nil {
			wg.Done()
			return
		}

		// create job communication channel for main core listener. MUST be created before calling wg.Done() to prevent
		// race condition when execution of "executeOnMain" before job channel is ready!
		ctx := &rt.mainCtx
		ctx.ch = make(chan *mainJob)
		log.Info("Ready for job listening on main core")
		wg.Done()

		// mainCoreJobListener will block until it stops running on main lcore, see [Thread.Stop]
		rt.mainCoreJobListener()
	}()
	wg.Wait()
	rt.running = true

	log.Info("swxruntime started!")

	return
}

// Stop sends signal to all threads to finish execution.
// Warning: it will block until all lcore threads finish execution.
func (rt *Runtime) Stop() (err error) {
	log.Info("Stopping swxruntime!")

	// stop DPDK SWX workers
	err = rt.ExecOnMain(func(ctx *MainCtx) error {
		return ThreadsStop()
	})
	if err != nil {
		return
	}

	log.Info("swxruntime worker threads stopped, stopping main thread!")

	// quit main LCore function
	err = rt.ExecOnMain(func(ctx *MainCtx) error {
		// TODO Does this work????
		ctx.done = true
		return nil
	})
	rt.running = false

	log.Info("swxruntime stopped (worker and main threads)!")
	return
}

func (rt *Runtime) IsRunning() bool {
	return rt.running
}
