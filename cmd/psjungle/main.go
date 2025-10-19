package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"psjungle/internal/psjungle"
)

// preprocessArgs handles special argument formats like -w2
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
