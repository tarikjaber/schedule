package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	Time           int
	CurrBlockIndex int
	DayBlocks      []Block
	Width          int
	Height         int
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

	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tickMsg:
		return m, tickCmd()
	}

	return m, nil
}

func tickCmd() tea.Cmd {
	// return tea.Tick(time.Minute*5, func(t time.Time) tea.Msg {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func divideAndRoundUp(number int, divisor int) int {
	return (number + divisor - 1) / divisor
}

func timeToMinInDay(time int) int {
	return time/100*60 + time%100
}

func secondsTo(toHour, toMinute int) int {
	now := time.Now()
	nowSecondsInDay := now.Hour()*60*60 + now.Minute()*60 + now.Second()
	toSecondsInDay := toHour*60*60 + toMinute*60

	return toSecondsInDay - nowSecondsInDay
}

func (m model) renderBlocks(dayBlocks []Block, timeToNext string) string {
	s := ""

	for i, dayBlock := range dayBlocks[:len(dayBlocks)-1] {
		timePadded := fmt.Sprintf("%04d", dayBlock.Time)
		taskStr := timePadded + " " + dayBlock.Name

		if i == m.CurrBlockIndex {
			s += currStyle.MarginLeft(2).Render(taskStr)
			s += timeToNext
		} else {
			s += regularStyle.MarginLeft(2).Render(taskStr)
		}
	}

	return s
}

func prettySecondsTo(toHour, toMinute int) string {
	secondsToBlock := secondsTo(toHour, toMinute)

	numHoursLeft := secondsToBlock / (60 * 60)
	numMinutesLeft := secondsToBlock/60 - numHoursLeft*60
	numSecondsLeft := secondsToBlock % 60

	result := ""
	if numHoursLeft > 0 {
		result += fmt.Sprintf("%dh ", numHoursLeft)
	}
	if numMinutesLeft > 0 {
		result += fmt.Sprintf("%dm ", numMinutesLeft)
	}
	result += fmt.Sprintf("%ds ", numSecondsLeft)

	return result
}

func (m model) View() string {
	dummyEnd := Block{
		Time: 2400,
		Name: "Free",
	}

	dayBlocks := append(m.DayBlocks, dummyEnd)

	nextBlockTime := dayBlocks[m.CurrBlockIndex+1].Time
	nextBlockHour := nextBlockTime / 100
	nextBlockMinute := nextBlockTime % 100

	prettySecondsToString := timeToNextStyle.Render(prettySecondsTo(nextBlockHour, nextBlockMinute))
	s := m.renderBlocks(dayBlocks, prettySecondsToString)

	return s
}

func runCli() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal: ", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There was an error: %v", err)
		os.Exit(1)
	}
}
