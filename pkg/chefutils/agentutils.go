// Copyright 2020 National Technology & Engineering Solutions of Sandia, LLC (NTESS).
// Under the terms of Contract DE-NA0003525 with NTESS,
// the U.S. Government retains certain rights in this software.

package chefutils

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/l50/mose/pkg/moseutils"

	"github.com/rs/zerolog/log"
)

// TargetAgents allows a user to select specific chef agents, or return them all as a []string
func TargetAgents(nodes []string, osTarget string) ([]string, error) {
	var targets []string
	if ans, err := moseutils.AskUserQuestion("Do you want to target specific chef agents? ", osTarget); ans && err == nil {
		reader := bufio.NewReader(os.Stdin)
		// Print the first discovered node (done for formatting purposes)
		log.Log().Msgf("%s", nodes[0])
		// Print the rest of the discovered nodes
		for _, node := range nodes[1:] {
			log.Log().Msgf(",%s", node)
		}
		log.Log().Msgf("\nPlease input the chef agents that you want to target using commas to separate them: ")
		text, _ := reader.ReadString('\n')
		targets = strings.Split(strings.TrimSuffix(text, "\n"), ",")
	} else if !ans && err == nil {
		// Target all of the agents
		return []string{"MOSEALL"}, nil
	} else if err != nil {
		return nil, errors.New("Quit")
	}
	return targets, nil
}
