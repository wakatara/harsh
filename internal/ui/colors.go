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

// PrintMuted prints text in a muted/subdued color (less bright but still readable)
func (cm *ColorManager) PrintMuted(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		// Use 256-color mode for a medium gray (245 is a nice muted gray)
		color.C256(245).Print(text)
	}
}

// PrintfMuted prints formatted text in a muted/subdued color
func (cm *ColorManager) PrintfMuted(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.C256(245).Printf(format, args...)
	}
}

// PrintDim prints text in dimmed/gray color (darker than muted)
func (cm *ColorManager) PrintDim(text string) {
	if cm.disabled {
		fmt.Print(text)
	} else {
		// Use 256-color mode for a darker gray (240 is quite dim)
		color.C256(240).Print(text)
	}
}

// PrintfDim prints formatted text in dimmed/gray color
func (cm *ColorManager) PrintfDim(format string, args ...interface{}) {
	if cm.disabled {
		fmt.Printf(format, args...)
	} else {
		color.C256(240).Printf(format, args...)
	}
}

// RenderDim returns text with dim color ANSI codes embedded
func (cm *ColorManager) RenderDim(text string) string {
	if cm.disabled {
		return text
	}
	return color.C256(240).Sprint(text)
}
