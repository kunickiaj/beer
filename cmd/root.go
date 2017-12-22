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
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var jiraConfig JiraConfig
var gerritConfig GerritConfig

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "beer",
	Short: "CLI for managing your JIRA / Gerrit / git workflow.",
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
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.beer.yaml)")
	RootCmd.PersistentFlags().Bool("dry-run", false, "Parses command syntax but does not make changes to JIRA or git")
	RootCmd.PersistentFlags().String("jira-server", "", "URL of JIRA server. Should end with slash")
	RootCmd.PersistentFlags().String("jira-username", "", "JIRA username")
	RootCmd.PersistentFlags().String("jira-password", "", "JIRA password")
	RootCmd.PersistentFlags().String("gerrit-url", "", "Gerrit SSH URL")

	viper.BindPFlag("jira.server", RootCmd.PersistentFlags().Lookup("jira-server"))
	viper.BindPFlag("jira.username", RootCmd.PersistentFlags().Lookup("jira-username"))
	viper.BindPFlag("jira.password", RootCmd.PersistentFlags().Lookup("jira-password"))
	viper.BindPFlag("gerrit.url", RootCmd.PersistentFlags().Lookup("gerrit-url"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".beer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".beer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	fmt.Println("All keys ", viper.AllKeys())

	if err := viper.UnmarshalKey("jira", &jiraConfig); err == nil {
		fmt.Println("Parsed JIRA config.")
	}

	if err := viper.UnmarshalKey("gerrit", &gerritConfig); err == nil {
		fmt.Println("Parsed Gerrit config.")
	}
}
