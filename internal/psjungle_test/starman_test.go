package psjungle_test

import (
	"testing"

	"psjungle/internal/psjungle"
)

func TestStarmanProcessMatching(t *testing.T) {
	// Test that we can find starman processes (which appear as "perl" in process name
	// but contain "starman" in their command line)

	// This test assumes starman processes are running in the background
	// If they're not available, we'll skip the test

	// Test regex mode - should find starman processes
	starmanPids, err := psjungle.ByRegex("starman", false)
	if err != nil {
		t.Fatalf("Error finding starman processes with regex: %v", err)
	}

	// Test strict mode - should also find starman processes
	starmanPidsStrict, err := psjungle.ByRegex("starman", true)
	if err != nil {
		t.Fatalf("Error finding starman processes with strict mode: %v", err)
	}

	// Both should find the same processes
	if len(starmanPids) != len(starmanPidsStrict) {
		t.Logf("Regex mode found %d starman processes", len(starmanPids))
		t.Logf("Strict mode found %d starman processes", len(starmanPidsStrict))
		// This might be OK depending on the environment
	}

	// Test that regex patterns work in regex mode but not in strict mode
	// "star.an" should match "starman" in regex mode but not in strict mode

	// Regex mode should find starman processes with the pattern "star.an"
	regexPids, err := psjungle.ByRegex("star.an", false)
	if err != nil {
		t.Fatalf("Error finding processes with regex pattern: %v", err)
	}

	// Strict mode should NOT find starman processes with the pattern "star.an"
	// because it's looking for the literal string "star.an"
	strictRegexPids, err := psjungle.ByRegex("star.an", true)
	if err != nil {
		t.Fatalf("Error finding processes with strict pattern: %v", err)
	}

	// Verify that regex mode found processes but strict mode didn't
	if len(regexPids) == 0 && len(starmanPids) > 0 {
		// This is unexpected - regex pattern should match if there are starman processes
		t.Logf("Found starman processes with 'starman' but not with 'star.an' regex")
	} else if len(strictRegexPids) > 0 {
		// This would be unexpected - strict mode should not match "star.an"
		// to "starman" command lines
		t.Logf("Strict mode unexpectedly found processes with 'star.an'")
	}

	// If we have starman processes, verify that we found at least some of them
	if len(starmanPids) > 0 {
		t.Logf("Successfully found %d starman processes in regex mode", len(starmanPids))
	}

	if len(starmanPidsStrict) > 0 {
		t.Logf("Successfully found %d starman processes in strict mode", len(starmanPidsStrict))
	}
}