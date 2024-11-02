package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/digilolnet/client3xui"
	"github.com/joho/godotenv"
	"github.com/tucnak/telebot"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	client := client3xui.New(client3xui.Config{
		Url:      os.Getenv("CLIENT_URL"),
		Username: os.Getenv("CLIENT_USERNAME"),
		Password: os.Getenv("CLIENT_PASSWORD"),
	})

	// Get server status
	status, err := client.ServerStatus(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(status.Msg)

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
	}

	app := NewBotHandler(bot, client)

	bot.Handle("/start", app.start)
	bot.Handle("/subscribe", app.subscribe)
	bot.Handle("/apps", app.showApps)
	bot.Handle(telebot.OnCallback, app.handleCallback)

	// Отображение главного меню при запуске бота или при любых других запросах
	bot.Handle(telebot.OnText, app.showMainMenu)

	bot.Start()

}
