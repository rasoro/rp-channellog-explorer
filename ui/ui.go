package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	focusIndex     int
	cursorMode     textinput.CursorMode
	paramInputs    []textinput.Model
	state          State
	searchingState searchingState
	searchSpinner  spinner.Model
	logList        list.Model
	db             *db.Queries
	err            error
}

var logData interface{}
var selectedChannel db.ChannelsChannel

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

var (
	docListStyle = lipgloss.NewStyle().Margin(1, 2)
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var (
	searchForm = ""
	listPanel  = ""
)

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
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlP:
			m.state = PromptParams
			return m, textinput.Blink
		}
	case tea.WindowSizeMsg:
		h, v := docListStyle.GetFrameSize()
		m.logList.SetSize(
			msg.Width-h,
			msg.Height-(v*2)-lipgloss.Height(searchForm))
	}

	m.logList, cmd = m.logList.Update(msg)
	return m, cmd
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
				for _, cl := range logData.([]db.ChannelsChannellog) {
					createdOn := cl.CreatedOn.Format("2006-01-02 15:04:05")
					items = append(items, item{
						desc: fmt.Sprintf(
							"%s | %s",
							createdOn,
							cl.Description,
						),
						title: fmt.Sprintf(
							"[%v] %v",
							cl.Method.String,
							// TruncateString(cl.Url.String, 40),
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
		h, v := docListStyle.GetFrameSize()
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

	if m.state != PromptParams && m.state != Searching {
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Top,
			components.InputStyle.Copy().Width(m.paramInputs[0].Width+4).Render(m.paramInputs[0].View()),
			components.InputStyle.Render(m.paramInputs[1].View()),
			components.InputStyle.Render(m.paramInputs[2].View()),
		)
	} else {
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Top,
			components.InputStyle.Copy().Width(m.paramInputs[0].Width+4).Render(m.paramInputs[0].View()),
			components.InputStyle.Render(m.paramInputs[1].View()),
			components.InputStyle.Render(m.paramInputs[2].View()),
			components.InputStyle.Render(*button),
		)
	}

	if m.state == Searching {
		searchForm = lipgloss.JoinHorizontal(
			lipgloss.Center,
			components.InputStyle.Copy().Width(m.paramInputs[0].Width+4).Render(m.paramInputs[0].View()),
			components.InputStyle.Render(m.paramInputs[1].View()),
			components.InputStyle.Render(m.paramInputs[2].View()),
			fmt.Sprintf("%s Searching", m.searchSpinner.View()),
		)
	}

	_, physicalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	if m.state == Listing {
		b.WriteString(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.JoinVertical(
					lipgloss.Top,
					docListStyle.Render(searchForm),
					docListStyle.Render(m.logList.View()),
				),
				// TODO: textpanel
			),
		)
		return b.String()
	} else {
		b.WriteString(docListStyle.Render(searchForm))
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
