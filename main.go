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

// make premium looking status bar
func makeBar(percent float64) string {
    totalBars := 10
    filledBars := int((percent / 100) * float64(totalBars))

    bar := ""
    for i := 0; i < totalBars; i++ {
        if i < filledBars {
            bar += "▰"
        } else {
            bar += "▱"
        }
    }
    return bar
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

// statsCommand → uptime + ping + Storage + RAM + CPU
func statsCommand(c telebot.Context) error {

    // Latency
    start := time.Now()
    msg, err := c.Bot().Send(c.Chat(), "📊 Fetching stats...")
    if err != nil {
        return err
    }
    latency := time.Since(start).Milliseconds()

    // Uptime
    uptime := formatDuration(time.Since(startTime))

    // CPU
    cpuPerc, _ := cpu.Percent(0, false)
    cpuUsage := cpuPerc[0]
    physicalCores, _ := cpu.Counts(false)
    logicalCores, _ := cpu.Counts(true)

    // RAM
    vm, _ := mem.VirtualMemory()

    // Storage
    diskStat, _ := disk.Usage("/")

    // Build clean decorated stats with bars
    stats := fmt.Sprintf(
        "📊 *System Performance Metrics*\n\n"+
            "⚡ *Latency:* `%dms`\n"+
            "⏱ *Uptime:* `%s`\n\n"+

            "💾 *Storage*\n"+
            "%s `%.2f%%`\n"+
            "└ (%s / %s)\n\n"+

            "🧠 *RAM*\n"+
            "%s `%.2f%%`\n"+
            "└ (%s / %s)\n\n"+

            "💻 *CPU*\n"+
            "%s `%.2f%%`\n"+
            "└ Cores: %d Physical | %d Logical\n",
        
        latency,
        uptime,

        makeBar(diskStat.UsedPercent),
        diskStat.UsedPercent,
        bytesToHuman(diskStat.Used),
        bytesToHuman(diskStat.Total),

        makeBar(vm.UsedPercent),
        vm.UsedPercent,
        bytesToHuman(vm.Used),
        bytesToHuman(vm.Total),

        makeBar(cpuUsage),
        cpuUsage,
        physicalCores,
        logicalCores,
    )

    _, _ = c.Bot().Edit(msg, stats, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})

    return nil
}

// info - cmd to get user info
func GetUserInfo(bot *telebot.Bot, c telebot.Context) error {

    sender := c.Sender()

    msg := fmt.Sprintf(
        "👤 *Your Telegram Info*\n\n"+
            "• *Name:* %s\n"+
            "• *ID:* `%d`\n"+
            "• *Username:* @%s",
        sender.FirstName,
        sender.ID,
        sender.Username,
    )

    return c.Send(
        msg,
        &telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
    )
}

func ytCommand(c telebot.Context) error {
	query := c.Args()
	if len(query) == 0 {
		return c.Send("Usage: /yt query or url")
	}

	text := strings.Join(query, " ")

	yt := youtube.Client{}
	var videoURL string

	if strings.Contains(text, "youtube.com") || strings.Contains(text, "youtu.be") {
		videoURL = text
	} else {
		results, err := yt.Search(text)
		if err != nil || len(results) == 0 {
			return c.Send("No results found.")
		}
		videoURL = "https://www.youtube.com/watch?v=" + results[0].ID
	}

	video, err := yt.GetVideo(videoURL)
	if err != nil {
		return c.Send("Failed to get video info.")
	}

	formats := video.Formats.Type("audio")
	if len(formats) == 0 {
		return c.Send("No audio stream available.")
	}

	streamURL, err := yt.GetStreamURL(video, &formats[0])
	if err != nil {
		return c.Send("Failed to extract audio URL.")
	}

	msg := fmt.Sprintf(
		"🎵 *YouTube Info*\n\n"+
			"*ID:* `%s`\n"+
			"*Title:* %s\n"+
			"*Duration:* %v\n"+
			"*Audio Link:* [Click Here](%s)",
		video.ID,
		video.Title,
		video.Duration,
		streamURL,
	)

	return c.Send(msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
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

        botUser := c.Bot().Me.Username     // username
        botName := c.Bot().Me.FirstName    // name shown
        botMention := "[" + botName + "](https://t.me/" + botUser + ")"

        caption := "👋 Welcome to " + botMention + "! Type /help to see all commands."

        return c.Send(
            caption,
            &telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
        )
    })

	// /help
	bot.Handle("/help", func(c telebot.Context) error {
		help := "📘 *Available Commands:*\n\n" +
			"/start - Welcome message\n" +
			"/help - Show help menu\n" +
			"/ping - Show latency\n" +
			"/stats - System stats\n" +
			"/id - Show your Telegram ID\n" +
		    "/info - Show your Telegram Info"

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

	bot.Handle("/info", func(c telebot.Context) error {
        return GetUserInfo(bot, c)
    })
	bot.Handle("/yt", ytCommand)

	log.Println("Bot is running…")
	bot.Start()
}
