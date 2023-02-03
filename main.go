package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type RaidLog struct {
	SeqNum      string   `json:"seqNum"`
	Time        string   `json:"Time"`
	Code        string   `json:"Code"`
	Class       int      `json:"Class"`
	Locale      string   `json:"Locale"`
	Description string   `json:"EventDescription"`
	Data        struct{} `json:"EventData"`
}

// this program should run forever and do:
// 1. convert storcli event output to logfmt or json
// 2. poll storcli for new events since last run
// 3. if new events, write that specific event(s) to the log file
// 4. ship with logrotation config
// 5. periodically wipe the storcli event log
// also need config file with like, storcli file location maybe? or just read from /etc/default/storcli_logger ?

// maybe use type=latest=N here?
func convertStorcliOutput() {
	storcliPath := "/usr/sbin/storcli"
	raidGroup := "/c0"
	// get event log
	out, err := exec.Command(storcliPath, raidGroup, "show events").Output()

	if err != nil {
		log.Fatal(err)
	}

	// split raid logs by seqNum as that seems to be unique to each event
	events := strings.Split(string(out), "seqNum:")
	// skip the first event as it's empty
	events = events[1:]

	// we need to add back the separator seqNum on the exact same line
	// can't use strings.SplitAfter as it will put seqNum and the sequence on separate lines
	for i, part := range events {
		events[i] = "seqNum:" + part
	}

	raidlogs := []RaidLog{}

	for _, event := range events {
		raidlog := RaidLog{}

		// split the individual raidlog into distinct events
		lines := strings.Split(event, "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "seqNum:") {
				raidlog.SeqNum = strings.TrimSpace(line[len("seqNum:"):])
			} else if strings.HasPrefix(line, "Time:") {
				raidlog.Time = strings.TrimSpace(line[len("Time:"):])
			} else if strings.HasPrefix(line, "Code:") {
				raidlog.Code = strings.TrimSpace(line[len("Code:"):])
			} else if strings.HasPrefix(line, "Class:") {
				// Class appears to be int
				class, _ := strconv.Atoi(strings.TrimSpace(line[len("Class:"):]))
				raidlog.Class = class
			} else if strings.HasPrefix(line, "Locale:") {
				raidlog.Locale = strings.TrimSpace(line[len("Locale:"):])
			} else if strings.HasPrefix(line, "Event Description:") {
				raidlog.Description = strings.TrimSpace(line[len("Event Description:"):])
			}

			raidlogs = append(raidlogs, raidlog)
		}
	}

	raidJSON, err := json.Marshal(raidlogs)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(raidJSON))

}

func main() {
	convertStorcliOutput()
}
