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
	Time string
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

			block := Block{
				Time: parts[0],
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

func sendSignalToWaybar() error {
	cmd := exec.Command("pkill", "-RTMIN+9", "waybar")
	err := cmd.Run()

	return err
}

func main() {
	notifyOnMobile, err := mobileNotifs()
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

	dummyStart := Block{
		Time: "0000",
		Name: "Free",
	}
	dummyEnd := Block{
		Time: "2400",
		Name: "Free",
	}

	currDayBlocks = append(currDayBlocks, dummyEnd)
	currDayBlocks = append([]Block{dummyStart}, currDayBlocks...)

	for i, currBlock := range currDayBlocks[1 : len(currDayBlocks)-1] {
		originalIndex := i + 1
		itemTime, err := strconv.Atoi(currBlock.Time)

		if err != nil {
			log.Fatalf("Could not convert schedule time %s to int. For item %s %s.", currBlock.Time, currBlock.Time, currBlock.Name)
		}

		if currTime == itemTime {
			nextTime := currDayBlocks[originalIndex+1].Time
			interval := fmt.Sprintf("%s-%s", currBlock.Time, nextTime)

			if notifyOnMobile {
				sendMobileNotification(currBlock.Name, interval)
			} else {
				sendDesktopNotification(currBlock.Name, interval)
			}
		}

		if currTime < itemTime && !notifyOnMobile {
			prevEvent := currDayBlocks[originalIndex-1].Name
			currBlockText := currBlock.Time + " " + prevEvent
			err := os.WriteFile("/home/tarik/.config/waybar/current_block", []byte(currBlockText), 644)

			if err != nil {
				log.Fatal(err.Error())
			}

			err = sendSignalToWaybar()
			if err != nil {
				log.Fatalf("Error sending signal to waybar: %v", err.Error())
			}

			break
		}
	}
}
