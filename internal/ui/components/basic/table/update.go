package table

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.Keys.Check(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

		if msg.String() == "?" {
			m.FullHelp = !m.FullHelp
			m.updateTableSize()
		}

	case tea.WindowSizeMsg:
		m.updateTableSize()
		m.BubbleTable.SetStyles(tableStyles())
		return m, nil
	}

	var tableCmd tea.Cmd
	m.BubbleTable, tableCmd = m.BubbleTable.Update(msg)
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}

	return m, tea.Batch(cmds...)
}
