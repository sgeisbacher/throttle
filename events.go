package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type CommandEventName int

const (
	COMMAND_STARTED CommandEventName = iota
	COMMAND_SUCCEEDED
	COMMAND_ABORTED
	COMMAND_FAILED
	COMMAND_ORPHANED
)

func (cen CommandEventName) String() string {
	return [...]string{"COMMAND_STARTED", "COMMAND_SUCCEEDED", "COMMAND_ABORTED", "COMMAND_FAILED", "COMMAND_ORPHANED"}[cen]
}

type CommandEvent struct {
	ID        string
	Name      CommandEventName
	Timestamp time.Time
	Command   string
}

func hashCommand(commandArgs []string) string {
	hash := md5.Sum([]byte(strings.Join(commandArgs, " ")))
	return hex.EncodeToString(hash[:])
}

func NewCommandEvent(name CommandEventName, commandArgs []string) CommandEvent {
	return CommandEvent{
		ID:        hashCommand(commandArgs),
		Name:      name,
		Timestamp: time.Now(),
		Command:   strings.Join(commandArgs, " "),
	}
}

func (ce CommandEvent) Serialize() string {
	return fmt.Sprintf("%s||%v||%v||%s", ce.ID, ce.Timestamp.Format(time.RFC3339), ce.Name, ce.Command)
}

func Deserialize(str string) (CommandEvent, error) {
	parts := strings.Split(str, "||")
	var name CommandEventName
	switch parts[2] {
	case "COMMAND_STARTED":
		name = COMMAND_STARTED
		break
	case "COMMAND_SUCCEEDED":
		name = COMMAND_SUCCEEDED
		break

	case "COMMAND_FAILED":
		name = COMMAND_FAILED
		break

	}
	dt, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return CommandEvent{}, err
	}
	return CommandEvent{
		ID:        parts[0],
		Timestamp: dt,
		Name:      name,
		Command:   strings.Join(parts[3:], " "),
	}, nil
}

func GetLastestEventForId(eventStore EventStore, id string) (CommandEvent, bool, error) {
	found := false
	events, err := eventStore.Load()
	if err != nil {
		return CommandEvent{}, found, err
	}
	var lastEvent CommandEvent
	for _, event := range events {
		if event.ID == id {
			lastEvent = event
			found = true
		}
	}
	return lastEvent, found, nil
}
