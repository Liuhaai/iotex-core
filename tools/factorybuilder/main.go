// Copyright (c) 2022 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "factorybuilder",
	Short: "factorybuilder is a command-line interface for build trie.db for archive mode",
}

func init() {
	RootCmd.AddCommand(cmdHeight)
	RootCmd.AddCommand(cmdSize)
	RootCmd.AddCommand(cmdStatus)
	RootCmd.HelpFunc()
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
