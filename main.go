package main

import (
	"nsy_chat_live/api"
	"nsy_chat_live/config"
	"nsy_chat_live/live"
	"time"
)

func main() {
	Init()

	live.StartLiveWatcher()

	for {
		time.Sleep(time.Hour)
	}
}

func Init() {
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}
	if err := api.InitHttp(); err != nil {
		panic(err)
	}
}
