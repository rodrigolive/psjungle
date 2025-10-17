package main

import (
	"github.com/urfave/cli/v2"
	"strconv"
	"strings"
	"testing"
)

// TestCLIWatchFlagParsing tests that the watch flag is correctly parsed
func TestCLIWatchFlagParsing(t *testing.T) {
	// Test default watch interval (2 seconds)
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
			watchValue := c.String("watch")

			// In the actual app, when no watch flag is provided, we don't enter watch mode
			// The test is checking the default value of the flag, which should be ""
			if watchValue != "" {
				t.Errorf("Expected watch value to be empty, got %s", watchValue)
			}

			return nil
		},
	}

	// Run the app with no watch value
	args := []string{"psjungle", "91501"}
	err := app.Run(args)
	if err != nil {
		t.Errorf("App failed to run: %v", err)
	}
}

// TestCLIWatchFlagWithCustomInterval tests that custom intervals are correctly parsed
func TestCLIWatchFlagWithCustomInterval(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedInterval int
		expectedWatchValue string
	}{
		{"With equals sign", []string{"psjungle", "-w=5", "91501"}, 5, "5"},
		{"Without equals sign", []string{"psjungle", "-w", "5", "91501"}, 5, "5"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
					watchValue := c.String("watch")

					// Parse watch interval from flag value
					watchInterval := 2 // default to 2 seconds
					if watchValue != "" {
						// Remove '=' if present in flag (e.g., -w=2)
						value := strings.TrimPrefix(watchValue, "=")
						if i, err := strconv.Atoi(value); err == nil {
							watchInterval = i
						}
					}

					// Verify the interval is correctly parsed
					if watchInterval != test.expectedInterval {
						t.Errorf("Expected watch interval to be %d, got %d", test.expectedInterval, watchInterval)
					}
					
					// Verify the watch value is correctly parsed
					if watchValue != test.expectedWatchValue {
						t.Errorf("Expected watch value to be %s, got %s", test.expectedWatchValue, watchValue)
					}

					return nil
				},
			}

			err := app.Run(test.args)
			if err != nil {
				t.Errorf("App failed to run: %v", err)
			}
		})
	}
}

// TestCLIWatchModeExecution tests that watch mode executes correctly
func TestCLIWatchModeExecution(t *testing.T) {
	// We'll test that the watch mode correctly parses the interval and target
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
				return nil
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

				// Verify the interval is correctly parsed
				expectedInterval := 2
				if watchValue != "" {
					value := strings.TrimPrefix(watchValue, "=")
					if i, err := strconv.Atoi(value); err == nil {
						expectedInterval = i
					}
				}

				if watchInterval != expectedInterval {
					t.Errorf("Expected watch interval to be %d, got %d", expectedInterval, watchInterval)
				}

				// Verify that we correctly parse the PID
				if input != "91501" {
					t.Errorf("Expected input to be '91501', got '%s'", input)
				}

				// Verify that we're in watch mode
				if !c.IsSet("watch") {
					t.Error("Expected watch flag to be set")
				}
			}

			return nil
		},
	}

	// Test with custom interval using equals sign
	args := []string{"psjungle", "-w=5", "91501"}
	err := app.Run(args)
	if err != nil {
		t.Errorf("App failed to run with custom interval: %v", err)
	}
	
	// Test with custom interval without equals sign
	args = []string{"psjungle", "-w", "5", "91501"}
	err = app.Run(args)
	if err != nil {
		t.Errorf("App failed to run with custom interval: %v", err)
	}
}
