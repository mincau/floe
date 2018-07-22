// +build integration_test

package hub

import (
	"testing"
	"time"

	"github.com/floeit/floe/event"
	"github.com/floeit/floe/exe/git"

	"github.com/floeit/floe/config"
	nt "github.com/floeit/floe/config/nodetype"
	"github.com/floeit/floe/store"
)

func TestRepoPoller(t *testing.T) {
	t.Parallel()

	s := store.NewMemStore()
	o := nt.Opts{
		"url":  "git@github.com:floeit/floe.git",
		"refs": "*",
	}
	tim := &timer{
		flow: config.FlowRef{
			ID:  "testflo",
			Ver: 1,
		},
	}
	q := &event.Queue{}
	c, err := config.ParseYAML(trigFlow)
	if err != nil {
		t.Fatal(err)
	}

	hub := &Hub{
		config: *c,
		runs:   newRunStore(s),
		queue:  q,
	}
	q.Register(hub)

	// make an observer that signals a chanel
	got := make(chan bool, 1)
	f := func(e event.Event) {
		if e.Tag == "sys.state" {
			got <- true
		}
	}
	q.Register(obs(f))

	p := newRepoPoller(s, "nodeID", o)

	// first call sets up the refs
	p.timer(nil, tim)

	// second call should produce no diffs
	p.timer(nil, tim)

	h := git.Hashes{}
	err = s.Load("refs/testflo/nodeID", &h)
	if err != nil {
		t.Fatal(err)
	}

	master, ok := h.Hashes["refs/heads/master"]
	if !ok {
		t.Fatal("master not found")
	}
	hashFromRepo := master.Hash
	master.Hash = "difference"
	h.Hashes["refs/heads/master"] = master
	err = s.Save("refs/flowID/nodeID", h)
	if err != nil {
		t.Fatal(err)
	}

	// this one will launch a pending flow
	p.timer(q, tim)

	select {
	case <-time.After(time.Second * 10):
		t.Fatal("no event")
	case <-got:
	}

	pend := pending{}
	err = s.Load("pending-list", &pend)
	if err != nil {
		t.Fatal(err)
	}
	if len(pend.Pends) != 1 {
		t.Fatal("should have got one pending")
	}
	if pend.Pends[0].Opts["hash"].(string) != hashFromRepo {
		t.Error("got bad hash ref")
	}
	if pend.Pends[0].Opts["branch"].(string) != "master" {
		t.Error("got bad branch")
	}
}

var trigFlow = []byte(`
flows:
    - id: testflo
      ver: 1
      
      triggers:
        - name: Commits
          type: poll-git
          opts:
            period: 10                                 # check every 10 seconds
            url: git@github.com:danmux/danmux-hugo.git # the repo to check
            refs: "refs/heads/*"                       # the refs pattern to match
            exclude-refs: "refs/heads/master"
`)
