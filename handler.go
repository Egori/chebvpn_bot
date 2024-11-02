package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digilolnet/client3xui"
	"github.com/tucnak/telebot"
)

type BotHandler struct {
	bot    *telebot.Bot
	client *client3xui.Client
}

func NewBotHandler(bot *telebot.Bot, client *client3xui.Client) *BotHandler {
	return &BotHandler{
		bot:    bot,
		client: client,
	}
}

func (h *BotHandler) start(m *telebot.Message) {
	h.bot.Send(m.Sender, "Добро пожаловать! Нажмите /subscribe, чтобы получить подписку на VPN. \nНажмите /apps, чтобы посмотреть список приложений для подключения.")
}

func (h *BotHandler) subscribe(m *telebot.Message) {
	link, err := h.addUserTo3xUI(m)
	if err != nil {
		h.bot.Send(m.Sender, "Ошибка при добавлении пользователя.")
	} else {
		h.bot.Send(m.Sender, "Ваша подписка готова: "+link)
	}
}

func (h *BotHandler) addUserTo3xUI(m *telebot.Message) (string, error) {
	// Prepare the client
	userID := m.Sender.ID
	userName := m.Sender.Username
	var subID string
	// Fetch inbounds from 3x-ui
	inboundsResp, err := h.client.GetInbounds(context.Background())
	if err != nil {
		h.bot.Send(m.Sender, "Failed to get inbounds: "+err.Error())
		return "", err
	}

	var targetInbound *client3xui.Inbound
	for _, inbound := range inboundsResp.Obj {
		if strings.HasPrefix(inbound.Remark, "chebvpn") && inbound.Protocol == "vless" {
			targetInbound = &inbound
			break
		}
	}

	if targetInbound == nil {
		h.bot.Send(m.Sender, "inbound with tag 'chebvpn' not found")
		return "", fmt.Errorf("inbound with tag 'chebvpn' not found")
	}

	var settings client3xui.VlessSetting
	err = json.Unmarshal([]byte(targetInbound.Settings), &settings)
	if err != nil {
		h.bot.Send(m.Sender, "Failed to get vless settings: "+err.Error())
		return "", err
	}

	for _, c := range settings.Clients {
		if c.ID == userName {
			subID = c.SubId
			break
		}
	}

	if subID == "" {
		client, err := h.addClientToInbound(userID, userName, targetInbound.ID)
		if err != nil {
			h.bot.Send(m.Sender, "Failed to add client: "+err.Error())
			return "", err
		}
		subID = client.SubID
	}

	subscriptionLink := fmt.Sprintf("https://proxy-m.duckdns.org:57010/sub216/%s", subID)
	return subscriptionLink, nil
}

func (h *BotHandler) addClientToInbound(userID int, userName string, inboundID int) (client3xui.XrayClient, error) {
	client := client3xui.XrayClient{
		ID:      userName,
		AlterID: uint(userID),
		Email:   userName + "-" + "i" + strconv.Itoa(inboundID),
		Enable:  true,
		SubID:   userName,
		LimitIP: uint(1),
	}

	// Add the client to the selected inbound
	_, err := h.client.AddClient(context.Background(), uint(inboundID), []client3xui.XrayClient{client})

	return client, err
}

func (h *BotHandler) showApps(m *telebot.Message) {

	inlineButtons := [][]telebot.InlineButton{
		{telebot.InlineButton{Text: "Windows", Data: "Windows"}},
		{telebot.InlineButton{Text: "MacOS", Data: "MacOS"}},
		{telebot.InlineButton{Text: "Linux", Data: "Linux"}},
		{telebot.InlineButton{Text: "Android", Data: "Android"}},
		{telebot.InlineButton{Text: "iOS", Data: "iOS"}},
	}

	h.bot.Send(m.Sender, "Выберите вашу операционную систему:", &telebot.ReplyMarkup{InlineKeyboard: inlineButtons})
}

func (h *BotHandler) handleCallback(c *telebot.Callback) {
	vpnClients, err := LoadVPNClients("apps.txt")
	if err != nil {
		log.Fatal("Не удалось загрузить VPN клиенты:", err)
	}

	var responseText string
	var links []string

	switch c.Data {
	case "Windows":
		responseText = "Вот ссылки на VPN клиенты для Windows"
		links = GetClientsByOS("Windows", vpnClients)
	case "MacOS":
		responseText = "Вот ссылки на VPN клиенты для MacOS"
		links = GetClientsByOS("MacOS", vpnClients)
	case "Linux":
		responseText = "Вот ссылки на VPN клиенты для Linux"
		links = GetClientsByOS("Linux", vpnClients)
	case "Android":
		responseText = "Вот ссылки на VPN клиенты для Android"
		links = GetClientsByOS("Android", vpnClients)
	case "iOS":
		responseText = "Вот ссылки на VPN клиенты для iOS"
		links = GetClientsByOS("iOS", vpnClients)
	default:
		responseText = "Неизвестная операционная система"
	}

	h.bot.Respond(c, &telebot.CallbackResponse{Text: responseText})
	h.bot.Send(c.Sender, strings.Join(links, "\n"))

}

func (h *BotHandler) showMainMenu(m *telebot.Message) {
	// Создаем кнопки для основных команд
	mainMenuButtons := [][]telebot.ReplyButton{
		{telebot.ReplyButton{Text: "/start"}},
		{telebot.ReplyButton{Text: "/subscribe"}},
		{telebot.ReplyButton{Text: "/apps", Action: func(c *telebot.Callback) { h.showApps(c.Message) }}},
	}

	// Отправляем клавиатуру с кнопками пользователю
	replyMarkup := &telebot.ReplyMarkup{
		ReplyKeyboard: mainMenuButtons,
	}

	h.bot.Send(m.Sender, "Выберите команду:", replyMarkup)
}
