package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/ui/colors"
)

var (
	highlight   = lipgloss.AdaptiveColor{Light: colors.Dark, Dark: colors.Light}
	InputBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "└",
		BottomRight: "┘",
	}
	InputStyle = lipgloss.NewStyle().
			Border(InputBorder, true).
			BorderForeground(highlight)
	LogStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Margin(1, 3, 0, 0).
			Padding(1, 2)
)
