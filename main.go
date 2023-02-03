package main

import (
	"flag"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type RaidLog struct {
	SeqNum      string `json:"seqNum"`
	Time        string `json:"Time"`
	Code        string `json:"Code"`
	Class       int    `json:"Class"`
	Locale      string `json:"Locale"`
	Description string `json:"EventDescription"`
	Data        string `json:"EventData"`
}

// this program should run forever and do:
// 1. convert storcli event output to logfmt or json
// 2. poll storcli for new events since last run
// 3. if new events, write that specific event(s) to the log file
// 4. ship with logrotation config
// 5. periodically wipe the storcli event log
// also need config file with like, storcli file location maybe? or just read from /etc/default/storcli_logger ?

// maybe use type=latest=N here?
func convertStorcliOutput(storcliPath string, raidGroup string, latest int) {

	// e.g storcli /c0 show events type=latest=5
	// gets last 5 events
	latestArg := "type=latest=" + strconv.Itoa(latest)
	// get event log
	out, err := exec.Command(storcliPath, raidGroup, "show events", latestArg).Output()

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

		for counter, line := range lines {
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
			} else if strings.HasPrefix(line, "Event Data:") {
				// eventData is everything after EventData up until "CLI Version = " (the last event)
				// so if we see CLI Version =, we've gone too far and should just discard everything
				eventData := strings.Split(strings.Join(lines[counter:], ";"), "CLI Version =")
				raidlog.Data = eventData[0]

			}

			raidlogs = append(raidlogs, raidlog)
		}
	}

	for _, raid := range raidlogs {
		log.WithFields(log.Fields{
			"SeqNum":            raid.SeqNum,
			"Time":              raid.Time,
			"Code":              raid.Code,
			"Class":             raid.Class,
			"Locale":            raid.Locale,
			"Event Description": raid.Description,
			"Event Data":        raid.Data,
		}).Info()

	}

}

func getMaxEvents(storcliPath string, raidGroup string) (latest int) {

	// get event log
	out, err := exec.Command(storcliPath, raidGroup, "show events").Output()
	if err != nil {
		log.Fatal(err)
	}
	latest = strings.Count(string(out), "seqNum")
	return latest
}

func main() {
	var (
		storcliPath = flag.String("storcli.path", "/usr/sbin/storcli",
			"Storcli Binary Path")
		raidGroup = flag.String("raidgroup", "/c0",
			"Storcli raid group (/cx)")
		logPath = flag.String("log.path", "/var/log/storcli.log",
			"Desired path to log to")
	)

	f, err := os.OpenFile(*logPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	// on program first start or restart
	max := getMaxEvents(*storcliPath, *raidGroup)
	convertStorcliOutput(*storcliPath, *raidGroup, max)
	var latest, newMax int
	for {
		latest = max
		if latest != 0 {
			convertStorcliOutput(*storcliPath, *raidGroup, latest)
		}
		// we only want to get events that are new since our last run
		newMax = getMaxEvents(*storcliPath, *raidGroup)
		max = newMax - max
		time.Sleep(20 * time.Second)
	}
}
