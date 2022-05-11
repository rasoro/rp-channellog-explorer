package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/internal/db"
	"github.com/rasoro/rp-channellog-explorer/ui/components"
	"golang.org/x/term"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#00DED2"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	void          = ""
)

func updateInputParams(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			// enter while submit buton was focused? loading and search
			if s == "enter" && m.focusIndex == len(m.paramInputs) {
				m.state = Searching
				m.searchingState = searchInit
				return m, m.searchSpinner.Tick
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else if s == "down" || s == "tab" {
				m.focusIndex++
			}

			if m.focusIndex > len(m.paramInputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.paramInputs)
			}

			cmds := make([]tea.Cmd, len(m.paramInputs))
			for i := 0; i <= len(m.paramInputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.paramInputs[i].Focus()
					m.paramInputs[i].PromptStyle = focusedStyle
					m.paramInputs[i].TextStyle = focusedStyle
					continue
				}
				// remove focused state
				m.paramInputs[i].Blur()
				m.paramInputs[i].PromptStyle = noStyle
				m.paramInputs[i].TextStyle = noStyle
			}
			return m, tea.Batch(cmds...)
		}
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.paramInputs))

	for i := range m.paramInputs {
		m.paramInputs[i], cmds[i] = m.paramInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func updateSearching(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.searchSpinner, cmd = m.searchSpinner.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	switch m.searchingState {
	case searchInit:
		m.searchingState = searchInProgress
		go func(m model) {
			time.Sleep(time.Second * 2)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			ch, err := m.db.GetChannel(ctx, m.paramInputs[0].Value())
			if err != nil {
				logData = fmt.Errorf("Error to get channel: %s", err)
				return
			}
			selectedChannel = ch
			clogs, err := m.db.GetChannelLogFromChannelID(ctx, ch.ID)
			if err != nil {
				logData = fmt.Errorf("Error to get channel logs: %s", err)
				return
			}
			logData = clogs
		}(m)
	case searchInProgress:
		if logData != nil {
			switch logData.(type) {
			case error:
				m.searchingState = searchErrored
				log.Fatal(logData)
			case []db.ChannelsChannellog:
				items := []list.Item{}
				for i, cl := range logData.([]db.ChannelsChannellog) {
					createdOn := cl.CreatedOn.Format("2006-01-02 15:04:05")
					var marker string
					if cl.IsError {
						marker = components.ErrorMark
					} else {
						marker = components.SuccessMark
					}
					items = append(items, item{
						desc: fmt.Sprintf(
							"%s | %s",
							createdOn,
							cl.Description,
						),
						title: fmt.Sprintf(
							"#%v %v - %v [%v] %v",
							i+1,
							marker,
							cl.ResponseStatus.Int32,
							cl.Method.String,
							cl.Url.String,
						),
					})
				}
				// m.logList = list.New(items, list.NewDefaultDelegate(), 0, 0)
				m.logList.SetItems(items)
				m.logList.Title = selectedChannel.Name.String
				// m.logList.Title = "Channel Logs"
				m.searchingState = searchSuccess
			}
		}
	case searchSuccess:
		m.state = Listing

		physicalWidth, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			log.Fatal(err)
		}
		h, v := components.DocListStyle.GetFrameSize()
		m.logList.SetSize(
			physicalWidth-h,
			physicalHeight-(v*2)-lipgloss.Height(searchForm))
		return m, textinput.Blink
	case searchErrored:
		// TODO: show error message
		m.state = PromptParams
		return m, nil
	}
	return m, cmd
}
