package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
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

	focusedButton = focusedStyle.Copy().Render(" Submit ")
	blurredButton = fmt.Sprintf(" %s ", noStyle.Render("Submit"))

	void = ""
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
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	switch m.searchingState {
	case searchInit:
		m.searchingState = searchInProgress
		m.err = nil
		go func(m model) {
			// time.Sleep(time.Second * 2)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			ch, err := m.db.GetChannel(ctx, m.paramInputs[0].Value())
			if err != nil {
				logData = fmt.Errorf("Error to get channel: %s", err)
				return
			}
			selectedChannel = ch

			after, _ := time.Parse("2006-01-02 15:04:05", m.paramInputs[1].Value())
			before, _ := time.Parse("2006-01-02 15:04:05", m.paramInputs[2].Value())

			args := db.GetChannelLogWithParamsParams{
				ChannelID: ch.ID,
				After:     after,
				Before:    before,
			}

			clogs, err := m.db.GetChannelLogWithParams(ctx, args)
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
				// log.Fatal(logData)
				m.err = logData.(error)
				logData = nil
				return m, cmd
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
					newItem := item{
						desc: fmt.Sprintf(
							"%s | %s",
							createdOn,
							cl.Description,
						),
						title: fmt.Sprintf(
							"#%v %v - ID:%v | %v | [%v] | %v",
							i+1,
							marker,
							cl.ID,
							cl.ResponseStatus.Int32,
							cl.Method.String,
							cl.Url.String,
						),
					}
					items = append(items, newItem)
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
		return m, cmd
	}
	return m, cmd
}

func newUUIDInput() textinput.Model {
	ci := textinput.NewModel()
	ci.Placeholder = "Channel UUID"
	ci.Focus()
	ci.CharLimit = 36
	ci.Width = 36
	ci.CursorStyle = cursorStyle
	ci.PromptStyle = focusedStyle
	ci.TextStyle = focusedStyle
	// ci.SetValue("b2abebe0-e8af-444c-9b8c-ccff91eb4ecc")
	ci.SetValue("cac4a1fe-0559-423e-97d6-f4a24f8d98cf")
	return ci
}

func newAfterDateInput(after time.Time) textinput.Model {
	ai := textinput.NewModel()
	ai.Placeholder = "After(yyyy-mm-dd)"
	ai.CursorStyle = cursorStyle
	ai.CharLimit = 19
	ai.Width = 19
	ai.SetValue(after.Format("2006-01-02 15:04:05"))
	return ai
}

func newBeforeDateInput(before time.Time) textinput.Model {
	bi := textinput.NewModel()
	bi.Placeholder = "Before(yyyy-mm-dd)"
	bi.CursorStyle = cursorStyle
	bi.CharLimit = 19
	bi.Width = 19
	bi.SetValue(before.Format("2006-01-02 15:04:05"))
	return bi
}

func buildSearchFormPanel(m model) string {

	button := &blurredButton
	if m.focusIndex == len(m.paramInputs) {
		button = &focusedButton
	}

	inputChannelUUID := components.InputStyle.Copy().Width(m.paramInputs[0].Width + 4).Render(m.paramInputs[0].View())
	inputDateAfter := components.InputStyle.Render(m.paramInputs[1].View())
	inputDateBefore := components.InputStyle.Render(m.paramInputs[2].View())

	switch m.state {
	case PromptParams:
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			inputChannelUUID,
			inputDateAfter,
			inputDateBefore,
			components.InputStyle.Render(*button),
		)
	case Searching:
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			inputChannelUUID,
			inputDateAfter,
			inputDateBefore,
			fmt.Sprintf("%s Searching", m.searchSpinner.View()),
		)
	case Listing:
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			inputChannelUUID,
			inputDateAfter,
			inputDateBefore,
		)
	}
	return ""
}

func promptView(m model) string {
	physicalWidth, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	var b strings.Builder
	searchForm = buildSearchFormPanel(m)
	b.WriteString(components.DocListStyle.Render(searchForm))
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(
			lipgloss.NewStyle().
				Width(physicalWidth).
				Align(lipgloss.Center).Render(m.err.Error()),
		)
	}
	b.WriteString(strings.Repeat("\n", physicalHeight-lipgloss.Height(b.String())-1))
	b.WriteString("  ")
	b.WriteString(cursorModeHelpStyle.Render("ctrl+c ") + helpStyle.Render("to quit • "))
	b.WriteString(cursorModeHelpStyle.Render("tab ") + helpStyle.Render("next input • "))
	b.WriteString(cursorModeHelpStyle.Render("shift+tab ") + helpStyle.Render("previous input • "))
	b.WriteString(cursorModeHelpStyle.Render("enter ") + helpStyle.Render("submit search"))
	return b.String()
}

func searchingView(m model) string {
	_, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	var b strings.Builder
	searchForm = buildSearchFormPanel(m)
	b.WriteString(components.DocListStyle.Render(searchForm))
	b.WriteString(strings.Repeat("\n", physicalHeight-lipgloss.Height(b.String())-1))
	b.WriteString(helpStyle.Render(" ctrl+c to quit"))
	return b.String()
}
