package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	nt "github.com/floeit/floe/config/nodetype"
)

const (
	// SubTagGood the tag to be used on events who's result is good
	SubTagGood = "good"
	// SubTagBad the tag to be used on events who's result is bad
	SubTagBad = "bad"
)

// NodeClass the type def for the classes a Node can be
type NodeClass string

// NodeClass values
const (
	NcTask    NodeClass = "task"
	NcMerge   NodeClass = "merge"
	NcTrigger NodeClass = "trigger"
)

// NodeRef uniquely identifies a Node across time (versions)
type NodeRef struct {
	Class NodeClass
	ID    string
}

func (n NodeRef) String() string {
	return fmt.Sprintf("%s.%s", n.Class, n.ID)
}

// node is the deserialised node whose set of fields cover all types of node
type node struct {
	flowRef    FlowRef // what flow is this node attached to
	Class      NodeClass
	Ref        NodeRef
	ID         string
	Name       string
	Listen     string
	Wait       []string // if used as a merge node this is an array of event tags to wait for
	Type       string
	Good       []int // the array of exit status codes considered a success
	IgnoreFail bool  `yaml:"ignore-fail"` // only ever send the good event cant be used in conjunction with UseStatus
	// use status if we don't send good or bad but the actual status code as an event
	// TODO - consider if mapping status codes to good and bad is all the complexity we need
	UseStatus bool    `yaml:"use-status"`
	Opts      nt.Opts // static config options
}

func (t *node) Execute(ws *nt.Workspace, opts nt.Opts, output chan string) (int, nt.Opts, error) {
	n := nt.GetNodeType(t.Type)
	if n == nil {
		return 255, nil, fmt.Errorf("no node type found: %s", t.Type)
	}
	// combine any event options with the overriding preset options from the config
	inOpts := nt.MergeOpts(opts, t.Opts)
	status, opts, err := n.Execute(ws, inOpts, output)
	if err != nil && t.IgnoreFail {
		err = nil
	}
	return status, opts, err
}

// Status will return the string to use on an event tag and a boolean to
// indicate if the status is considered good
func (t *node) Status(status int) (string, bool) {
	// always good if ignore fail
	if t.IgnoreFail {
		return SubTagGood, true
	}
	// is this code considered a success
	good := false
	// no specific good statuses so consider 0 success, all others fail
	if len(t.Good) == 0 {
		good = status == 0
	} else {
		for _, s := range t.Good {
			if s == status {
				good = true
				break
			}
		}
	}
	// use specific exit statuses TODO - is it an overcomplication
	if t.UseStatus {
		return strconv.Itoa(status), good
	}
	// or binary result
	if good {
		return SubTagGood, true
	}
	return SubTagBad, false
}

func (t *node) FlowRef() FlowRef {
	return t.flowRef
}

func (t *node) NodeRef() NodeRef {
	return t.Ref
}

func (t *node) TypeOfNode() string {
	return t.Type
}

func (t *node) Waits() int {
	return len(t.Wait)
}

func (t *node) GetTag(subTag string) string {
	return fmt.Sprintf("%s.%s.%s", t.Class, t.Ref.ID, subTag)
}

func (t *node) matchedTriggers(eType string, opts *nt.Opts) bool {
	// trigger matches must always have opts
	if opts == nil {
		return false
	}
	if t.Type != eType {
		return false
	}
	n := nt.GetNodeType(eType)
	// if there is no type registered then there is no matching logic
	if n == nil {
		return true
	}
	// compare config options with the event options
	return n.Match(t.Opts, *opts)
}

func (t *node) matched(tag string) bool {
	// match on the Listen
	if t.Listen != "" && t.Listen == tag {
		return true
	}
	// or if any tags in the the Wait list match (merge nodes only)
	for _, wt := range t.Wait {
		if wt == tag {
			return true
		}
	}
	return false
}

func (t *node) setName(n string) {
	t.Name = n
}
func (t *node) setID(i string) {
	t.ID = i
}
func (t *node) name() string {
	return t.Name
}
func (t *node) id() string {
	return t.ID
}

func (t *node) zero(defaultClass NodeClass, flow FlowRef) error {
	if err := zeroNID(t); err != nil {
		return err
	}
	t.flowRef = flow
	if t.Class == "" {
		t.Class = defaultClass
	}

	t.Ref = NodeRef{
		Class: t.Class,
		ID:    t.ID,
	}

	// node specific checks
	switch t.Class {
	case NcTask:
		if len(t.Wait) != 0 {
			return errors.New("task nodes can not have waits")
		}
	case NcMerge:
		// default to waiting for all of them if not specified
		if t.Type == "" {
			t.Type = "all"
		}
		if t.Listen != "" {
			return errors.New("merge nodes can not have listen set")
		}
	}

	// not entirely sure what CastOpts was supposed to do
	// possibly set up default node Opts as specific node related struct fields?
	// n := nt.GetNodeType(t.Type)
	// if n == nil {
	// 	return nil
	// }
	// n.CastOpts(&t.Opts)

	return nil
}

type nid interface {
	setID(string)
	setName(string)
	name() string
	id() string
}

// zeroNid checks and sets ID from name, or name from ID
func zeroNID(n nid) error {
	name := trimNIDs(n.name())
	id := strings.ToLower(trimNIDs(n.id()))

	if name == "" && id == "" {
		return errors.New("id and name can not both be empty")
	}
	if id == "" {
		id = idFromName(name)
	}
	if strings.ContainsAny(id, " .") {
		return errors.New("a specified id can not contain spaces or full stops")
	}
	if name == "" {
		name = nameFromID(id)
	}

	n.setID(id)
	n.setName(name)
	return nil
}

// trim trailing spaces and dots and hyphens
func trimNIDs(s string) string {
	return strings.Trim(s, " .-")
}

// idFromName makes a file and URL/HTML friendly ID from the name.
func idFromName(name string) string {
	s := strings.Split(strings.ToLower(strings.TrimSpace(name)), " ")
	ns := strings.Join(s, "-")
	s = strings.Split(ns, ".")
	return strings.Join(s, "-")
}

func nameFromID(id string) string {
	s := strings.Split(strings.ToLower(strings.TrimSpace(id)), "-")
	return strings.Join(s, " ")
}
