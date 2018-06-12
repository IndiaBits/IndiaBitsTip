package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"strings"
	"github.com/jinzhu/gorm"
	"log"
)

func InitTelegramCommands(bot *telebot.Bot) {
	bot.Handle("/help",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.HelpHandler()
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, message)
	})

	bot.Handle("/register",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.RegisterHandler(tmessage)
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, message)
	})

	bot.Handle("/address",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.GetAddressHandler(tmessage)
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, message)
	})

	bot.Handle("/balance",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.BalanceHandler(tmessage)
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, message)
	})

	bot.Handle("/withdraw",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}

		user, err := findUser(tmessage.Sender.Username)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				response := "You must be registered to use the bot"
				bot.Send(tmessage.Sender, response)
				UpdateResponse(response, message)
			} else {
				log.Println(err)
				response := "Something went wrong."
				bot.Send(tmessage.Sender, response)
				UpdateResponse(response, message)
			}
		}

		BalanceMutexes[user.Username].Lock()

		response := message.WithdrawHandler(tmessage)

		BalanceMutexes[user.Username].Unlock()
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, message)
	})

	bot.Handle(telebot.OnText, func(tmessage *telebot.Message) {
		if strings.Index(strings.ToUpper(tmessage.Text),"TIP") == 0 {
			message, err := NewMessage(tmessage)
			if err != nil {
				response := err.Error()
				bot.Send(tmessage.Sender, response)
			}
			response := message.TipHandler(tmessage)
			bot.Send(tmessage.Sender, response)
		}
	})



}