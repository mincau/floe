package exe

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// RunOutput executes the command in a bash process capturing the output and
// returning it in the string slice
func RunOutput(env []string, wd, cmd string, args ...string) ([]string, int) {
	var output []string

	out := make(chan string)
	rangeDone := make(chan bool)
	go func() {
		for t := range out {
			output = append(output, t)
		}
		rangeDone <- true
	}()

	status := Run(out, env, wd, cmd, args...)

	<-rangeDone

	return output, status
}

// Run executes the command in a bash process
func Run(out chan string, env []string, wd, cmd string, args ...string) int {
	out <- cmd + " " + strings.Join(args, " ")
	out <- ""

	if wd != "" {
		// make sure working directory is in place
		if err := os.MkdirAll(wd, 0700); err != nil {
			out <- err.Error()
			return 1
		}
	}

	eCmd := exec.Command(cmd, args...)

	eCmd.Env = os.Environ()
	eCmd.Env = append(eCmd.Env, env...)

	// this is mandatory
	eCmd.Dir = wd

	// safely aggregate both to a single reader
	pr, pw := io.Pipe()
	eCmd.Stdout = pw
	eCmd.Stderr = pw

	// start scanning from the common pipe
	scanDone := make(chan bool)
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			out <- scanner.Text()
		}
		if e := scanner.Err(); e != nil {
			out <- "scanning output failed with: " + e.Error()
		}
		scanDone <- true
	}()

	err := eCmd.Start()
	if err != nil {
		out <- err.Error()
		out <- ""
		close(out)
		return 1
	}

	err = eCmd.Wait()

	// close the writer pipe
	e := pw.Close()
	if e != nil {
		panic("not sure how this particular close could error" + err.Error())
	}

	// wait to be sure scanner is fully complete
	<-scanDone
	close(out)

	if err != nil {
		exitCode := 1
		if msg, ok := err.(*exec.ExitError); ok {
			if status, ok := msg.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		// we prefer to return 0 for good or one for bad
		return exitCode
	}

	return 0
}
