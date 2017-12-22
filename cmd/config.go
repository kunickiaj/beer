package cmd

// JiraConfig configuration structure for JIRA
type JiraConfig struct {
	server   string
	username string
	password string
}

// GerritConfig configuration structure for gerrit
type GerritConfig struct {
	url string
}
