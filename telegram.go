package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"strings"
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
	})

	bot.Handle("/register",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.RegisterHandler(tmessage)
		bot.Send(tmessage.Sender, response)
	})

	bot.Handle("/address",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.GetAddressHandler(tmessage)
		bot.Send(tmessage.Sender, response)
	})

	bot.Handle("/balance",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.HelpHandler()
		bot.Send(tmessage.Sender, response)
	})

	bot.Handle("/withdraw",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.HelpHandler()
		bot.Send(tmessage.Sender, response)
	})

	bot.Handle("/withdraw",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.HelpHandler()
		bot.Send(tmessage.Sender, response)
	})

	bot.Handle(telebot.OnText, func(tmessage *telebot.Message) {
		if strings.Index(strings.ToUpper(tmessage.Text),"TIP") == 0 {
			message, err := NewMessage(tmessage)
			if err != nil {
				response := err.Error()
				bot.Send(tmessage.Sender, response)
			}
			response := message.HelpHandler()
			bot.Send(tmessage.Sender, response)
		}
	})



}