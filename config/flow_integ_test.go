// +build integration_test

package config

import "testing"

func TestGetFromGit(t *testing.T) {
	b, out, err := getFromGit("flowID", "git@github.com:floeit/floe.git/dev/floe.yml", "master", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
	if len(b) < 100 {
		t.Error("did not get the file")
	}
	if len(out) < 2 {
		t.Error("did not get git output", len(out))
	}
}
