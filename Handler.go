package main

import (
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"os"
	"sync"
)

type Tip struct {
	Id int
	FromId int
	ToId int
	MessageId int
	Amount float64
}

func (tip *Tip) Create() error {
	DB.Create(tip)
	return DB.Error
}

func (message *Message) HelpHandler() string {
	help_text := ""
	help_text = "Welcome to the lightning tipbot. Start by sending /register to register an account and start using the bot."
	help_text += "\n\nCommands:"
	help_text += "\n\n\\register: Register an account. Make sure you have a telegram username. Your funds are associated with your telegram username so withdraw all your funds if you decide to change your telegram username"
	help_text += "\n\n\\address <amount>: Get your deposit address"
	help_text += "\n\n\\withdraw <pay_req>: Withdraw your coins over lightning network(10 SAT Fees). DO NOT USE THE SAME PAYMENT REQUEST TWICE"
	help_text += "\n\n\\balance: To check your balance"
	help_text += "\n\n\\tip <amount>: Reply to any message with tip <amount> and the sender of the message will be tipped with the specified amount"
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
		user.Address, err = generateAddress()
		user.Update()
		if err != nil {
			log.Println(err)
			return "Something went wrong."
		}
	}

	return "Your address is " + user.Address
}

func (message *Message) BalanceHandler(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "You must be registered to use the bot"
		} else {
			log.Println(err)
			return "Something went wrong."
		}
	}
	return strconv.FormatFloat(user.Balance,'f', 8, 64) + " BTC"
}

func (message *Message) WithdrawHandler(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "You must be registered to use the bot"
		} else {
			log.Println(err)
			return "Something went wrong."
		}
	}

	data := strings.Split(tmessage.Payload," ")
	if len(data) < 2 {
		return "Please provide both address and amount"
	}

	address, err := getAddress(data[0])
	if err != nil {
		return "Please enter a valid bitcoin address"
	}

	amount, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		return "Please enter a valid amount"
	}

	withdrawal_fee, err := strconv.ParseFloat(os.Getenv("WITHDRAWAL_FEE"),  64)
	if err != nil {
		log.Println(err)
		return "Something went wrong"
	}

	if user.Balance < (amount + withdrawal_fee) {
		return "Insufficient balance"
	}

	minimum_withdrawal_amount, err := strconv.ParseFloat(os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW"), 64)
	if err != nil {
		log.Println(err)
		return "Something went wrong"
	}

	if amount < minimum_withdrawal_amount {
		return "Amount is less than the minimum amount required to withdraw " + os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW")
	}

	withdrawal_amount, err := btcutil.NewAmount(amount)
	if err != nil {
		log.Println(err)
		return "Please enter a valid amount"
	}

	/*fees, err := strconv.ParseFloat(data[2], 64)
	if err != nil {
		return "Please enter a valid fees"
	}

	fee_amount, err := btcutil.NewAmount(fees)
	if err != nil {
		log.Println(err)
		return "Please enter a valid fees"
	}*/

	transaction := Transaction{
		UserId: user.Id,
		Type: 2,
		Amount: amount,
		MessageId: message.Id,
		Confirmed:0,
		Address:address.String(),
	}
	if err := transaction.Create(); err != nil {
		log.Println(err)
		return "Something went wrong"
	}

	user.Balance = user.Balance - amount - withdrawal_fee
	err = user.Update()
	if err != nil {
		log.Println(err)
		return "Something went wrong"
	}

	/*err = Client.SetTxFee(fee_amount)
	if err != nil {
		log.Println(err)
		return "Something went wrong"
	}*/

	tx, err := Client.SendToAddress(address, withdrawal_amount)
	if err != nil {
		user.Balance = user.Balance + amount + withdrawal_fee
		user.Update()
		log.Println(err)
		return "Something went wrong"
	}

	transaction.TransactionId = tx.String()
	err = transaction.Update()
	if err != nil {
		log.Println(err)
	}

	return tx.String()
}

func (message *Message) TipHandler(tmessage *telebot.Message) string {

	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "You must be registered to use the bot"
		} else {
			log.Println(err)
			return "Something went wrong."
		}
	}

	if tmessage.ReplyTo == nil {
		return "You need to reply to the message you want to tip for"
	}

	otheruser, err := findUser(tmessage.ReplyTo.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "The user must be registered to receive tips"
		} else {
			log.Println(err)
			return "Something went wrong."
		}
	}

	if user.Username == otheruser.Username {
		return "You cannot tip yourself"
	}

	data := strings.Split(tmessage.Text," ")
	if len(data) != 2 {
		return "Incorrect format"
	}

	amount, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		return "Please enter correct amount"
	}

	if(BalanceMutexes[user.Username] == nil) {
		BalanceMutexes[user.Username] = &sync.Mutex{}
	}

	if(BalanceMutexes[otheruser.Username] == nil) {
		BalanceMutexes[otheruser.Username] = &sync.Mutex{}
	}

	BalanceMutexes[user.Username].Lock()
	BalanceMutexes[otheruser.Username].Lock()

	if user.Balance < amount {
		BalanceMutexes[user.Username].Unlock()
		BalanceMutexes[otheruser.Username].Unlock()
		return "Insufficient balance"
	}

	user.Balance = user.Balance - amount
	user.Update()

	otheruser.Balance = otheruser.Balance + amount
	otheruser.Update()

	tip := Tip{
		FromId: user.Id,
		ToId: otheruser.Id,
		Amount:amount,
		MessageId: message.Id,
	}

	err = tip.Create()
	if err != nil {
		log.Println(err)

		user.Balance = user.Balance + amount
		user.Update()

		otheruser.Balance = otheruser.Balance - amount
		otheruser.Update()

		BalanceMutexes[user.Username].Unlock()
		BalanceMutexes[otheruser.Username].Unlock()

		return "Something went wrong"
	}

	BalanceMutexes[user.Username].Unlock()
	BalanceMutexes[otheruser.Username].Unlock()

	return "@" + user.Username + " tipped " + data[1] + " btc to " + "@" + otheruser.Username
}

func findUser(username string) (*User, error) {
	user := &User{
		Username: username,
	}
	user, err := user.First()
	return user, err
}

func findUserByAddress(address string) (*User, error) {
	user := &User{
		Address: address,
	}
	user, err := user.First()
	return user, err
}

func getAddress(addr string) (btcutil.Address, error) {
	address, err := btcutil.DecodeAddress(addr, &chaincfg.Params{})
	if err != nil {
		return nil, err
	}
	return address, nil
}