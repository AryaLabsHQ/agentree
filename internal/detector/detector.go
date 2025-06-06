// Package detector provides auto-detection for package managers and build tools
package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PackageManager represents a detected package manager with its commands
type PackageManager struct {
	Name     string
	Detected bool
	Commands []string
}

// DetectSetupCommands auto-detects package manager and returns setup commands
func DetectSetupCommands(dir string) []string {
	var commands []string

	// Node.js package managers
	if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
		commands = append(commands, "pnpm install")
		if hasNpmScript(dir, "build") {
			commands = append(commands, "pnpm build")
		}
		return commands
	}

	if _, err := os.Stat(filepath.Join(dir, "package-lock.json")); err == nil {
		commands = append(commands, "npm install")
		if hasNpmScript(dir, "build") {
			commands = append(commands, "npm run build")
		}
		return commands
	}

	if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
		commands = append(commands, "yarn install")
		if hasNpmScript(dir, "build") {
			commands = append(commands, "yarn build")
		}
		return commands
	}

	// Other languages
	if _, err := os.Stat(filepath.Join(dir, "Cargo.lock")); err == nil {
		return []string{"cargo build"}
	}

	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return []string{"go mod download"}
	}

	if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
		return []string{"pip install -r requirements.txt"}
	}

	if _, err := os.Stat(filepath.Join(dir, "Gemfile.lock")); err == nil {
		return []string{"bundle install"}
	}

	return commands
}

// hasNpmScript checks if a package.json has a specific script
func hasNpmScript(dir, script string) bool {
	packagePath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	_, exists := pkg.Scripts[script]
	return exists
}