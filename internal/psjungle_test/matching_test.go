package psjungle_test

import (
	"os/exec"
	"testing"
	"time"

	"psjungle/internal/psjungle"
)

func TestByRegexMatching(t *testing.T) {
	// Test that ByRegex matches on both process name and command line

	// Look for "bash" processes with regex matching
	bashPids, err := psjungle.ByRegex("bash", false)
	if err != nil {
		t.Fatalf("Error finding bash processes: %v", err)
	}

	// We should find at least some bash processes
	if len(bashPids) == 0 {
		t.Log("No bash processes found, but continuing with other tests")
	}

	// Test that we don't find non-existent processes
	foobazPids, err := psjungle.ByRegex("foobaz", false)
	if err != nil {
		t.Fatalf("Error in ByRegex for non-existent process: %v", err)
	}

	if len(foobazPids) != 0 {
		t.Errorf("Expected 0 PIDs for 'foobaz', got %d: %v", len(foobazPids), foobazPids)
	}
}

func TestByRegexMatchesCommandLine(t *testing.T) {
	// Start a test process with a unique command line that we can search for
	// We'll use a command that includes a unique identifier in its command line
	uniqueID := "psjungle_test_12345"

	// Start a process that will have our unique ID in its command line
	cmd := exec.Command("sh", "-c", "sleep 10; echo "+uniqueID)
	cmd.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make sure we clean up
	defer cmd.Process.Kill()

	// Now search for processes that contain our unique ID
	// This should find our test process via command line matching
	pids, err := psjungle.ByRegex(uniqueID, false)
	if err != nil {
		t.Fatalf("Error finding processes by command line: %v", err)
	}

	// We should find exactly one process (our test process)
	if len(pids) != 1 {
		t.Errorf("Expected 1 PID for unique command line '%s', got %d: %v", uniqueID, len(pids), pids)
	}

	// The found PID should match our test process
	if len(pids) > 0 && pids[0] != cmd.Process.Pid {
		t.Errorf("Expected PID %d for our test process, got %d", cmd.Process.Pid, pids[0])
	}
}

func TestRegexMatching(t *testing.T) {
	// Test that ByRegex works with command line patterns

	// Look for processes with bash in their command line
	bashPids, err := psjungle.ByRegex("bash", false)
	if err != nil {
		t.Fatalf("Error finding bash processes with regex: %v", err)
	}

	// We should find at least some bash processes
	if len(bashPids) == 0 {
		t.Log("No bash processes found for regex test")
	} else {
		// Verify we found some processes
		t.Logf("Found %d bash processes with regex", len(bashPids))
	}
}