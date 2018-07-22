package hub

import (
	"testing"
	"time"

	"github.com/floeit/floe/exe/git"

	"github.com/floeit/floe/config"
	nt "github.com/floeit/floe/config/nodetype"
	"github.com/floeit/floe/event"
)

type obs func(e event.Event)

func (o obs) Notify(e event.Event) {
	o(e)
}

func TestTimers(t *testing.T) {
	t.Parallel()

	q := &event.Queue{}

	// make an observer that signals a chanel
	got := make(chan bool, 1)
	f := func(e event.Event) {
		if e.Tag == "sys.state" {
			got <- true
		}
	}
	q.Register(obs(f))

	ts := newTimers(q)

	ts.register(config.FlowRef{
		ID:  "test-flow",
		Ver: 1,
	}, "test-node", nt.Opts{
		"period": 1,
	}, startFlowTrigger)

	select {
	case <-time.After(time.Second * 2):
		t.Fatal("no event")
	case <-got:
	}
}

func TestDiffRefs(t *testing.T) {
	old := git.Hashes{
		RepoURL: "foo",
	}
	new := git.Hashes{
		RepoURL: "bar",
	}
	diff := changedRefs(old, new)
	if len(diff.Hashes) != 0 {
		t.Error("diff of different repos must be zero")
	}

	old = git.Hashes{
		RepoURL: "foo",
		Hashes: map[string]git.Ref{
			"a": git.Ref{
				Hash: "aaa",
			},
			"b": git.Ref{
				Hash: "bbb",
			},
		},
	}
	new = git.Hashes{
		RepoURL: "foo",
		Hashes:  map[string]git.Ref{},
	}

	// 2 old 0 new
	diff = changedRefs(old, new)
	if len(diff.Hashes) != 0 {
		t.Error("changes when no new ones should be zero")
	}

	// 2 new ones
	diff = changedRefs(new, old)
	if len(diff.Hashes) != 2 {
		t.Error("2 new ones should produce 2 changes", len(diff.Hashes))
	}

	// 2 the same one new
	new.Hashes["a"] = git.Ref{Hash: "aaa"}
	new.Hashes["b"] = git.Ref{Hash: "bbb"}
	new.Hashes["c"] = git.Ref{Hash: "ccc"}
	diff = changedRefs(old, new)
	if len(diff.Hashes) != 1 {
		t.Error("changes with 2 the same and 1 new should be 1", len(diff.Hashes))
	}

	// 1 the same 1 changed and one new
	new.Hashes["a"] = git.Ref{Hash: "aaa"}
	new.Hashes["b"] = git.Ref{Hash: "bbc"}
	new.Hashes["c"] = git.Ref{Hash: "ccc"}
	diff = changedRefs(old, new)
	if len(diff.Hashes) != 2 {
		t.Error("changes with 1 the same, 1 changed and 1 new should be 2", len(diff.Hashes))
	}
	if diff.RepoURL != "foo" {
		t.Error("repo url wrong", diff.RepoURL)
	}
	if diff.Hashes["b"].Hash != "bbc" {
		t.Error("got wrong changed hash", diff.Hashes["b"].Hash)
	}
}
