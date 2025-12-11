package main

import (
	"log"
	"strings"
	"sync"
	"time"

	tb "gopkg.in/telebot.v3"
)

// in-memory maps (not persistent)
var (
	welcomeMsgMu sync.Mutex
	welcomeMsgs  = make(map[int64]int)   // userID -> messageID (welcome msg)

	subsMu sync.Mutex
	subs   = make(map[int64]struct{}) // subscribed user IDs for updates
)

func StartBot() {
	pref := tb.Settings{
		Token:  BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create reply markup and buttons
	markup := &tb.ReplyMarkup{}
	updatesBtn := tb.Btn{Text: "📢 Updates"}
	supportBtn := tb.Btn{Text: "💬 Support"}
	aboutBtn := tb.Btn{Text: "ℹ️ About"}
	closeBtn := tb.Btn{Text: "❌ Close"}
	nextBtn := tb.Btn{Text: "➡️ Next"}
	backBtn := tb.Btn{Text: "⬅️ Back"}
	subBtn := tb.Btn{Text: "🔔 Subscribe"}
	unsubBtn := tb.Btn{Text: "🔕 Unsubscribe"}

	// Inline layout: page 1 rows (Updates, Support), (About, Subscribe)
	markup.Inline(
		markup.Row(updatesBtn, supportBtn),
		markup.Row(aboutBtn, subBtn),
		markup.Row(closeBtn),
	)

	// Page 2 layout (About details + Back)
	page2 := &tb.ReplyMarkup{}
	page2.Inline(
		page2.Row(backBtn),
	)

	// /start handler — sends welcome message with keyboard, auto-deletes previous welcome
	bot.Handle("/start", func(c tb.Context) error {
		uid := c.Sender().ID

		// Delete previous welcome (if any)
		welcomeMsgMu.Lock()
		if mid, ok := welcomeMsgs[uid]; ok {
			_ = bot.Delete(&tb.Message{ID: mid, Chat: &tb.Chat{ID: uid}})
			delete(welcomeMsgs, uid)
		}
		welcomeMsgMu.Unlock()

		welcome := "👋 *Welcome to Arushi Bot!*\n\n" +
			"I am a fast Golang bot. Use the buttons below to get updates or contact support.\n\n" +
			"✨ *Quick features*\n" +
			"• Fast responses\n" +
			"• Inline menus\n" +
			"• Owner broadcast (admins only)\n\n" +
			"Type /help for commands."

		// send message with markup and store message id
		msg, err := c.Send(welcome, &tb.SendOptions{
			ParseMode:   tb.ModeMarkdown,
			ReplyMarkup: markup,
		})
		if err == nil {
			welcomeMsgMu.Lock()
			welcomeMsgs[uid] = msg.ID
			welcomeMsgMu.Unlock()
		}
		return err
	})

	// Updates button — shows channel and subscribe option
	markup.Handle(&updatesBtn, func(c tb.Context) error {
		uid := c.Sender().ID
		// reply with channel link
		text := "📢 *Updates Channel:*\n" + Channel + "\n\n" +
			"Use Subscribe to get broadcast messages directly."
		return c.Edit(text, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: markup})
	})

	// Support button — show support link
	markup.Handle(&supportBtn, func(c tb.Context) error {
		text := "💬 *Support:*\n" + Support
		return c.Edit(text, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: markup})
	})

	// About button — open page 2 (About details)
	markup.Handle(&aboutBtn, func(c tb.Context) error {
		aboutText := "ℹ️ *About Arushi Bot*\n\n" +
			"This Golang bot is built for speed and simplicity.\n\n" +
			"• Language: Go (Golang)\n" +
			"• Library: telebot.v3\n" +
			"• Owner: " + formatID(OwnerID) + "\n\n" +
			"Use the Back button to return."
		// send page2 with back button
		return c.Edit(aboutText, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: page2})
	})

	// Back button (returns to main markup)
	page2.Handle(&backBtn, func(c tb.Context) error {
		text := "👋 Back to menu. Use the buttons below."
		return c.Edit(text, &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: markup})
	})

	// Close button — delete the welcome message (the message the user pressed)
	markup.Handle(&closeBtn, func(c tb.Context) error {
		// delete the message that contains the buttons
		_ = c.Delete()
		// also remove from welcomeMsgs map if present
		uid := c.Sender().ID
		welcomeMsgMu.Lock()
		delete(welcomeMsgs, uid)
		welcomeMsgMu.Unlock()
		return nil
	})

	// Subscribe button — add user to subscribers
	markup.Handle(&subBtn, func(c tb.Context) error {
		uid := c.Sender().ID
		subsMu.Lock()
		subs[uid] = struct{}{}
		subsMu.Unlock()
		return c.Respond(&tb.CallbackResponse{Text: "Subscribed to updates ✅", ShowAlert: false})
	})

	// Unsubscribe button — remove user
	markup.Handle(&unsubBtn, func(c tb.Context) error {
		uid := c.Sender().ID
		subsMu.Lock()
		delete(subs, uid)
		subsMu.Unlock()
		return c.Respond(&tb.CallbackResponse{Text: "Unsubscribed 🔕", ShowAlert: false})
	})

	// /help - list commands
	bot.Handle("/help", func(c tb.Context) error {
		help := "📘 *Available Commands:*\n\n" +
			"/start - Start the bot and show menu\n" +
			"/help - Show this help\n" +
			"/ping - Check bot status\n" +
			"/broadcast <text> - Owner only (send to subscribers)\n"
		return c.Send(help, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
	})

	// /ping
	bot.Handle("/ping", func(c tb.Context) error {
		return c.Send("🏓 Pong! Bot is alive.")
	})

	// /broadcast owner-only - sends message to all subscribers
	bot.Handle("/broadcast", func(c tb.Context) error {
		if c.Sender().ID != OwnerID {
			return c.Reply("⛔ You are not allowed to use this command.")
		}
		// get payload after command
		payload := strings.TrimSpace(strings.TrimPrefix(c.Message().Text, "/broadcast"))
		if payload == "" {
			return c.Reply("Usage: /broadcast <message>")
		}
		// send to all subscribers (simple sequential send)
		subsMu.Lock()
		targets := make([]int64, 0, len(subs))
		for id := range subs {
			targets = append(targets, id)
		}
		subsMu.Unlock()

		sent := 0
		for _, id := range targets {
			_, err := bot.Send(&tb.Chat{ID: id}, payload)
			if err == nil {
				sent++
			} else {
				// if sending failed (user blocked bot), remove from subscribers
				subsMu.Lock()
				delete(subs, id)
				subsMu.Unlock()
			}
			// small delay to avoid hitting limits
			time.Sleep(200 * time.Millisecond)
		}
		return c.Replyf("✅ Broadcast completed. Sent to %d users.", sent)
	})

	// Generic text handler (echo)
	bot.Handle(tb.OnText, func(c tb.Context) error {
		user := c.Sender().FirstName
		msg := c.Text()
		reply := "You said: " + msg + "\nNice to meet you, " + user + " 😊"
		return c.Send(reply)
	})

	log.Println("Bot is running...")
	bot.Start()
}

// helper to format owner id as mention or id string
func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
