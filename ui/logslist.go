package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/internal/db"
	"github.com/rasoro/rp-channellog-explorer/ui/components"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func updateListing(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlP:
			m.state = PromptParams
			return m, textinput.Blink
		case tea.KeyEnter:
			m.state = Inspecting
			content, ok := logData.([]db.ChannelsChannellog)
			if !ok {
				log.Fatal(ok)
				return m, tea.Quit
			}
			m.viewport.SetContent(content[m.logList.Index()].Response.String)
			return m, textinput.Blink
		}
	case tea.WindowSizeMsg:
		h, v := components.DocListStyle.GetFrameSize()
		m.logList.SetSize(
			msg.Width-h,
			msg.Height-(v*2)-lipgloss.Height(searchForm))
	}

	m.logList, cmd = m.logList.Update(msg)
	return m, cmd
}
