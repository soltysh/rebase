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
	// Checkout the specified remote
	Checkout(remote string) error
	// CreateBranch creates a named branch based on remote
	CreateBranch(name, remote string) error
	// CherryPick invokes the cherry-pick command
	CherryPick(sha string) error
	// Merge remote branch
	Merge(remote string) error
}

// OpenGit opens path as a git repository, ensuring that remotes contain
// both upstream kubernetes and openshift remotes properly configured.
func OpenGit(path string) (Git, error) {
	klog.V(2).Infof("Using %s as git repository", path)
	repository, err := gitv5.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	gitRepo := &git{repository: repository, path: path}
	klog.V(2).Infof("Checking if openshift and upstream remotes are configured..")
	if err := gitRepo.checkRemotes(); err != nil {
		return nil, err
	}
	return gitRepo, nil
}

type git struct {
	path       string
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
		// TODO: add auto-updating remotes if the above are missing, there's CreateRemote function
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
		return nil, err
	}

	commit, err := git.repository.TagObject(tagHash.Hash())
	if err != nil {
		return nil, err
	}

	o := &gitv5.LogOptions{Since: &commit.Tagger.When, Order: gitv5.LogOrderCommitterTime}
	iter, err := git.repository.Log(o)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	commits := make([]*gitv5object.Commit, 0)
	iter.ForEach(func(c *gitv5object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	return commits, nil
}

// Checkout the specified remote
func (git *git) Checkout(remote string) error {
	return git.runGit("checkout", remote)
}

// Commit returns commit for a given has
// TODO: can we pass has as a string?
func (git *git) Commit(hash plumbing.Hash) (*gitv5object.Commit, error) {
	return git.repository.CommitObject(hash)
}

// CreateBranch creates a named branch based on remote
func (git *git) CreateBranch(name, remote string) error {
	return git.runGit("checkout", "-b", name, remote)
}

// Merge remote branch
func (git *git) Merge(remote string) error {
	return git.runGit("merge", "--strategy", "ours", remote, "--no-edit")
}

// CherryPick invokes the cherry-pick command
func (git *git) CherryPick(sha string) error {
	return git.runGit("cherry-pick", sha)
}

func (git *git) runGit(args ...string) error {
	// cmd := exec.Command("git", "merge", "--strategy", "ours", remote, "--no-edit")
	cmd := exec.Command("git", args...)
	klog.V(2).Infof("Invoking %s...", cmd)
	cmd.Dir = git.path
	var (
		output []byte
		err    error
	)
	output, err = cmd.CombinedOutput()
	klog.V(3).Infof(string(output))
	return err
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
