// Package ui provides consistent, colorized output for gmh.
package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	// Cyan is for headers and info.
	Cyan = color.New(color.FgCyan)
	// Bold is for emphasis.
	Bold = color.New(color.Bold)
	// Green is for success.
	Green = color.New(color.FgGreen)
	// Yellow is for warnings.
	Yellow = color.New(color.FgYellow)
	// Red is for errors.
	Red = color.New(color.FgRed)
	// Gray is for muted text.
	Gray = color.New(color.FgHiBlack)
	// NoColor disables colors if true.
	NoColor = false
)

func init() {
	if os.Getenv("GMH_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
		NoColor = true
	}
}

// Info prints an info message.
func Info(format string, args ...interface{}) {
	if NoColor {
		fmt.Printf("==> "+format+"\n", args...)
		return
	}
	Cyan.Printf("==> ")
	fmt.Printf(format+"\n", args...)
}

// OK prints a success message.
func OK(format string, args ...interface{}) {
	if NoColor {
		fmt.Printf("✅ "+format+"\n", args...)
		return
	}
	Green.Print("✅ ")
	fmt.Printf(format+"\n", args...)
}

// Warn prints a warning message.
func Warn(format string, args ...interface{}) {
	if NoColor {
		fmt.Printf("⚠️  "+format+"\n", args...)
		return
	}
	Yellow.Print("⚠️  ")
	fmt.Printf(format+"\n", args...)
}

// Fail prints a failure message.
func Fail(format string, args ...interface{}) {
	if NoColor {
		fmt.Printf("❌ "+format+"\n", args...)
		return
	}
	Red.Print("❌ ")
	fmt.Printf(format+"\n", args...)
}

// Step prints a sub-step under a header.
func Step(format string, args ...interface{}) {
	if NoColor {
		fmt.Printf("  • "+format+"\n", args...)
		return
	}
	Gray.Print("  • ")
	fmt.Printf(format+"\n", args...)
}

// Header prints a section header.
func Header(format string, args ...interface{}) {
	if NoColor {
		fmt.Printf("\n"+format+"\n", args...)
		fmt.Println(strings.Repeat("─", 60))
		return
	}
	fmt.Println()
	Bold.Printf(format+"\n", args...)
	Gray.Println(strings.Repeat("─", 60))
}
