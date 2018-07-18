package git

import (
	"strings"

	"github.com/floeit/floe/exe"
)

type logger interface {
	Info(...interface{})
	Debug(...interface{})
	Error(...interface{})
}

// Ref contains details of a git reference
type Ref struct {
	Name string
	Type string
	Hash string
}

// Hashes stores the result of a GitLS
type Hashes struct {
	RepoURL string
	Hashes  map[string]Ref
}

// Ls list a remote repo
func Ls(log logger, url, pattern string) (*Hashes, bool) {
	if pattern == "" {
		pattern = "refs/*"
	}
	gitOut, status := exe.RunOutput(log, "", "git", "ls-remote", url, pattern)
	if status != 0 {
		return nil, false
	}
	latestHash := &Hashes{
		RepoURL: url,
	}

	// drop the command and blank line
	parseGitResponse(gitOut[2:], latestHash)
	return latestHash, true
}

func parseGitResponse(lines []string, hashes *Hashes) {
	// map the lines by branch
	hashes.Hashes = map[string]Ref{}
	for _, l := range lines { // from 2 onwards 1 = command 0 = empty
		sl := strings.Fields(l)

		if len(sl) < 2 {
			continue
		}

		dp := strings.Split(sl[1], "/")
		if len(dp) < 3 || dp[0] != "refs" {
			continue
		}

		ty := "branch"
		switch {
		case strings.HasPrefix(dp[1], "pull"):
			ty = "pull"
		case strings.HasPrefix(dp[1], "tag"):
			ty = "tag"
		}
		name := dp[2]
		name = strings.TrimSuffix(name, "^{}")
		hashes.Hashes[sl[1]] = Ref{
			Name: name,
			Type: ty,
			Hash: sl[0],
		}

	}
}
