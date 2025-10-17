package psjungle_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/shirou/gopsutil/v3/process"

	"psjungle/internal/psjungle"
)

func TestBuildTreePrefixRoot(t *testing.T) {
	rootProc, err := process.NewProcess(1)
	if err != nil {
		t.Skip("unable to access PID 1:", err)
	}

	rootNode := &psjungle.ProcessNode{
		Process:  rootProc,
		Children: []*psjungle.ProcessNode{},
		Depth:    0,
		IsTarget: false,
		Parent:   nil,
	}

	prefix := psjungle.BuildTreePrefix(rootNode, nil)
	if prefix != "" {
		t.Fatalf("expected empty prefix for root node, got %q", prefix)
	}
}

func TestBuildTreePrefixChildGlyphs(t *testing.T) {
	rootProc, err := process.NewProcess(1)
	if err != nil {
		t.Skip("unable to access PID 1:", err)
	}

	currentProc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Skip("unable to access current process:", err)
	}

	rootNode := &psjungle.ProcessNode{
		Process:  rootProc,
		Children: []*psjungle.ProcessNode{},
		Depth:    0,
		IsTarget: false,
		Parent:   nil,
	}

	childNode := &psjungle.ProcessNode{
		Process:  currentProc,
		Children: []*psjungle.ProcessNode{},
		Depth:    1,
		IsTarget: true,
		Parent:   rootNode,
	}

	rootNode.Children = append(rootNode.Children, childNode)

	prefix := psjungle.BuildTreePrefix(childNode, nil)
	if prefix != "" && !regexp.MustCompile(`[├└]`).MatchString(prefix) {
		t.Fatalf("expected prefix to contain tree glyphs, got %q", prefix)
	}
}
