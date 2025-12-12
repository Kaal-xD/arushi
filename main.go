package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gopkg.in/telebot.v3"
)

var startTime = time.Now() // for uptime calculation

// formatDuration converts time.Duration to readable uptime
func formatDuration(d time.Duration) string {
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

// bytesToHuman converts bytes to KB/MB/GB/etc
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

// pingCommand → sends latency
func pingCommand(c telebot.Context) error {
	start := time.Now()
	msg, err := c.Bot().Send(c.Chat(), "🏓 Pinging...")
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()

	_, _ = c.Bot().Edit(msg,
		fmt.Sprintf("🏓 Pong! `%dms`", latency),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
	)

	return nil
}

// statsCommand → uptime + CPU + RAM + Storage + Ping
func statsCommand(c telebot.Context) error {

	// Test latency
	start := time.Now()
	msg, err := c.Bot().Send(c.Chat(), "📊 Collecting system metrics...")
	if err != nil {
		return err
	}
	latency := time.Since(start).Milliseconds()

	// CPU Usage %
	cpuPercentList, _ := cpu.Percent(0, false)
	cpuPercent := cpuPercentList[0]

	// CPU Cores
	coresPhysical, _ := cpu.Counts(false)
	coresLogical, _ := cpu.Counts(true)

	// Memory
	vm, _ := mem.VirtualMemory()

	// Disk (root partition)
	diskStat, _ := disk.Usage("/") // Linux, Termux, Ubuntu, VPS

	// Uptime
	uptime := formatDuration(time.Since(startTime))

	stats := fmt.Sprintf(
		"*📊 System Performance Metrics*\n\n"+
			"🏓 *Latency:* `%dms`\n"+
			"⏱ *Uptime:* `%s`\n\n"+
			"💻 *CPU Usage:* `%.2f%%`\n"+
			"🧩 *CPU Cores:* `%d physical` | `%d logical`\n\n"+
			"🧠 *RAM:* `%.2f%%` (%s / %s)\n\n"+
			"💾 *Storage:* `%.2f%%`\n"+
			"• Used: %s\n"+
			"• Free: %s\n"+
			"• Total: %s\n",
		latency,
		uptime,
		cpuPercent,
		coresPhysical,
		coresLogical,
		vm.UsedPercent,
		bytesToHuman(vm.Used),
		bytesToHuman(vm.Total),
		diskStat.UsedPercent,
		bytesToHuman(diskStat.Used),
		bytesToHuman(diskStat.Free),
		bytesToHuman(diskStat.Total),
	)

	_, _ = c.Bot().Edit(msg, stats, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})

	return nil
}

func main() {
	pref := telebot.Settings{
		Token:  BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// /start command
	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send("👋 Welcome to *Arushi Bot*! Type /help to see all commands.",
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// /help
	bot.Handle("/help", func(c telebot.Context) error {
		help := "📘 *Available Commands:*\n\n" +
			"/start - Welcome message\n" +
			"/help - Show help menu\n" +
			"/ping - Show latency\n" +
			"/stats - System stats (CPU, RAM, Storage, Cores, Uptime)\n" +
			"/id - Show your Telegram ID"

		return c.Send(help, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// /ping
	bot.Handle("/ping", pingCommand)

	// /stats
	bot.Handle("/stats", statsCommand)

	// /id (works everywhere)
	bot.Handle("/id", func(c telebot.Context) error {
		uid := fmt.Sprint(c.Sender().ID)
		return c.Send("🆔 *Your Telegram ID:* `"+uid+"`",
			&telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	})

	// Reply to any text message
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		user := c.Sender().FirstName
		return c.Send("You said: " + c.Text() + "\nNice to meet you, " + user + " 😊")
	})

	log.Println("Bot is running…")
	bot.Start()
}
