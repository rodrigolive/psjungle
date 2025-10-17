package psjungle

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/urfave/cli/v2"
)

// ProcessNode represents a node in the process tree
type ProcessNode struct {
	Process  *process.Process
	Children []*ProcessNode
	Depth    int
	IsTarget bool
	Parent   *ProcessNode
}

// getAllProcesses returns a map of all processes indexed by PID
func getAllProcesses() (map[int32]*process.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	procMap := make(map[int32]*process.Process)
	for _, proc := range processes {
		procMap[proc.Pid] = proc
	}

	return procMap, nil
}

// buildProcessTree recursively builds a process tree starting from the given PID
func buildProcessTree(targetPid int, currentPid int32, depth int, procMap map[int32]*process.Process, visited map[int32]bool) *ProcessNode {
	// Avoid infinite loops
	if visited[currentPid] {
		return nil
	}
	visited[currentPid] = true

	// Try to get process from procMap first, then directly
	proc, exists := procMap[currentPid]
	if !exists {
		// Process not found in our map, but it might still exist
		// Let's try to get it directly
		p, err := process.NewProcess(currentPid)
		if err != nil {
			return nil
		}
		proc = p
	}

	node := &ProcessNode{
		Process:  proc,
		Children: []*ProcessNode{},
		Depth:    depth,
		IsTarget: int(currentPid) == targetPid,
		Parent:   nil,
	}

	// Instead of using proc.Children(), we'll iterate through all processes
	// and find those that have currentPid as their parent
	for _, p := range procMap {
		// Skip if we've already visited this process
		if visited[p.Pid] {
			continue
		}

		// Get the parent PID of this process
		ppid, err := p.Ppid()
		if err != nil {
			// Skip processes where we can't get the parent PID
			continue
		}

		// If the parent PID matches currentPid, this is a child
		if ppid == currentPid {
			childNode := buildProcessTree(targetPid, p.Pid, depth+1, procMap, visited)
			if childNode != nil {
				childNode.Parent = node
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node
}

// findParentChain finds the chain of parent processes for a target PID
func findParentChain(targetPid int32, procMap map[int32]*process.Process) ([]int32, error) {
	chain := []int32{}
	currentPid := targetPid

	// Traverse up the parent chain until we reach PID 1 or find a cycle
	for currentPid > 1 {
		// Try to get process from procMap first, then directly
		proc, exists := procMap[currentPid]
		if !exists {
			// Process not found in our map, but it might still exist
			// Let's try to get it directly
			p, err := process.NewProcess(currentPid)
			if err != nil {
				// Process doesn't exist, break the chain
				break
			}
			proc = p
		}

		ppid, err := proc.Ppid()
		if err != nil {
			break
		}

		// Prevent infinite loops
		if ppid <= 1 || ppid == currentPid {
			if ppid == 1 {
				chain = append(chain, ppid)
			}
			break
		}

		chain = append(chain, ppid)
		currentPid = ppid
	}

	// Reverse the chain so it goes from root to target
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// buildFocusedTree builds a tree focused on the target process and its ancestors/descendants
func buildFocusedTree(targetPid int, procMap map[int32]*process.Process) *ProcessNode {
	targetProc, exists := procMap[int32(targetPid)]
	if !exists {
		// Try to get target process directly
		p, err := process.NewProcess(int32(targetPid))
		if err != nil {
			return nil
		}
		targetProc = p
	}

	// Find parent chain up to PID 1
	parentChain, err := findParentChain(int32(targetPid), procMap)
	if err != nil {
		// If we can't get parents, just build a tree with the target process and its children
		node := &ProcessNode{
			Process:  targetProc,
			Children: []*ProcessNode{},
			Depth:    0,
			IsTarget: true,
			Parent:   nil,
		}
		visited := make(map[int32]bool)
		addChildren(node, procMap, visited, 1)
		return node
	}

	// If parentChain is empty, build a minimal tree
	if len(parentChain) == 0 {
		node := &ProcessNode{
			Process:  targetProc,
			Children: []*ProcessNode{},
			Depth:    0,
			IsTarget: true,
			Parent:   nil,
		}
		visited := make(map[int32]bool)
		addChildren(node, procMap, visited, 1)
		return node
	}

	// Build the direct chain from PID 1 to target process
	var rootNode *ProcessNode
	var currentNode *ProcessNode

	// Start with PID 1
	rootProc, exists := procMap[1]
	if !exists {
		p, err := process.NewProcess(1)
		if err != nil {
			// If we can't get PID 1, just build a tree with the target process and its children
			node := &ProcessNode{
				Process:  targetProc,
				Children: []*ProcessNode{},
				Depth:    0,
				IsTarget: true,
				Parent:   nil,
			}
			visited := make(map[int32]bool)
			addChildren(node, procMap, visited, 1)
			return node
		}
		rootProc = p
	}

	rootNode = &ProcessNode{
		Process:  rootProc,
		Children: []*ProcessNode{},
		Depth:    0,
		IsTarget: false,
		Parent:   nil,
	}
	currentNode = rootNode

	// Build nodes for each process in the parent chain (excluding PID 1 which we already created)
	for _, pid := range parentChain {
		// Skip PID 1 as we already created it
		if pid == 1 {
			continue
		}

		// Skip if this is the target PID, we'll add it separately
		if int(pid) == targetPid {
			continue
		}

		proc, exists := procMap[pid]
		if !exists {
			p, err := process.NewProcess(pid)
			if err != nil {
				break
			}
			proc = p
		}

		node := &ProcessNode{
			Process:  proc,
			Children: []*ProcessNode{},
			Depth:    0,
			IsTarget: false,
			Parent:   currentNode,
		}

		currentNode.Children = append(currentNode.Children, node)
		currentNode = node
	}

	// Add the target node
	targetNode := &ProcessNode{
		Process:  targetProc,
		Children: []*ProcessNode{},
		Depth:    0,
		IsTarget: true,
		Parent:   currentNode,
	}
	currentNode.Children = append(currentNode.Children, targetNode)
	currentNode = targetNode

	// Add children of the target process
	visited := make(map[int32]bool)
	addChildren(targetNode, procMap, visited, targetNode.Depth+1)

	// Adjust depths to be positive starting from root
	adjustDepths(rootNode, 0)

	return rootNode
}

// addChildren recursively adds direct children to a node
func addChildren(node *ProcessNode, procMap map[int32]*process.Process, visited map[int32]bool, depth int) {
	// Avoid infinite loops
	if visited[node.Process.Pid] {
		return
	}
	visited[node.Process.Pid] = true

	currentPid := node.Process.Pid

	// Iterate through all processes to find direct children
	for _, proc := range procMap {
		// Skip already visited processes
		if visited[proc.Pid] {
			continue
		}

		ppid, err := proc.Ppid()
		if err != nil {
			continue
		}

		// If the parent PID matches currentPid, this is a direct child
		if ppid == currentPid {
			childNode := &ProcessNode{
				Process:  proc,
				Children: []*ProcessNode{},
				Depth:    depth,
				IsTarget: false,
				Parent:   node,
			}

			node.Children = append(node.Children, childNode)
			// Recursively add children of this child
			addChildren(childNode, procMap, visited, depth+1)
		}
	}
}

// adjustDepths adjusts node depths to start from 0 at the root
func adjustDepths(node *ProcessNode, depth int) {
	node.Depth = depth
	for _, child := range node.Children {
		adjustDepths(child, depth+1)
	}
}

// formatMemory formats memory usage in a human-readable way
func formatMemory(memoryKB uint64) string {
	if memoryKB < 1000 {
		return fmt.Sprintf("%dKB", memoryKB)
	} else if memoryKB < 1000000 {
		// For MB, show at least 3 digits with up to 2 decimal places
		mb := float64(memoryKB) / 1000.0
		if mb < 10 {
			return fmt.Sprintf("%.2fMB", mb)
		} else {
			return fmt.Sprintf("%.1fMB", mb)
		}
	} else {
		// For GB, show at least 3 digits with up to 2 decimal places
		gb := float64(memoryKB) / 1000000.0
		if gb < 10 {
			return fmt.Sprintf("%.2fGB", gb)
		} else {
			return fmt.Sprintf("%.1fGB", gb)
		}
	}
}

// BuildTreePrefix creates a tree prefix with Unicode characters for visual representation
func BuildTreePrefix(node *ProcessNode, nextSiblings []*ProcessNode) string {
	if node.Depth == 0 {
		return ""
	}

	var prefix strings.Builder
	// Create a slice of ancestors from root to parent of current node
	ancestors := []*ProcessNode{}
	current := node.Parent
	for current != nil {
		ancestors = append([]*ProcessNode{current}, ancestors...)
		current = current.Parent
	}

	// Build the prefix by traversing ancestors
	for i := 0; i < len(ancestors); i++ {
		if i == len(ancestors)-1 {
			// This is the parent of the current node
			// Check if there are more siblings (children after this node)
			if len(nextSiblings) > 0 {
				prefix.WriteString("├── ")
			} else {
				prefix.WriteString("└── ")
			}
		} else {
			// For intermediate ancestors, draw vertical line if there are more siblings
			hasSiblings := false
			if i+1 < len(ancestors) {
				parent := ancestors[i]
				child := ancestors[i+1]
				for _, c := range parent.Children {
					if c == child {
						break
					}
					hasSiblings = true
				}
			}
			if hasSiblings {
				prefix.WriteString("│   ")
			} else {
				prefix.WriteString("    ")
			}
		}
	}

	return prefix.String()
}

// printNodeWithTree prints the process tree nodes with proper indentation
func printNodeWithTree(node *ProcessNode, targetPid int, nextSiblings []*ProcessNode) {
	prefix := BuildTreePrefix(node, nextSiblings)

	// Get process info
	pid := node.Process.Pid
	name, _ := node.Process.Name()
	cmdline, _ := node.Process.Cmdline()
	cpuPercent, _ := node.Process.CPUPercent()
	memInfo, _ := node.Process.MemoryInfo()
	var rss uint64
	if memInfo != nil {
		rss = memInfo.RSS / 1024 // Convert to KB
	}

	if cmdline == "" {
		cmdline = name
	}

	// Format memory usage in human-readable way
	memStr := formatMemory(rss)

	// Print the process with highlighting if it's the target PID
	// Format similar to ps aux: PID, CPU%, MEM%, COMMAND
	if node.IsTarget {
		fmt.Printf("%s\033[32m%d %.1f %s %s\033[0m\n", prefix, pid, cpuPercent, memStr, cmdline)
	} else {
		fmt.Printf("%s%d %.1f %s %s\n", prefix, pid, cpuPercent, memStr, cmdline)
	}

	// Print children with proper tree characters
	for i, child := range node.Children {
		// Create a slice of siblings for this child (all children of the same parent)
		// Mark which siblings come after this child
		var siblings []*ProcessNode
		for j := i + 1; j < len(node.Children); j++ {
			siblings = append(siblings, node.Children[j])
		}
		printNodeWithTree(child, targetPid, siblings)
	}
}

// findTargetNode recursively finds the target node in the tree
func findTargetNode(node *ProcessNode, targetPid int) *ProcessNode {
	if int(node.Process.Pid) == targetPid {
		return node
	}

	for _, child := range node.Children {
		if targetNode := findTargetNode(child, targetPid); targetNode != nil {
			return targetNode
		}
	}

	return nil
}

// pstreeBoth displays the process tree for a given PID using gopsutil
func pstreeBoth(targetPid int) error {
	// Get all processes
	procMap, err := getAllProcesses()
	if err != nil {
		return fmt.Errorf("failed to get all processes: %v", err)
	}

	// Check if target process exists
	_, err = process.NewProcess(int32(targetPid))
	if err != nil {
		return fmt.Errorf("target process %d not found", targetPid)
	}

	// Build the focused tree containing the target process, its ancestors, and descendants
	tree := buildFocusedTree(targetPid, procMap)
	if tree == nil {
		return fmt.Errorf("failed to build process tree for PID %d", targetPid)
	}

	// Print the entire tree (it's already focused)
	printNodeWithTree(tree, targetPid, []*ProcessNode{})
	return nil
}

// collectProcessTreePids recursively collects all PIDs in a process tree
func collectProcessTreePids(node *ProcessNode) []int {
	var pids []int
	pids = append(pids, int(node.Process.Pid))

	for _, child := range node.Children {
		childPids := collectProcessTreePids(child)
		pids = append(pids, childPids...)
	}

	return pids
}

// getProcessTreePids builds a tree for the target PID and returns all PIDs in that tree
func getProcessTreePids(targetPid int) []int {
	// Get all processes
	procMap, err := getAllProcesses()
	if err != nil {
		return []int{targetPid}
	}

	// Build the focused tree
	tree := buildFocusedTree(targetPid, procMap)
	if tree == nil {
		return []int{targetPid}
	}

	// Collect all PIDs in the tree
	return collectProcessTreePids(tree)
}

// runPstree dispatches based on user input and prints matching trees.
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
		port := strings.TrimPrefix(input, ":")
		portNum, convErr := strconv.Atoi(port)
		if convErr != nil || portNum < 0 || portNum > 65535 {
			return fmt.Errorf("invalid port '%s'", port)
		}
		pids, err = ByPort(uint32(portNum))
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(input, "/") {
		// Regex pattern matching
		pattern := input[1:]
		pids, err = ByRegex(pattern)
		if err != nil {
			return err
		}
	} else {
		// Process name matching
		pids, err = ByName(input)
		if err != nil {
			return err
		}
	}

	if len(pids) == 0 {
		fmt.Println("No processes found")
		os.Exit(1)
	}

	// For port matching, we want to avoid showing duplicate trees
	// Keep track of processes already shown in a tree
	var shownPids map[int]bool
	if strings.HasPrefix(input, ":") {
		shownPids = make(map[int]bool)
	}

	// Display process trees for all matching PIDs
	for _, pid := range pids {
		// Skip if this process is already part of a previously displayed tree (for port matching)
		if strings.HasPrefix(input, ":") && shownPids[pid] {
			continue
		}

		if len(pids) > 1 {
			fmt.Printf("Process tree for PID %d:\n", pid)
		}
		if err := pstreeBoth(pid); err != nil {
			fmt.Printf("Error for PID %d: %v\n", pid, err)
		}

		// For port matching, mark all processes in this tree as shown
		if strings.HasPrefix(input, ":") {
			treePids := getProcessTreePids(pid)
			for _, treePid := range treePids {
				shownPids[treePid] = true
			}
		}

		if len(pids) > 1 {
			fmt.Println()
		}
	}

	return nil
}

// NewApp builds the CLI application configuration.
func NewApp() *cli.App {
	app := &cli.App{
		Name:      "psjungle",
		Usage:     "Display process trees for PIDs, ports, process names, or regex patterns",
		UsageText: "psjungle [options] [PID|:port|name|/pattern]\n\nEXAMPLES:\n   psjungle 1234               Display process tree for PID 1234\n   psjungle :8080              Display process trees for processes listening on port 8080\n   psjungle node               Display process trees for processes with \"node\" in their name\n   psjungle \"/node.*8080\"       Display process trees for processes matching regex pattern\n   psjungle -w 1234            Watch process tree for PID 1234 (refresh every 2 seconds)\n   psjungle -w=5 1234          Watch process tree for PID 1234 (refresh every 5 seconds)\n\nOutput format: PID CPU% MemoryUsage CommandLine\nMemory usage is shown in human-readable format (KB/MB/GB)",
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
			}

			// Normal mode
			if err := runPstree(input); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			return nil
		},
	}

	return app
}

// Run executes the CLI application with provided args.
func Run(args []string) error {
	return NewApp().Run(args)
}
