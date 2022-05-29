package ui

import (
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
	logContent     LogContent
	currentLog     *db.ChannelsChannellog
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

func initialModel(db *db.Queries) model {
	paramInputs := make([]textinput.Model, 0)

	after := time.Date(2022, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	before := time.Now()

	ci := newUUIDInput()
	paramInputs = append(paramInputs, ci)

	ai := newAfterDateInput(after)
	paramInputs = append(paramInputs, ai)

	bi := newBeforeDateInput(before)
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

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	switch m.state {
	case PromptParams:
		return promptView(m)
	case Searching:
		return searchingView(m)
	case Listing:
		return listingView(m)
	case Inspecting:
		return inspectLogView(m)
	}

	return ""
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
