package main

import (
	"errors"
	"fmt"
	_ "fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"subinfobot/handler"
	"subinfobot/utils"
	"time"
)

type Subinfo struct {
	Link       string
	ExpireTime string
	TimeRemain string
	Upload     string
	Download   string
	Used       string
	Total      string
	Expired    int //0:not Expired,1:Expired,2:unknown
	Available  int //0:Available,1:unavailable,2:unknown
	DataRemain string
}

func getSinf(link string) (error, Subinfo) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", link, nil)
	req.Header.Add("User-Agent", "ClashforWindows/0.19.21")
	if err != nil {
		return err, Subinfo{}
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err, Subinfo{}
	}
	if res.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("獲取失敗，傑哥的房間返回了代碼%s", strconv.Itoa(res.StatusCode))), Subinfo{}
	}
	if sinfo := res.Header["Subscription-Userinfo"]; sinfo == nil {
		return errors.New("未獲取到訂閱詳細信息，傑哥覺得該訂閱可能已經到期或者已被刪除"), Subinfo{}
	} else {
		sinf := Subinfo{Link: link}
		sinfmap := make(map[string]int64)
		parseExp := regexp.MustCompile("[A-Za-z]+=[0-9]+")
		sslice := parseExp.FindAllStringSubmatch(sinfo[0], -1)
		for _, val := range sslice {
			kvslice := strings.Split(val[0], "=")
			if len(kvslice) == 2 {
				i, err := strconv.ParseInt(kvslice[1], 10, 64)
				if err == nil {
					sinfmap[kvslice[0]] = i
				}
			}
		}
		if upload, oku := sinfmap["upload"]; oku {
			sinf.Upload = utils.FormatFileSize(upload)
		} else {
			sinf.Upload = "看起來傑哥不懂"
		}
		if download, okd := sinfmap["download"]; okd {
			sinf.Download = utils.FormatFileSize(download)
		} else {
			sinf.Download = "看起來傑哥不懂"
		}
		if total, okt := sinfmap["total"]; okt {
			sinf.Total = utils.FormatFileSize(total)
			down, oka := sinfmap["download"]
			up, okb := sinfmap["upload"]
			if (oka) && (okb) {
				sinf.Used = utils.FormatFileSize(up + down)
				remain := total - (up + down)
				if remain >= 0 {
					if remain > 0 {
						sinf.Available = 0
						sinf.DataRemain = utils.FormatFileSize(remain)
					} else {
						sinf.Available = 1
						sinf.DataRemain = utils.FormatFileSize(remain)
					}
				} else {
					sinf.Available = 1
					sinf.DataRemain = "逾量" + utils.FormatFileSize(int64(math.Abs(float64(remain))))
				}
			} else {
				sinf.Available = 2
				sinf.DataRemain = "看起來傑哥不懂"
			}
		} else {
			sinf.Available = 2
			sinf.Total = "看起來傑哥不懂"
		}
		if exp, oke := sinfmap["expire"]; oke {
			//get expire time and remain time
			timeStamp := time.Now().Unix()
			timeExp := time.Unix(exp, 0)
			sinf.ExpireTime = timeExp.String()
			if timeStamp >= exp {
				sinf.Expired = 1
				sinf.Available = 1
				remain := timeExp.Sub(time.Now())
				if remain.Hours() > 24 {
					sinf.TimeRemain = "逾期<code>" + strconv.Itoa(int(math.Floor(remain.Hours()/24))) + "天" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Hours()))%24)))) + "小时" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Minutes()))%60)))) + "分" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Seconds()))%60)))) + "秒" + "</code>"
				} else if remain.Minutes() > 60 {
					sinf.TimeRemain = "逾期<code>" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Hours()))%24)))) + "小时" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Minutes()))%60)))) + "分" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Seconds()))%60)))) + "秒" + "</code>"
				} else if remain.Seconds() > 60 {
					sinf.TimeRemain = "逾期<code>" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Minutes()))%60)))) + "分" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Seconds()))%60)))) + "秒" + "</code>"
				} else {
					sinf.TimeRemain = "逾期<code>" + strconv.Itoa(int(math.Floor(remain.Seconds()))) + "秒" + "</code>"
				}
			} else {
				sinf.Expired = 0
				remain := timeExp.Sub(time.Now())
				if remain.Hours() > 24 {
					sinf.TimeRemain = "傑哥說距離到期還有<code>" + strconv.Itoa(int(math.Floor(remain.Hours()/24))) + "天" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Hours()))%24)))) + "小时" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Minutes()))%60)))) + "分" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Seconds()))%60)))) + "秒" + "</code>"
				} else if remain.Minutes() > 60 {
					sinf.TimeRemain = "傑哥說距離到期還有<code>" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Hours()))%24)))) + "小时" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Minutes()))%60)))) + "分" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Seconds()))%60)))) + "秒" + "</code>"
				} else if remain.Seconds() > 60 {
					sinf.TimeRemain = "傑哥說距離到期還有<code>" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Minutes()))%60)))) + "分" + strconv.Itoa(int(math.Floor(float64(int(math.Floor(remain.Seconds()))%60)))) + "秒" + "</code>"
				} else {
					sinf.TimeRemain = "傑哥說距離到期還有<code>" + strconv.Itoa(int(math.Floor(remain.Seconds()))) + "秒" + "</code>"
				}
			}
		} else {
			sinf.ExpireTime = "看起來傑哥不懂"
			sinf.TimeRemain = "傑哥覺得可能是無限時長訂閱或者傑哥的房間抽抽了呢"
		}
		return nil, sinf
	}
}
func subInfoMsg(link string, update *tgbotapi.Update, bot *tgbotapi.BotAPI, msg *tgbotapi.MessageConfig) {
	msg.Text = "🏗️讓我看看..."
	msg.ReplyToMessageID = update.Message.MessageID
	sres, err := handler.SendMsg(bot, msg)
	handler.HandleError(err)
	if err == nil {
		err, sinf := getSinf(link)
		handler.HandleError(err)
		if err != nil {
			_, err := handler.EditMsg(fmt.Sprintf("<strong>❌獲取失敗</strong>\n\n傑哥獲取訂閱<code>%s</code>時發生錯誤:\n<code>%s</code>", sinf.Link, err), "html", bot, sres)
			handler.HandleError(err)
			if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
				_, _ = handler.DelMsgWithTimeOut(10*time.Second, bot, sres)
			}
		} else {
			var resMsg string
			if sinf.Expired == 0 && sinf.Available == 0 {
				resMsg = "✅傑哥說該訂閱有效"
			}
			if sinf.Expired == 2 || sinf.Available == 2 {
				resMsg = "❓傑哥不知道該訂閱狀態"
			}
			if sinf.Expired == 1 || sinf.Available == 1 {
				resMsg = "❌傑哥說該訂閱不可用"
			}
			_, err = handler.EditMsg(fmt.Sprintf("<strong>%s</strong>\n<strong>訂閱鏈接:</strong><code>%s</code>\n<strong>總流量:</strong><code>%s</code>\n<strong>剩餘流量:</strong><code>%s</code>\n<strong>已上傳:</strong><code>%s</code>\n<strong>已下載:</strong><code>%s</code>\n<strong>該訂閱將於<code>%s</code>過期,%s</strong>", resMsg, sinf.Link, sinf.Total, sinf.DataRemain, sinf.Upload, sinf.Download, sinf.ExpireTime, sinf.TimeRemain), "html", bot, sres)
			handler.HandleError(err)
			if update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup" {
				_, _ = handler.DelMsgWithTimeOut(10*time.Second, bot, sres)
			}
		}
	}
}
