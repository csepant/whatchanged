package output

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spinnerModel struct {
	spinner  spinner.Model
	message  string
	quitting bool
}

type spinnerDoneMsg struct{}
type spinnerMessageMsg string

func newSpinnerModel(message string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	if !noColor {
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	}
	return spinnerModel{
		spinner: s,
		message: message,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case spinnerDoneMsg:
		m.quitting = true
		return m, tea.Quit
	case spinnerMessageMsg:
		m.message = string(msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("%s %s\n", m.spinner.View(), m.message)
}

// RunWithSpinner runs a function while showing a spinner. The function receives
// a callback to update the spinner message.
func RunWithSpinner(initialMsg string, fn func(updateMsg func(string)) error) error {
	if noColor {
		fmt.Printf("  %s\n", initialMsg)
		return fn(func(msg string) {
			fmt.Printf("  %s\n", msg)
		})
	}

	var fnErr error
	p := tea.NewProgram(newSpinnerModel(initialMsg))

	go func() {
		fnErr = fn(func(msg string) {
			p.Send(spinnerMessageMsg(msg))
		})
		p.Send(spinnerDoneMsg{})
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	return fnErr
}
