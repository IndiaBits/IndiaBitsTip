package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"time"
	"log"
	"github.com/jinzhu/gorm"
)

type Message struct {
	Id int
	UserId int
	Username string
	Message string
	Response string
	ReceivedAt time.Time
	RepliedAt time.Time
}

func (message *Message) create() error {
	err := DB.Create(message)
	if err.Error != nil {
		log.Printf("Error occured while creating message: %v",err)
	}
	return err.Error
}


func (message *Message) update() error {
	err := DB.Save(message)
	if err.Error != nil {
		log.Printf("Error occured while updating message: %v",err)
	}
	return err.Error
}

func NewMessage(tmessage *telebot.Message) (*Message, error) {
	message, err := storeMessage(tmessage)
	if err != nil {
		err := err
		return nil, err
	}
	return message, nil
}

func storeMessage(tmessage *telebot.Message) (*Message, error) {
	message := Message{}
	message.Username = tmessage.Sender.Username

	user := &User{
		Username: message.Username,
	}
	user, err :=  user.First()
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil,err
		}
	}

	message.UserId = user.Id
	message.Message = tmessage.Text
	message.ReceivedAt = time.Now()

	err = message.create()
	if err != nil {
		return nil,err
	}
	return &message,nil
}


func UpdateResponse(response string, message Message) {
	message.Response = response
	message.RepliedAt = time.Now()
	message.update()
}