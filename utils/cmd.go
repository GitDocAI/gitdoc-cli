package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func ExecuteCmd(c string, dir string, arg ...string) (string, bool) {
	cmd := exec.Command(c, arg...)

	cmd.Dir = dir
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	out, err := cmd.Output()
	if err != nil {
		return string(out[:]) + err.Error(), true
	}

	if string(out[:]) == "" {
		return "Command executed successfully but no output", false
	}

	return string(out[:]), false
}

// outputWhitelistPatterns contains patterns that should be shown in the output
// All other lines will be filtered out
var outputWhitelistPatterns = []string{
	"error",
	"Error",
	"ERROR",
	"failed",
	"Failed",
	"FAILED",
	"✓",
	"✗",
	"Local:",
	"Network:",
	"Ready in",
	"http://",
	"https://",
	"Port",
	"Starting",
	"Listening",
}

// outputBlacklistPatterns contains patterns that should never be shown
// These take precedence over whitelist patterns
var outputBlacklistPatterns = []string{
	"nextra]",
	"Init git repository failed",
	"Only log non-404 errors",
	"} catch (error) {",
}

// shouldShowLine checks if a line should be shown based on whitelist patterns
func shouldShowLine(line string) bool {
	// Always hide empty lines
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Check blacklist first - these should never be shown
	for _, pattern := range outputBlacklistPatterns {
		if strings.Contains(line, pattern) {
			return false
		}
	}

	// Check if line matches any whitelist pattern
	for _, pattern := range outputWhitelistPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// extractLocalURL extracts the localhost URL from a line like "- Local:        http://localhost:3000"
func extractLocalURL(line string) string {
	// Pattern to match URLs after "Local:"
	re := regexp.MustCompile(`Local:\s+(https?://localhost:\d+|https?://127\.0\.0\.1:\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// filterOutput reads from a reader and writes to a writer, showing only whitelisted lines
func filterOutput(reader io.Reader, writer io.Writer, stopSpinner chan bool, spinnerControl chan string, clearLine chan bool, browserOpened *bool) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if shouldShowLine(line) {
			// Pause spinner and clear line before printing
			if spinnerControl != nil {
				spinnerControl <- "pause"
				time.Sleep(50 * time.Millisecond) // Wait for spinner to stop
			}
			if clearLine != nil {
				clearLine <- true
				time.Sleep(10 * time.Millisecond) // Wait for line to clear
			}

			fmt.Fprintln(writer, line)

			// Check for localhost URL and open browser
			if browserOpened != nil && !*browserOpened {
				if url := extractLocalURL(line); url != "" {
					*browserOpened = true
					go func() {
						time.Sleep(2 * time.Second)
						_ = openBrowser(url)
					}()
				}
			}

			// Stop spinner only when "Ready" appears
			if stopSpinner != nil && strings.Contains(line, "✓ Ready in") {
				stopSpinner <- true
			} else if spinnerControl != nil {
				// Resume spinner if not stopping
				spinnerControl <- "resume"
			}
		}
	}
}

func RunBlockingWithTerminalIO(command string, dir string, custom_envs map[string]any, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	env := os.Environ()
	for k, v := range custom_envs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
	}()

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Track if browser has been opened
	browserOpened := false

	// Filter output in separate goroutines (no spinner)
	go filterOutput(stdoutPipe, os.Stdout, nil, nil, nil, &browserOpened)
	go filterOutput(stderrPipe, os.Stderr, nil, nil, nil, nil)

	// Wait for command to finish
	err = cmd.Wait()

	signal.Stop(sigs)
	close(sigs)

	return err
}
