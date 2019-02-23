/*
	package errors is define some error for this system.
*/

package errors

import (
	"fmt"
)

type ErrorS struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func (e ErrorS) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

func New(format string, args ...interface{}) ErrorS {
	return ErrorS{Code: 1, Msg: fmt.Sprintf(format, args)}
}
