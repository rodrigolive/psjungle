package psjungle

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
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

// createMinimalTree creates a minimal process tree with just the target process and its children
func createMinimalTree(targetProc *process.Process, procMap map[int32]*process.Process) *ProcessNode {
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

// getRootProcessNode gets or creates the root process node (PID 1)
func getRootProcessNode(procMap map[int32]*process.Process) (*ProcessNode, error) {
	rootProc, exists := procMap[1]
	if !exists {
		p, err := process.NewProcess(1)
		if err != nil {
			return nil, err
		}
		rootProc = p
	}

	rootNode := &ProcessNode{
		Process:  rootProc,
		Children: []*ProcessNode{},
		Depth:    0,
		IsTarget: false,
		Parent:   nil,
	}

	return rootNode, nil
}

// buildParentChainNodes builds the chain of parent nodes from root to target
func buildParentChainNodes(rootNode *ProcessNode, parentChain []int32, targetPid int, procMap map[int32]*process.Process) *ProcessNode {
	currentNode := rootNode

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

	return currentNode
}

// addTargetNode adds the target process node to the tree
func addTargetNode(currentNode *ProcessNode, targetProc *process.Process) *ProcessNode {
	targetNode := &ProcessNode{
		Process:  targetProc,
		Children: []*ProcessNode{},
		Depth:    0,
		IsTarget: true,
		Parent:   currentNode,
	}

	currentNode.Children = append(currentNode.Children, targetNode)
	return targetNode
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
		return createMinimalTree(targetProc, procMap)
	}

	// If parentChain is empty, build a minimal tree
	if len(parentChain) == 0 {
		return createMinimalTree(targetProc, procMap)
	}

	// Get root process node (PID 1)
	rootNode, err := getRootProcessNode(procMap)
	if err != nil {
		// If we can't get PID 1, just build a tree with the target process and its children
		return createMinimalTree(targetProc, procMap)
	}

	// Build the chain of parent nodes from root to target
	currentNode := buildParentChainNodes(rootNode, parentChain, targetPid, procMap)

	// Add the target node
	targetNode := addTargetNode(currentNode, targetProc)
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

// parseSignal parses a signal string and returns the corresponding syscall.Signal
func parseSignal(signalStr string) (syscall.Signal, error) {
	switch strings.ToLower(signalStr) {
	case "", "term":
		return syscall.SIGTERM, nil
	case "hup":
		return syscall.SIGHUP, nil
	case "int":
		return syscall.SIGINT, nil
	case "kill":
		return syscall.SIGKILL, nil
	default:
		// Try to parse as a number
		if sigNum, err := strconv.Atoi(signalStr); err == nil {
			// Validate that it's a valid signal number
			if sigNum >= 0 && sigNum <= 64 { // Most systems have signals in this range
				return syscall.Signal(sigNum), nil
			}
		}
		return 0, fmt.Errorf("invalid signal: %s", signalStr)
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

// defineFlags returns the CLI flags for the psjungle application
func defineFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "watch",
			Aliases: []string{"w"},
			Value:   "",
			Usage:   "Watch mode with refresh interval. Use formats like -w=2, -w2, or -w 2 for 2 seconds refresh. Provide a PID/port/name to watch (only watches the first target when multiple targets are specified).",
		},
		&cli.BoolFlag{
			Name:    "flat",
			Aliases: []string{"f"},
			Value:   false,
			Usage:   "Flat mode - removes Unicode tree indentation and lists processes left-aligned",
		},
		&cli.BoolFlag{
			Name:    "strict",
			Aliases: []string{"s"},
			Value:   false,
			Usage:   "Strict mode - treats input as exact string to match, not as regex pattern",
		},
		&cli.StringFlag{
			Name:    "host",
			Aliases: []string{"H"},
			Value:   "",
			Usage:   "Filter port connections by specific host (e.g., 127.0.0.1 or 0.0.0.0). Only applies to :port syntax.",
		},
		&cli.StringFlag{
			Name:    "kill",
			Aliases: []string{"k"},
			Value:   "",
			Usage:   "Send signal to matching processes. Use formats like -k, -k=9, -k term. Only sends signal after displaying tree.",
		},
	}
}

