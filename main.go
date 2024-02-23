package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"regexp"
	"strings"
	"subinfobot/handler"
	"time"
)

var (
	version string
	commit  string
	logger  = log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)
)

func main() {
	logger.Printf("Subbot %s start.", version)
	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		logger.Panic(fmt.Sprintf("Connect failed. %s", err))
	}
	bot.Debug = false
	logger.Printf("Connected with name %s.", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			if !update.Message.IsCommand() {
				if update.Message.Chat.IsPrivate() {
					linkReg := regexp.MustCompile("(http|https){0,1}://[^\\x{4e00}-\\x{9fa5}\\n\\r\\s]{3,}")
					if linkReg.MatchString(update.Message.Text) {
						slice := linkReg.FindAllStringSubmatch(update.Message.Text, -1)
						go subInfoMsg(slice[0][0], &update, bot, &msg)
					} else {
						msg.Text = "❌傑哥沒有在你發送的內容中找到任何有效信息哦！"
						msg.ReplyToMessageID = update.Message.MessageID
						_, err := handler.SendMsg(bot, &msg)
						handler.HandleError(err)
					}
				}
			}
			switch update.Message.Command() {
			case "start":
				if update.Message.Chat.IsPrivate() {
					msg.ParseMode = "html"
					msg.Text = "<strong>🌈傑哥覺得你是完全都不懂哦！</strong> \n\n 📖命令列表: \n/start 開始\n/si 獲取訂閱鏈接的詳細信息\n/about 關於\n\n歡迎加入<a href=\"https://t.me/nodpai\">@nodpai</a>來改善此bot!\n"
					_, err := handler.SendMsg(bot, &msg)
					handler.HandleError(err)
				}
			case "about":
				msg.ParseMode = "html"
				msg.Text = fmt.Sprintf("<strong>Subinfo Bot</strong>\n\nSubinfo Bot是一款由Golang編寫的開源輕量訂閱查詢Bot。體積小巧，無需任何第三方運行時，即點即用。 \n\n<strong>Github:<a href=\"https://github.com/ThekingMX1998/subinfobot\">https://github.com/ThekingMX1998/subinfobot</a></strong>")
				_, err := handler.SendMsg(bot, &msg)
				handler.HandleError(err)
			case "si":
				msg.ParseMode = "html"
				commandSlice := strings.Split(update.Message.Text, " ")
				if len(commandSlice) < 2 {
					msg.Text = "❌傑哥覺得你的參數有問題，請檢查後再試"
					msg.ReplyToMessageID = update.Message.MessageID
					res, err := handler.SendMsg(bot, &msg)
					handler.HandleError(err)
					if err == nil {
						if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
							_, _ = handler.DelMsgWithTimeOut(10*time.Second, bot, res)
						}
					}
				} else if strings.HasPrefix(commandSlice[1], "http://") || strings.HasPrefix(commandSlice[1], "https://") {
					go subInfoMsg(commandSlice[1], &update, bot, &msg)
				} else {
					msg.Text = "❌傑哥覺得你的鏈接有問題，請檢查後再試"
					msg.ReplyToMessageID = update.Message.MessageID
					res, err := handler.SendMsg(bot, &msg)
					handler.HandleError(err)
					if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
						_, _ = handler.DelMsgWithTimeOut(10*time.Second, bot, res)
					}
				}
			default:
			}
		}
	}
}
