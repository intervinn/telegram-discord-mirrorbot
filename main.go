package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

var (
	telegram *tele.Bot
	discord  *dgo.Session

	tgChannel int64
	dsChannel string

	err error
)

func isEmpty(str string) bool {
	return strings.ReplaceAll(str, " ", "") == ""
}

func initTelegram() {
	telegram, err = tele.NewBot(tele.Settings{
		Token:  os.Getenv("TELEGRAM"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalln("failed to init telegram: ", err)
	}

	telegram.Handle("/send", func(ctx tele.Context) error {
		fmt.Println(ctx.Chat().ID)
		ch, err := discord.Channel("1051934380924338186")
		if err != nil {
			return err
		}

		msg := ctx.Text()[5:]
		if isEmpty(msg) {
			return nil
		}

		content := fmt.Sprintf("%s - Telegram\n%s", ctx.Message().Sender.Username, msg)

		_, err = discord.ChannelMessageSend(ch.ID, content)
		return err
	})

	fmt.Println("tg is running")
	telegram.Start()
}

func initDiscord() {
	discord, err = dgo.New("Bot " + os.Getenv("DISCORD"))
	channel := dsChannel

	discord.Identify.Intents = 513

	discord.AddHandler(func(s *dgo.Session, m *dgo.MessageCreate) {
		if m.ChannelID != channel {
			return
		}

		if !strings.HasPrefix(m.Content, ".tg") {
			return
		}

		msg := m.Content[3:]
		content := fmt.Sprintf("%s - Discord\n%s", m.Author.Username, msg)
		chat, err := telegram.ChatByID(tgChannel)
		if err != nil {
			return
		}

		telegram.Send(chat, content)
	})

	discord.AddHandler(func(s *dgo.Session, r *dgo.Ready) {
		fmt.Println("bot is ready")
	})

	err = discord.Open()
	if err != nil {
		log.Fatalln("Failed to start discord: ", err)
	}
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalln("failed to load .env", err)
	}

	dsChannel = os.Getenv("DISCORD_CHANNEL")
	tgc, err := strconv.Atoi(os.Getenv("TELEGRAM_CHANNEL"))
	if err != nil {
		panic(err)
	}
	tgChannel = int64(tgc)

	go initTelegram()
	go initDiscord()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
