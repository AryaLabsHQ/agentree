// Package main provides a mock Claude CLI for testing the multiplexer
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Mock responses to simulate Claude behavior
var responses = []struct {
	prompt string
	response string
	tokens struct {
		input  int
		output int
	}
}{
	{
		prompt: "hello",
		response: "Hello! I'm Claude, an AI assistant created by Anthropic. How can I help you today?",
		tokens: struct{ input, output int }{input: 12, output: 24},
	},
	{
		prompt: "test",
		response: "I understand you want to test something. What specifically would you like to test?",
		tokens: struct{ input, output int }{input: 8, output: 20},
	},
	{
		prompt: "implement",
		response: "I'll help you implement that. Let me analyze the requirements and create a solution...\n\n" +
			"Here's my implementation approach:\n" +
			"1. First, I'll examine the existing code structure\n" +
			"2. Then, I'll design the solution\n" +
			"3. Finally, I'll implement and test it\n\n" +
			"Let me start working on this...",
		tokens: struct{ input, output int }{input: 15, output: 87},
	},
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "code" {
		runInteractiveMode()
	} else {
		fmt.Println("Mock Claude CLI - Use 'mock-claude code' for interactive mode")
	}
}

func runInteractiveMode() {
	scanner := bufio.NewScanner(os.Stdin)
	rand.Seed(time.Now().UnixNano())
	
	// Print welcome message
	fmt.Println("\n\x1b[1;36mClaude Code\x1b[0m (Mock Version)")
	fmt.Println("\x1b[90mPress Ctrl+D to exit\x1b[0m\n")
	
	for {
		// Print prompt
		fmt.Print("\x1b[1;32mHuman:\x1b[0m ")
		
		// Read input
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		// Simulate thinking time
		fmt.Print("\n\x1b[1;33mAssistant:\x1b[0m ")
		fmt.Print("\x1b[90m(thinking...)\x1b[0m\r")
		time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
		fmt.Print("\x1b[1;33mAssistant:\x1b[0m              \r") // Clear thinking message
		fmt.Print("\x1b[1;33mAssistant:\x1b[0m ")
		
		// Find matching response
		response, tokens := findResponse(input)
		
		// Type out response with delays
		typeOut(response)
		
		// Print token usage
		fmt.Printf("\n\n\x1b[90mTokens: Input: %d, Output: %d\x1b[0m\n\n", 
			tokens.input, tokens.output)
	}
	
	fmt.Println("\n\x1b[90mGoodbye!\x1b[0m")
}

func findResponse(input string) (string, struct{ input, output int }) {
	lowerInput := strings.ToLower(input)
	
	// Check for matching prompts
	for _, r := range responses {
		if strings.Contains(lowerInput, r.prompt) {
			return r.response, r.tokens
		}
	}
	
	// Default response
	return fmt.Sprintf("I understand you're asking about '%s'. Let me help you with that...", input),
		struct{ input, output int }{input: len(input) / 5, output: 25}
}

func typeOut(text string) {
	words := strings.Fields(text)
	for i, word := range words {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(word)
		fmt.Print("\x1b[90m\u2588\x1b[0m") // Cursor
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		fmt.Print("\b ") // Remove cursor
	}
}

// Additional features to test
func init() {
	// Handle special environment variables
	if os.Getenv("MOCK_CLAUDE_SLOW") == "1" {
		// Slow mode for testing timeouts
		time.Sleep(5 * time.Second)
	}
	
	if os.Getenv("MOCK_CLAUDE_CRASH") == "1" {
		// Crash mode for testing error handling
		panic("Mock crash for testing")
	}
}