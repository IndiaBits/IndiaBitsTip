package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"strings"
	"os"
	"math/rand"
	"log"
)

func InitTelegramCommands(bot *telebot.Bot) {

	confirmBtn := telebot.InlineButton{
		Unique: "confirm_withdrawal",
		Text: "Confirm",
	}

	cancelBtn := telebot.InlineButton{
		Unique: "cancel_withdrawal",
		Text: "Cancel",
	}

	bot.Handle("/help",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.HelpHandler()
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, *message)
	})

	bot.Handle("/register",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.RegisterHandler(tmessage)
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, *message)
	})

	bot.Handle("/address",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.GetAddressHandler(tmessage)
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, *message)
	})

	bot.Handle("/balance",func(tmessage *telebot.Message) {
		message, err := NewMessage(tmessage)
		if err != nil {
			response := err.Error()
			bot.Send(tmessage.Sender, response)
		}
		response := message.BalanceHandler(tmessage)
		bot.Send(tmessage.Sender, response)
		UpdateResponse(response, *message)
	})

	bot.Handle("/withdraw",func(tmessage *telebot.Message) {
		validation_error := withdrawalValidations(tmessage)
		if validation_error != "ok" {
			bot.Send(tmessage.Sender, validation_error)
			return
		}

		random_string := RandomString(32)
		confirmBtn.Data = random_string
		cancelBtn.Data = random_string
		confirmKeys := [][]telebot.InlineButton{
			{confirmBtn},
			{cancelBtn},
		}
		withdrawal_confirmations[random_string] = tmessage

		response := "Each withdrawal currently costs " + os.Getenv("WITHDRAWAL_FEE") + " btc. Are you sure you want to withdraw?"
		_, err := bot.Send(tmessage.Sender, response, &telebot.ReplyMarkup{
			InlineKeyboard: confirmKeys,
		})
		if err != nil {
			log.Print(err)
		}
	})

	bot.Handle(telebot.OnText, func(tmessage *telebot.Message) {
		if strings.Index(strings.ToUpper(tmessage.Text),"TIP") == 0 {
			message, err := NewMessage(tmessage)
			if err != nil {
				response := err.Error()
				bot.Send(tmessage.Sender, response)
			}
			response := message.TipHandler(tmessage)
			bot.Send(tmessage.Chat, response)
			UpdateResponse(response, *message)
		}
	})

	bot.Handle(&confirmBtn, func(c *telebot.Callback) {
		message, ok := withdrawal_confirmations[c.Data]
		if !ok {
			bot.Send(c.Sender,"Withdrawal already cancelled or already processed")
			return
		}

		ProcessWithdrawal(bot, message)
		delete(withdrawal_confirmations, c.Data)
	})

	bot.Handle(&cancelBtn, func(c *telebot.Callback) {
		_, ok := withdrawal_confirmations[c.Data]
		if !ok {
			bot.Send(c.Sender,"Withdrawal already cancelled or already processed")
			return
		}

		delete(withdrawal_confirmations, c.Data)
		bot.Send(c.Sender,"Withdrawal cancelled")
	})

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(n int) string {
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}
	return string(output)
}