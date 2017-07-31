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

	"github.com/docker/docker/cli"
	"github.com/spf13/cobra"
)

var brewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Work on an existing JIRA or create a new ticket. Not specifying an ISSUE_ID creates a new JIRA.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("brew called with arg", args[0])
	},
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
