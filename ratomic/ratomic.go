package ratomic

import (
	"context"
	"fmt"
	"time"
)

type optionsCtxKey string
type hybridDriverCtxKey string

type ratomic struct {
	driver  Driver
	options Options
}

func WithRatomic(ctx context.Context, driver Driver, options Options) context.Context {
	newCtx := context.WithValue(ctx, optionsCtxKey(""), &ratomic{
		driver:  driver,
		options: options,
	})

	if options.UseFilter {
		newCtx = context.WithValue(newCtx, hybridDriverCtxKey(""), NewLocalDriver(driver.KeyPrefix(), 0))
	}

	return newCtx
}

type Options struct {
	NumRetry    int
	Delay       time.Duration
	UseFilter   bool // first checks local locks if true.
	FilterDelay time.Duration
}

func driver(ctx context.Context) Driver {
	val := ctx.Value(optionsCtxKey(""))
	if val == nil {
		panic("WithRatomic must be called.")
	}

	return val.(*ratomic).driver
}

func options(ctx context.Context) Options {
	val := ctx.Value(optionsCtxKey(""))
	if val == nil {
		return Options{
			NumRetry:    10,
			Delay:       50 * time.Millisecond,
			UseFilter:   true,
			FilterDelay: 10 * time.Millisecond,
		}
	}

	return val.(*ratomic).options
}

// obtain locks.
// atomic, key order not matter if you use Redis.
func Lock(ctx context.Context, keys ...string) *RatomicError {
	dri := driver(ctx)

	keysWithPrefix := make([]string, 0, len(keys))
	for i := range keys {
		keysWithPrefix = append(keysWithPrefix, dri.KeyPrefix().Merge(keys[i]))
	}

	options := options(ctx)

	if ld, ok := ctx.Value(hybridDriverCtxKey("")).(Driver); ok {
		for i := 0; i < 1+options.NumRetry; i++ {
			replyCnt, _ := ld.MSetNX(keysWithPrefix...)

			if isAlreadyLocked(replyCnt) {
				time.Sleep(options.FilterDelay)
				continue
			}

			defer ld.Del(keysWithPrefix...)
			break
		}
	}

	var lastError *RatomicError
	for i := 0; i < 1+options.NumRetry; i++ {
		replyCnt, err := dri.MSetNX(keysWithPrefix...)
		if err != nil {
			lastError = newRatomicError(err, false, "")
			if err.ShouldRetry == false {
				break
			}
		}

		if isAlreadyLocked(replyCnt) {
			lastError = newRatomicError(ErrBusy, true, "")
			time.Sleep(options.Delay)
			continue
		} else {
			// has the lock here.
			lastError = nil
			break
		}
	}
	if lastError != nil {
		return lastError
	}

	return nil
}

// release locks.
// atomic, key order not matter if you use Redis.
func Unlock(ctx context.Context, keys ...string) *RatomicError {
	dri := driver(ctx)

	keysWithPrefix := make([]string, 0, len(keys))
	for i := range keys {
		keysWithPrefix = append(keysWithPrefix, dri.KeyPrefix().Merge(keys[i]))
	}

	retryConf := options(ctx)

	var lastError *RatomicError
	for i := 0; i < 1+retryConf.NumRetry; i++ {
		replyCnt, err := dri.Del(keysWithPrefix...)
		if err != nil {
			lastError = newRatomicError(err, false, "")
			if err.ShouldRetry == false {
				break
			} else {
				continue
			}
		}

		if int64(len(keys)) != replyCnt {
			fmt.Printf("ratomic: key length is %d but exists locks of just %d of them. the remaining locks have been deleted.\n", len(keys), replyCnt)
			lastError = newRatomicError(ErrCountNotMatch, false, fmt.Sprintf("key length is %d but exists locks of just %d of them. the remaining locks have been deleted.", len(keys), replyCnt))
			break
		} else {
			break
		}
	}
	if lastError != nil {
		return lastError
	}

	return nil
}

func isAlreadyLocked(setNxResult int64) bool {
	return setNxResult < 1
}

func isReleaseSuccess(delResult int64) bool {
	return delResult > 0
}
