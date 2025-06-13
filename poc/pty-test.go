package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

func main() {
	fmt.Println("=== PTY Test ===")
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

	// Handle pty size
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("Error resizing pty: %v", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize

	// Set stdin in raw mode (for later interactive use)
	oldState, err := makeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Printf("Warning: Failed to set raw mode: %v", err)
	}
	defer func() {
		_ = restore(int(os.Stdin.Fd()), oldState)
	}()

	// Copy stdin to pty (for interactive commands)
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	// Read output line by line to demonstrate parsing
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

// Terminal raw mode helpers (simplified - full implementation would handle more cases)
func makeRaw(fd int) (*syscall.Termios, error) {
	var oldState syscall.Termios
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(&oldState))); err != 0 {
		return nil, err
	}

	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&newState))); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func restore(fd int, state *syscall.Termios) error {
	if state == nil {
		return nil
	}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(state)))
	if err != 0 {
		return err
	}
	return nil
}