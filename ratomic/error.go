package ratomic

import "errors"

var (
	ErrDriverError   = errors.New("ERR_DRIVER_ERROR")
	ErrCountNotMatch = errors.New("ERR_COUNT_NOT_MATCH")
	ErrBusy          = errors.New("ERR_BUSY")
)

func NewDriverError(err error, shouldRetry bool) *DriverError {
	return &DriverError{
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

type DriverError struct {
	ShouldRetry bool
	Err         error
}

func (de *DriverError) Error() string {
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
