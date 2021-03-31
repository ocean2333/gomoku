package main

import (
	"encoding/json"
	"log"
	"testing"
)

func TestResult(t *testing.T) {
	data := "{\"color\":\"\",\"message\":\"系统：您已落子，请等待对手落子！\",\"bout\":false,\"xy\":\"saf\"}"
	res := &Result{}
	err := json.Unmarshal([]byte(data), res)
	if err != nil {
		t.Error(err)
	}
	log.Println(res.Message)
}
