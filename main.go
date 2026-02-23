package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	anim "github.com/krisavdome/disshkographia/ascii"
)

const (
	host = "0.0.0.0"
	port = "22"
)

type Model struct {
	Anim              anim.Model
	width             int
	height            int
	hasDarkBackground bool
}

func (m Model) Init() tea.Cmd {
	return m.Anim.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newAnim, cmd := m.Anim.Update(msg)
	m.Anim = newAnim.(anim.Model)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, cmd
	}

	return m, cmd
}

func (m Model) View() string {
	content := m.Anim.View()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting ssh server", "host", host, "port", port)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("stopping ssh server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	renderer := bubbletea.MakeRenderer(s)

	lipgloss.SetDefaultRenderer(renderer)

	m := Model{
		Anim:              anim.NewWithDefaults(),
		width:             pty.Window.Width,
		height:            pty.Window.Height,
		hasDarkBackground: renderer.HasDarkBackground(),
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
