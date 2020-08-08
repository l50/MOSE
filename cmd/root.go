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
	"github.com/l50/mose/pkg/system"
	"github.com/l50/mose/pkg/userinput"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	UserInput userinput.UserInput
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "github.com/master-of-servers/mose",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/.settings.yaml)")

	rootCmd.PersistentFlags().StringP("osarch", "a", "", "Architecture that the target CM tool is running on")
	rootCmd.PersistentFlags().StringP("cmd", "c", "", "Command to run on the targets")
	rootCmd.PersistentFlags().Bool("debug", false, "Display debug output")
	rootCmd.PersistentFlags().Int("exfilport", 443, "Port used to exfil data from chef server (default 443 with ssl, 9090 without)")
	rootCmd.PersistentFlags().StringP("filepath", "f", "", "Output binary locally at <filepath>")
	rootCmd.PersistentFlags().StringP("fileupload", "u", "", "File upload option")
	rootCmd.PersistentFlags().StringP("localip", "l", "", "Local IP Address")
	rootCmd.PersistentFlags().StringP("payloadname", "m", "my_cmd", "Name for backdoor payload")
	rootCmd.PersistentFlags().StringP("ostarget", "o", "linux", "Operating system that the target CM tool is on")
	rootCmd.PersistentFlags().Int("websrvport", 443, "Port used to serve payloads on (default 443 with ssl, 8090 without)")
	rootCmd.PersistentFlags().String("remoteuploadpath", "/root/.definitelynotevil", "Remote file path to upload a script to (used in conjunction with -fu)")
	rootCmd.PersistentFlags().StringP("rhost", "r", "", "Set the remote host for /etc/hosts in the chef workstation container (format is hostname:ip)")
	rootCmd.PersistentFlags().Bool("ssl", false, "Serve payload over TLS")
	rootCmd.PersistentFlags().Int("tts", 60, "Number of seconds to serve the payload")

	path := system.Gwd()
	rootCmd.PersistentFlags().String("payloads", filepath.Join(path, "payloads"), "Location of payloads output by mose")
	rootCmd.PersistentFlags().String("basedir", path, "Location of payloads output by mose")

	viper.BindPFlag("osarch", rootCmd.PersistentFlags().Lookup("osarch"))
	viper.BindPFlag("cmd", rootCmd.PersistentFlags().Lookup("cmd"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("exfilport", rootCmd.PersistentFlags().Lookup("exfilport"))
	viper.BindPFlag("filepath", rootCmd.PersistentFlags().Lookup("filepath"))
	viper.BindPFlag("fileupload", rootCmd.PersistentFlags().Lookup("fileupload"))
	viper.BindPFlag("localip", rootCmd.PersistentFlags().Lookup("localip"))
	viper.BindPFlag("payloadname", rootCmd.PersistentFlags().Lookup("payloadname"))
	viper.BindPFlag("ostarget", rootCmd.PersistentFlags().Lookup("ostarget"))
	viper.BindPFlag("websrvport", rootCmd.PersistentFlags().Lookup("websrvport"))
	viper.BindPFlag("remoteuploadpath", rootCmd.PersistentFlags().Lookup("remoteuploadpath"))
	viper.BindPFlag("rhost", rootCmd.PersistentFlags().Lookup("rhost"))
	viper.BindPFlag("ssl", rootCmd.PersistentFlags().Lookup("ssl"))
	viper.BindPFlag("tts", rootCmd.PersistentFlags().Lookup("tts"))

	viper.BindPFlag("payloads", rootCmd.PersistentFlags().Lookup("payloads"))
	viper.BindPFlag("basedir", rootCmd.PersistentFlags().Lookup("basedir"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find current directory.
		cur, err := os.Getwd()
		if err != nil {
			log.Error().Err(err).Msg("")
			os.Exit(1)
		}

		// Search config in home directory with name ".github.com/master-of-servers/mose" (without extension).
		viper.AddConfigPath(cur)
		viper.SetConfigType("yaml")
		viper.SetConfigName("settings")
	}
	//log.Debug().Msg(viper.ConfigFileUsed())
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Error reading in config file")
	}
	//log.Debug().Msgf("Using config file:", viper.ConfigFileUsed())

	err := viper.Unmarshal(&UserInput)

	if UserInput.Cmd == "" && UserInput.FileUpload == "" {
		log.Fatal().Msg("You must specify a CM target and a command or file to upload.")
	}

	if UserInput.Cmd != "" && UserInput.FileUpload != "" {
		log.Fatal().Msg("You must specify a CM target, a command or file to upload, and an operating system.")
	}

	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling config file")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if UserInput.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
