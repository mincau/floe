package config

import (
	"testing"

	nt "github.com/floeit/floe/config/nodetype"
)

func TestNodeExec(t *testing.T) {
	output := make(chan string)
	captured := make(chan bool)
	cl := 0
	go func() {
		for l := range output {
			println(l)
			cl++
		}
		captured <- true
	}()
	n := &node{
		// what flow is this node attached to
		Class: "task",
		Type:  "exec",
		Opts: nt.Opts{
			"shell": "export",
			"env":   []string{"DAN=fart"},
		},
	}

	status, _, err := n.Execute(&nt.Workspace{}, nt.Opts{}, output)
	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Error("wrong status", status)
	}
	close(output)
	<-captured
}
