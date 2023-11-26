package interval

import (
	"errors"
	"time"
)

var errMaxRetriesReached = errors.New("exceeded retry limit")

var MaxRetries = 20

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// Do keeps trying the function until the second argument
// returns false, or no error is returned.
func Do(sec time.Duration, fn Func) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if MaxRetries != -1 && attempt > MaxRetries {
			return errMaxRetriesReached
		}
		if err != nil {
			time.Sleep(sec * time.Second) // wait x seconds
		}
	}
	return err
}

// IsMaxRetries checks whether the error is due to hitting the
// maximum number of retries or not.
func IsMaxRetries(err error) bool {
	return err == errMaxRetriesReached
}
