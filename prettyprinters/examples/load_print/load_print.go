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

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/asteris-llc/converge/prettyprinters/graphviz/providers"
)

func main() {
	log.SetOutput(ioutil.Discard)
	g, err := load.Load(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	printer := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), providers.ResourcePreparer()))
	dotCode, err := printer.Show(g)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(dotCode)
}
