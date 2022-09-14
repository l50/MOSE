// Copyright 2020 National Technology & Engineering Solutions of Sandia, LLC (NTESS).
// Under the terms of Contract DE-NA0003525 with NTESS,
// the U.S. Government retains certain rights in this software.

package cmd

import (
	"github.com/rs/zerolog/log"

	"os"

	"github.com/master-of-servers/mose/pkg/chefutils"
	"github.com/master-of-servers/mose/pkg/moseutils"

	"github.com/spf13/cobra"
)

// CMTARGETCHEF specifies the CM tool that we are targeting.
const CMTARGETCHEF = "chef"

// chefCmd represents the chef command
var chefCmd = &cobra.Command{
	Use:   "chef",
	Short: "Create MOSE payload for chef",
	Long:  `Create MOSE payload for chef`,
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
}
