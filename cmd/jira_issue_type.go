package cmd

import "fmt"

type jiraIssueType string

var jiraIssueTypes = map[string]bool{
	"Bug":         true,
	"New Feature": true,
	"Task":        true,
	"Improvement": true,
}

func newJiraIssueType(val string, p *string) *jiraIssueType {
	*p = val
	return (*jiraIssueType)(p)
}

func (t *jiraIssueType) Set(val string) error {
	if !jiraIssueTypes[val] {
		return fmt.Errorf("%s is not a valid JIRA issue type", val)
	}
	*t = jiraIssueType(val)
	return nil
}

func (t *jiraIssueType) String() string {
	return string(*t)
}

func (t *jiraIssueType) Type() string {
	keys := make([]string, 0, len(jiraIssueTypes))
	for k := range jiraIssueTypes {
		keys = append(keys, k)
	}
	return fmt.Sprint(keys)
}
