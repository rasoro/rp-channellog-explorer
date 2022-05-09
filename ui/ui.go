package ui

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/ui/colors"
	"github.com/rasoro/rp-channellog-explorer/ui/components"
	"golang.org/x/term"
)

func NewProgram() *tea.Program {
	return tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)
}

type State int
type searchingState int

const (
	PromptParams State = iota
	Searching
	NotFound
	Listing
	Errored
)

const (
	searchInit searchingState = iota
	searchInProgress
	searchSuccess
	searchErrored
)

type errMsg error

type model struct {
	focusIndex       int
	cursorMode       textinput.CursorMode
	uuidChannelInput textinput.Model
	afterInput       textinput.Model
	beforeInput      textinput.Model
	paramInputs      []textinput.Model
	state            State
	searchingState   searchingState
	searchSpinner    spinner.Model
	err              error
}

var logData *string

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

func initialModel() model {
	paramInputs := make([]textinput.Model, 0)

	ci := textinput.NewModel()
	ci.Placeholder = "Channel UUID"
	ci.Focus()
	ci.CharLimit = 32
	ci.Width = 32
	ci.CursorStyle = cursorStyle
	ci.PromptStyle = focusedStyle
	ci.TextStyle = focusedStyle
	paramInputs = append(paramInputs, ci)

	ai := textinput.NewModel()
	ai.Placeholder = "After(yyyy-mm-dd)"
	ai.CursorStyle = cursorStyle
	ai.CharLimit = 18
	ai.Width = 18
	paramInputs = append(paramInputs, ai)

	bi := textinput.NewModel()
	bi.Placeholder = "Before(yyyy-mm-dd)"
	bi.CursorStyle = cursorStyle
	bi.CharLimit = 18
	bi.Width = 18
	paramInputs = append(paramInputs, bi)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Primary))

	return model{
		uuidChannelInput: ci,
		afterInput:       ai,
		beforeInput:      bi,
		paramInputs:      paramInputs,
		state:            PromptParams,
		searchSpinner:    s,
	}
}

func (m model) Init() tea.Cmd {
	// return m.searchSpinner.Tick
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case PromptParams:
		return updateInputParams(msg, m)
	case Searching:
		return updateSearching(msg, m)
	case Listing:
		return updateListing(msg, m)
	default:
	}
	return m, nil
}

func updateListing(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
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
		go func() {
			time.Sleep(time.Second * 2)
			result := "{}"
			logData = &result
		}()
	case searchInProgress:
		if logData != nil {
			m.searchingState = searchSuccess
		}
	case searchSuccess:
		m.state = Listing
		return m, textinput.Blink
	case searchErrored:
		return m, nil
	}
	return m, cmd
}

func updateInputParams(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+p":
			return m, nil
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// enter while submit buton was focused? loading and search
			if s == "enter" && m.focusIndex == len(m.paramInputs) {
				// TODO: change status to loading and search channel logs
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

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	var b strings.Builder

	button := &blurredButton
	if m.focusIndex == len(m.paramInputs) {
		button = &focusedButton
	}

	searchForm := ""

	if m.state != PromptParams && m.state != Searching {
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Top,
			components.InputStyle.Copy().Width(35).Render(m.paramInputs[0].View()),
			components.InputStyle.Render(m.paramInputs[1].View()),
			components.InputStyle.Render(m.paramInputs[2].View()),
		)
	} else {
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Top,
			components.InputStyle.Copy().Width(35).Render(m.paramInputs[0].View()),
			components.InputStyle.Render(m.paramInputs[1].View()),
			components.InputStyle.Render(m.paramInputs[2].View()),
			components.InputStyle.Render(*button),
		)
	}

	if m.state == Searching {
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Center,
			components.InputStyle.Copy().Width(35).Render(m.paramInputs[0].View()),
			components.InputStyle.Render(m.paramInputs[1].View()),
			components.InputStyle.Render(m.paramInputs[2].View()),
			fmt.Sprintf("%s Searching", m.searchSpinner.View()),
		)
	}

	b.WriteString(searchForm)

	_, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	// b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(strings.Repeat("\n", physicalHeight-lipgloss.Height(b.String())-1))
	b.WriteString(helpStyle.Render(" ctrl+c to quit"))
	return b.String()
}
