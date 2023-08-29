package main

import (
	"net"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pion/stun"
)

const botToken = "bot_token_here"
const author = "admin_tg_login_here"
const portToPunch = 31338
const stunServer = "stun.sipnet.ru:3478"

func main() {
	ip := "unknown"

	go func() {
		var err error = nil
		for {
			func() {
				var remoteEP *net.UDPAddr
				remoteEP, err = net.ResolveUDPAddr("udp", stunServer)
				if err != nil {
					ip = err.Error()
					return
				}
				localEP := net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: portToPunch}

				var conn *net.UDPConn
				conn, err = net.DialUDP("udp", &localEP, remoteEP)
				if err != nil {
					ip = err.Error()
					return
				}
				defer conn.Close()
				var c *stun.Client
				c, err = stun.NewClient(conn)
				if err != nil {
					ip = err.Error()
					return
				}
				defer c.Close()
				message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
				for {
					ip = "pending"
					c.Do(message, func(res stun.Event) {
						if res.Error != nil {
							err = res.Error
							return
						}
						var xorAddr stun.XORMappedAddress
						if er := xorAddr.GetFrom(res.Message); er != nil {
							err = er
							return
						}
						ip = xorAddr.String()
					})
					if err != nil {
						ip = err.Error()
						break
					}
					time.Sleep(time.Minute)
				}
			}()
			time.Sleep(time.Second * 15)
		}
	}()

	for ; ; time.Sleep(time.Second * 15) {
		bot, err := tg.NewBotAPI(botToken)
		if err != nil {
			continue
		}
		u := tg.NewUpdate(0)
		u.Timeout = 60

		updates, err := bot.GetUpdatesChan(u)
		if err != nil {
			continue
		}

		for update := range updates {
			if update.Message == nil {
				continue
			}
			if update.Message.From.UserName != author {
				continue
			}
			if update.Message.Text == "/vpn" {
				msg := tg.NewMessage(update.Message.Chat.ID, ip)
				bot.Send(msg)
			}
		}
	}
}
