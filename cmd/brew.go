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
	"io/ioutil"
	"os"
	"regexp"

	jira "github.com/andygrunwald/go-jira"
	"github.com/docker/docker/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/libgit2/git2go.v26"
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

const testingStatusKey = "Testing Status"
const docImpactKey = "Doc Impact"

var description string
var docImpact bool
var issueType jiraIssueType
var projectKey string
var summary string
var testingStatus bool

func init() {
	RootCmd.AddCommand(brewCmd)

	brewCmd.Flags().StringVarP(&projectKey, "project", "p", "", "JIRA project key, e.g. SDC, SDCE")
	brewCmd.Flags().VarP(&issueType, "issue-type", "t", "Issue type to create")
	brewCmd.Flags().StringVarP(&summary, "summary", "s", "", "Issue summary")
	brewCmd.Flags().StringVarP(&description, "description", "d", "", "Issue detailed description. If not specified defaults to summary")
	brewCmd.Flags().BoolVarP(&docImpact, "doc-impact", "x", false, "When included, sets the Doc Impact field to 'Yes'")
	brewCmd.Flags().BoolVarP(&testingStatus, "testing-status", "q", false, "When present, indicates extended testing is required.")
}

func brew(cmd *cobra.Command, args []string) {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	username := viper.GetString("jira.username")
	password := viper.GetString("jira.password")
	jiraURL := viper.GetString("jira.url")

	jiraClient, _ := jira.NewClient(nil, jiraURL)
	jiraClient.Authentication.SetBasicAuth(username, password)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repo, err := git.OpenRepository(cwd)
	if err != nil {
		panic(err)
	}

	defer repo.Free()

	var issue *jira.Issue
	if len(args) > 0 {
		issueKey := args[0]

		// Fetch details for existing issue
		issue, _, err = jiraClient.Issue.Get(issueKey, nil)
		if err != nil {
			fmt.Printf("Error fetching issue: %s\n", err)
			return
		}

		if dryRun {
			return
		}

	} else {
		// Creating a new JIRA
		if summary == "" {
			fmt.Println("When creating a new issue, an issue summary is required.")
			return
		}

		if description == "" {
			description = summary
		}

		if dryRun {
			fmt.Printf("Dry Run -- Summary: %s, Description %s\n", summary, description)
			return
		}

		// Create the issue
		if len(projectKey) == 0 {
			projectKey, err = getProjectKey(repo)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		metaProject, err := createMetaProject(jiraClient, projectKey)
		if err != nil {
			fmt.Println(err)
			return
		}

		metaIssueType, err := createMetaIssueType(metaProject, string(issueType))
		if err != nil {
			fmt.Println(err)
			return
		}

		fieldsConfig := make(map[string]string)
		fieldsConfig["Project"] = projectKey
		fieldsConfig["Issue Type"] = string(issueType)
		fieldsConfig["Summary"] = summary
		fieldsConfig["Description"] = description

		fields, err := metaIssueType.GetAllFields()
		if err != nil {
			fmt.Println(err)
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
			fmt.Println(err)
			return
		}
		fmt.Println(issue)

		var res *jira.Response
		issue, res, err = jiraClient.Issue.Create(issue)
		if err != nil {
			fmt.Println(err)
			fmt.Println(bodyToString(res))
			return
		}
		issue, _, err = jiraClient.Issue.Get(issue.Key, nil)
		if err != nil {
			fmt.Printf("error fetching issue details: %s\n", err)
			return
		}
	}

	err = checkout(repo, issue)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func bodyToString(res *jira.Response) string {
	bytes, _ := ioutil.ReadAll(res.Body)
	bodyStr := string(bytes)
	return bodyStr
}

func checkout(repo *git.Repository, issue *jira.Issue) error {
	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}

	// Check only for local branches
	branch, err := repo.LookupBranch(issue.Key, git.BranchLocal)
	newBranch := false
	// If it doesn't exist then create it
	if branch == nil || err != nil {
		newBranch = true

		head, err := repo.Head()
		if err != nil {
			return err
		}

		headCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		branch, err = repo.CreateBranch(issue.Key, headCommit, false)
		if err != nil {
			return err
		}
	}

	defer branch.Free()

	// Get tree for the branch
	commit, err := repo.LookupCommit(branch.Target())
	if err != nil {
		return err
	}

	defer commit.Free()

	tree, err := repo.LookupTree(commit.TreeId())
	if err != nil {
		return err
	}

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		return err
	}

	// Set the head to point to the new branch
	repo.SetHead("refs/heads/" + issue.Key)

	headCommit, err := repo.LookupCommit(branch.Target())
	if err != nil {
		return err
	}

	signature, err := repo.DefaultSignature()
	if err != nil {
		return err
	}

	if newBranch {
		fmt.Printf("\nfields: %+v", issue)
		commitMessage := fmt.Sprintf("%s. %s", issue.Key, issue.Fields.Summary)
		_, err = repo.CreateCommit("refs/heads/"+issue.Key, signature, signature, commitMessage, tree, headCommit)
	}
	return nil
}

func getProjectKey(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		return "", err
	}

	var depth uint
	for depth < 5 {
		re := regexp.MustCompile("^([a-zA-Z]{3,})(-[0-9]+)")
		message := commit.Message()
		match := re.FindStringSubmatch(message)
		if len(match) == 3 && len(match[1]) > 0 {
			fmt.Printf("Inferred project key: %s, override with --project if incorrect\n", match[1])
			return match[1], nil
		}
		depth++
		commit = commit.Parent(0)
	}

	defer commit.Free()
	return "", errors.New("Wasn't able to infer a project key")
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