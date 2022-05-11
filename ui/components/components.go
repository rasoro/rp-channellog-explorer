package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/ui/colors"
)

var (
	highlight   = lipgloss.AdaptiveColor{Light: colors.Dark, Dark: colors.Light}
	csuccess    = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	cerrors     = lipgloss.AdaptiveColor{Light: "#bf4343", Dark: "#f57373"}
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
	SuccessMark = lipgloss.NewStyle().SetString("✓").
			Foreground(csuccess).
			PaddingRight(1).
			String()
	ErrorMark = lipgloss.NewStyle().SetString("✗").
			Foreground(cerrors).
			PaddingRight(1).
			String()
	DocListStyle = lipgloss.NewStyle().Margin(1, 2)
)
