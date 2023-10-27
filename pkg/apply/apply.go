package apply

import (
	"fmt"
	"time"

	"github.com/openshift/rebase/pkg/carry"
	"github.com/openshift/rebase/pkg/git"
)

type Apply struct {
	log           *carry.Log
	from          string
	repositoryDir string
}

func NewApply(from, repositoryDir string) *Apply {
	return &Apply{
		log:           carry.NewLog(from, repositoryDir),
		from:          from,
		repositoryDir: repositoryDir,
	}
}

func (c *Apply) Run() error {
	// this applies the steps from https://github.com/openshift/kubernetes/blob/master/REBASE.openshift.md
	repository, err := git.OpenGit(c.repositoryDir)
	if err != nil {
		return err
	}
	// TODO: add fetching remotes
	commits, err := c.log.GetCommits(repository)
	if err != nil {
		return fmt.Errorf("Error reading carries: %w", err)
	}
	branchName := fmt.Sprintf("rebase-%s", time.Now().Format(time.DateOnly))
	if err := repository.CreateBranch(branchName, "refs/remotes/upstream/master"); err != nil {
		return fmt.Errorf("Error creating rebase branch: %w", err)
	}
	if err := repository.Merge("openshift/master"); err != nil {
		return fmt.Errorf("Error creating rebase branch: %w", err)
	}
	for _, c := range commits {
		if err := repository.CherryPick(c.Hash.String()); err != nil {
			// stop on first error for now...
			return err
		}
	}
	return nil
}
