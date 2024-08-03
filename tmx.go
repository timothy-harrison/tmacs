package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func getWd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't get working directory")
	}
	return path
}

func WalkMatch(pattern string) []string {
	var matches []string
	err := filepath.Walk(getWd(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return []string{"No Matching Files"}
	}
	return matches
}

type model struct {
	header         string
	currentCommand []string
	choices        []string
	cursor         int
	runningCommand bool
	textInput textinput.Model
	filePicker filepicker.Model
	selectedFile string
	openFilePicker bool
	err error
}

type runCmd bool

func getOptions(command string) []string {
	switch command {
	case "mdrun":
		return []string{"-deffnm", "-o", "-c", "-s"}
	case "grompp":
		return []string{"-f", "-c", "-r", "-p", "-n", "-maxwarn"}
	default:
		return []string{}
	}
}

func initialModel() model {

	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 20

	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.FileAllowed = true


	return model{
		header:         "- tmacs -\nA wrapper for gromacs written by Timothy Harrison\n",
		currentCommand: []string{"gmx"},
		choices:        []string{"grompp", "mdrun", "solvate", "genion", "pdb2gmx", "editconf", "make_ndx"},
		runningCommand: false,
		textInput: ti,
		filePicker: fp,
		selectedFile: "",
		openFilePicker: false,
		err: nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case runCmd:
		m.runningCommand = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 && !m.textInput.Focused() {
				m.cursor--
			} else {
				m.cursor = len(m.choices)-1
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 && !m.textInput.Focused() {
				m.cursor++
			}
		case " ", "enter":
			if m.textInput.Focused() {
				m.currentCommand = append(m.currentCommand, m.textInput.Value())
				m.textInput.Blur()
				return m, nil
			} else {
				m.currentCommand = append(m.currentCommand, m.choices[m.cursor])
				m.cursor = 0
			}
		case "+":
			m.runningCommand = true
			return m, RunCommand(m.currentCommand)
		case "_":
			//m.choices = []string{"grompp", "mdrun", "solvate", "genion", "pdb2gmx", "editconf", "make_ndx"}
			m.cursor = 0
			m.currentCommand = []string{"gmx"}
		case "b":
			if m.runningCommand {
				m.runningCommand = false
				m = initialModel()
				return m, nil
			}
		case "backspace":
			if m.openFilePicker {
				m.openFilePicker = false
				return m, nil
			}
			if len(m.currentCommand) == 1 {
				return m, nil
			} else if len(m.currentCommand) == 2 {
				m.currentCommand = []string{"gmx"}
			} else {
				m.currentCommand = m.currentCommand[:len(m.currentCommand)-1]
				m.cursor = 0
			}
		case "r":
			if len(m.currentCommand) >= 2 {
				return m, m.textInput.Focus()
			}
		case "esc":
			if m.textInput.Focused() {
				m.textInput.Blur()
			}
			return m, nil
		case ":":
			m.openFilePicker = true
		}

	}

	if m.openFilePicker {
		m.filePicker, cmd = m.filePicker.Update(msg)
		return m, cmd
	}

	if m.textInput.Focused() {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	// If the cmd hasn't been selected, show the commands.
	if len(m.currentCommand) == 1 {
		m.choices = []string{"grompp", "mdrun", "solvate", "genion", "pdb2gmx", "editconf", "make_ndx"}
		return m, nil
	} else if len(m.currentCommand) == 2 {
		m.choices = getOptions(m.currentCommand[1])
	} else {
		switch m.currentCommand[len(m.currentCommand)-1] {
		case "-f":
			m.choices = WalkMatch("*.mdp")
		case "-c", "-r":
			m.choices = WalkMatch("*.pdb")
		case "-p":
			m.choices = WalkMatch("*.top")
		case "-deffnm":
			m.choices = WalkMatch("*.tpr")
		case "-s":
			m.choices = WalkMatch("*.tpr")
		default:
			m.choices = getOptions(m.currentCommand[1])
		}
	}
	return m, nil
}

func (m model) View() string {

	var s string

	if m.openFilePicker {
		var s strings.Builder
		s.WriteString("\n  ")
		if m.err != nil {
			s.WriteString(m.filePicker.Styles.DisabledFile.Render(m.err.Error()))
		} else if m.selectedFile == "" {
			s.WriteString("Pick a file:")
		} else {
			s.WriteString("Selected file: " + m.filePicker.Styles.Selected.Render(m.selectedFile))
		}
		s.WriteString("\n\n" + m.filePicker.View() + "\n")
		s.WriteString("\nPress backspace to go back")
		return s.String()
	}

	s = m.header
	s += fmt.Sprintf("\nWorking Dir: %s\n\n", getWd())

	if m.runningCommand {
		s += fmt.Sprintf("Running Command:\n\n%s\n\n", GetCommand(m.currentCommand))
		s += "Press b to go back\nPress q to quit\n\n"
		return s
	}

	s += "Current Command:\n"
	for _, cmd := range m.currentCommand {
		s += fmt.Sprintf("%s ", cmd)
	}
	s += "\n\n"

	if m.textInput.Focused() {
		s += "\n" + m.textInput.View() + "\n\n"
		s += "Press ctrl-c to quit.\n"
		return s
	}

	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	if len(m.currentCommand) >= 2 {
		s += "\nPress r to enter a custom option.\n"
		s += "Press : to open the file picker\n"
		s += "Press _ to reset the comand.\n"
		s += "Press + to run the command.\n"
	}
	s += "\nPress ctrl-c to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func GetCommand(cmd []string) string {
	command := ""
	for _, c := range cmd {
		command += fmt.Sprintf(" %s", c)
	}
	return command
}

func RunCommand(command []string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("gmx", command[1:]...)
		if err := cmd.Run(); err != nil {
			return runCmd(false)
		}

		return runCmd(true)
	}
}
