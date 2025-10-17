package psjungle_test

import (
	"testing"

	"github.com/urfave/cli/v2"

	"psjungle/internal/psjungle"
)

func findStringFlag(flags []cli.Flag, name string) *cli.StringFlag {
	for _, f := range flags {
		if stringFlag, ok := f.(*cli.StringFlag); ok {
			if stringFlag.Name == name {
				return stringFlag
			}
			for _, alias := range stringFlag.Aliases {
				if alias == name {
					return stringFlag
				}
			}
		}
	}
	return nil
}

func TestWatchFlagDefault(t *testing.T) {
	app := psjungle.NewApp()
	flag := findStringFlag(app.Flags, "watch")
	if flag == nil {
		t.Fatalf("watch flag not found on app")
	}

	if flag.Value != "" {
		t.Fatalf("expected watch flag default to be empty string, got %q", flag.Value)
	}
}

func TestRunInvalidPort(t *testing.T) {
	app := psjungle.NewApp()
	originalExiter := cli.OsExiter
	defer func() { cli.OsExiter = originalExiter }()
	cli.OsExiter = func(int) {}

	err := app.Run([]string{"psjungle", ":notaport"})
	if err == nil {
		t.Fatalf("expected error when running with invalid port")
	}

	exitErr, ok := err.(cli.ExitCoder)
	if !ok {
		t.Fatalf("expected cli.ExitCoder, got %T", err)
	}

	if exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
	}
}
