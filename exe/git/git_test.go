package git

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {

	stashAndGit := `8b650c4565361eab3ba0cd21295c169e55087257    refs/heads/reverted_certs
a6767754ef433da45f8096d06ba1eb7c8de16e3f    refs/heads/hoombe
9906c112ddd51934593fe5f97cba8bbc86e916b2    refs/heads/hoom-nag-message
5793e9c730a1b464539e304ba61f7e647622ba40    refs/heads/foo-1531910961902
5fe9cc9bb4bb6da45c112b390219a58bcb0543d6    refs/pull-requests/6638/from
c47f2ec90df29fc4502d7364b635a1454636f2e4    refs/pull-requests/6649/from
d433e9bb255c529450881d18dca62506b525bff2    refs/pull-requests/6649/merge
855b90cb574bd0945c4e7aa35189cf431abc3cd0    refs/pull-requests/6650/from
bd8d2a6589ec23e455d35689207f3a5319c5f62a    refs/tags/v_4.6.911
461f5ead415c6fa457dcfdc774855cdc9ecc1e3e    refs/tags/v_5.6.912
d046e4898b8df7e45fc1e63905a4a950a96452bf    refs/tags/_4.6.913
0db621b7f0cf8dd4545f930f000d8d2f41c65607    refs/tags/ver_4.6.92
0db621b7f0cf8dd4545f930f000d8d2f41c65607    refs/tags/v_4.6.92^{}
ef0f5274afae6d4f36ee29fa61d4398ad8a6567c	HEAD
ef0f5274afae6d4f36ee29fa61d4398ad8a6567c	refs/heads/master
fbf6240b17cd4aeedd070b6b5461395602708ace	refs/heads/poll-git-changes
97a574fa05056609f5746afaae42e083477e06cc	refs/pull/1/head
68aebba3d722f158eb59b3cd3f573bd2cf152bba	refs/pull/2/head
fbf6240b17cd4aeedd070b6b5461395602708ace	refs/pull/3/head
978ae2def696424956ee03f367233ee20cedc8ad	refs/tags/v0.1
	`

	st := strings.Split(stashAndGit, "\n")
	h := Hashes{}
	parseGitResponse(st, &h)

	if len(h.Hashes) != len(st)-2 {
		t.Fatal("HEAD should be ignored")
	}

	exp := map[string]Ref{
		// stash ones
		"refs/heads/hoom-nag-message":  Ref{Name: "hoom-nag-message", Type: "branch", Hash: "9906c112ddd51934593fe5f97cba8bbc86e916b2"},
		"refs/pull-requests/6650/from": Ref{Name: "6650", Type: "pull", Hash: "855b90cb574bd0945c4e7aa35189cf431abc3cd0"},
		"refs/tags/ver_4.6.92":         Ref{Name: "ver_4.6.92", Type: "tag", Hash: "0db621b7f0cf8dd4545f930f000d8d2f41c65607"},

		// git hub ones
		"refs/heads/master": Ref{Name: "master", Type: "branch", Hash: "ef0f5274afae6d4f36ee29fa61d4398ad8a6567c"},
		"refs/pull/2/head":  Ref{Name: "2", Type: "pull", Hash: "68aebba3d722f158eb59b3cd3f573bd2cf152bba"},
		"refs/tags/v0.1":    Ref{Name: "v0.1", Type: "tag", Hash: "978ae2def696424956ee03f367233ee20cedc8ad"},
	}

	for n, ex := range exp {
		got := h.Hashes[n]
		if ex.Name != got.Name || ex.Hash != got.Hash || ex.Type != got.Type {
			t.Errorf("[%s] - parsing failed n:%s h:%s t:%s != n:%s h:%s t:%s", n, ex.Name, ex.Hash, ex.Type, got.Name, got.Hash, got.Type)
		}
	}

}
