// Copyright © 2017 Adam Kunicki <kunickiaj@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/99designs/keyring"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	debugMode    bool
	cfgFile      string
	config Config
)

var version string // set at build time

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "beer",
	Version: version,
	Short:   "CLI for managing your JIRA / Gerrit / git workflow.",
	Long: `Beer Review is a CLI for managing your JIRA <--> Gerrit workflow.

It can be used to create new tickets, work on existing ones, and submit reviews.

For example:
	$ beer brew ABC-123

	This will branch from the current HEAD and set the branch name to ABC-123 and add
	an empty commit with the JIRA ID and JIRA summary text as the first line of
	the commit message.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.WithField("error", err).Fatal("Fatal Error")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.beer.yaml)")
	RootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enables debug messages")
	RootCmd.PersistentFlags().Bool("dry-run", false, "Parses command syntax but does not make changes to JIRA or git")
	RootCmd.PersistentFlags().String("jira-url", "", "URL of JIRA server. Should end with slash")
	RootCmd.PersistentFlags().String("jira-username", "", "JIRA username")
	RootCmd.PersistentFlags().String("jira-password", "", "JIRA password")
	RootCmd.PersistentFlags().String("gerrit-url", "", "Gerrit SSH URL")
	RootCmd.PersistentFlags().String("review-tool", "gerrit", "Tool for publishing reviews, e.g. Gerrit")

	_ = viper.BindPFlag("jira.url", RootCmd.PersistentFlags().Lookup("jira-url"))
	_ = viper.BindPFlag("jira.username", RootCmd.PersistentFlags().Lookup("jira-username"))
	_ = viper.BindPFlag("jira.password", RootCmd.PersistentFlags().Lookup("jira-password"))
	_ = viper.BindPFlag("gerrit.url", RootCmd.PersistentFlags().Lookup("gerrit-url"))
	_ = viper.BindPFlag("reviewTool", RootCmd.PersistentFlags().Lookup("review-tool"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Init log level
	if debugMode {
		log.SetLevel(log.DebugLevel)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.WithField("error", err).Fatal("Unable to find home directory")
			os.Exit(1)
		}

		// Search config in home directory with name ".beer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".beer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.WithField("config", viper.ConfigFileUsed()).Debug("Using config file")
	}

	log.WithField("config_keys", viper.AllKeys()).Debug("Configuration keys")

	if err := viper.Unmarshal(&config); err == nil {
		log.WithField("config", config).Debug("Parsed config")
	}

	// last step, check os keychain for credentials
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: "beer", // ref: https://github.com/99designs/keyring/issues/44
	})

	// if users have existing config files with a password, let's inform them to migrate their config
	if len(config.Jira.Password) > 0 {
		log.Warn("Plaintext password detected in beer config, please remove it from the file.")
		log.Warn("You will be prompted for your password so that it can be stored securely in your OS keychain instead.")
	}

	i, err := ring.Get("jira-password")
	if errors.Is(err, keyring.ErrKeyNotFound) {
		password, _ := credentials()
		i = keyring.Item{
			Key: "jira-password",
			Data: []byte(password),
		}
		_ = ring.Set(i)
	} else if err != nil {
		log.WithError(err).Fatal("Unable to access keychain")
	}

	config.Jira.Password = string(i.Data)
}

func credentials() (string, error) {
	fmt.Print("Enter Jira Password or API token: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}
