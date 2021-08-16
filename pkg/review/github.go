package review

type GitHubReview struct {
	Meta
}

func NewGitHubReview(title string, description string, reviewers []string, baseBranch string, isDraft bool) Review {
	return &GitHubReview{
		Meta: Meta{
			Title:       title,
			Description: description,
			Reviewers:   reviewers,
			BaseBranch:  baseBranch,
			IsDraft:     isDraft,
		},
	}
}

func (g GitHubReview) Publish() error {
	return &NotImplementedError{}
}

func (g GitHubReview) Merge() error {
	return &NotImplementedError{}
}
