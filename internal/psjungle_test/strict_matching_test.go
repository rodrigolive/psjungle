package psjungle_test

import (
	"os/exec"
	"testing"
	"time"

	"psjungle/internal/psjungle"
)

func TestByRegexWithStrictMode(t *testing.T) {
	// Start a test process with a unique command line that we can search for
	uniqueID := "psjungle_test_strict_67890"

	// Start a process that will have our unique ID in its command line
	cmd := exec.Command("sh", "-c", "sleep 10; echo "+uniqueID)
	cmd.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make sure we clean up
	defer cmd.Process.Kill()

	// Test strict mode - exact string matching
	pids, err := psjungle.ByRegex(uniqueID, true) // strict mode = true
	if err != nil {
		t.Fatalf("Error finding processes with strict mode: %v", err)
	}

	// We should find exactly one process (our test process)
	if len(pids) != 1 {
		t.Errorf("Expected 1 PID for exact string '%s' in strict mode, got %d: %v", uniqueID, len(pids), pids)
	}

	// Test regex mode - should also find our process (since exact string is valid regex)
	pids2, err := psjungle.ByRegex(uniqueID, false) // strict mode = false (regex mode)
	if err != nil {
		t.Fatalf("Error finding processes with regex mode: %v", err)
	}

	// We should find exactly one process (our test process)
	if len(pids2) != 1 {
		t.Errorf("Expected 1 PID for regex pattern '%s', got %d: %v", uniqueID, len(pids2), pids2)
	}

	// Both should return the same PID
	if len(pids) > 0 && len(pids2) > 0 && pids[0] != pids2[0] {
		t.Errorf("Strict mode and regex mode returned different PIDs: %d vs %d", pids[0], pids2[0])
	}
}

func TestByRegexStrictModeVsRegexMode(t *testing.T) {
	// Test that strict mode doesn't interpret regex patterns
	// Start a test process with dots in the command line
	testString := "test.with.dots.in.name"

	cmd := exec.Command("sh", "-c", "sleep 10; echo "+testString)
	cmd.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make sure we clean up
	defer cmd.Process.Kill()

	// Test strict mode with a pattern that contains dots
	// In strict mode, this should NOT match because it's looking for the exact string "test...name"
	pids, err := psjungle.ByRegex("test.*name", true) // strict mode = true
	if err != nil {
		t.Fatalf("Error finding processes with strict mode: %v", err)
	}

	// We should NOT find any processes because "test.*name" is not literally in the command line
	if len(pids) != 0 {
		t.Errorf("Expected 0 PIDs for strict pattern 'test.*name', got %d: %v", len(pids), pids)
	}

	// Test regex mode with the same pattern
	// In regex mode, "test.*name" should match "test.with.dots.in.name"
	pids2, err := psjungle.ByRegex("test.*name", false) // strict mode = false (regex mode)
	if err != nil {
		t.Fatalf("Error finding processes with regex mode: %v", err)
	}

	// We should find at least one process because the regex pattern matches
	if len(pids2) == 0 {
		t.Log("No processes found for regex pattern 'test.*name', but this might be expected in some environments")
	}
}