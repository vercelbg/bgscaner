package table

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Aliases for upstream table types
type (
	Column = table.Column
	Row    = table.Row
)

// Model wraps a Bubble Tea table with responsive layout, key bindings, and concurrency safety.
type Model struct {
	mu sync.RWMutex

	id     ui.ComponentID
	name   string
	Title  string
	Layout *layout.Layout

	Help     help.Model
	FullHelp bool

	BubbleTable table.Model
	Keys        KeyMap

	colsWidth []int
	paddingY  int
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component ID.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose performs cleanup when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// New creates a new table model.
func New(title string, cols []table.Column, rows []table.Row, lay *layout.Layout) *Model {
	m := &Model{
		id:       ui.NewComponentID(),
		name:     "table",
		Title:    title,
		Layout:   lay,
		Help:     help.New(),
		Keys:     defaultKeys(),
		paddingY: 0,
	}
	m.BubbleTable = m.createTable(rows, cols)

	m.mu.Lock()
	m.updateTableSizeLocked()
	m.mu.Unlock()

	return m
}

// SetPaddingY sets vertical padding and updates table size.
func (m *Model) SetPaddingY(padding int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paddingY = padding
	m.updateTableSizeLocked()
}

// ActionKey defines a key binding and its action.
type ActionKey struct {
	Keys      []string
	ShortHelp string
	FullHelp  string
	Cmd       tea.Cmd
}

// NewKey creates a new action key definition.
func NewKey(keys []string, shortHelp, fullHelp string, cmd tea.Cmd) ActionKey {
	ks := make([]string, len(keys))
	for i, k := range keys {
		switch k {
		case "up", "↑":
			ks[i] = "↑"
		case "down", "↓":
			ks[i] = "↓"
		case "left", "←":
			ks[i] = "←"
		case "right", "→":
			ks[i] = "→"
		default:
			ks[i] = k
		}
	}

	kstr := strings.Join(ks, "/")
	if kstr == "" {
		kstr = "?"
	}
	if shortHelp != "" {
		shortHelp = fmt.Sprintf("%s %s", kstr, shortHelp)
	}
	if fullHelp != "" {
		fullHelp = fmt.Sprintf("%s %s", kstr, fullHelp)
	}

	return ActionKey{
		Keys:      keys,
		ShortHelp: shortHelp,
		FullHelp:  fullHelp,
		Cmd:       cmd,
	}
}

// SetKeys replaces the current key bindings.
func (m *Model) SetKeys(keys ...ActionKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Keys = defaultKeys(keys...)
}

// KeyMap stores registered key bindings.
type KeyMap struct {
	Actions []ActionKey
}

// Add appends a new key binding.
func (k *KeyMap) Add(a ActionKey) {
	k.Actions = append(k.Actions, a)
}

// Check returns the command associated with a key message.
func (k KeyMap) Check(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	keyStr := keyMsg.String()
	for _, a := range k.Actions {
		if slices.Contains(a.Keys, keyStr) {
			return a.Cmd
		}
	}
	return nil
}

// ShortHelp returns key bindings for the condensed help view.
func (k KeyMap) ShortHelp() []key.Binding {
	var bindings []key.Binding
	for _, a := range k.Actions {
		if a.ShortHelp == "" {
			continue
		}
		bindings = append(bindings, key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp(a.ShortHelp, ""),
		))
	}
	return bindings
}

// FullHelp returns key bindings for the expanded help view in columns.
func (k KeyMap) FullHelp() [][]key.Binding {
	if len(k.Actions) == 0 {
		return nil
	}
	colCount := 4
	cols := make([][]key.Binding, colCount)
	for i, a := range k.Actions {
		if a.FullHelp == "" {
			continue
		}
		binding := key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp("", a.FullHelp),
		)
		col := i % colCount
		cols[col] = append(cols[col], binding)
	}
	return cols
}

