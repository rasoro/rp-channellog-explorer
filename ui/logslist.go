package ui

import (
	"log"
	"strings"

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
			currentContent := content[m.logList.Index()]

			m.currentLog = &currentContent

			request := currentContent.Request.String
			response := currentContent.Response.String

			reqH, reqD := FormatRequestResponse(request)
			reqHMap := trimMultilineString(reqH)
			m.logContent.Request = Request{reqHMap, reqD}

			resH, resD := FormatRequestResponse(response)
			resHMap := trimMultilineString(resH)
			m.logContent.Response = Response{resHMap, resD}

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

func trimMultilineString(in string) string {
	inList := strings.Split(in, "\n")
	var str string
	for _, st := range inList {
		str += strings.TrimSpace(st) + "\n"
	}
	return str
}
