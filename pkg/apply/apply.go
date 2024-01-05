package apply

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/openshift/rebase/pkg/carry"
	"github.com/openshift/rebase/pkg/git"
	"github.com/openshift/rebase/pkg/github"
	"github.com/openshift/rebase/pkg/utils"
	"k8s.io/klog/v2"
)

type Apply struct {
	log           *carry.Log
	from          string
	repositoryDir string
}

const (
	carryAction = "<carry>"
	dropAction  = "<drop>"
	skipPatch   = "<skip>"
)

var (
	actionRE = regexp.MustCompile(`UPSTREAM: (?P<action>[<>\w]+):`)
)

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
	// TODO:
	// 1. add fetching remotes
	// 2. checkout upstream/master and print its sha
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
		klog.V(2).Infof("Processing %s: %q", c.Hash.String(), utils.FormatMessage(c.Message))
		action := actionFromMessage(utils.FormatMessage(c.Message))
		if number, err := strconv.Atoi(action); err == nil {
			merged, err := github.IsMerged(number)
			if err != nil {
				// TODO: abort only after 2-3 errors, maybe?
				return fmt.Errorf("Failed reading merge state for %s: %q: %w", c.Hash.String(), utils.FormatMessage(c.Message), err)
			}
			if merged {
				klog.V(2).Infof("Skipping commit %s - merged upstream.", c.Hash.String())
				continue
			}
			// in all other cases we just continue to carry a patch
			action = carryAction
		}
		switch action {
		case carryAction:
			if err := carryFlow(repository, c); err != nil {
				// TODO: abort only after 2-3 errors, maybe?
				return err
			}
		case dropAction:
			klog.Warningf("Skipping drop commit %s.", c.Hash.String())
		default:
			klog.Errorf("Unkown action on commit %s: %s", c.Hash.String(), action)
		}
	}
	return nil
}

// carryFlow implements the carry action
func carryFlow(repository git.Git, commit *object.Commit) error {
	klog.V(2).Infof("Initiating carry flow for %s...", commit.Hash.String())
	if err := repository.CherryPick(commit.Hash.String()); err == nil {
		return nil
	}
	klog.Infof("Encountered problems picking %s:", commit.Hash.String())
	if err := repository.Status(); err != nil {
		return err
	}
	if err := repository.AbortCherryPick(); err != nil {
		return err
	}
	klog.V(2).Infof("Looking for a fixed carry")
	patch, skip, err := findFixedCarry(commit.Hash.String())
	if err != nil {
		// if the cherry-pick failed and there's no fixed carry try using:
		// git cherry-pick --strategy=recursive --strategy-option theirs
		if err := repository.RetryCherryPick(commit.Hash.String()); err == nil {
			klog.Warningf("Carry https://github.com/openshift/kubernetes/commit/%s was picked auto-magically \\o/ - make sure to double check it!", commit.Hash.String())
			return nil
		}
		if err := repository.AbortCherryPick(); err != nil {
			return err
		}
		klog.Errorf("Carry https://github.com/openshift/kubernetes/commit/%s requires manual intervention!", commit.Hash.String())
		return err
	}
	if skip {
		klog.Infof("Found skip patch %s.", patch)
		return nil
	}
	klog.Infof("Found %s, applying...", patch)
	if err := repository.Apply(patch); err != nil {
		if err := repository.AbortApply(); err != nil {
			klog.Errorf("Aborting apply failed: %v", err)
		}
		// if the apply failed, try using 3-way merge before failing
		if err := repository.Apply3Way(patch); err == nil {
			klog.Warningf("Current fix https://github.com/soltysh/rebase/tree/main/carries/%s was picked auto-magically \\o/ - make sure to double check it!", commit.Hash.String())
			return nil
		}
		if err := repository.AbortApply(); err != nil {
			klog.Errorf("Aborting apply failed: %v", err)
		}
		klog.Errorf("The current fix stopped working https://github.com/soltysh/rebase/tree/main/carries/%s and requires manual intervention!",
			commit.Hash.String())
		klog.Errorf("The original carry was https://github.com/openshift/kubernetes/commit/%s", commit.Hash.String())
		return err
	}
	return nil
}

// actionFromMessage parses the upstream action from commit message, returning
// which action to take on a commit
func actionFromMessage(message string) string {
	matches := actionRE.FindStringSubmatch(message)
	lastIndex := actionRE.SubexpIndex("action")
	if lastIndex < 0 {
		return ""
	}
	return matches[lastIndex]
}

// findFixedCarry looks for fixed carry patches. Returns path to a file containing
// the carry, information whether to skip it or not and an error.
func findFixedCarry(carrySha string) (string, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, err
	}
	carryPath := path.Join(cwd, "carries", carrySha)
	fileInfo, err := os.Stat(carryPath)
	if err != nil {
		return "", false, err
	}
	// empty fixed carry informs the patch was mislabeled
	return carryPath, fileInfo.Size() == 0, nil
}
