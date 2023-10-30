package apply

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/openshift/rebase/pkg/carry"
	"github.com/openshift/rebase/pkg/git"
	"k8s.io/klog/v2"
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
		if err := repository.CherryPick(c.Hash.String()); err == nil {
			continue
		}
		klog.Infof("Encountered problems picking %s:", c.Hash.String())
		// print git status
		if err := repository.Status(); err != nil {
			return err
		}
		if err := repository.AbortCherryPick(); err != nil {
			return err
		}
		klog.Infof("Looking for fixed carry for %s...", c.Hash.String())
		patch, err := findFixedCarry(c.Hash.String())
		if err != nil {
			klog.Infof("Carry https://github.com/openshift/kubernetes/commit/%s requires manual intervention!", c.Hash.String())
			return err
		}
		klog.Infof("Found %s, applying...", patch)
		if err := repository.Apply(patch); err != nil {
			return err
		}
	}
	return nil
}

// findFixedCarry looks for fixed carry patches. Returns path to a file containing
// the carry or error.
func findFixedCarry(carrySha string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	carryPath := path.Join(cwd, "carries", carrySha)
	if _, err := os.Stat(carryPath); err != nil {
		return "", err
	}
	return carryPath, nil
}
