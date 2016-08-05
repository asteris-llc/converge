package helpers

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

func SquashCheck(status1 string, willChange1 bool, err1 error, status2 string, willChange2 bool, err2 error) (string, bool, error) {
	s, c := fmt.Sprintf("%s\n%s", status1, status2), willChange1 || willChange2
	e := err1
	if e != nil {
		if err2 != nil {
			e = multierror.Append(err1, err2)
		}
	} else {
		if err2 != nil {
			e = err2
		}
	}
	return s, c, e
}
