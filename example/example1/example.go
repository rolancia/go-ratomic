package main

import (
	"context"
	"fmt"
	"github.com/rolancia/go-ratomic/ratomic"
	"time"
)

func main() {
	ctx := context.Background()

	localDriver := ratomic.NewLocalDriver("lock", 5*time.Millisecond)
	ctx = ratomic.WithRatomic(ctx, localDriver, ratomic.RetryConfig{
		NumRetry: 5,
		Delay:    1000 * time.Millisecond,
	})

	var syncVarKey = "user1"

	// first locking
	err := ratomic.Lock(ctx, syncVarKey)
	if err != nil {
		fmt.Println(err.Err.Error(), err.Hint)
	}
	fmt.Println("you've got the lock here.")

	// second locking, the lock doesn't be released.
	// error will occur.
	err = ratomic.Lock(ctx, syncVarKey)
	if err != nil {
		if err.Err == ratomic.ErrDriverError {
			fmt.Println("your redis error.")
		} else if err.Err == ratomic.ErrBusy {
			fmt.Println("busy though retried.")
		}
	}

	// release first lock
	err = ratomic.Unlock(ctx, syncVarKey)
	if err != nil {
		if err.Err == ratomic.ErrDriverError {
			fmt.Println("your redis error.")
		} else if err.Err == ratomic.ErrCountNotMatch {
			fmt.Println("lock count and released count mismatched. maybe your code problem. (deadlock)")
		}
	}

	// third locking, the lock released before, it would be okay.
	err = ratomic.Lock(ctx, syncVarKey)
	if err == nil {
		fmt.Println("you've got the lock.")
	}
	defer ratomic.Unlock(ctx, syncVarKey) // never use without error checking. you will not find deadlock.
}
