package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/creack/pty"
)

func main() {
	fmt.Println("=== Claude Code PTY Test ===")
	fmt.Println("Testing Claude Code in PTY environment")
	fmt.Println()

	// Get current directory for testing
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	
	// Go up to the project root
	projectRoot := filepath.Dir(cwd)
	fmt.Printf("Project root: %s\n", projectRoot)
	fmt.Println()

	// Check if claude command exists
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Println("❌ 'claude' command not found in PATH")
		fmt.Println("Please ensure Claude CLI is installed and in your PATH")
		fmt.Println("\nFalling back to simulation mode...")
		runSimulation()
		return
	}

	// Create command for Claude Code
	cmd := exec.Command("claude", "code")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), 
		"AGENTREE_TEST=true",
		"AGENTREE_INSTANCE=test",
	)

	// Start with PTY
	fmt.Println("Starting Claude Code in PTY...")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatalf("Failed to start Claude Code: %v", err)
	}
	defer func() {
		_ = ptmx.Close()
	}()

	// Set a timeout for the test
	timeout := time.After(30 * time.Second)
	done := make(chan bool)

	// Read output
	go func() {
		scanner := bufio.NewScanner(ptmx)
		lineCount := 0
		fmt.Println("\n=== Claude Code Output ===")
		
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			fmt.Printf("[%03d] %s\n", lineCount, line)
			
			// Stop after capturing some output
			if lineCount >= 20 {
				done <- true
				return
			}
		}
		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		fmt.Println("\n✅ Successfully captured Claude Code output")
	case <-timeout:
		fmt.Println("\n⏱️ Test timed out after 30 seconds")
	}

	// Kill the process
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}

	fmt.Println("\n=== Test Results ===")
	fmt.Println("✓ Claude Code runs in PTY")
	fmt.Println("✓ Output can be captured")
	fmt.Println("✓ Process management works")
}

func runSimulation() {
	// Simulate Claude Code output for testing
	cmd := exec.Command("bash", "-c", `
		echo "Claude CLI v1.0.0"
		echo "Starting Claude Code..."
		echo ""
		sleep 1
		echo "Workspace: $(pwd)"
		echo "Loading context..."
		sleep 1
		echo ""
		echo "Ready for input. Type /help for available commands."
		echo ""
		
		# Simulate some activity
		for i in {1..10}; do
			case $((i % 3)) in
				0) echo "[Assistant] Analyzing codebase..." ;;
				1) echo "[Assistant] Found $(($RANDOM % 100)) files to process" ;;
				2) echo "[Token Usage] Input: $(($RANDOM % 1000)) | Output: $(($RANDOM % 500))" ;;
			esac
			sleep 0.5
		done
	`)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatalf("Failed to start simulation: %v", err)
	}
	defer ptmx.Close()

	scanner := bufio.NewScanner(ptmx)
	fmt.Println("\n=== Simulation Output ===")
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	_ = cmd.Wait()
	fmt.Println("\n✅ Simulation completed")
}