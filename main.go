package main

import (
	"log"
	"time"

	tb "gopkg.in/telebot.v3"
)

func main() {
	pref := tb.Settings{
		Token:  BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// MAIN MENU BUTTONS
	menu := &tb.ReplyMarkup{}
	btnUpdates := menu.Data("📢 Updates", "updates")
	btnSupport := menu.Data("💬 Support", "support")
	btnAbout := menu.Data("ℹ️ About", "about")
	btnClose := menu.Data("❌ Close", "close")

	menu.Inline(
		menu.Row(btnUpdates, btnSupport),
		menu.Row(btnAbout),
		menu.Row(btnClose),
	)

	// ABOUT PAGE BUTTONS
	aboutMenu := &tb.ReplyMarkup{}
	btnBack := aboutMenu.Data("⬅️ Back", "back")
	aboutMenu.Inline(aboutMenu.Row(btnBack))

	// /start command
	bot.Handle("/start", func(c tb.Context) error {

		text := "👋 *Welcome to Arushi Bot!*\n\n" +
			"Use the menu below to navigate.\n\n" +
			"✨ *Features*\n" +
			"• Updates\n" +
			"• Support\n" +
			"• About\n"

		return c.Send(text, &tb.SendOptions{
			ParseMode:   tb.ModeMarkdown,
			ReplyMarkup: menu,
		})
	})

	// CALLBACK HANDLERS (CORRECT FOR TELEBOT v3)

	bot.Handle(&btnUpdates, func(c tb.Context) error {
		return c.Edit("📢 *Updates Channel:*\n"+Channel,
			&tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu})
	})

	bot.Handle(&btnSupport, func(c tb.Context) error {
		return c.Edit("💬 *Support:*\n"+Support,
			&tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu})
	})

	bot.Handle(&btnAbout, func(c tb.Context) error {
		text := "ℹ️ *About Arushi Bot*\n\n" +
			"• Language: Go (Golang)\n" +
			"• Library: telebot.v3\n" +
			"• Fast & Lightweight\n\n" +
			"Tap Back to return."
		return c.Edit(text, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: aboutMenu})
	})

	bot.Handle(&btnBack, func(c tb.Context) error {
		return c.Edit("👋 Back to main menu.",
			&tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu})
	})

	bot.Handle(&btnClose, func(c tb.Context) error {
		return c.Delete()
	})

	// /ping
	bot.Handle("/ping", func(c tb.Context) error {
		return c.Send("🏓 Pong!")
	})

	// simple echo
	bot.Handle(tb.OnText, func(c tb.Context) error {
		return c.Send("You said: " + c.Text())
	})

	log.Println("Bot running…")
	bot.Start()
}
