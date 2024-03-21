package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

type Block struct {
	Time int
	Name string
}

type DayBlocks struct {
	Weekdays []int
	Blocks   []Block
}

func getDayBlocks() []DayBlocks {
	file := os.Getenv("HOME") + "/schedule/bo"

	data, err := os.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	fileString := strings.TrimSpace(string(data))

	days := strings.Split(fileString, "\n\n")

	var trimmedDays []string

	for _, element := range days {
		trimmedDays = append(trimmedDays, strings.TrimSpace(element))
	}

	var dayBlocks []DayBlocks

	for _, element := range trimmedDays {
		lines := strings.Split(element, "\n")
		daysString := lines[0]

		var daysInts []int

		for _, dayChar := range daysString {
			day, err := strconv.Atoi(string(dayChar))
			if err != nil {
				log.Fatalf("Error converting day '%c' to integer: %v", dayChar, err)
			}
			daysInts = append(daysInts, day)
		}

		var blocks []Block

		for _, line := range lines[1:] {
			parts := strings.SplitN(line, " ", 2)
			intTime, err := strconv.Atoi(parts[0])

			if err != nil {
				log.Fatalf("Could not convert time %s to int", parts[0])
			}

			block := Block{
				Time: intTime,
				Name: parts[1],
			}
			blocks = append(blocks, block)
		}

		schedule := DayBlocks{
			Weekdays: daysInts,
			Blocks:   blocks,
		}
		dayBlocks = append(dayBlocks, schedule)
	}

	return dayBlocks
}

func sendMobileNotification(blockName string, interval string) error {
	client := &http.Client{}

	body := []byte(interval)
	req, err := http.NewRequest("POST", "https://ntfy.sh/yobas_secras", bytes.NewBuffer(body))

	if err != nil {
		return fmt.Errorf("Error creating request: %w", err)
	}

	req.Header.Add("Title", blockName)

	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("Error sending request: %w", err)
	}

	defer resp.Body.Close()

	return nil
}

func sendDesktopNotification(blockName string, interval string) error {
	err := beeep.Notify(blockName, interval, "/home/tarik/Pictures/notif.png")
	if err != nil {
		return fmt.Errorf("Error sending notification: %w", err)
	}
	return nil
}

func mobileNotifs() (bool, error) {
	args := os.Args

	if len(args) < 2 {
		return false, fmt.Errorf("No CLI arguments provided")
	}

	command := args[1]

	switch command {
	case "desktop":
		return false, nil
	case "mobile":
		return true, nil
	default:
		return false, fmt.Errorf("invalid command: %s; must be 'desktop' or 'mobile'", command)
	}
}

func updateWaybarFile(currBlockName string, endTime string) error {
	fileText := currBlockName + " " + endTime

	err := os.WriteFile("/home/tarik/.config/waybar/current_block", []byte(fileText), 644)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = sendSignalToWaybar()
	if err != nil {
		log.Fatalf("Error sending signal to waybar: %v", err.Error())
	}
	return nil
}

func sendSignalToWaybar() error {
	cmd := exec.Command("pkill", "-RTMIN+9", "waybar")
	err := cmd.Run()

	return err
}

func processBlockStart(currDayBlocks []Block, currTime int) (blockName string, interval string, taskStarting bool) {
	for i, block := range currDayBlocks {
		if currTime == block.Time {
			nextTime := currDayBlocks[i+1].Time
			interval := fmt.Sprintf("%04d-%04d", block.Time, nextTime)
			return block.Name, interval, true
		}
	}
	return "", "", false
}

func processBlockCurrent(currDayBlocks []Block, currTime int) (blockName string, endTime string, err error) {
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
			event := currDayBlocks[i-1].Name
			paddedNextEventTime := fmt.Sprintf("%04d", block.Time)

			return event, paddedNextEventTime, nil
		}
	}

	return "", "", fmt.Errorf("No matching event found for current time: %d", currTime)
}

func main() {
	mobile, err := mobileNotifs()
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()

	currWeekday := int(now.Weekday())
	currWeekday = ((currWeekday + 6) % 7) + 1

	currTime, err := strconv.Atoi(now.Format("1504"))

	if err != nil {
		log.Fatal(err)
	}

	schedules := getDayBlocks()

	var currDayBlocks []Block
	for _, schedule := range schedules {
		if slices.Contains(schedule.Weekdays, currWeekday) {
			currDayBlocks = schedule.Blocks
		}
	}

	if currDayBlocks == nil {
		log.Fatalf("Current day %d not found in schedule.", currWeekday)
	}

	startingBlockName, startingBlockInterval, taskStarting := processBlockStart(currDayBlocks, currTime)

	if mobile {
		if taskStarting {
			sendMobileNotification(startingBlockName, startingBlockInterval)
		}
	} else {
		if taskStarting {
			sendDesktopNotification(startingBlockName, startingBlockInterval)
		}

		currBlockName, endTime, err := processBlockCurrent(currDayBlocks, currTime)
		if err != nil {
			log.Fatal(err)
		}

		updateWaybarFile(currBlockName, endTime)

		if err != nil {
			log.Fatal(err)
		}
	}
}
