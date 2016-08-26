// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// Prettyprint the errors
func multiErrorPrinter(errs []error) string {
	fmt.Printf("%+v\n", errs)
	errString := ""
	for _, err := range errs {
		errString = errString + "\n\terror: " + err.Error()
	}
	return fmt.Sprintf("%d errors occured!\n%s", len(errs), errString)
}
