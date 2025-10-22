package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

var (
	// Color definitions
	cyan    = color.New(color.FgCyan).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
)

// PrintBanner prints the GitDocAI banner
func PrintBanner() {
	myFigure := figure.NewColorFigure("GitDocAI", "slant", "cyan", true)
	myFigure.Print()
	fmt.Println()
}

// PrintSuccess prints a success message with a checkmark
func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", green("✓"), message)
}

// PrintError prints an error message with an X mark
func PrintError(message string) {
	fmt.Printf("%s %s\n", red("✗"), message)
}

// PrintWarning prints a warning message with a warning icon
func PrintWarning(message string) {
	fmt.Printf("%s %s\n", yellow("⚠"), message)
}

// PrintInfo prints an info message with an info icon
func PrintInfo(message string) {
	fmt.Printf("%s %s\n", blue("ℹ"), message)
}

// PrintStep prints a step message with a pointer icon
func PrintStep(message string) {
	fmt.Printf("%s %s\n", cyan("▸"), message)
}

// PrintHeader prints a header with decoration
func PrintHeader(title string) {
	line := strings.Repeat("─", len(title)+4)
	fmt.Printf("\n%s\n", cyan(line))
	fmt.Printf("%s %s %s\n", cyan("│"), bold(title), cyan("│"))
	fmt.Printf("%s\n\n", cyan(line))
}

// PrintSection prints a section title
func PrintSection(title string) {
	fmt.Printf("\n%s %s\n", magenta("●"), bold(title))
}

// PrintKeyValue prints a key-value pair
func PrintKeyValue(key, value string) {
	fmt.Printf("  %s: %s\n", cyan(key), value)
}

// PrintList prints a list item
func PrintList(item string) {
	fmt.Printf("  %s %s\n", green("•"), item)
}

// PrintDivider prints a decorative divider
func PrintDivider() {
	fmt.Println(cyan("─────────────────────────────────────────────────"))
}

// PrintCommand prints a command being executed
func PrintCommand(cmd string) {
	fmt.Printf("%s %s\n", yellow("$"), cmd)
}

// PrintURL prints a URL in a highlighted format
func PrintURL(label, url string) {
	fmt.Printf("  %s %s\n", cyan("→"), fmt.Sprintf("%s: %s", label, blue(url)))
}

// PrintLogo prints a small logo with version
func PrintLogo(version string) {
	fmt.Printf("%s %s\n", bold(cyan("GitDocAI")), yellow("v"+version))
}

// NewSpinner creates a new spinner with a custom message
func NewSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("cyan")
	return s
}

// RunWithSpinner runs a function with a spinner
func RunWithSpinner(message string, fn func() error) error {
	s := NewSpinner(message)
	s.Start()
	err := fn()
	s.Stop()
	fmt.Print("\r\033[K") // Clear the spinner line
	return err
}