// BuildTreePrefix creates a tree prefix with Unicode characters for visual representation
func BuildTreePrefix(node *ProcessNode, nextSiblings []*ProcessNode, flatMode bool) string {
	if flatMode {
		return ""
	}
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
func printNodeWithTree(node *ProcessNode, targetPid int, nextSiblings []*ProcessNode, flatMode bool) {
	prefix := BuildTreePrefix(node, nextSiblings, flatMode)

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
		printNodeWithTree(child, targetPid, siblings, flatMode)
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
func pstreeBoth(targetPid int, flatMode bool) error {
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
	printNodeWithTree(tree, targetPid, []*ProcessNode{}, flatMode)
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

// getDirectChildrenPids recursively gets direct children PIDs for a process
func getDirectChildrenPids(parentPid int32, procMap map[int32]*process.Process) []int {
	var children []int
	for _, proc := range procMap {
		ppid, err := proc.Ppid()
		if err != nil {
			continue
		}
		if ppid == parentPid {
			children = append(children, int(proc.Pid))
			// Recursively get grandchildren
		 grandchildren := getDirectChildrenPids(proc.Pid, procMap)
		 children = append(children, grandchildren...)
		}
	}
	return children
}

// getProcessTreePids builds a tree for the target PID and returns all PIDs in that tree
func getProcessTreePids(targetPid int) []int {
	// Get all processes
	procMap, err := getAllProcesses()
	if err != nil {
		return []int{targetPid}
	}

	// Create a focused set of PIDs that are more relevant for comparison
	// Include the target process, its parent, and all its descendants
	pids := []int{targetPid}

	// Get the target process
	targetProc, exists := procMap[int32(targetPid)]
	if !exists {
		// Try to get target process directly
		p, err := process.NewProcess(int32(targetPid))
		if err != nil {
			return []int{targetPid}
		}
		targetProc = p
	}

	// Get parent PID
	ppid, err := targetProc.Ppid()
	if err == nil {
		pids = append(pids, int(ppid))
	}

	// Get all direct children and descendants
	children := getDirectChildrenPids(int32(targetPid), procMap)
	pids = append(pids, children...)

	return pids
}

// parseInputs determines which processes to display trees for based on input arguments.
// Returns a list of PIDs to process.
func parseInputs(inputs []string, strictMode bool, host string) ([]int, error) {
	var allPids []int
	var err error

	// If we have multiple inputs, treat them all as PIDs
	if len(inputs) > 1 {
		for _, input := range inputs {
			// Check if input is a PID (only numbers)
			if regexp.MustCompile(`^\d+$`).MatchString(input) {
				pid, convErr := strconv.Atoi(input)
				if convErr != nil {
					return nil, fmt.Errorf("invalid PID '%s'", input)
				}
				allPids = append(allPids, pid)
			} else {
				// For non-numeric inputs when we have multiple arguments,
				// we could extend this but for now we'll treat them as invalid
				return nil, fmt.Errorf("invalid PID '%s'", input)
			}
		}
	} else if len(inputs) == 1 {
		// Single input - use existing logic
		input := inputs[0]
		var pids []int

		// Check if input is a PID (only numbers)
		if regexp.MustCompile(`^\d+$`).MatchString(input) {
			pid, convErr := strconv.Atoi(input)
			if convErr != nil {
				return nil, fmt.Errorf("invalid PID '%s'", input)
			}
			pids = []int{pid}
		} else if strings.HasPrefix(input, ":") {
			// Port matching
			port := strings.TrimPrefix(input, ":")
			portNum, convErr := strconv.Atoi(port)
			if convErr != nil || portNum < 0 || portNum > 65535 {
				return nil, fmt.Errorf("invalid port '%s'", port)
			}
			pids, err = ByPort(uint32(portNum), host)
			if err != nil {
				return nil, err
			}
		} else {
			// Regex or strict string matching
			pids, err = ByRegex(input, strictMode)
			if err != nil {
				return nil, err
			}
		}

		allPids = pids
	} else {
		return nil, fmt.Errorf("no input provided")
	}

	if len(allPids) == 0 {
		fmt.Println("No processes found")
		os.Exit(1)
	}

	return allPids, nil
}

// runPstree dispatches based on user input and prints matching trees.
// When multiple inputs are provided, they are all treated as PIDs.
// Returns the list of PIDs that were processed.
func runPstree(inputs []string, flatMode bool, strictMode bool, host string) ([]int, error) {
	allPids, err := parseInputs(inputs, strictMode, host)
	if err != nil {
		return nil, err
	}

	// For multiple PIDs, we want to avoid showing duplicate trees
	// Keep track of processes already shown in a tree
	shownPids := make(map[int]bool)

	// Display process trees for all PIDs, but avoid duplicates
	// Keep track of which PIDs we actually displayed trees for
	return displayProcessTrees(allPids, flatMode, shownPids)
}

// appUsageText contains the extensive usage documentation for psjungle
const appUsageText = `psjungle [options] [PID|:port|pattern]...

EXAMPLES:
   psjungle 1234               Display process tree for PID 1234
   psjungle :8080              Display process trees for processes listening on port 8080
   psjungle :8080 --host 127.0.0.1  Display process trees for processes listening on port 8080 on localhost only
   psjungle :8080 --host 0.0.0.0    Display process trees for processes listening on port 8080 on all interfaces
   psjungle node               Display process trees for processes matching "node" (regex pattern)
   psjungle "node.*8080"        Display process trees for processes matching regex pattern
   psjungle -s "node.*8080"    Display process trees for processes with exact string "node.*8080" in name or command line
   psjungle 1234 5678          Display process trees for multiple PIDs (intelligently shows separate trees only when needed)
   psjungle 1234 5678 9012     Display process trees for three PIDs
   psjungle 1 1234 4321        Display process trees for root process and two other PIDs
   psjungle -w 1234            Watch process tree for PID 1234 (refresh every 2 seconds)
   psjungle -w=5 :3000          Watch processes listening on port 3000 (refresh every 5 seconds)
   psjungle -w2 1234           Watch process tree for PID 1234 (refresh every 2 seconds)
   psjungle -s -w2 starman     Watch process trees for processes with "starman" in name or command line
   psjungle -k 1234            Display process tree for PID 1234 and send SIGTERM to it
   psjungle -k=9 :8080         Display process trees for processes on port 8080 and send SIGKILL to them
   psjungle -k hup node        Display process trees for processes matching "node" and send SIGHUP to them

By default, patterns are treated as regex. Use the -s/--strict flag to match exact strings.
When multiple arguments are provided, they are all treated as PIDs and psjungle intelligently
shows separate process trees only when needed (when PIDs are not in the same process tree).
Use the --host flag to filter port connections by specific host. Only applies to :port syntax.
Use the --kill/-k flag to send signals to matching processes after displaying trees.

Output format: PID CPU% Memory CommandLine
Memory usage is shown in human-readable format (KB/MB/GB). Processes are highlighted in green.`

// NewApp builds the CLI application configuration.
func NewApp() *cli.App {
	app := &cli.App{
		Name:      "psjungle",
		Usage:     "Display process trees for PIDs, ports, or patterns (regex by default, strict string with -s flag)",
		UsageText: appUsageText,
		Flags:     defineFlags(),
		Action: func(c *cli.Context) error {
			flatMode := c.Bool("flat")
			strictMode := c.Bool("strict")

			// Get all inputs (all non-flag arguments)
			inputs := make([]string, c.NArg())
			for i := 0; i < c.NArg(); i++ {
				inputs[i] = c.Args().Get(i)
			}

			host := c.String("host")
			killValue := c.String("kill")

			// Check if watch flag was explicitly set
			if c.IsSet("watch") {
				// Validate arguments for watch mode
				if c.NArg() < 1 {
					cli.ShowAppHelp(c)
					return cli.Exit("Watch mode requires at least one target PID/port/name", 1)
				}
				return handleWatchMode(c, inputs, flatMode, strictMode, host, killValue)
			}

			return handleNormalMode(c, inputs, flatMode, strictMode, host, killValue)
		},
	}

	return app
}

// handleWatchMode processes the watch mode functionality
func handleWatchMode(c *cli.Context, inputs []string, flatMode bool, strictMode bool, host string, killValue string) error {
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

	var killSignal syscall.Signal
	var useKill bool

	// Parse kill signal if kill flag is set
	if c.IsSet("kill") {
		var err error
		killSignal, err = parseSignal(killValue)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Error parsing signal: %v", err), 1)
		}
		useKill = true
	}

	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")
		// Print status line with all targets
		fmt.Printf("Every %.1fs: psjungle -w%s", float64(watchInterval), watchValue)
		if strictMode {
			fmt.Print(" -s")
		}
		if host != "" {
			fmt.Printf(" --host %s", host)
		}
		if useKill {
			if killValue == "" {
				fmt.Print(" -k")
			} else {
				fmt.Printf(" -k=%s", killValue)
			}
		}
		for _, input := range inputs {
			fmt.Printf(" %s", input)
		}
		fmt.Println()
		fmt.Println()
		// Run pstree and get the list of processed PIDs
		processedPids, err := runPstree(inputs, flatMode, strictMode, host)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		// If kill flag is set, send signal to processed PIDs
		if useKill {
			// Send signal to all processed PIDs
			for _, pid := range processedPids {
				proc, err := process.NewProcess(int32(pid))
				if err != nil {
					fmt.Printf("Warning: Could not create process object for PID %d: %v\n", pid, err)
					continue
				}

				// Send the signal
				if err := proc.SendSignal(killSignal); err != nil {
					fmt.Printf("Warning: Could not send signal to PID %d: %v\n", pid, err)
				} else {
					fmt.Printf("Sent signal %d to PID %d\n", killSignal, pid)
				}
			}
		}
		time.Sleep(time.Duration(watchInterval) * time.Second)
	}
}

// handleNormalMode processes the normal (non-watch) mode functionality
func handleNormalMode(c *cli.Context, inputs []string, flatMode bool, strictMode bool, host string, killValue string) error {
	// If we get here and have no arguments, show help
	if c.NArg() < 1 {
		cli.ShowAppHelp(c)
		return cli.Exit("", 1)
	}

	// Run pstree and get the list of processed PIDs
	processedPids, err := runPstree(inputs, flatMode, strictMode, host)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	// If kill flag is set, send signal to processed PIDs
	if c.IsSet("kill") {
		signal, err := parseSignal(killValue)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Error parsing signal: %v", err), 1)
		}

		// Send signal to all processed PIDs
		for _, pid := range processedPids {
			proc, err := process.NewProcess(int32(pid))
			if err != nil {
				fmt.Printf("Warning: Could not create process object for PID %d: %v\n", pid, err)
				continue
			}

			// Send the signal
			if err := proc.SendSignal(signal); err != nil {
				fmt.Printf("Warning: Could not send signal to PID %d: %v\n", pid, err)
			} else {
				fmt.Printf("Sent signal %d to PID %d\n", signal, pid)
			}
		}
	}

	return nil
}

