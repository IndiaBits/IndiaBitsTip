package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"github.com/jinzhu/gorm"
)

func (message *Message) HelpHandler() string {
	help_text := ""
	help_text = "Welcome to the lightning tipbot. Start by sending /register to register an account and start using the bot."
	help_text += "\n\nCommands:"
	help_text += "\n\n\\register: Register an account. Make sure you have a telegram username. Your funds are associated with your telegram username so withdraw all your funds if you decide to change your telegram username"
	help_text += "\n\n\\address <amount>: Get your deposit address"
	help_text += "\n\n\\withdraw <pay_req>: Withdraw your coins over lightning network(10 SAT Fees). DO NOT USE THE SAME PAYMENT REQUEST TWICE"
	help_text += "\n\n\\balance: To check your balance"
	help_text += "\n\n\\tip <amount>: Reply to any message with tip <amount> and the sender of the message will be tipped with the specified amount"

	UpdateResponse(help_text, message)
	return help_text
}

func (message *Message) RegisterHandler(tmessage *telebot.Message) string {
	if tmessage.Sender.Username == "" {
		return "You need to have a username to use this bot."
	}

	user, err := findUser(tmessage.Sender.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Println(err)
		response := "Something went wrong."
		return response
	}

	if user.Id != 0 {
		return "You are already registered"
	}

	err = user.Register()
	if err != nil {
		log.Println(err.Error())
		return "Something went wrong."
	} else {
		return "Successfully registered"
	}
}

func (message *Message) GetAddressHandler(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "You must be registered to use the bot"
		} else {
			log.Println(err)
			return "Something went wrong."
		}
	}

	if user.Address == "" {
		user.Address, err = generateAddress(user.Username)
		user.Update()
		if err != nil {
			log.Println(err)
			return "Something went wrong."
		}
	}

	return "Your address is " + user.Address
}

func findUser(username string) (*User, error) {
	user := &User{
		Username: username,
	}
	user, err := user.Find()
	return user, err
}