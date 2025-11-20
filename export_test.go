package main

import (
	"strings"
	"testing"
)

func Test_export(t *testing.T) {
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
	f, err := export(tickets)
	if err != nil {
		t.Error(err)
	}
	err = f.SaveAs("tickets.xlsx")
	if err != nil {
		t.Error(err)
	}
}

func Test_exportAll(t *testing.T) {
	cfg := searchConfig{
		Days:     365 * 7,
		Password: "",
		Username: "",
	}
	mails, err := search(cfg)
	if err != nil {
		t.Error(err)
	}
	var tickets []Ticket
	for _, m := range mails {
		ts, err := parse(m)
		if err != nil {
			continue
		} else {
			tickets = append(tickets, ts...)
		}
	}
	f, err := export(tickets)
	if err != nil {
		t.Error(err)
	}
	err = f.SaveAs("tickets.xlsx")
	if err != nil {
		t.Error(err)
	}
}
