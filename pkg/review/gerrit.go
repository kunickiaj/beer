package review

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	log "github.com/sirupsen/logrus"
)

type GerritReview struct {
	Meta
}

func NewGerritReview(title string, description string, reviewers []string, baseBranch string, isDraft bool) Review {
	return &GerritReview{
		Meta: Meta{
			Title:       title,
			Description: description,
			Reviewers:   reviewers,
			BaseBranch:  baseBranch,
			IsDraft:     isDraft,
		},
	}
}

func (g GerritReview) Publish() error {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repo, err := git.PlainOpenWithOptions(cwd, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		panic(err)
	}

	var ref string
	if g.IsDraft {
		ref = fmt.Sprintf("%s%%wip", g.BaseBranch)
	} else {
		ref = g.BaseBranch
	}

	head, _ := repo.Head()
	refspec := fmt.Sprintf("%s:refs/for/%s", head.Name(), ref)
	if len(g.Reviewers) > 0 {
		refspec = fmt.Sprintf("%s%%r=%s", refspec, strings.Join(g.Reviewers, ",r="))
	}
	log.WithField("refspec", refspec).Debug("Using refspec")

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(refspec)},
	})
	if err != nil {
		log.WithError(err).Error("Error publishing review")
		os.Exit(1)
	}
	log.Info("Published review")
	return nil
}

func (g GerritReview) Merge() error {
	return nil
}
