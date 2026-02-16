package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	anim "github.com/krisavdome/diskographia/ascii"
)

type Model struct {
	Anim   anim.Model
	width  int
	height int
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
	p := tea.NewProgram(Model{Anim: anim.NewWithDefaults()}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
