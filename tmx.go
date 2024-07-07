package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
}

func getOptions (command string) []string {
	switch command {
	case "mdrun":
		return []string{"-deffnm", "-o", "-c"}
	case "grompp":
		return []string{"-f", "-c", "-r", "-p", "-n", "-maxwarn"}
	default:
		return []string{}
	}
}

func initialModel() model {
	return model{
		header:         "\n\n- tmacs, a wrapper for gromacs written by Timothy Harrison -\n\n",
		currentCommand: []string{"gmx"},
		choices:        []string{"grompp", "mdrun", "solvate", "genion", "pdb2gmx", "editconf", "make_ndx"},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			m.currentCommand = append(m.currentCommand, m.choices[m.cursor])
			m.cursor = 0
		case "enter":
			cmd := getCommand(m.currentCommand)
			fmt.Printf("\n\n%s\n\n", cmd)
			return m, tea.Quit
		}
	}

	// If the cmd hasn't been selected, show the commands.
	if len(m.currentCommand) == 1 {
		m.choices = []string{"grompp", "mdrun", "solvate", "genion", "pdb2gmx", "editconf", "make_ndx"}
		return m, nil
	} else if len(m.currentCommand) == 2 {
		m.choices = getOptions(m.currentCommand[1])
	} else {
		switch m.currentCommand[len(m.currentCommand) - 1] {
		case "-f":
			m.choices = WalkMatch("*.mdp")
		case "-c", "-r":
			m.choices = WalkMatch("*.pdb")
		case "-p":
			m.choices = WalkMatch("*.top")
		case "-deffnm":
			m.choices = WalkMatch("*.tpr")
		default:
			m.choices = getOptions(m.currentCommand[1])
		}
	}
	return m, nil
}

func (m model) View() string {
	s := m.header
	s += fmt.Sprintf("\nWorking Dir: %s\n\n", getWd())
	s += "Current Command:\n"
	for _, cmd := range m.currentCommand {
		s += fmt.Sprintf("%s\n", cmd)
	}
	s += "\n"

	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress enter to run the command.\n"
	s += "\nPress r to enter a custom option.\n"
	s += "\nPress q to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func getCommand(cmd []string) string {
	command := ""
	for _, c := range cmd {
		command += fmt.Sprintf(" %s", c)
	}
	return command
}
