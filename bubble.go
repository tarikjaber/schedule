package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	Time           int
	CurrBlockIndex int
	DayBlocks      []Block
}

type tickMsg time.Time

func getCurrBlockIndex(currDayBlocks []Block, currTime int) (int, error) {
	dummyStart := Block{
		Time: 0,
		Name: "Free",
	}
	dummyEnd := Block{
		Time: 2400,
		Name: "Free",
	}

	currDayBlocks = append(currDayBlocks, dummyEnd)
	currDayBlocks = append([]Block{dummyStart}, currDayBlocks...)

	for i, block := range currDayBlocks {
		if currTime < block.Time {
			return i - 2, nil
		}
	}

	return 0, fmt.Errorf("No matching event found for current time: %d", currTime)
}

func getCurrModel() model {
	now := time.Now()

	currWeekday := int(now.Weekday())
	currWeekday = ((currWeekday + 6) % 7) + 1

	currTime, err := strconv.Atoi(now.Format("1504"))

	if err != nil {
		log.Fatal(err)
	}

	currDayBlocks := getCurrDayBlocks(currWeekday)
	currBlockIndex, err := getCurrBlockIndex(currDayBlocks, currTime)

	if err != nil {
		log.Fatalf("Curr block index could not be found. Err: %v.", err)
	}

	return model{
		Time:           currTime,
		CurrBlockIndex: currBlockIndex,
		DayBlocks:      currDayBlocks,
	}
}

func initialModel() model {
	return getCurrModel()
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.ClearScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		}
	case tickMsg:
		return getCurrModel(), tickCmd()
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func tickCmd() tea.Cmd {
	// return tea.Tick(time.Minute*5, func(t time.Time) tea.Msg {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() string {
	s := ""
	var currStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#76ABAE")).
		PaddingTop(1).
		PaddingLeft(4).
		Width(22)
	var regStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EEEEEE")).
		PaddingTop(1).
		PaddingLeft(4).
		Width(22)

	for i, dayBlock := range m.DayBlocks {
		taskStr := strconv.Itoa(dayBlock.Time) + " " + dayBlock.Name
		if i == m.CurrBlockIndex {
			s += currStyle.Render(taskStr) + "\n"
		} else {
			s += regStyle.Render(taskStr) + "\n"
		}
	}

	return s
}

func runCli() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There was an error: %v", err)
		os.Exit(1)
	}
}
