package multiplex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)


// TokenTracker tracks token usage for an instance
type TokenTracker struct {
	usage      TokenUsage
	mu         sync.RWMutex
	listeners  []func(TokenUsage)
	
	// Parsing patterns
	patterns   *TokenPatterns
}

// TokenPatterns holds regex patterns for token extraction
type TokenPatterns struct {
	// Claude CLI patterns
	claudePattern    *regexp.Regexp
	claudeJSONPattern *regexp.Regexp
	
	// API response patterns
	apiUsagePattern  *regexp.Regexp
	
	// Cost patterns
	costPattern      *regexp.Regexp
	
	// Additional patterns
	humanPattern     *regexp.Regexp
	assistantPattern *regexp.Regexp
	totalPattern     *regexp.Regexp
}

// NewTokenTracker creates a new token tracker
func NewTokenTracker() *TokenTracker {
	return &TokenTracker{
		patterns: &TokenPatterns{
			// Pattern for Claude CLI token output
			claudePattern: regexp.MustCompile(`(?i)tokens?\s*[:\-\s]\s*input\s*[:\-\s]?\s*(\d+)\s*[,\s]*output\s*[:\-\s]?\s*(\d+)`),
			
			// Pattern for JSON-formatted usage
			claudeJSONPattern: regexp.MustCompile(`"usage"\s*:\s*\{[^}]*"input_tokens"\s*:\s*(\d+)[^}]*"output_tokens"\s*:\s*(\d+)`),
			
			// Pattern for API response
			apiUsagePattern: regexp.MustCompile(`"usage":\s*\{[^}]+\}`),
			
			// Pattern for cost estimation
			costPattern: regexp.MustCompile(`(?i)cost\s*[:\-\s]\s*\$?([0-9.]+)`),
			
			// Additional patterns for different output formats
			humanPattern:     regexp.MustCompile(`(?i)Human:\s*(\d+)\s*tokens?`),
			assistantPattern: regexp.MustCompile(`(?i)Assistant:\s*(\d+)\s*tokens?`),
			totalPattern:     regexp.MustCompile(`(?i)Total\s*tokens?\s*(?:used)?\s*[:\-\s]\s*(\d+)`),
		},
		listeners: make([]func(TokenUsage), 0),
	}
}

// ParseOutput parses output for token usage information
func (tt *TokenTracker) ParseOutput(data []byte) {
	// Try different parsing strategies
	if tt.tryParseClaudeFormat(data) {
		return
	}
	
	if tt.tryParseJSONFormat(data) {
		return
	}
	
	if tt.tryParseAPIFormat(data) {
		return
	}
	
	// Try line-by-line parsing
	tt.tryParseLines(data)
}

// tryParseClaudeFormat tries to parse Claude CLI format
func (tt *TokenTracker) tryParseClaudeFormat(data []byte) bool {
	matches := tt.patterns.claudePattern.FindSubmatch(data)
	if len(matches) >= 3 {
		input, err1 := strconv.ParseInt(string(matches[1]), 10, 64)
		output, err2 := strconv.ParseInt(string(matches[2]), 10, 64)
		
		if err1 == nil && err2 == nil {
			tt.updateUsage(input, output)
			return true
		}
	}
	return false
}

// tryParseJSONFormat tries to parse JSON format
func (tt *TokenTracker) tryParseJSONFormat(data []byte) bool {
	matches := tt.patterns.claudeJSONPattern.FindSubmatch(data)
	if len(matches) >= 3 {
		input, err1 := strconv.ParseInt(string(matches[1]), 10, 64)
		output, err2 := strconv.ParseInt(string(matches[2]), 10, 64)
		
		if err1 == nil && err2 == nil {
			tt.updateUsage(input, output)
			return true
		}
	}
	return false
}

// tryParseAPIFormat tries to parse API response format
func (tt *TokenTracker) tryParseAPIFormat(data []byte) bool {
	if matches := tt.patterns.apiUsagePattern.Find(data); matches != nil {
		var usage struct {
			Usage struct {
				InputTokens  int64 `json:"input_tokens"`
				OutputTokens int64 `json:"output_tokens"`
			} `json:"usage"`
		}
		
		// Extract JSON object
		start := bytes.Index(data, []byte("{"))
		end := bytes.LastIndex(data, []byte("}"))
		if start >= 0 && end > start {
			jsonData := data[start : end+1]
			if err := json.Unmarshal(jsonData, &usage); err == nil && usage.Usage.InputTokens > 0 {
				tt.updateUsage(usage.Usage.InputTokens, usage.Usage.OutputTokens)
				return true
			}
		}
	}
	return false
}

// tryParseLines tries line-by-line parsing
func (tt *TokenTracker) tryParseLines(data []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var humanTokens, assistantTokens int64
	var foundAny bool
	
	for scanner.Scan() {
		line := scanner.Bytes()
		
		// Try each format on individual lines
		if tt.tryParseClaudeFormat(line) {
			return
		}
		if tt.tryParseJSONFormat(line) {
			return
		}
		
		// Check for Human/Assistant patterns
		if matches := tt.patterns.humanPattern.FindSubmatch(line); len(matches) >= 2 {
			if val, err := strconv.ParseInt(string(matches[1]), 10, 64); err == nil {
				humanTokens = val
				foundAny = true
			}
		}
		
		if matches := tt.patterns.assistantPattern.FindSubmatch(line); len(matches) >= 2 {
			if val, err := strconv.ParseInt(string(matches[1]), 10, 64); err == nil {
				assistantTokens = val
				foundAny = true
			}
		}
		
		// Check for total tokens
		if matches := tt.patterns.totalPattern.FindSubmatch(line); len(matches) >= 2 {
			if val, err := strconv.ParseInt(string(matches[1]), 10, 64); err == nil {
				// If we have total but not input/output, assume it's all output
				if humanTokens == 0 && assistantTokens == 0 {
					assistantTokens = val
				}
				foundAny = true
			}
		}
		
		// Check for cost information
		if matches := tt.patterns.costPattern.FindSubmatch(line); len(matches) >= 2 {
			if cost, err := strconv.ParseFloat(string(matches[1]), 64); err == nil {
				tt.mu.Lock()
				tt.usage.EstimatedCost = cost
				tt.mu.Unlock()
			}
		}
	}
	
	// Update if we found any tokens
	if foundAny && (humanTokens > 0 || assistantTokens > 0) {
		tt.updateUsage(humanTokens, assistantTokens)
	}
}

