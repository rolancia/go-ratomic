package ratomic

import (
	"context"
	"fmt"
	"time"
)

type ratomicCtxKey string

type ratomic struct {
	driver      Driver
	retryConfig RetryConfig
}

func WithRatomic(ctx context.Context, driver Driver, retryConfig RetryConfig) context.Context {
	newCtx := context.WithValue(ctx, ratomicCtxKey(""), &ratomic{
		driver:      driver,
		retryConfig: retryConfig,
	})

	return newCtx
}

type RetryConfig struct {
	NumRetry int
	Delay    time.Duration
}

func driver(ctx context.Context) Driver {
	val := ctx.Value(ratomicCtxKey(""))
	if val == nil {
		panic("WithRatomic must be called.")
	}

	return val.(*ratomic).driver
}

func retryConfig(ctx context.Context) RetryConfig {
	val := ctx.Value(ratomicCtxKey(""))
	if val == nil {
		return RetryConfig{
			NumRetry: 10,
			Delay:    50 * time.Millisecond,
		}
	}

	return val.(*ratomic).retryConfig
}

// obtain locks.
// atomic, key order not matter if you use Redis.
func Lock(ctx context.Context, keys ...string) *RatomicError {
	dri := driver(ctx)

	keysWithPrefix := make([]string, 0, len(keys))
	for i := range keys {
		keysWithPrefix = append(keysWithPrefix, dri.KeyPrefix().Merge(keys[i]))
	}

	retryConf := retryConfig(ctx)

	var lastError *RatomicError
	for i := 0; i < 1+retryConf.NumRetry; i++ {
		replyCnt, err := dri.MSetNX(keysWithPrefix...)
		if err != nil {
			lastError = newRatomicError(err, false, "")
			if err.ShouldRetry == false {
				break
			}
		}

		if isAlreadyLocked(replyCnt) {
			lastError = newRatomicError(ErrBusy, true, "")
			time.Sleep(retryConf.Delay)
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

	retryConf := retryConfig(ctx)

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
