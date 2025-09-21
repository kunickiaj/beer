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
	"io"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"

	jira "github.com/andygrunwald/go-jira"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gogitConfig "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var brewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Work on an existing JIRA or create a new ticket. Not specifying an ISSUE_ID creates a new JIRA.",
	Long:  ``,
	Run:   brew,
	Args:  cobra.MaximumNArgs(1),
}

type FileType struct {
	Name      string
	Extension string
}

const (
	testingStatusKey = "Testing Status"
	docImpactKey     = "Doc Impact"
)

var description string
var docImpact bool
var issueType string
var components []string
var labels []string
var projectKey string
var summary string
var testingStatus bool
var autoMetadata bool

func init() {
	RootCmd.AddCommand(brewCmd)

	brewCmd.Flags().StringVarP(&projectKey, "project", "p", "", "JIRA project key, e.g. SDC, SDCE")
	brewCmd.Flags().StringVarP(&issueType, "issue-type", "t", "Bug", "Issue type to create, e.g. Bug, 'New Feature', etc. This varies by project.")
	brewCmd.Flags().StringVarP(&summary, "summary", "s", "", "Issue summary")
	brewCmd.Flags().StringVarP(&description, "description", "d", "", "Issue detailed description. If not specified defaults to summary")
	brewCmd.Flags().BoolVarP(&docImpact, "doc-impact", "x", false, "When included, sets the Doc Impact field to 'Yes'")
	brewCmd.Flags().BoolVarP(&testingStatus, "testing-status", "q", false, "When present, indicates extended testing is required.")
	brewCmd.Flags().StringSliceVarP(&components, "components", "c", nil, "Sets the components field of the issue. Can be a comma separated list.")
	brewCmd.Flags().StringSliceVarP(&labels, "labels", "l", nil, "Sets the labels field of the issue. Can be a comma separated list.")
	brewCmd.Flags().BoolVarP(&autoMetadata, "auto", "a", false, "Enable automatic metadata detection for components and/or labels.")
}

func brew(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	transport := jira.BasicAuthTransport{
		Username:  config.Jira.Username,
		Password:  config.Jira.Password,
	}
	jiraClient, _ := jira.NewClient(transport.Client(), config.Jira.URL)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		panic(err)
	}

	// Get user struct for logged in user
	jiraUser, _, err := jiraClient.User.GetSelf()
	if err != nil {
		panic(err)
	}

	var issue *jira.Issue
	if len(args) > 0 {
		issueKey := args[0]

		// Fetch details for existing issue
		issue, _, err = jiraClient.Issue.Get(issueKey, nil)
		if err != nil {
			log.WithField("error", err).Fatal("Error fetching issue")
			return
		}

		if dryRun {
			return
		}

		// Ensure issue is assigned to self
		assignee := map[string]interface{}{"fields": map[string]interface{}{"assignee": jiraUser}}

		response, err := jiraClient.Issue.UpdateIssue(issue.ID, assignee)
		if err != nil {
			log.
				WithError(err).
				WithField("response", bodyToString(response)).
				Fatal("Failed to update issue")
			return
		}

	} else {
		// Creating a new JIRA
		if summary == "" {
			log.Error("When creating a new issue, an issue summary is required.")
			return
		}

		if description == "" {
			description = summary
		}

		if dryRun {
			log.WithFields(log.Fields{"summary": summary, "description": description}).Info("Dry Run")
			return
		}

		// Create the issue
		if len(projectKey) == 0 {
			projectKey, err = getProjectKey(repo)
			if err != nil {
				log.WithField("error", err).Fatal()
				return
			}
		}

		metaProject, err := createMetaProject(jiraClient, projectKey)
		if err != nil {
			log.WithField("error", err).Fatal()
			return
		}

		metaIssueType, err := createMetaIssueType(metaProject, issueType)
		if err != nil {
			log.WithField("error", err).Fatal()
			return
		}

		fieldsConfig := map[string]string{
			"Project":     projectKey,
			"Issue Type":  issueType,
			"Summary":     summary,
			"Description": description,
			"Assignee":    assignee(jiraUser),
		}

		fields, err := metaIssueType.GetAllFields()
		if err != nil {
			log.WithField("error", err).Fatal()
			return
		}

		if _, ok := fields[testingStatusKey]; ok {
			testingStatusStr := "Not Required"
			if testingStatus {
				testingStatusStr = "Required"
			}
			fieldsConfig[testingStatusKey] = testingStatusStr
		}

		if _, ok := fields[docImpactKey]; ok {
			docImpactStr := "No"
			if docImpact {
				docImpactStr = "Yes"
			}
			fieldsConfig[docImpactKey] = docImpactStr
		}

		issue, err = jira.InitIssueWithMetaAndFields(metaProject, metaIssueType, fieldsConfig)
		if err != nil {
			log.WithField("error", err).Fatal()
			return
		}
		log.WithField("issue", issue).Debug("Initialized Issue")

		numComponents := len(components)
		if numComponents > 0 {
			jiraComponents := make([]*jira.Component, numComponents)
			for i, v := range components {
				jiraComponents[i] = &jira.Component{
					Name: v,
				}
			}
			issue.Fields.Components = jiraComponents
		}

		issue.Fields.Labels = labels


		issue = appendAutomaticMetadata(repo, issue)

		var res *jira.Response
		issue, res, err = jiraClient.Issue.Create(issue)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "response": bodyToString(res)}).Fatal()
			return
		}
		issue, _, err = jiraClient.Issue.Get(issue.Key, nil)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "key": issue.Key}).Fatal("Error fetching issue details")
			return
		}
	}

	err = checkout(repo, issue)
	if err != nil {
		log.WithField("error", err).Fatal()
		return
	}
}

