package review

type Review interface {
	Publish() error
	Merge() error
}

type Meta struct {
	Title       string   // First line of the commit message
	Description string   // The rest of the commit message or PR description
	Reviewers   []string // Reviewers to notify
	BaseBranch  string   // The branch to merge into
	IsDraft     bool     // Is this a draft request?
}
