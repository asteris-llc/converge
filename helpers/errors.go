package helpers

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

func MultiErrorAppend(errs ...error) error {
	//Filter out all the nil errors
	nonNilErrs := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}
	if len(nonNilErrs) == 0 {
		return nil
	} else if len(nonNilErrs) == 1 {
		return nonNilErrs[0]
	} else {
		e := multierror.Append(errs[0], errs[1:]...)
		e.ErrorFormat = multiErrorPrinter
		return e
	}
}
func multiErrorPrinter(errs []error) string {
	fmt.Printf("%+v\n", errs)
	errString := ""
	for _, err := range errs {
		errString = errString + "\n\terror: " + err.Error()
	}
	return fmt.Sprintf("%d errors occured!\n%s", len(errs), errString)
}
