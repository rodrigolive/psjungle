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


// ByRegex returns PIDs whose command line or name matches the provided pattern.
// If strict is true, performs exact substring matching. Otherwise, treats pattern as regex.
func ByRegex(pattern string, strict bool) ([]int, error) {
	var re *regexp.Regexp
	var err error

	if !strict {
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
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
	target := strings.ToLower(pattern)

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

		if strict {
			// Strict matching - exact substring match in either cmdline or name
			cmdline, err := proc.CmdlineWithContext(ctx)
			if err != nil || cmdline == "" {
				// Fallback to process name if full command line unavailable
				name, nameErr := proc.NameWithContext(ctx)
				if nameErr != nil || name == "" {
					continue
				}
				if !strings.Contains(strings.ToLower(name), target) {
					continue
				}
			} else if !strings.Contains(strings.ToLower(cmdline), target) {
				// Try matching against process name when command line does not match
				name, nameErr := proc.NameWithContext(ctx)
				if nameErr != nil || name == "" || !strings.Contains(strings.ToLower(name), target) {
					continue
				}
			}
		} else {
			// Regex matching
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
// By default, it looks for listening connections on all hosts.
func ByPort(port uint32, host string) ([]int, error) {
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

		// Check port match first
		portMatch := (conn.Laddr.Port == port || conn.Raddr.Port == port)
		if !portMatch {
			continue
		}

		// Check if we're filtering by host
		if host != "" {
			// For host filtering, we only check listening connections
			if conn.Status != "LISTEN" {
				continue
			}
			// Check if the local address matches the specified host
			// Handle special cases: "*" means all hosts, so it should match any host filter
			hostMatch := (conn.Laddr.IP == host || conn.Laddr.IP == "*" ||
				(host == "127.0.0.1" && conn.Laddr.IP == "localhost") ||
				(host == "localhost" && conn.Laddr.IP == "127.0.0.1"))
			if !hostMatch {
				continue
			}
		} else {
			// Default behavior: look for listening connections on all hosts
			// Skip non-listening connections unless they're communicating on the specified port
			if conn.Status != "LISTEN" && conn.Laddr.Port != port && conn.Raddr.Port != port {
				continue
			}
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
