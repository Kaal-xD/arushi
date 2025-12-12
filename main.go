package main

import (
	"log"
	"time"
	"fmt"

	"gopkg.in/telebot.v3"
)

func pingCommand(c telebot.Context) error {
	return c.Send("🏓 Pong! Bot is alive.")
}

func main() {
	pref := telebot.Settings{
		Token: BotToken, // same as before
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
			"/ping - Check bot status\n" +
			"/id - Show your Telegram ID\n"

		return c.Send(text, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// /ping command
	bot.Handle("/ping", pingCommand)

	// /id → only works in private, silent in groups
	bot.Handle("/id", func(c telebot.Context) error {
		if !c.Message().Private() {
			return nil // do nothing in group
		}

		uid := fmt.Sprint(c.Sender().ID)
		return c.Send("🆔 *Your Telegram ID:* `" + uid + "`",
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
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
