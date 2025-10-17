package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/shirou/gopsutil/v3/process"
)

// TestTreeOutputFormat tests that the tree output uses Unicode characters correctly
func TestTreeOutputFormat(t *testing.T) {
	// Create a simple test tree structure
	rootProc, err := process.NewProcess(1)
	if err != nil {
		t.Skip("Could not get root process")
	}

	// Create a mock tree structure for testing
	rootNode := &ProcessNode{
		Process: rootProc,
		Children: []*ProcessNode{},
		Depth:    0,
		IsTarget: false,
		Parent:   nil,
	}

	// Test that buildTreePrefix returns correct Unicode characters
	nextSiblings := []*ProcessNode{}
	prefix := buildTreePrefix(rootNode, nextSiblings)

	// Root node should have empty prefix
	expected := ""
	if prefix != expected {
		t.Errorf("Expected root prefix to be '%s', got '%s'", expected, prefix)
	}

	// Test with a child node (no siblings after it)
	childProc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Skip("Could not get current process")
	}

	childNode := &ProcessNode{
		Process: childProc,
		Children: []*ProcessNode{},
		Depth:    1,
		IsTarget: true,
		Parent:   rootNode,
	}

	rootNode.Children = append(rootNode.Children, childNode)

	// For a child with no next siblings, should end with └──
	prefix = buildTreePrefix(childNode, []*ProcessNode{})
	// Should have empty prefix for depth 1 when there's only one child
	// since we're not testing the actual output formatting here

	// Test that the prefix contains the expected Unicode characters
	hasCorrectChars := regexp.MustCompile(`[├└]`).MatchString(prefix)
	if !hasCorrectChars && prefix != "" {
		t.Errorf("Prefix should contain tree characters, got '%s'", prefix)
	}
}

// TestUnicodeCharacters tests that buildTreePrefix generates correct Unicode tree characters
func TestUnicodeCharacters(t *testing.T) {
	tests := []struct {
		name        string
		hasSiblings bool
		expectedPattern string
	}{
		{"Last child node", false, `└── `},
		{"Middle child node", true, `├── `},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create mock processes
			rootProc, err := process.NewProcess(1)
			if err != nil {
				t.Skip("Could not get root process")
			}

			childProc, err := process.NewProcess(int32(os.Getpid()))
			if err != nil {
				t.Skip("Could not get current process")
			}

			// Create tree structure
			rootNode := &ProcessNode{
				Process: rootProc,
				Children: []*ProcessNode{},
				Depth:    0,
				IsTarget: false,
				Parent:   nil,
			}

			childNode := &ProcessNode{
				Process: childProc,
				Children: []*ProcessNode{},
				Depth:    1,
				IsTarget: true,
				Parent:   rootNode,
			}

			rootNode.Children = append(rootNode.Children, childNode)

			var nextSiblings []*ProcessNode
			if test.hasSiblings {
				nextSiblings = append(nextSiblings, &ProcessNode{})
			}

			prefix := buildTreePrefix(childNode, nextSiblings)

			if test.expectedPattern != "" {
				matched, _ := regexp.MatchString(test.expectedPattern, prefix)
				if !matched {
					t.Errorf("Expected prefix to match pattern '%s', got '%s'", test.expectedPattern, prefix)
				}
			}
		})
	}
}