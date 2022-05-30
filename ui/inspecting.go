package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasoro/rp-channellog-explorer/ui/components"
	"golang.org/x/term"
)

const useHighPerformanceRender = false

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
			m.inspectReady = false
			return m, nil
		}
	case tea.WindowSizeMsg:
		if !m.inspectReady {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.MouseWheelDelta = 9
			m.viewport.YPosition = 0
			m.viewport.HighPerformanceRendering = useHighPerformanceRender
			m.viewport.SetContent(inspectLogView(m))
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
		m.viewport.MouseWheelDelta = 9
		m.viewport.YPosition = 0
		m.viewport.HighPerformanceRendering = useHighPerformanceRender
		m.viewport.SetContent(inspectLogView(m))
		m.inspectReady = true
		if useHighPerformanceRender {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func inspectLogView(m model) string {
	var b strings.Builder
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))

	requestBar := components.HorizontalStatusBar(
		width,
		"REQUEST",
		m.currentLog.Url.String,
		m.currentLog.Method.String,
		strings.TrimSpace(m.currentLog.Description),
	)
	b.WriteString(
		requestBar,
	)
	b.WriteString(
		lipgloss.NewStyle().Width(width).Render(
			lipgloss.JoinVertical(lipgloss.Top,
				components.DocListStyle.Render(
					m.logContent.Request.Headers,
				),
				components.DocListStyle.Render(
					m.logContent.Request.Data,
				),
			),
		),
	)
	responseBar := components.HorizontalStatusBar(
		width,
		"RESPONSE",
		m.currentLog.Url.String,
		m.currentLog.Method.String,
		fmt.Sprint(m.currentLog.ResponseStatus.Int32),
	)
	b.WriteString(responseBar)
	b.WriteString(
		lipgloss.NewStyle().Width(width).Render(
			lipgloss.JoinVertical(lipgloss.Top,
				components.DocListStyle.Render(
					m.logContent.Response.Headers,
				),
				components.DocListStyle.Render(
					m.logContent.Response.Data,
				),
			),
		),
	)

	return b.String()
}

func FormatRequestResponse(r string) (string, string) {
	if strings.Contains(r, "application/json") {
		breakLineIndex := strings.Index(r, "\n{")
		if breakLineIndex == -1 {
			return r, ""
		}
		topContent := r[:breakLineIndex]
		bottomContent := r[breakLineIndex:]
		bottomFormated, err := JsonPrettyPrint(bottomContent)
		if err != nil {
			return topContent, bottomContent
		}
		return topContent, bottomFormated
	} else {
		breakLineIndex := strings.Index(r, "\n\n")
		if breakLineIndex == -1 {
			return r, ""
		}
		topContent := r[:breakLineIndex]
		bottomContent := r[breakLineIndex:]
		bcfmt, err := QueryPrettyPrint(bottomContent)
		if err != nil {
			return topContent, bottomContent
		}
		return topContent, bcfmt
	}
}

func JsonPrettyPrint(in string) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "    ")
	return out.String(), err
}

func QueryPrettyPrint(in string) (string, error) {
	params, err := url.ParseQuery(in)
	if err != nil {
		return "", err
	}
	var out string
	for k, v := range params {
		out = out + fmt.Sprintf("%s = %s\n", k, v)
	}
	return out, nil
}

type LogContent struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Headers string `json:"headers"`
	Data    string `json:"data"`
}

type Response struct {
	Headers string `json:"headers"`
	Data    string `json:"data"`
}
