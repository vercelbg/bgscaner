package outboundmenu

import (
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ImportMethod defines how the user wants to add the outbound configurations.
type ImportMethod string

const (
	MethodLink ImportMethod = "link"
	MethodJSON ImportMethod = "json"
)

// MsgSelectImportMethod is fired when a selection is made, carrying the chosen method argument.
type MsgSelectImportMethod struct {
	Method ImportMethod
}

// Model represents the outbound integration selection menu component.
type Model struct {
	id     ui.ComponentID
	name   string
	menu   ui.Component
	Layout *layout.Layout
}

// New instantiates an initialized outbound menu component layout.
func New(layout *layout.Layout) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "Outbound Menu",
		Layout: layout,
	}
	m.menu = newMenu(layout)
	return m
}

// ID returns the component's unique structural registration identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the descriptive name of the menu component.
func (m *Model) Name() string {
	return m.name
}

// OnClose executes lifecycle cleanup operations when the component is popped from the view stack.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode establishes active keyboard mapping and state profile rules for the environment layout.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// Init initializes the component scope upon activation.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ── Private Helpers ──────────────────────────────────────────────────────────

// newMenu builds the basic structural component menu options along with their argument messages.
func newMenu(layout *layout.Layout) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▶",
			"Via Link",
			"l",
			func() tea.Msg {
				return MsgSelectImportMethod{Method: MethodLink}
			},
		),
		menu.NewMenuItem(
			"☰",
			"Select Json",
			"j",
			func() tea.Msg {
				return MsgSelectImportMethod{Method: MethodJSON}
			},
		),
	}

	return menu.New(items, "Addition Method", layout)
}

