package main

import (
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

func main() {
	pref := telebot.Settings{
		Token: BotToken, // <- replace with fresh token
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// /start command
	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send("👋 Hello! Welcome to Arushi Bot.\nType /help to see all commands.")
	})

	// /help command
	bot.Handle("/help", func(c telebot.Context) error {
		text := "📘 *Available Commands:*\n\n" +
			"/start - Start the bot\n" +
			"/help - Show this help menu\n" +
			"/ping - Check bot status\n"

		return c.Send(text, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// /ping command
	bot.Handle("/ping", func(c telebot.Context) error {
		return c.Send("🏓 Pong! Bot is alive.")
	})

	// Reply to any text message
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		user := c.Sender().FirstName
		msg := c.Text()

		reply := "You said: " + msg + "\nNice to meet you, " + user + " 😊"

		return c.Send(reply)
	})

	log.Println("Bot is running...")
	bot.Start()
}
