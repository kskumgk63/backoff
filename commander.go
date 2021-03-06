package backoff

import (
	"errors"
	"fmt"
	"time"
)

// Commander .
type Commander struct{ options }

// NewCommander .
func NewCommander(opts ...Option) Commander {
	var (
		defaultTimeout       = 1*time.Minute + 5*time.Second
		defaultTimeoutErrMsg = "Ends the exponential backoff because of timeout"
		defaultMaxWaitTime   = 32 * time.Second
		defaultDebugMode     = false
		defaultErrPrint      = func(err error) {
			fmt.Println(err)
		}
		defaultAbortLoop = func(error) bool { return false }
		defaultTimePrint = func(d time.Duration) {
			fmt.Printf("waiting %fs...\n", d.Seconds())
		}
	)
	options := options{
		timeout:           defaultTimeout,
		timeoutErrMessage: defaultTimeoutErrMsg,
		maxWaitTime:       defaultMaxWaitTime,
		debugMode:         defaultDebugMode,
		debugPrint:        defaultErrPrint,
		abortLoop:         defaultAbortLoop,
		timePrint:         defaultTimePrint,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return Commander{options}
}

// Exec f() in the backoff
func (cmd Commander) Exec(f func() error) error {
	done := make(chan struct{})
	go cmd.backoffLoop(done, f)

	select {
	case <-done:
		return nil
	case <-time.After(cmd.timeout):
		return errors.New(cmd.timeoutErrMessage)
	}
}

func (cmd Commander) backoffLoop(done chan struct{}, f func() error) {
	exponent := 1
	for {
		err := f()
		if err == nil {
			done <- struct{}{}
			return
		}
		if cmd.abortLoop(err) {
			if cmd.debugMode {
				cmd.debugPrint(err)
			}
			done <- struct{}{}
			return
		}
		exponentSecond := time.Duration(pow2(exponent)) * time.Second
		d := min(exponentSecond+randomMilliSecond(), cmd.maxWaitTime)
		if cmd.debugMode {
			cmd.debugPrint(err)
			cmd.timePrint(d)
		}
		time.Sleep(d)
		exponent++
	}
}
