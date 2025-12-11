package main

import (
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

func main() {
	pref := telebot.Settings{
		Token:  "2027709241:AAEcqro-6YUIMLk0ip11cnRnM-g-oGVIxd8",
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send("Hello, your Go bot is running on Ubuntu 22.04 🚀")
	})

	bot.Start()
}
