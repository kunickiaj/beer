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

	jira "github.com/andygrunwald/go-jira"
	"github.com/docker/docker/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var brewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Work on an existing JIRA or create a new ticket. Not specifying an ISSUE_ID creates a new JIRA.",
	Long:  ``,
	Run:   brew,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequiresMaxArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
}

var issueTypeVar jiraIssueType

func init() {
	RootCmd.AddCommand(brewCmd)

	brewCmd.Flags().VarP(&issueTypeVar, "issue-type", "t", "Issue type to create")
}

func brew(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	username := viper.GetString("jira.username")
	password := viper.GetString("jira.password")
	server := viper.GetString("jira.server")

	jiraClient, _ := jira.NewClient(nil, server)
	jiraClient.Authentication.SetBasicAuth(username, password)

	if len(args) == 0 {
		// Creating a new JIRA
		if dryRun {
			fmt.Println("Would create new issue, but this is a dry run.")
		}
	}

	issueID := args[0]

	// Fetch details for existing issue
	issue, _, err := jiraClient.Issue.Get(issueID, nil)
	if err != nil {
		fmt.Println("Error fetching issue:", err)
	} else {
		fmt.Println("Fetched issue details:", issue.Key, issue.Fields.Summary)
	}
}
