package src

import (
	"context"
	"sort"
	"time"

	gonet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

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
