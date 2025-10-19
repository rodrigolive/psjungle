package psjungle_test

import (
	"os/exec"
	"testing"
	"time"

	"psjungle/internal/psjungle"
)

func TestByRegexExactStringMatching(t *testing.T) {
	// Test that strict mode matches exact strings within command lines
	// and doesn't interpret regex patterns, while regex mode does interpret patterns

	// Start a test process with a command line containing "cla starman worker"
	// This simulates the actual starman process scenario you described
	testCmdLine := "cla starman worker"
	cmd := exec.Command("sh", "-c", "sleep 10; echo "+testCmdLine)
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make sure we clean up
	defer cmd.Process.Kill()

	// Test strict mode with exact substring "starman"
	// This SHOULD match the process with "cla starman worker" in command line
	// It should also match any existing starman processes
	pids, err := psjungle.ByRegex("starman", true) // strict mode = true
	if err != nil {
		t.Fatalf("Error finding processes with strict mode: %v", err)
	}

	// We should find at least one process (our test process and possibly existing starman processes)
	foundOurProcess := false
	for _, pid := range pids {
		if pid == cmd.Process.Pid {
			foundOurProcess = true
			break
		}
	}

	if !foundOurProcess {
		t.Errorf("Expected to find our test process PID %d for exact string 'starman' in strict mode, but got PIDs: %v", cmd.Process.Pid, pids)
	}

	// Test strict mode with pattern "star.an"
	// This should NOT match our "cla starman worker" process because it's looking for literal "star.an"
	pids2, err := psjungle.ByRegex("star.an", true) // strict mode = true
	if err != nil {
		t.Fatalf("Error finding processes with strict pattern: %v", err)
	}

	// Our process should NOT be found because "star.an" is not literally in the command line
	for _, pid := range pids2 {
		if pid == cmd.Process.Pid {
			t.Errorf("Strict mode unexpectedly found our test process with pattern 'star.an'")
			break
		}
	}

	// Test regex mode with pattern "star.an"
	// This SHOULD match our "cla starman worker" process because "star.an" regex matches "starman"
	pids3, err := psjungle.ByRegex("star.an", false) // strict mode = false (regex mode)
	if err != nil {
		t.Fatalf("Error finding processes with regex pattern: %v", err)
	}

	// We should find our process because "star.an" regex matches "starman"
	foundOurProcessRegex := false
	for _, pid := range pids3 {
		if pid == cmd.Process.Pid {
			foundOurProcessRegex = true
			break
		}
	}

	if !foundOurProcessRegex {
		t.Errorf("Expected to find our test process PID %d with regex pattern 'star.an', but got PIDs: %v", cmd.Process.Pid, pids3)
	}
}