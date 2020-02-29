package main

import (
	"bufio"
	"os"
)

type EventStore struct {
	Path string
}

func (e EventStore) Load() ([]CommandEvent, error) {
	file, err := os.Open(e.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commandEvents []CommandEvent
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		event, err := Deserialize(scanner.Text())
		if err != nil {
			return nil, err
		}
		commandEvents = append(commandEvents, event)
	}
	return commandEvents, nil
}

func (e EventStore) Persist(commandEvent CommandEvent) error {
	file, err := os.OpenFile(e.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(commandEvent.Serialize() + "\n")
	if err != nil {
		return err
	}
	return nil
}
