package ui

import (
	"fmt"

	"github.com/gookit/color"
)

// ColorManager handles color output configuration
type ColorManager struct {
	disabled bool
}

// NewColorManager creates a new color manager
func NewColorManager(noColor bool) *ColorManager {
	cm := &ColorManager{disabled: noColor}
	if noColor {
		color.Disable()
	}
	return cm
}

// SetNoColor enables or disables color output
func (cm *ColorManager) SetNoColor(disabled bool) {
	cm.disabled = disabled
	if disabled {
		color.Disable()
	} else {
		color.Enable = true
	}
}

// IsDisabled returns whether color output is disabled
func (cm *ColorManager) IsDisabled() bool {
	return cm.disabled
}

// PrintBold prints text in bold
func (cm *ColorManager) PrintBold(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		color.Bold.Print(text)
	}
}

// PrintlnBold prints text in bold with newline
func (cm *ColorManager) PrintlnBold(text string) {
	if cm.disabled {
		fmt.Println(text)
	} else {
		color.Bold.Println(text)
	}
}

// PrintfBold prints formatted text in bold
func (cm *ColorManager) PrintfBold(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.Bold.Printf(format, args...)
	}
}

// PrintGreen prints text in green
func (cm *ColorManager) PrintGreen(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		color.FgGreen.Print(text)
	}
}

// PrintfGreen prints formatted text in green
func (cm *ColorManager) PrintfGreen(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.FgGreen.Printf(format, args...)
	}
}

// PrintRed prints text in red
func (cm *ColorManager) PrintRed(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		color.FgRed.Print(text)
	}
}

// PrintfRed prints formatted text in red
func (cm *ColorManager) PrintfRed(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.FgRed.Printf(format, args...)
	}
}

// PrintYellow prints text in yellow
func (cm *ColorManager) PrintYellow(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		color.FgYellow.Print(text)
	}
}

// PrintfYellow prints formatted text in yellow
func (cm *ColorManager) PrintfYellow(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.FgYellow.Printf(format, args...)
	}
}

// PrintBlue prints text in blue
func (cm *ColorManager) PrintBlue(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		color.FgBlue.Print(text)
	}
}

// PrintfBlue prints formatted text in blue
func (cm *ColorManager) PrintfBlue(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.FgBlue.Printf(format, args...)
	}
}
