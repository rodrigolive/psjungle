package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func getPidsByPort(port string) ([]int, error) {
	cmd := exec.Command("lsof", "-i", ":"+port, "-t")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get PIDs by port: %v", err)
	}

	var pids []int
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err == nil {
			pids = append(pids, pid)
		}
	}

	return pids, nil
}

func getPidsByName(name string) ([]int, error) {
	cmd := exec.Command("pgrep", "-i", name)
	output, err := cmd.Output()
	if err != nil {
		// If no processes found, pgrep returns exit code 1
		if _, ok := err.(*exec.ExitError); ok {
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to get PIDs by name: %v", err)
	}

	var pids []int
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err == nil {
			pids = append(pids, pid)
		}
	}

	return pids, nil
}

func getPidsByRegex(pattern string) ([]int, error) {
	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	// Get all processes with ps
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %v", err)
	}

	var pids []int
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	// Skip the header line
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				pid, err := strconv.Atoi(fields[1])
				if err == nil {
					pids = append(pids, pid)
				}
			}
		}
	}

	return pids, nil
}

func pstreeBoth(targetPid int) error {
	cmd := exec.Command("pstree", "-w", "-p", strconv.Itoa(targetPid))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		// Replace zero-padded PIDs with non-padded PIDs and highlight target PID
		pidPattern := fmt.Sprintf(`\b0*%d\b`, targetPid)
		targetPidRegex := regexp.MustCompile(pidPattern)

		// First highlight the target PID
		highlightedLine := targetPidRegex.ReplaceAllStringFunc(line, func(match string) string {
			return fmt.Sprintf("\033[32m%d\033[0m", targetPid)
		})

		// Then remove leading zeros from all other PIDs
		cleanedLine := regexp.MustCompile(`\b0+(\d+)`).ReplaceAllStringFunc(highlightedLine, func(match string) string {
			// Extract the number part and convert it back without leading zeros
			numStr := regexp.MustCompile(`0*(\d+)`).FindStringSubmatch(match)[1]
			num, _ := strconv.Atoi(numStr)
			// If this is our target PID, it's already highlighted, so return as is
			if num == targetPid {
				return match
			}
			// Otherwise, just remove leading zeros
			return strconv.Itoa(num)
		})
		fmt.Println(cleanedLine)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %v", err)
	}

	return nil
}

func runPstree(input string) error {
	var pids []int
	var err error

	// Check if input is a PID (only numbers)
	if regexp.MustCompile(`^\d+$`).MatchString(input) {
		pid, convErr := strconv.Atoi(input)
		if convErr != nil {
			return fmt.Errorf("invalid PID '%s'", input)
		}
		pids = []int{pid}
	} else if strings.HasPrefix(input, ":") {
		// Port matching
		port := input[1:]
		pids, err = getPidsByPort(port)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(input, "/") {
		// Regex pattern matching
		pattern := input[1:]
		pids, err = getPidsByRegex(pattern)
		if err != nil {
			return err
		}
	} else {
		// Process name matching
		pids, err = getPidsByName(input)
		if err != nil {
			return err
		}
	}

	if len(pids) == 0 {
		fmt.Println("No processes found")
		os.Exit(1)
	}

	// Display process trees for all matching PIDs
	for _, pid := range pids {
		if len(pids) > 1 {
			fmt.Printf("Process tree for PID %d:\n", pid)
		}
		if err := pstreeBoth(pid); err != nil {
			fmt.Printf("Error for PID %d: %v\n", pid, err)
		}
		if len(pids) > 1 {
			fmt.Println()
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "psjungle",
		Usage: "Display process trees for PIDs, ports, process names, or regex patterns",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Value:   "",
				Usage:   "Watch mode with refresh interval (use -w=2 or -w2 for 2 seconds refresh, then provide PID/port/name)",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				cli.ShowAppHelp(c)
				return cli.Exit("", 1)
			}

			input := c.Args().First()
			watchValue := c.String("watch")

			// Check if watch flag was explicitly set
			if c.IsSet("watch") {
				// Parse watch interval from flag value
				watchInterval := 2 // default to 2 seconds
				if watchValue != "" {
					// Remove '=' if present in flag (e.g., -w=2)
					value := strings.TrimPrefix(watchValue, "=")
					if i, err := strconv.Atoi(value); err == nil {
						watchInterval = i
					}
				}

				// Watch mode
				for {
					// Clear screen
					fmt.Print("\033[H\033[2J")
					// Print status line
					if watchValue == "" {
						fmt.Printf("Every 2.0s: psjungle -w %s\n\n", input)
					} else {
						fmt.Printf("Every %.1fs: psjungle -w%s %s\n\n", float64(watchInterval), watchValue, input)
					}
					if err := runPstree(input); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					time.Sleep(time.Duration(watchInterval) * time.Second)
				}
			} else {
				// Normal mode
				if err := runPstree(input); err != nil {
					return cli.Exit(err.Error(), 1)
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
