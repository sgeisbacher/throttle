package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var minutes int

func init() {
	flag.IntVar(&minutes, "m", 0, "throttle for x minutes")
}

func main() {
	flag.Parse()

	eventStore := EventStore{Path: "/tmp/eventstore.txt"}
	commandArgs := flag.Args()

	if len(commandArgs) == 0 {
		log.Fatal("no command found")
	}

	commandID := hashCommand(commandArgs)
	latestCmdEvent, found, err := GetLastestEventForId(eventStore, commandID)
	if err != nil {
		fmt.Printf("error while loading events: %v\n", err)
	}

	// todo check against system-uptime to see if aborted
	if found {
		if latestCmdEvent.Name == COMMAND_STARTED {
			log.Fatal("error: command is currently running ...")
		}
		minSinceLastRun := time.Since(latestCmdEvent.Timestamp).Minutes()
		fmt.Printf("minSinceLastRun: %v\n", time.Since(latestCmdEvent.Timestamp))
		if minSinceLastRun < float64(minutes) {
			log.Fatal("throttling ...")
		}
	}

	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = eventStore.Persist(NewCommandEvent(COMMAND_STARTED, commandArgs))
	if err != nil {
		fmt.Printf("error while persisting started-event: %v\n", err)
	}

	err = cmd.Run()
	if err != nil {
		err = eventStore.Persist(NewCommandEvent(COMMAND_FAILED, commandArgs))
		if err != nil {
			fmt.Printf("error while persisting failed-event: %v\n", err)
		}
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
				return
			}
		} else {
			log.Fatalf("run error: %v", err)
		}
	}

	err = eventStore.Persist(NewCommandEvent(COMMAND_SUCCEEDED, commandArgs))
	if err != nil {
		fmt.Printf("error while persisting suceeded-event: %v\n", err)
	}
}
