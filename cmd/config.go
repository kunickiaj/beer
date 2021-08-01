package cmd

import "strings"

type Config struct {
	Jira JiraConfig
	Gerrit GerritConfig
	GitHub GithubConfig
	ReviewTool ReviewTool
}

type ReviewTool string

func (r ReviewTool) Normalize() ReviewTool {
	return ReviewTool(strings.ToLower(string(r)))
}

const (
	Gerrit ReviewTool = "gerrit"
	GitHub ReviewTool = "github"
)

type Defaults struct {
	Branch string
	ReviewTool ReviewTool
}
// JiraConfig configuration structure for JIRA
type JiraConfig struct {
	URL      string
	Username string
	Password string
}

// GerritConfig configuration structure for gerrit
type GerritConfig struct {
	URL string
}

type GithubConfig struct {

}