// displayProcessTrees shows process trees for all PIDs, avoiding duplicates
// Returns the list of PIDs that were processed (had trees displayed)
func displayProcessTrees(allPids []int, flatMode bool, shownPids map[int]bool) ([]int, error) {
	// Display process trees for all PIDs, but avoid duplicates
	// Keep track of which PIDs we actually displayed trees for
	var processedPids []int
	firstTree := true
	for _, pid := range allPids {
		// Skip if process doesn't exist
		_, err := process.NewProcess(int32(pid))
		if err != nil {
			fmt.Printf("Process %d not found\n", pid)
			continue
		}

		// Get all PIDs in this process's tree
		treePids := getProcessTreePids(pid)

		// Check if any PID in this tree has already been shown
		alreadyShown := false
		for _, treePid := range treePids {
			if shownPids[treePid] {
				alreadyShown = true
				break
			}
		}

		// If this tree hasn't been shown yet, display it
		if !alreadyShown {
			if !firstTree {
				fmt.Println()
			}
			if len(allPids) > 1 {
				fmt.Printf("Process tree for PID %d:\n", pid)
			}
			if err := pstreeBoth(pid, flatMode); err != nil {
				fmt.Printf("Error for PID %d: %v\n", pid, err)
			}

			// Add the target PID to our processed list
			processedPids = append(processedPids, pid)

			// Mark all processes in this tree as shown
			for _, treePid := range treePids {
				shownPids[treePid] = true
			}

			firstTree = false
		}
	}

	return processedPids, nil
}

// Run executes the CLI application with provided args.
func Run(args []string) error {
	return NewApp().Run(args)
}
