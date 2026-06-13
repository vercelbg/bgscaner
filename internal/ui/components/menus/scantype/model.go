package scantype

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"bgscan/internal/core/config"
	"bgscan/internal/core/scanner"
	"bgscan/internal/core/xray"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/basic/notice"
	scannerUi "bgscan/internal/ui/components/scanner"
	"bgscan/internal/ui/components/tables/outbounds"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
)

type Model struct {
	id           ui.ComponentID
	name         string
	layout       *layout.Layout
	input        string
	xrayTemplate string
	menu         ui.Component
	closeScanner bool
	scanner      *scanner.Scanner
}

func New(layout *layout.Layout, input string) *Model {
	m := &Model{
		id:           ui.NewComponentID(),
		name:         "Scan Menu",
		layout:       layout,
		input:        input,
		closeScanner: true,
	}

	m.menu = menu.New([]menu.MenuItem{
		menu.NewMenuItem("▦", "ICMP Scan", "i", m.open(scanner.ICMP_SCAN)),
		menu.NewMenuItem("≡", "TCP Scan", "t", m.open(scanner.TCP_SCAN)),
		menu.NewMenuItem("▦", "HTTP Scan", "h", m.open(scanner.HTTP_SCAN)),
		menu.NewMenuItem("#", "DNS Scan", "d", m.open(scanner.RESOLVE_SCAN)),
		menu.NewMenuItem("▦", "Xray Scan", "x", m.openXrayTemplates()),
	}, "Select Scan Type", layout)

	return m
}

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) Mode() env.Mode     { return env.NormalMode }

func (m *Model) OnClose() tea.Cmd {
	if m.closeScanner && m.scanner != nil {
		m.scanner.Close()
	}
	return nil
}

// ------------------------------------------------------------
// Scanner entry
// ------------------------------------------------------------

func (m *Model) open(mode scanner.ScanMode) tea.Cmd {
	return func() tea.Msg {
		scn, err := m.createScanner(mode, m.input)
		if err != nil {
			return m.errorCmd("scanner error", err.Error())
		}

		m.closeScanner = false
		return ui.OpenComponentCmd(scannerUi.New(m.layout, 10_000, scn))()
	}
}

func (m *Model) openXrayTemplates() tea.Cmd {
	return ui.OpenComponentCmd(
		outbounds.New(m.layout, "select outbound", func(xof *xray.XrayOutboundsFile) tea.Cmd {
			m.xrayTemplate = xof.Name
			return m.open(scanner.XRAY_SCAN)
		}),
	)
}

// ------------------------------------------------------------
// Scanner builder
// ------------------------------------------------------------

func (m *Model) createScanner(mode scanner.ScanMode, input string) (*scanner.Scanner, error) {
	ctx := context.Background()
	scn := scanner.NewScanner(ctx, input)

	if mode == scanner.XRAY_SCAN {
		return m.buildXrayScanner(ctx, scn)
	}

	if mode == scanner.RESOLVE_SCAN {
		return m.buildResolveScanner(ctx, scn)
	}

	stage, err := m.buildStage(ctx, scn, mode)
	if err != nil {
		return nil, err
	}

	scn.AddStage(stage)
	return scn, nil
}

func (m *Model) buildStage(ctx context.Context, scn *scanner.Scanner, mode scanner.ScanMode) (scanner.StageConfig, error) {
	switch mode {
	case scanner.TCP_SCAN:
		return scn.BuildTCPStage(ctx)
	case scanner.ICMP_SCAN:
		return scn.BuildICMPStage(ctx)
	case scanner.HTTP_SCAN:
		return scn.BuildHTTPStage(ctx)
	default:
		return scn.BuildTCPStage(ctx)
	}
}

// ------------------------------------------------------------
// Special scanners
// ------------------------------------------------------------

func (m *Model) buildResolveScanner(ctx context.Context, scn *scanner.Scanner) (*scanner.Scanner, error) {
	if stage, err := scn.BuildResolveStage(ctx); err != nil {
		return nil, err
	} else {
		scn.AddStage(stage)
	}

	if config.GetDNS().DNSTT.Enabled {
		if stage, err := scn.BuildDNSTTStage(ctx); err == nil {
			scn.AddStage(stage)
		} else {
			return nil, err
		}
	}

	if config.GetDNS().SlipStream.Enabled {
		if stage, err := scn.BuildSlipStreamStage(ctx); err == nil {
			scn.AddStage(stage)
		} else {
			return nil, err
		}
	}

	return scn, nil
}

func (m *Model) buildXrayScanner(ctx context.Context, scn *scanner.Scanner) (*scanner.Scanner, error) {
	cfg := config.GetXray()

	pre := map[string]func() error{
		"tcp": func() error {
			s, err := scn.BuildTCPStage(ctx)
			if err != nil {
				return err
			}
			scn.AddStage(s)
			return nil
		},
		"icmp": func() error {
			s, err := scn.BuildICMPStage(ctx)
			if err != nil {
				return err
			}
			scn.AddStage(s)
			return nil
		},
		"http": func() error {
			s, err := scn.BuildHTTPStage(ctx)
			if err != nil {
				return err
			}
			scn.AddStage(s)
			return nil
		},
	}

	scanType := strings.ToLower(cfg.PreScanType)
	if fn, ok := pre[scanType]; ok {
		if err := fn(); err != nil {
			return nil, fmt.Errorf("pre-scan failed: %w", err)
		}
	}

	xrayStage, err := scn.BuildXrayStage(ctx, m.xrayTemplate)
	if err != nil {
		return nil, fmt.Errorf("xray stage failed: %w", err)
	}

	scn.AddStage(xrayStage)
	return scn, nil
}

// ------------------------------------------------------------
// UI helpers
// ------------------------------------------------------------

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}
