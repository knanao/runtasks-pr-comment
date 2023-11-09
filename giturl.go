package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	GITHUB_HOST string = "github.com"
)

type gitURL interface {
	Host() string
	Owner() string
	Repository() string
	PullRequest() int
}

func newGitURL(v string) (gitURL, error) {
	url, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	host := url.Hostname()

	switch host {
	case GITHUB_HOST:
		return newGithubURL(url), nil
	default:
		// TODO: Support GitHub Enterprise. Currently GitHub is only supported.
		return nil, fmt.Errorf("unsupported host: %s", host)
	}
}

type githubURL struct {
	host        string
	owner       string
	repository  string
	pullRequest int
}

func newGithubURL(url *url.URL) *githubURL {
	var (
		paths      = strings.Split(url.Path[1:], "/")
		owner      = paths[0]
		repository = paths[1]
		number, _  = strconv.Atoi(paths[3])
	)
	return &githubURL{
		host:        GITHUB_HOST,
		owner:       owner,
		repository:  repository,
		pullRequest: number,
	}
}

func (g *githubURL) Host() string {
	return g.host
}

func (g *githubURL) Owner() string {
	return g.owner
}

func (g *githubURL) Repository() string {
	return g.repository
}

func (g *githubURL) PullRequest() int {
	return g.pullRequest
}
