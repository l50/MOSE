/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/rs/zerolog/log"

	"github.com/l50/mose/pkg/chefutils"
	"github.com/l50/mose/pkg/moseutils"
	"os"

	"github.com/spf13/cobra"
)

const CMTARGETCHEF = "chef"

// chefCmd represents the chef command
var chefCmd = &cobra.Command{
	Use:   "chef",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		UserInput.CMTarget = CMTARGETCHEF
		UserInput.SetLocalIP()
		UserInput.GenerateParams()
		UserInput.GeneratePayload()
		UserInput.StartTakeover()
		ans, err := moseutils.AskUserQuestion("Is your target a chef workstation? ", UserInput.OSTarget)
		if err != nil {
			log.Fatal().Err(err).Msg("Quitting")
		}
		if ans {
			log.Info().Msg("Nothing left to do locally, continue all remaining activities on the target workstation.")
			os.Exit(0)
		}

		ans, err = moseutils.AskUserQuestion("Is your target a chef server? ", UserInput.OSTarget)
		if err != nil {
			log.Fatal().Msg("Quitting")
		}
		if ans {
			chefutils.SetupChefWorkstationContainer(UserInput)
			os.Exit(0)
		} else {
			log.Error().Msg("Invalid chef target")
		}
	},
}

func init() {
	rootCmd.AddCommand(chefCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chefCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chefCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
