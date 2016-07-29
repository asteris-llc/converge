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

package file

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
)

type statResult struct {
	Stat os.FileInfo
	Err  error
}

//ValidatePath checks that this filepath is usable.
//Used to screen mistakes like the file not existing
func ValidatePath(path string) (statResult, error) {
	info, err := os.Stat(path)
	result := statResult{info, err}
	return result, every(path, []FileOp{
		stat2interface(FileExist),
		stat2interface(ValidStat),
	}, true)
}

func FileExist(result statResult) error {
	if os.IsNotExist(result.Err) {
		return fmt.Errorf("file %q does not exit\n", result.Stat.Name())
	}
	return nil
}

func ValidStat(result statResult) error {
	return result.Err
}

//every checks that some interface is valid for every function
type FileOp func(thing interface{}) error

func every(thing interface{}, checks []FileOp, failFast bool) error {

	var merr error
	for _, op := range checks {
		err := op(thing)
		if err != nil {
			merr := multierror.Append(merr, err)
			if failFast {
				return merr
			}
		}
	}
	return merr
}

//str2interface takes a function that checks a string and returns a func that
// checks an interface
func stat2interface(Func func(result statResult) error) func(thing interface{}) error {
	return func(thing interface{}) error {
		result, ok := thing.(statResult)
		if !ok {
			return fmt.Errorf("cannot check: %+v with function: %+v. Expected \"string\" found %T",
				thing,
				Func,
				thing,
			)
		}
		return Func(result)
	}
}

//str2interface takes a function that checks a string and returns a func that
// checks an interface
func str2interface(Func func(path string) error) func(thing interface{}) error {
	return func(thing interface{}) error {
		str, ok := thing.(string)
		if !ok {
			return fmt.Errorf("cannot check: %+v with function: %+v. Expected \"string\" found %T",
				thing,
				Func,
				thing,
			)
		}
		return Func(str)
	}
}
