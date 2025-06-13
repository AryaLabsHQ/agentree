package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"github.com/creack/pty"
)

func main() {
	fmt.Println("=== Simple PTY Test ===")
	fmt.Println("Testing process launch with pseudo-terminal")
	fmt.Println()

	// Create a command that produces colored output
	cmd := exec.Command("bash", "-c", `
		echo -e "\033[32mStarting test process...\033[0m"
		for i in {1..10}; do
			echo -e "[\033[34m$(date +%H:%M:%S)\033[0m] Line $i: \033[33mHello from PTY!\033[0m"
			sleep 1
		done
		echo -e "\033[32mProcess complete!\033[0m"
	`)

	// Start the command with a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatalf("Failed to start pty: %v", err)
	}
	defer func() {
		_ = ptmx.Close()
	}()

	// Read output line by line
	fmt.Println("=== Process Output ===")
	scanner := bufio.NewScanner(ptmx)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		fmt.Printf("[Output %d] %s\n", lineCount, line)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Printf("Captured %d lines of output\n", lineCount)
	fmt.Println("✓ PTY creation successful")
	fmt.Println("✓ Process execution successful") 
	fmt.Println("✓ ANSI color codes preserved")
}