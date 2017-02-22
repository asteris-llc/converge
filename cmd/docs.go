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

package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

const hugoHeaderTemplate = `---
date: %q
title: %q
slug: %q
menu: { main: { parent: commands } }

---
`

func hugoFilePrepender(filename string) string {
	now := time.Now().Format(time.RFC3339)

	name := filepath.Base(filename)

	title := strings.Replace(strings.TrimSuffix(name, path.Ext(name)), "_", " ", -1)
	if title == "converge" { // need to change this to avoid a duplicate command
		title = "converge (root)"
	}

	slugReplacer := strings.NewReplacer(
		" ", "_",
		"(", "",
		")", "",
	)

	return fmt.Sprintf(
		hugoHeaderTemplate,
		now,
		title,
		slugReplacer.Replace(title)+"_command",
	)
}

func hugoLinkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "{{< ref \"" + strings.ToLower(base) + ".md\" >}}"
}

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:    "docs",
	Hidden: true,
	Short:  "Generate markdown documentation for commands",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll(viper.GetString("path"), os.FileMode(0755)); err != nil {
			logrus.WithError(err).Fatal("could not create man tree path")
		}

		if err := doc.GenMarkdownTreeCustom(RootCmd, viper.GetString("path"), hugoFilePrepender, hugoLinkHandler); err != nil {
			logrus.WithError(err).Fatal("could not generate markdown docs tree")
		}
	},
}

func init() {
	docsCmd.Flags().String("path", "markdown-docs", "path to generate docs into")

	genCmd.AddCommand(docsCmd)
}
