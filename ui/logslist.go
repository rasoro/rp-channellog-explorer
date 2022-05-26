package ui

import (
	"fmt"
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
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+p":
			m.state = PromptParams
			return m, textinput.Blink
		case "enter":
			m.state = Inspecting
			content, ok := logData.([]db.ChannelsChannellog)
			if !ok {
				log.Fatal(ok)
				return m, tea.Quit
			}
			request := content[m.logList.Index()].Request.String
			response := content[m.logList.Index()].Response.String
			requestFmtd, err := FormatRequestResponse4(request)
			if err != nil {
				requestFmtd = request
			}
			responseFmtd, err := FormatRequestResponse4(response)
			if err != nil {
				responseFmtd = response
			}
			ctt := fmt.Sprintf("%s\n\n%s", requestFmtd, responseFmtd)
			m.inspectContent = ctt
			m.viewport.SetContent(m.inspectContent)
			return m, tea.Batch(textinput.Blink /*, viewport.Sync(m.viewport)*/)
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
