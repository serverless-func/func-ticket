package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func Test_parse(t *testing.T) {
	text := ``

	mails := strings.Split(text, "\n")
	var tickets []Ticket
	for _, mail := range mails {
		if len(mail) == 0 {
			continue
		}
		ts, err := parse(mail)
		if err != nil {
			continue
		} else {
			tickets = append(tickets, ts...)
		}
	}
	ticketBytes, _ := json.Marshal(tickets)
	fmt.Println(string(ticketBytes))
}
