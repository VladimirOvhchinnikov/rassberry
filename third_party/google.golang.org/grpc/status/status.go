package status

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

func Errorf(code codes.Code, format string, a ...interface{}) error {
	return fmt.Errorf("%v: "+format, append([]interface{}{code}, a...)...)
}
