package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"gopkg.in/telebot.v3"
)

var startTime = time.Now() // for uptime calculation

// safeDelete tries to delete the user's message and ignores any error.
func safeDelete(c telebot.Context) {
	if err := c.Delete(); err != nil {
		// ignore delete error (like Python's try: ... except: pass)
	}
}

// formatDuration converts a time.Duration to a human readable string.
func formatDuration(d time.Duration) string {
	// round to seconds
	secs := int(d.Seconds())
	days := secs / 86400
	secs %= 86400
	hours := secs / 3600
	secs %= 3600
	mins := secs / 60
	secs %= 60

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, mins, secs)
	}
	if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%02dm %02ds", mins, secs)
	}
	return fmt.Sprintf("%02ds", secs)
}

// pingCommand sends a temporary message to measure latency then edits it to show ms.
func pingCommand(c telebot.Context) error {
	safeDelete(c)

	start := time.Now()
	msg, err := c.Bot().Send(c.Chat(), "🏓 Pinging...")
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()

	_, _ = c.Bot().Edit(msg, fmt.Sprintf("🏓 Pong! `%dms`", latency),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
	)

	return nil
}

// statsCommand shows ping latency + uptime + CPU% + RAM usage
func statsCommand(c telebot.Context) error {
	safeDelete(c)

	// measure ping (send then edit)
	start := time.Now()
	msg, err := c.Bot().Send(c.Chat(), "📊 Gathering stats...")
	if err != nil {
		return err
	}
	latency := time.Since(start).Milliseconds()

	// CPU percent (instant snapshot)
	cpuPercents, err := cpu.Percent(0, false)
	cpuPercent := float64(0)
	if err == nil && len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	// Memory info
	vm, err := mem.VirtualMemory()
	memUsedPercent := float64(0)
	memUsed := uint64(0)
	memTotal := uint64(0)
	if err == nil {
		memUsedPercent = vm.UsedPercent
		memUsed = vm.Used
		memTotal = vm.Total
	}

	uptime := formatDuration(time.Since(startTime))

	// Build stats message
	stats := fmt.Sprintf(
		"*Bot Stats*\n\n"+
			"🏓 Latency: `%dms`\n"+
			"⏱ Uptime: `%s`\n\n"+
			"💻 CPU Usage: `%.2f%%`\n"+
			"🧠 RAM Usage: `%.2f%%` (%s / %s)\n",
		latency,
		uptime,
		cpuPercent,
		memUsedPercent,
		bytesToHuman(memUsed),
		bytesToHuman(memTotal),
	)

	_, _ = c.Bot().Edit(msg, stats, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})

	return nil
}

// bytesToHuman converts bytes to a compact human readable string.
func bytesToHuman(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(b) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB"}[exp]
	return fmt.Sprintf("%.2f %s", value, suffix)
}

func main() {
	pref := telebot.Settings{
		Token: BotToken, // <-- keeps your original BotToken variable
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// /start command
	bot.Handle("/start", func(c telebot.Context) error {
		safeDelete(c)
		return c.Send("👋 Hello! Welcome to Arushi Bot.\nType /help to see all commands.")
	})

	// /help command
	bot.Handle("/help", func(c telebot.Context) error {
		safeDelete(c)

		text := "📘 *Available Commands:*\n\n" +
			"/start - Start the bot\n" +
			"/help - Show this help menu\n" +
			"/ping - Check bot status (latency)\n" +
			"/stats - Show bot stats (uptime, CPU, RAM, latency)\n" +
			"/id - Show your Telegram ID\n"

		return c.Send(text, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// /ping command
	bot.Handle("/ping", pingCommand)

	// /stats command
	bot.Handle("/stats", statsCommand)

	// /id → only works in private, silent in groups
	bot.Handle("/id", func(c telebot.Context) error {
		safeDelete(c)

		if !c.Message().Private() {
			return nil // silent in groups
		}

		uid := fmt.Sprint(c.Sender().ID)
		return c.Send("🆔 *Your Telegram ID:* `"+uid+"`",
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// Reply to any text message
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		safeDelete(c)

		user := c.Sender().FirstName
		msg := c.Text()

		reply := "You said: " + msg + "\nNice to meet you, " + user + " 😊"

		return c.Send(reply)
	})

	log.Println("Bot is running...")
	bot.Start()
}
