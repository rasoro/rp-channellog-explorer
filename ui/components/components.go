package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/ui/colors"
)

var (
	highlight   = lipgloss.AdaptiveColor{Light: colors.Dark, Dark: colors.Light}
	lowlight    = lipgloss.AdaptiveColor{Light: colors.Light, Dark: colors.Dark}
	csuccess    = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	cerrors     = lipgloss.AdaptiveColor{Light: "#bf4343", Dark: "#f57373"}
	InputBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
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

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	StatusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	HBTitleStyle = lipgloss.NewStyle().
			Inherit(StatusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	HBValStyle  = lipgloss.NewStyle().Inherit(StatusBarStyle)
	HBInfoStyle = StatusNugget.Copy().
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	HBDetailStyle = StatusNugget.Copy().Background(lipgloss.Color("#6124DF"))
)

func HorizontalStatusBar(width int, key, textContent, info, detail string) string {
	w := lipgloss.Width

	statusKey := HBTitleStyle.Render(key)
	statusInfo := HBInfoStyle.Render(info)
	statusDetail := HBDetailStyle.Render(detail)

	statusContent := HBValStyle.Copy().
		Width(width - w(statusKey) - w(statusInfo) - w(statusDetail) - 4).
		Render(textContent)

	return DocListStyle.Render(
		StatusBarStyle.Width(width - 4).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				statusKey,
				statusContent,
				statusInfo,
				statusDetail,
			),
		),
	)
}