func defaultKeys(keys ...ActionKey) KeyMap {
	const spacebar = " "
	km := KeyMap{}
	km.Add(NewKey([]string{"up", "k"}, "up", "Move up", nil))
	km.Add(NewKey([]string{"down", "j"}, "down", "Move down", nil))
	km.Add(NewKey([]string{"b", "pgup"}, "", "Page up", nil))
	km.Add(NewKey([]string{"f", "pgdown", spacebar}, "", "Page down", nil))
	km.Add(NewKey([]string{"u", "ctrl+u"}, "", "½ page up", nil))
	km.Add(NewKey([]string{"d", "ctrl+d"}, "", "½ page down", nil))
	km.Add(NewKey([]string{"home", "g"}, "", "Go to start", nil))
	km.Add(NewKey([]string{"end", "G"}, "", "Go to end", nil))
	for _, keyConfig := range keys {
		km.Add(keyConfig)
	}
	km.Add(NewKey([]string{"?"}, "help", "Toggle help", nil))
	km.Add(NewKey([]string{"q", "esc"}, "quit", "Quit", nil))
	return km
}

// NewRowTime formats a timestamp for display.
func NewRowTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

// NewRowBool returns "yes" or "no" for a boolean value.
func NewRowBool(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

// NewTimeDurationRow returns a human-readable duration from a past time.
func NewTimeDurationRow(from time.Time) string {
	if from.IsZero() {
		return "-"
	}
	d := time.Since(from)
	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return d.Truncate(time.Second).String()
	case d < time.Hour:
		return d.Truncate(time.Minute).String()
	default:
		return d.Truncate(time.Hour).String()
	}
}

// AppendRow adds a new row to the table safely.
func (m *Model) AppendRow(row table.Row) {
	m.mu.Lock()
	defer m.mu.Unlock()

	newRow := make(table.Row, len(row))
	copy(newRow, row)

	rows := append([]table.Row(nil), m.BubbleTable.Rows()...)
	rows = append(rows, newRow)
	m.BubbleTable.SetRows(rows)
}

// Mode implements ui.Component.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// Private helpers

func (m *Model) createTable(rows []table.Row, cols []table.Column) table.Model {
	m.mu.Lock()
	width := m.tableWidthLocked()
	m.mu.Unlock()

	total := 0
	for _, col := range cols {
		total += col.Width
	}
	if total > 0 {
		ratio := float64(width) / float64(total)
		m.mu.Lock()
		m.colsWidth = make([]int, len(cols))
		for i := range cols {
			m.colsWidth[i] = cols[i].Width
			cols[i].Width = int(float64(cols[i].Width) * ratio)
		}
		m.mu.Unlock()
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(max(1, len(rows))),
	)
	t.SetStyles(tableStyles())
	return t
}

func (m *Model) updateTableSizeLocked() {
	if m.Layout == nil || m.Layout.Body.Height == 0 || m.Layout.Body.Width == 0 {
		return
	}

	width := m.tableWidthLocked()
	helpHeight := lipgloss.Height(m.renderHelpView())
	titleHeight := lipgloss.Height(m.renderTitle())

	height := max(1, m.Layout.Body.Height-helpHeight-titleHeight-m.paddingY)
	cols := m.BubbleTable.Columns()
	if len(cols) == 0 {
		return
	}

	total := 0
	for _, w := range m.colsWidth {
		total += w
	}
	if total <= 0 {
		return
	}

	ratio := float64(width) / float64(total)
	for i := range cols {
		cols[i].Width = int(ratio * float64(m.colsWidth[i]))
	}

	m.BubbleTable.SetColumns(cols)
	m.BubbleTable.SetHeight(height)
}

func (m *Model) updateTableSize() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateTableSizeLocked()
}

func (m *Model) tableWidthLocked() int {
	if m.Layout == nil || m.Layout.Body.Width == 0 {
		return 80
	}
	return min(80, m.Layout.Body.Width-10)
}

