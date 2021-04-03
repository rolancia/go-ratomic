package ratomic

import "errors"

type errorType uint64

var (
	ErrDriverError   = errors.New("ERR_DRIVER_ERROR")
	ErrCountNotMatch = errors.New("ERR_COUNT_NOT_MATCH")
	ErrBusy          = errors.New("ERR_BUSY")
)

func NewDriverError(err error, shouldRetry bool) *driverError {
	return &driverError{
		ShouldRetry: shouldRetry,
		Err:         err,
	}
}

func newRatomicError(err error, shouldRetry bool, hint string) *RatomicError {
	return &RatomicError{
		ShouldRetry: shouldRetry,
		Err:         err,
		Hint:        hint,
	}
}

type driverError struct {
	ShouldRetry bool
	Err         error
}

func (de *driverError) Error() string {
	return de.Err.Error()
}

type RatomicError struct {
	ShouldRetry bool
	Err         error
	Hint        string
}

func (re *RatomicError) Error() string {
	return re.Err.Error()
}