// updateUsage updates the token usage
func (tt *TokenTracker) updateUsage(input, output int64) {
	tt.mu.Lock()
	tt.usage.InputTokens += input
	tt.usage.OutputTokens += output
	tt.usage.TotalTokens = tt.usage.InputTokens + tt.usage.OutputTokens
	tt.usage.LastUpdated = time.Now()
	
	// Estimate cost (simplified pricing model)
	// TODO: Make this configurable based on model
	inputCost := float64(tt.usage.InputTokens) * 0.003 / 1000  // $3 per 1M tokens
	outputCost := float64(tt.usage.OutputTokens) * 0.015 / 1000 // $15 per 1M tokens
	tt.usage.EstimatedCost = inputCost + outputCost
	
	usage := tt.usage
	tt.mu.Unlock()
	
	// Notify listeners
	for _, listener := range tt.listeners {
		listener(usage)
	}
}

// GetUsage returns current token usage
func (tt *TokenTracker) GetUsage() TokenUsage {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.usage
}

// Reset resets token usage
func (tt *TokenTracker) Reset() {
	tt.mu.Lock()
	tt.usage = TokenUsage{
		LastUpdated: time.Now(),
	}
	tt.mu.Unlock()
}

// AddListener adds a usage update listener
func (tt *TokenTracker) AddListener(listener func(TokenUsage)) {
	tt.listeners = append(tt.listeners, listener)
}

// TokenLimiter manages token limits across instances
type TokenLimiter struct {
	totalLimit int64
	instances  map[string]*InstanceTokens
	mu         sync.RWMutex
}

// InstanceTokens tracks per-instance token allocation
type InstanceTokens struct {
	Allocated int64
	Used      int64
	Remaining int64
}

// NewTokenLimiter creates a new token limiter
func NewTokenLimiter(totalLimit int64) *TokenLimiter {
	return &TokenLimiter{
		totalLimit: totalLimit,
		instances:  make(map[string]*InstanceTokens),
	}
}

// AllocateTokens allocates tokens to an instance
func (tl *TokenLimiter) AllocateTokens(instanceID string, tokens int64) bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	// Calculate total allocated
	var totalAllocated int64
	for _, inst := range tl.instances {
		totalAllocated += inst.Allocated
	}
	
	// Check if allocation is possible
	if totalAllocated+tokens > tl.totalLimit {
		return false
	}
	
	// Allocate tokens
	if _, exists := tl.instances[instanceID]; !exists {
		tl.instances[instanceID] = &InstanceTokens{}
	}
	
	tl.instances[instanceID].Allocated = tokens
	tl.instances[instanceID].Remaining = tokens
	
	return true
}

// UseTokens consumes tokens from an instance's allocation
func (tl *TokenLimiter) UseTokens(instanceID string, tokens int64) bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	inst, exists := tl.instances[instanceID]
	if !exists {
		return false
	}
	
	if inst.Remaining < tokens {
		return false
	}
	
	inst.Used += tokens
	inst.Remaining -= tokens
	
	return true
}

// GetInstanceTokens returns token information for an instance
func (tl *TokenLimiter) GetInstanceTokens(instanceID string) *InstanceTokens {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	
	if inst, exists := tl.instances[instanceID]; exists {
		// Return a copy
		return &InstanceTokens{
			Allocated: inst.Allocated,
			Used:      inst.Used,
			Remaining: inst.Remaining,
		}
	}
	
	return nil
}

// GetTotalUsage returns total token usage across all instances
func (tl *TokenLimiter) GetTotalUsage() (allocated, used, remaining int64) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	
	for _, inst := range tl.instances {
		allocated += inst.Allocated
		used += inst.Used
		remaining += inst.Remaining
	}
	
	return
}

// FormatTokenUsage formats token usage for display
func FormatTokenUsage(usage TokenUsage) string {
	var parts []string
	
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		parts = append(parts, fmt.Sprintf("In: %s", formatNumber(usage.InputTokens)))
		parts = append(parts, fmt.Sprintf("Out: %s", formatNumber(usage.OutputTokens)))
		parts = append(parts, fmt.Sprintf("Total: %s", formatNumber(usage.TotalTokens)))
	}
	
	if usage.EstimatedCost > 0 {
		parts = append(parts, fmt.Sprintf("~$%.4f", usage.EstimatedCost))
	}
	
	if len(parts) == 0 {
		return "No usage"
	}
	
	return strings.Join(parts, " | ")
}

// formatNumber formats a number with commas
func formatNumber(n int64) string {
	if n < 1000 {
		return strconv.FormatInt(n, 10)
	}
	
	// Simple comma formatting
	str := strconv.FormatInt(n, 10)
	var result []byte
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(digit))
	}
	
	return string(result)
}