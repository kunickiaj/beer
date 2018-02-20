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

	"github.com/docker/docker/cli"
	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v26"
)

var tasteCmd = &cobra.Command{
	Use:   "taste",
	Short: "Push a review for the current branch, optionally specifying reviewers by email.",
	Long:  ``,
	Run:   taste,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cli.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		return nil
	},
}

var draft bool
var reviewers []string

func init() {
	RootCmd.AddCommand(tasteCmd)

	tasteCmd.Flags().BoolVarP(&draft, "draft", "d", false, "Setting this flag will post a draft review")
	tasteCmd.Flags().StringArrayVarP(&reviewers, "reviewers", "r", nil, "Comma separated list of email ids of reviewers to add")
}

func taste(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	isDraft, _ := cmd.Flags().GetBool("draft")
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
	if isDraft {
		ref = "drafts"
	} else {
		ref = "for"
	}

	refspec := fmt.Sprintf("HEAD:refs/%s/%s", ref, "master")
	if len(reviewers) > 0 {
		refspec = fmt.Sprintf("%s%%r=%s", refspec, strings.Join(reviewers, ",r="))
	}

	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		panic(err)
	}

	callbacks := &git.RemoteCallbacks{
		CredentialsCallback:      credentialsCallback,
		CertificateCheckCallback: certificateCheckCallback,
	}

	err = remote.ConnectPush(callbacks, nil, make([]string, 0))
	if err != nil {
		panic(err)
	}

	pushOpts := &git.PushOptions{}
	refspecs := []string{refspec}
	err = remote.Push(refspecs, pushOpts)

}

func credentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	ret, cred := git.NewCredSshKeyFromAgent(username)
	return git.ErrorCode(ret), &cred
}

func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}
