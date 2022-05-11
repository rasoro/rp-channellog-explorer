package ui

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/internal/db"
	"github.com/rasoro/rp-channellog-explorer/ui/colors"
	"github.com/rasoro/rp-channellog-explorer/ui/components"
	"golang.org/x/term"
)

func NewProgram(dbq *db.Queries) *tea.Program {
	return tea.NewProgram(
		initialModel(dbq),
		tea.WithAltScreen(),
	)
}

type GlobalState int
type SearchingState int

const (
	PromptParams GlobalState = iota
	Searching
	NotFound
	Listing
	Errored
	Inspecting
)

const (
	searchInit SearchingState = iota
	searchInProgress
	searchSuccess
	searchErrored
)

type model struct {
	focusIndex     int
	cursorMode     textinput.CursorMode
	paramInputs    []textinput.Model
	state          GlobalState
	searchingState SearchingState
	searchSpinner  spinner.Model
	logList        list.Model
	inspectContent string
	inspectReady   bool
	viewport       viewport.Model
	db             *db.Queries
	err            error
}

var logData interface{}
var selectedChannel db.ChannelsChannel

var (
	searchForm = ""
	listPanel  = ""
)

const useHighPerformanceRender = false

func initialModel(db *db.Queries) model {
	paramInputs := make([]textinput.Model, 0)

	nowString := time.Now().Format("2006-01-02") + " 00:00:00"
	after, _ := time.Parse("2006-01-02 00:00:00", nowString)
	before := after.AddDate(0, 0, 1).Add(time.Nanosecond * -1)

	ci := textinput.NewModel()
	ci.Placeholder = "Channel UUID"
	ci.Focus()
	ci.CharLimit = 36
	ci.Width = 36
	ci.CursorStyle = cursorStyle
	ci.PromptStyle = focusedStyle
	ci.TextStyle = focusedStyle
	// ci.SetValue("cac4a1fe-0559-423e-97d6-f4a24f8d98cf")
	paramInputs = append(paramInputs, ci)

	ai := textinput.NewModel()
	ai.Placeholder = "After(yyyy-mm-dd)"
	ai.CursorStyle = cursorStyle
	ai.CharLimit = 19
	ai.Width = 19
	ai.SetValue(after.Format("2006-01-02 15:04:05"))
	paramInputs = append(paramInputs, ai)

	bi := textinput.NewModel()
	bi.Placeholder = "Before(yyyy-mm-dd)"
	bi.CursorStyle = cursorStyle
	bi.CharLimit = 19
	bi.Width = 19
	bi.SetValue(before.Format("2006-01-02 15:04:05"))
	paramInputs = append(paramInputs, bi)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Primary))

	logList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	logList.Title = "Channel Logs"
	logList.SetShowHelp(false)

	return model{
		paramInputs:   paramInputs,
		state:         PromptParams,
		searchSpinner: s,
		db:            db,
		logList:       logList,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case PromptParams:
		return updateInputParams(msg, m)
	case Searching:
		return updateSearching(msg, m)
	case Inspecting:
		return updateInspecting(msg, m)
	case Listing:
		return updateListing(msg, m)
	default:
	}
	return m, nil
}

func updateInspecting(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.state = Listing
			return m, nil
		}
	case tea.WindowSizeMsg:
		if !m.inspectReady {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.YPosition = 0
			m.viewport.HighPerformanceRendering = useHighPerformanceRender
			m.viewport.SetContent(m.inspectContent)
			m.inspectReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}

		if useHighPerformanceRender {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	if !m.inspectReady {
		physicalWidth, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			log.Fatal(err)
		}
		m.viewport = viewport.New(physicalWidth, physicalHeight)
		m.viewport.YPosition = 0
		m.viewport.HighPerformanceRendering = useHighPerformanceRender
		content, ok := logData.([]db.ChannelsChannellog)
		if !ok {
			log.Fatal(ok)
			return m, tea.Quit
		}
		m.viewport.SetContent(content[m.logList.Index()].Response.String)
		m.inspectReady = true
		if useHighPerformanceRender {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	_, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	var b strings.Builder

	button := &blurredButton
	if m.focusIndex == len(m.paramInputs) {
		button = &focusedButton
	}

	inputChannelUUID := components.InputStyle.Copy().Width(m.paramInputs[0].Width + 4).Render(m.paramInputs[0].View())
	inputDateAfter := components.InputStyle.Render(m.paramInputs[1].View())
	inputDateBefore := components.InputStyle.Render(m.paramInputs[2].View())

	switch m.state {
	case PromptParams:
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Center,
			inputChannelUUID,
			inputDateAfter,
			inputDateBefore,
			components.InputStyle.Render(*button),
		)
	case Searching:
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Center,
			inputChannelUUID,
			inputDateAfter,
			inputDateBefore,
			fmt.Sprintf("%s Searching", m.searchSpinner.View()),
		)
	case Listing:
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Center,
			inputChannelUUID,
			inputDateAfter,
			inputDateBefore,
		)
	case Inspecting:
		searchForm = ""
	}

	if m.state == Listing {
		b.WriteString(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.JoinVertical(
					lipgloss.Top,
					components.DocListStyle.Render(searchForm),
					components.DocListStyle.Render(m.logList.View()),
				),
			),
		)
		return b.String()
	} else if m.state == Inspecting {
		if !m.inspectReady {
			b.WriteString(
				"\n  Initializing...",
			)
		} else {
			b.WriteString(
				m.viewport.View(),
			)
		}
	} else {
		b.WriteString(components.DocListStyle.Render(searchForm))
		b.WriteString(strings.Repeat("\n", physicalHeight-lipgloss.Height(b.String())-1))
		b.WriteString(helpStyle.Render(" ctrl+c to quit"))
	}
	return b.String()
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return string([]rune(str)[:length])
}
