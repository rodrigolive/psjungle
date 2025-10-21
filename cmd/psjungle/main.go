package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"psjungle/internal/psjungle"
)

// isKnownSignal checks if a string is a known signal name
func isKnownSignal(s string) bool {
	knownSignals := map[string]bool{
		"term": true,
		"hup":  true,
		"int":  true,
		"kill": true,
		"stop": true,
		"cont": true,
		"usr1": true,
		"usr2": true,
	}
	_, exists := knownSignals[strings.ToLower(s)]
	return exists
}

// preprocessArgs handles special argument formats like -w2 and -k
func preprocessArgs(args []string) []string {
	processed := make([]string, 0, len(args)*2) // Pre-allocate with some extra space

	i := 0
	for i < len(args) {
		arg := args[i]

		// Skip the program name (first argument)
		if i == 0 {
			processed = append(processed, arg)
			i++
			continue
		}

		// Check if this is a -w flag with concatenated number
		if strings.HasPrefix(arg, "-w") && len(arg) > 2 {
			// Extract the part after -w
			value := arg[2:]
			// Check if it's a valid number
			if matched, _ := regexp.MatchString(`^\d+$`, value); matched {
				// Split into separate flag and value
				processed = append(processed, "-w", value)
				i++
				continue
			}
		}

		// Check if this is a -k flag with concatenated value
		if strings.HasPrefix(arg, "-k") && len(arg) > 2 {
			// Extract the part after -k
			value := arg[2:]
			// Handle equals format
			if strings.HasPrefix(value, "=") && len(value) > 1 {
				value = value[1:]
			}
			// Split into separate flag and value
			processed = append(processed, "-k", value)
			i++
			continue
		}

		// Handle the case where -w is followed by a non-flag argument
		// (which should be treated as the target, not as the flag value)
		if arg == "-w" && i+1 < len(args) {
			nextArg := args[i+1]
			// If the next argument is not a flag (doesn't start with -),
			// then it should be treated as the target, not as the flag value
			if !strings.HasPrefix(nextArg, "-") {
				// Add the -w flag with default empty value
				processed = append(processed, "-w", "")
				// Add the next arg as the first non-flag argument (target)
				processed = append(processed, nextArg)
				i += 2 // Skip both the -w flag and the next argument
				continue
			}
		}

		// Handle -k flag
		if arg == "-k" {
			// Check if there's a next argument
			if i+1 < len(args) {
				nextArg := args[i+1]
				// Check if the next argument is a known signal name
				// If so, use it as the signal value
				if isKnownSignal(nextArg) {
					processed = append(processed, "-k", nextArg)
					i += 2 // Skip both the -k flag and the signal argument
					continue
				}
				// Check if the next argument is a number (signal number)
				if matched, _ := regexp.MatchString(`^\d+$`, nextArg); matched {
					processed = append(processed, "-k", nextArg)
					i += 2 // Skip both the -k flag and the signal number
					continue
				}
				// If the next argument is not a flag and not a known signal/number,
				// treat it as target pattern and add -k with default empty value (SIGTERM)
				if !strings.HasPrefix(nextArg, "-") {
					processed = append(processed, "-k", "")
					processed = append(processed, nextArg)
					i += 2 // Skip both the -k flag and the target pattern
					continue
				}
			}
			// If no next argument or next argument is a flag, treat as standalone -k (SIGTERM)
			processed = append(processed, "-k", "")
			i++
			continue
		}

		// If not a special case, just add the argument as is
		processed = append(processed, arg)
		i++
	}

	return processed
}

func main() {
	processedArgs := preprocessArgs(os.Args)
	if err := psjungle.Run(processedArgs); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
