package nodetype

import (
	"fmt"
	"net"
	"net/http"
	"testing"
)

func TestLinkLocation(t *testing.T) {

	fixs := []struct {
		loc string
		exp string
	}{
		{
			loc: "",
			exp: "{{ws}}/foo.txt",
		},
		{
			loc: "/what/",
			exp: "/what/foo.txt",
		},
		{
			loc: "/what",
			exp: "/what",
		},
		{
			loc: "where/",
			exp: "{{ws}}/where/foo.txt",
		},
		{
			loc: "where.xx",
			exp: "{{ws}}/where.xx",
		},
	}

	for i, f := range fixs {
		got := linkLocation(f.loc, "foo.txt")
		if f.exp != got {
			t.Errorf("%d) expected:<%s> got: <%s>", i, f.exp, got)
		}
	}
}

func TestFetch(t *testing.T) {
	portCh := make(chan int)

	go serveFiles(portCh)

	port := <-portCh

	success := []string{`(100.00%)`, `Downloading`, "200 OK"}
	fail := []string{"404", "Not Found"}
	fixtures := []struct {
		url      string
		algo     string
		checksum string
		anyError bool
		expected []string
	}{
		{ // simple dl
			url:      fmt.Sprintf("http://127.0.0.1:%d/get-file.txt", port),
			expected: success,
		},
		{ // dl with sha256 check
			url:      fmt.Sprintf("http://127.0.0.1:%d/get-file.txt", port),
			algo:     "sha256",
			checksum: "864d6473d56d235de9ffb9d404e76f23e4d596ce77eae5b7ce5106f454fa7ee4 get-file.txt",
			expected: success,
		},
		{ // dl with sha1 check
			url:      fmt.Sprintf("http://127.0.0.1:%d/get-file.txt", port),
			algo:     "sha1",
			checksum: "bb3357153aa8e2c0b22fef75a7f21969abb7c2b4",
			expected: success,
		},
		{ // dl with sha256 check
			url:      fmt.Sprintf("http://127.0.0.1:%d/get-file.txt", port),
			algo:     "md5",
			checksum: "f35ff35df6efc82e474e97eaf10e7ff6",
			expected: success,
		},
		{ // good dl bad checksum
			url:      fmt.Sprintf("http://127.0.0.1:%d/get-file.txt", port),
			algo:     "sha256",
			checksum: "badfeeda", // hex compatible clearly crap checksum
			anyError: true,
			expected: []string{"Download failed", "checksum"},
		},
		{ // bad dl
			url:      fmt.Sprintf("http://127.0.0.1:%d/wont_be_found", port),
			anyError: true,
			expected: fail,
		},
		{ // good external check
			url:      "https://dl.google.com/go/go1.10.2.src.tar.gz",
			algo:     "sha256",
			checksum: "6264609c6b9cd8ed8e02ca84605d727ce1898d74efa79841660b2e3e985a98bd go1.10.2.src.tar.gz",
			expected: success,
		},
	}

	for i, fx := range fixtures {
		opts := Opts{
			"url":           fx.url,
			"checksum":      fx.checksum,
			"checksum-algo": fx.algo,
			"location":      "tmpdl/",
		}
		testNode(t, fmt.Sprintf("fetch test: %d", i), fetch{}, opts, fx.expected, fx.anyError)
	}
}

// simple local server that returns a bit of content
func serveFiles(portChan chan int) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/get-file.txt", func(w http.ResponseWriter, r *http.Request) {
		message := "this is a file with known hashes"
		w.Write([]byte(message))
	})
	portChan <- listener.Addr().(*net.TCPAddr).Port
	http.Serve(listener, mux)
}
