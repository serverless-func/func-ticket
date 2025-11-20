package main

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func Test_decodeSubject(t *testing.T) {
	ds, err := base64.StdEncoding.DecodeString("5L2g55qE6K6i6ZiF57ut5pyf")
	if err != nil {
		t.Error(err)
	}
	t.Log(string(ds))
}

func Test_search(t *testing.T) {
	cfg := searchConfig{
		Days:     365,
		Password: "",
		Username: "",
	}
	mails, err := search(cfg)
	if err != nil {
		t.Error(err)
	}
	for _, m := range mails {
		fmt.Println(m)
	}
}
