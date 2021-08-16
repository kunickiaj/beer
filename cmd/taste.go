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
	"github.com/spf13/viper"

	"github.com/kunickiaj/beer/pkg/review"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tasteCmd = &cobra.Command{
	Use:   "taste",
	Short: "Push a review for the current branch, optionally specifying reviewers by email.",
	Long:  ``,
	Run:   taste,
	Args:  cobra.ExactArgs(0),
}

var wip bool
var reviewers []string

func init() {
	RootCmd.AddCommand(tasteCmd)

	tasteCmd.Flags().BoolVar(&wip, "wip", false, "Setting this flag will post a WIP review")
	tasteCmd.Flags().StringSliceVarP(&reviewers, "reviewers", "r", nil, "Comma separated list of email ids of reviewers to add")
	tasteCmd.Flags().String("branch", "main", "Target branch for review")

	_ = viper.BindPFlag("defaults.branch", tasteCmd.Flags().Lookup("branch"))
}

func taste(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	isWIP, _ := cmd.Flags().GetBool("wip")
	targetBranch := viper.GetString("defaults.branch")
	log.WithField("targetBranch", targetBranch).Debug("Determined target branch for comparison")
	reviewers, err := cmd.Flags().GetStringSlice("reviewers")

	if err != nil {
		log.WithError(err).Fatal("Could not parse reviewers")
	}

	log.WithField("reviewers", reviewers).Debug("Parsed reviewers")

	var r review.Review
	switch config.ReviewTool.Normalize() {
	case Gerrit:
		r = review.NewGerritReview("title", "description", reviewers, targetBranch, isWIP)
	case GitHub:
		r = review.NewGitHubReview("title", "description", reviewers, targetBranch, isWIP)
	default:
		log.WithField("reviewTool", config.ReviewTool).Fatal("This review tool is not yet supported")
	}

	if dryRun {
		return
	}

	err = r.Publish()
	if err != nil {
		log.WithError(err).Error("Failed to publish review")
	}
}
