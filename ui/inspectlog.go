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
		m.viewport.SetContent(currentContent)
		m.inspectReady = true
		if useHighPerformanceRender {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func FormatRequestResponse(r string) (string, error) {
	breakLineIndex := strings.Index(r, "\n\n")
	if breakLineIndex == -1 {
		return r, nil
	}
	topContent := r[:breakLineIndex]
	bottomConntent := r[breakLineIndex:]
	bcfmt, err := JsonPrettyPrint(bottomConntent)
	if err != nil {
		bcfmt, err := QueryPrettyPrint(bottomConntent)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s\n\n%s", topContent, bcfmt), nil
	}
	return fmt.Sprintf("%s\n\n%s", topContent, bcfmt), nil
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
