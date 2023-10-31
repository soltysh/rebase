package github

import (
	"context"
	"os"

	"github.com/google/go-github/v56/github"
	"k8s.io/klog/v2"
)

func IsMerged(number int) (bool, error) {
	client := github.NewClient(nil)
	if token := os.Getenv("GITHUB_TOKEN"); len(token) > 0 {
		client = client.WithAuthToken(token)
	} else {
		klog.V(3).Infof("Using the default github token, which might rate limit your requests!")
	}
	isMerged, response, err := client.PullRequests.IsMerged(context.Background(), "kubernetes", "kubernetes", number)
	klog.V(3).Infof("Remaining rate with current token is %s", response.Rate.String())
	return isMerged, err
}