func assignee(jiraUser *jira.User) string {
	if len(jiraUser.AccountID) > 0 {
		return jiraUser.AccountID
	} else {
		return jiraUser.Name
	}
}

func appendAutomaticMetadata(repo *git.Repository, issue *jira.Issue) *jira.Issue {
	if !autoMetadata {
		return issue
	}

	w, _ := repo.Worktree()
	files, err := w.Filesystem.ReadDir(".")
	if err != nil {
		log.WithError(err).Warn("metadata inference failed")
		return issue
	}

	terraformType := FileType{
		Name:      "Terraform",
		Extension: ".tf",
	}
	if containsFileType(files, terraformType) {
		component := &jira.Component{
			Name: terraformType.Name,
		}
		issue.Fields.Components = append(issue.Fields.Components, component)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		log.WithError(err).Warn("metadata inference failed")
		return issue
	}


	repoName := strings.TrimSuffix(path.Base(remote.Config().URLs[0]), ".git")
	repoComponent := &jira.Component{
		Name: repoName,
	}

	issue.Fields.Components = append(issue.Fields.Components, repoComponent)
	return issue
}

func containsFileType(files []os.FileInfo, fileType FileType) bool {
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), fileType.Extension) {
			return true
		}
	}
	return false
}

func bodyToString(res *jira.Response) string {
	bytes, _ := io.ReadAll(res.Body)
	bodyStr := string(bytes)
	return bodyStr
}

func checkout(repo *git.Repository, issue *jira.Issue) error {
	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	branch := fmt.Sprintf("refs/heads/%s", issue.Key)
	b := plumbing.ReferenceName(branch)

	// First try to checkout branch
	newBranch := false
	err = workTree.Checkout(&git.CheckoutOptions{Create: false, Force: false, Keep: true, Branch: b})
	if err != nil {
		// didn't exist so try to create it
		newBranch = true
		err = workTree.Checkout(&git.CheckoutOptions{Create: true, Force: false, Keep: true, Branch: b})
	}

	if err != nil {
		return err
	}

	if newBranch {
		// If an issue description was provided, that isn't simply a repeat of the summary, we'll automatically include
		// it in the commit message after the break.
		extendedDescription := ""
		if issue.Fields.Summary != issue.Fields.Description {
			extendedDescription = fmt.Sprintf("\n\n%s", issue.Fields.Description)
		}
		commitMessage := fmt.Sprintf("%s. %s%s", issue.Key, issue.Fields.Summary, extendedDescription)
		usr, err := user.Current()
		if err != nil {
			return err
		}
		gitConfig, err := os.Open(path.Join(usr.HomeDir, ".gitconfig"))
		if err != nil {
			return err
		}
		decoder := gogitConfig.NewDecoder(gitConfig)
		decodedConfig := gogitConfig.New()
		err = decoder.Decode(decodedConfig)

		if err != nil {
			return err
		}

		userSection := decodedConfig.Section("user")
		_, err = workTree.Commit(commitMessage, &git.CommitOptions{
			Author: &object.Signature{
				Name:  userSection.Option("name"),
				Email: userSection.Option("email"),
				When:  time.Now(),
			},
		})
		return err
	}
	return nil
}

func getProjectKey(repo *git.Repository) (string, error) {
	ref, err := repo.Head()
	if err != nil {
		return "", errors.Wrap(err, "couldn't get HEAD reference when searching for project key")
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("couldn't get git log from hash '%s'", ref.Hash().String()))
	}

	var depth uint
	commit, err := cIter.Next()
	for depth < 5 {
		if err != nil {
			return "", errors.Wrap(err, "log iterator had no next commit")
		}
		re := regexp.MustCompile("^([a-zA-Z]{3,})(-[0-9]+)")
		message := commit.Message
		log.WithField("message", commit.Message).Debug("Searching commit message...")
		match := re.FindStringSubmatch(message)
		if len(match) == 3 && len(match[1]) > 0 {
			log.WithField("project_key", match[1]).Info("Inferred project key; override with --project if incorrect")
			return match[1], nil
		}
		depth++
		commit, err = cIter.Next()
	}

	return "", errors.Wrap(err, "wasn't able to infer a project key")
}

func createMetaProject(jira *jira.Client, projectKey string) (*jira.MetaProject, error) {
	meta, _, err := jira.Issue.GetCreateMeta(projectKey)
	if err != nil {
		return nil, err
	}

	metaProject := meta.GetProjectWithKey(projectKey)
	if metaProject == nil {
		return nil, fmt.Errorf("could not find project with key %s", projectKey)
	}

	return metaProject, nil
}

func createMetaIssueType(metaProject *jira.MetaProject, issueType string) (*jira.MetaIssueType, error) {
	MetaIssueType := metaProject.GetIssueTypeWithName(issueType)
	if MetaIssueType == nil {
		return nil, fmt.Errorf("could not find issuetype %s, available types are %#v", issueType, getAllIssueTypeNames(metaProject))
	}
	return MetaIssueType, nil
}

func getAllIssueTypeNames(project *jira.MetaProject) []string {
	var foundIssueTypes []string
	for _, m := range project.IssueTypes {
		foundIssueTypes = append(foundIssueTypes, m.Name)
	}
	return foundIssueTypes
}
