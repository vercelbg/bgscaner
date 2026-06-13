package outboundmenu

import "github.com/charmbracelet/lipgloss"

func (m *Model) View() string {
	return lipgloss.NewStyle().Padding(0, 5).Render(m.menu.View())
}
