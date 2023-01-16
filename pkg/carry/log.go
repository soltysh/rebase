package carry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/rebase/pkg/git"
	"k8s.io/klog/v2"
)

const (
	rebaseMarker   = `Merge remote-tracking branch 'openshift/master' into`
	mergeMarker    = `Merge pull request #`
	upstreamPrefix = "UPSTREAM: "
)

type CarriesLogger struct {
	from          string
	repositoryDir string
}

func NewLog(from, repositoryDir string) *CarriesLogger {
	return &CarriesLogger{
		from:          from,
		repositoryDir: repositoryDir,
	}
}

func (c *CarriesLogger) Run() error {
	repository, err := git.OpenGit(c.repositoryDir)
	if err != nil {
		return err
	}
	commits, err := repository.LogFromTag(c.from)
	if err != nil {
		return err
	}
	sort.Sort(git.CommitsByDate(commits))
	foundRebaseMarker := false
	for _, c := range commits {
		if !foundRebaseMarker {
			if strings.Contains(c.Message, rebaseMarker) {
				klog.V(2).Infof("Found rebase marker at %s", c)
				foundRebaseMarker = true
			}
			continue
		}
		if !strings.Contains(c.Message, upstreamPrefix) {
			continue
		}
		if strings.Contains(c.Message, mergeMarker) {
			// TODO: check if the commit being brought by merge commit is already included, this currently produces duplicates
			// for some of the commits, but is required since some commits might be created before the rebase landed,
			// but merged afterwards, so the --sice in log will skip them, a good example from 4.11/1.24.is:
			// 2022-06-09 00:31:24 -0400 -0400 OpenShift Merge Robot Merge pull request #1229 from rphillips/backports/109103 cb7147853d28e94e1e32674d535e53aec4d9946f
			// 2022-03-29 23:53:02 +0800 +0800 DingShujie UPSTREAM: 109103: cpu manager policy set to none, no one remove container id from container map, lea ed4d3f61aaccbc2fbe383c4d6b9614e8d2ad3e16
			for _, hash := range c.ParentHashes {
				ci, err := repository.LogHash(hash)
				if err != nil {
					klog.Errorf("error reading commit %s: %v", hash, err)
					continue
				}
				if strings.Contains(ci.Message, mergeMarker) {
					continue
				}
				fmt.Printf("%s\t%-25s\t%s\t%s\n", ci.Author.When, ci.Author.Name, formatMessage(ci.Message), ci.Hash.String())
			}
			continue
		}
		fmt.Printf("%s\t%-25s\t%s\t%s\n", c.Author.When, c.Author.Name, formatMessage(c.Message), c.Hash.String())
	}
	return nil
}

func formatMessage(message string) string {
	msg := strings.TrimSpace(message)
	max := len(msg)
	if max > 100 {
		max = 100
	}
	newline := strings.Index(msg, "\n")
	if newline > 0 && max > newline {
		max = newline
	}
	return msg[:max]
}
