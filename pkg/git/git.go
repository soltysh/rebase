package git

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"

	gitv5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitv5object "github.com/go-git/go-git/v5/plumbing/object"
)

// Git provides an interface for interacting with a git repository.
type Git interface {
	// LogFromTag returns a list of carry commits from provided tag
	LogFromTag(tag string) ([]*gitv5object.Commit, error)
	// Commit returns commit for a given has
	Commit(hash plumbing.Hash) (*gitv5object.Commit, error)
	// CherryPick invokes the cherry-pick command
	CherryPick(sha string) error
}

// OpenGit opens path as a git repository, ensuring that remotes contain
// both upstream kubernetes and openshift remotes properly configured.
func OpenGit(path string) (Git, error) {
	klog.V(2).Infof("Using %s as git repository", path)
	repository, err := gitv5.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	gitRepo := &git{repository: repository}
	klog.V(2).Infof("Checking if openshift and upstream remotes are configured..", path)
	if err := gitRepo.checkRemotes(); err != nil {
		return nil, err
	}
	return gitRepo, nil
}

type git struct {
	repository *gitv5.Repository
}

// checkRemotes ensures both openshift and upstream remotes are properly configured
func (git *git) checkRemotes() error {
	for _, remote := range []struct {
		name string
		path string
	}{
		{
			name: "openshift",
			path: "github.com:openshift/kubernetes.git",
		},
		{
			name: "upstream",
			path: "github.com:kubernetes/kubernetes.git",
		},
	} {
		gitRemote, err := git.repository.Remote(remote.name)
		if err != nil {
			return err
		}
		config := gitRemote.Config()
		// URLs the URLs of a remote repository. It must be non-empty. Fetch will
		// always use the first URL, while push will use all of them.
		if len(config.URLs) == 0 {
			return fmt.Errorf("no fetch URLs, remote=%s", remote.name)
		}
		fetchURL := config.URLs[0]
		if !strings.Contains(fetchURL, remote.path) {
			return fmt.Errorf("fetch URL does not match, remote=%s path=%s", remote.name, remote.path)
		}
		klog.V(2).Infof("%s -> %s - OK", remote.name, fetchURL)
	}
	return nil
}

// LogFromTag returns a list of carry commits from provided tag
func (git *git) LogFromTag(tag string) ([]*gitv5object.Commit, error) {
	tagHash, err := git.repository.Tag(tag)
	if err != nil {
		return nil, fmt.Errorf("git log failed reading tag %q: %w", tag, err)
	}

	commit, err := git.repository.TagObject(tagHash.Hash())
	if err != nil {
		return nil, fmt.Errorf("git log failed reading tag 1 %q: %w", tag, err)
	}

	o := &gitv5.LogOptions{Since: &commit.Tagger.When, Order: gitv5.LogOrderCommitterTime}
	iter, err := git.repository.Log(o)
	if err != nil {
		return nil, fmt.Errorf("git log failed since %q: %w", &commit.Tagger.When, err)
	}
	defer iter.Close()

	commits := make([]*gitv5object.Commit, 0)
	iter.ForEach(func(c *gitv5object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	return commits, nil
}

// Commit returns commit for a given has
// TODO: can we pass has as a string?
func (git *git) Commit(hash plumbing.Hash) (*gitv5object.Commit, error) {
	return git.repository.CommitObject(hash)
}

// CherryPick invokes the cherry-pick command
func (git *git) CherryPick(sha string) error {
	// skipping --strategy-option=ours
	cmd := exec.Command("git", "cherry-pick", "--allow-empty", sha)

	var stdoutStderr []byte
	var err error

	klog.InfoS("executing cherry-pick", "command", cmd.String())
	defer func() {
		if len(stdoutStderr) > 0 {
			defer klog.Infof(">>>>>>>>>>>>>>>>>>>> OUTPUT: END >>>>>>>>>>>>>>>>>>>>>>\n")
			klog.Infof("<<<<<<<<<<<<<<<<<<<< OUTPUT: START <<<<<<<<<<<<<<<<<<<<\n%s", stdoutStderr)
		}
	}()

	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git cherry-pick failed: %w", err)
	}
	return nil
}

// CommitsByDate sorts a list of commits by date
type CommitsByDate []*gitv5object.Commit

func (s CommitsByDate) Len() int      { return len(s) }
func (s CommitsByDate) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s CommitsByDate) Less(i, j int) bool {
	iDate := s[i].Committer.When
	jDate := s[j].Committer.When
	return iDate.Before(jDate)
}
