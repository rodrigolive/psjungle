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

	for i, arg := range args {
		// Skip the program name (first argument)
		if i == 0 {
			processed = append(processed, arg)
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
				continue
			}
		}

		// If not a special case, just add the argument as is
		processed = append(processed, arg)
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
