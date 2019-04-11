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
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v27"
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
}

func taste(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	isWIP, _ := cmd.Flags().GetBool("wip")
	reviewers, _ := cmd.Flags().GetStringArray("reviewers")

	if dryRun {
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repo, err := git.OpenRepository(cwd)
	if err != nil {
		panic(err)
	}

	defer repo.Free()

	var ref string
	if isWIP {
		ref = "master%wip"
	} else {
		ref = "master"
	}

	refspec := fmt.Sprintf("HEAD:refs/for/%s", ref)
	if len(reviewers) > 0 {
		refspec = fmt.Sprintf("%s%%r=%s", refspec, strings.Join(reviewers, ",r="))
	}
	log.WithField("refspec", refspec).Debug("Using refspec")

	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		panic(err)
	}
	log.WithField("remote", remote).Debug("Discovered remote")

	pushOpts := &git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
			SidebandProgressCallback: transportMessageCallback,
		},
	}
	refspecs := []string{refspec}
	err = remote.Push(refspecs, pushOpts)
	if err != nil {
		log.WithError(err).Error("Error pushing review")
		os.Exit(1)
	}
	log.Info("Pushed review")
}

func credentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	ret, cred := git.NewCredSshKeyFromAgent(username)
	return git.ErrorCode(ret), &cred
}

func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}

func transportMessageCallback(str string) git.ErrorCode {
	fmt.Printf(str)
	return 0
}
