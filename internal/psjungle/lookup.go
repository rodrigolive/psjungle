package psjungle

import (
	"context"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	gonet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ByName returns PIDs whose process name contains the provided term (case-insensitive).
func ByName(name string) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	target := strings.ToLower(name)
	seen := make(map[int32]struct{})
	var matches []int

	for _, proc := range procs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		displayName, err := proc.NameWithContext(ctx)
		if err != nil || displayName == "" {
			continue
		}

		if strings.Contains(strings.ToLower(displayName), target) {
			pid := proc.Pid
			if _, ok := seen[pid]; !ok {
				seen[pid] = struct{}{}
				matches = append(matches, int(pid))
			}
		}
	}

	sort.Ints(matches)
	return matches, nil
}

// ByRegex returns PIDs whose command line matches the provided regex pattern.
func ByRegex(pattern string) ([]int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[int32]struct{})
	var matches []int
	currentPid := int32(os.Getpid())

	for _, proc := range procs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Skip the current process (psjungle itself)
		if proc.Pid == currentPid {
			continue
		}

		cmdline, err := proc.CmdlineWithContext(ctx)
		if err != nil || cmdline == "" {
			// Fallback to process name if full command line unavailable
			name, nameErr := proc.NameWithContext(ctx)
			if nameErr != nil || name == "" {
				continue
			}
			if !re.MatchString(name) {
				continue
			}
		} else if !re.MatchString(cmdline) {
			// Try matching against process name when command line does not match
			name, nameErr := proc.NameWithContext(ctx)
			if nameErr != nil || name == "" || !re.MatchString(name) {
				continue
			}
		}

		pid := proc.Pid
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		matches = append(matches, int(pid))
	}

	sort.Ints(matches)
	return matches, nil
}

// ByPort returns PIDs that have a connection bound to or communicating with the given port.
func ByPort(port uint32) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conns, err := gatherConnections(ctx)
	if err != nil {
		return nil, err
	}

	matches, seen := []int{}, make(map[int32]struct{})
	for _, conn := range conns {
		if conn.Pid == 0 {
			continue
		}
		if conn.Laddr.Port != port && conn.Raddr.Port != port {
			continue
		}
		if _, ok := seen[conn.Pid]; ok {
			continue
		}
		seen[conn.Pid] = struct{}{}
		matches = append(matches, int(conn.Pid))
	}

	sort.Ints(matches)
	return matches, nil
}

func gatherConnections(ctx context.Context) ([]gonet.ConnectionStat, error) {
	if conns, err := gonet.ConnectionsWithContext(ctx, "inet"); err == nil {
		return conns, nil
	}

	// Fallback: iterate processes and inspect their connections.
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var (
		all     []gonet.ConnectionStat
		lastErr error
	)

	for _, proc := range procs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		conns, err := proc.ConnectionsWithContext(ctx)
		if err != nil {
			lastErr = err
			continue
		}
		all = append(all, conns...)
	}

	if len(all) == 0 && lastErr != nil {
		return nil, lastErr
	}

	return all, nil
}
