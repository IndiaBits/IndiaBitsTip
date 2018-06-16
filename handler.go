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
	"github.com/funyug/bitcoin-tipbot/emoji"
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
	help_text = emoji.Emoji("white_check_mark") + " Welcome to the Indiabits tipbot. Start by sending /register to create an account and start using the bot."
	help_text += "\n\nCommands:"
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " /register: Register an account. Make sure you have a telegram username. Your funds are associated with your telegram username so withdraw all your funds if you decide to change your telegram username"
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " /address: Get your bitcoin deposit address"
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " /withdraw <address> <amount>: Withdraw coins to a bitcoin address. Withdrawal fee: " + os.Getenv("WITHDRAWAL_FEE") + " BTC. Minimum amount: " + os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW")
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " /balance: Check your balance"
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " /tip <amount>: Reply to any message with tip <amount> and the sender of the message will be tipped with the specified amount"
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " /help for this help menu"
	help_text += "\n\n" + emoji.Emoji("information_source") + " Tips are offchain hence no fees for tipping users and database is maintained by one of the IndiaBits Admin."
	help_text += "\n" + emoji.Emoji("heavy_minus_sign") + " Supports tip amount upto 8 decimal amount/points"
	help_text += "\n" + emoji.Emoji("warning") + " Its not recommended to use tipbot as a wallet or to exchange large amounts."
	return help_text
}

func (message *Message) RegisterHandler(tmessage *telebot.Message) string {
	if tmessage.Sender.Username == "" {
		return emoji.Emoji("information_source") + " You need to have a username to use this bot."
	}

	user, err := findUser(tmessage.Sender.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Println(err)
		response := emoji.Emoji("no_entry_sign") + " Something went wrong."
		return response
	}

	if user.Id != 0 {
		return emoji.Emoji("information_source") + " @" + user.Username + " is already registered!"
	}

	err = user.Register()
	if err != nil {
		log.Println(err.Error())
		return emoji.Emoji("no_entry_sign") + " Something went wrong."
	} else {
		return emoji.Emoji("ballot_box_with_check") + " Successfully registered"
	}
}

func (message *Message) GetAddressHandler(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return emoji.Emoji("information_source") + " You must be registered to use the bot"
		} else {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	if user.Address == "" {
		user.Address, err = generateAddress()
		user.Update()
		if err != nil {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	return emoji.Emoji("information_source") + " Your address is " + user.Address
}

func (message *Message) BalanceHandler(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return emoji.Emoji("information_source")+" You must be registered to use the bot"
		} else {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	transaction := Transaction{
		Type:1,
		Confirmed:2,
		UserId:user.Id,
	}

	unconfirmed_balance := 0.00

	transactions,err := transaction.Find()
	for _, transaction = range transactions {
		unconfirmed_balance += transaction.Amount
	}

	confirmed_balance_text := emoji.Emoji("ballot_box_with_check") + " Balance: " + strconv.FormatFloat(user.Balance,'f', 8, 64) + " BTC\n"
	unconfirmed_balance_text := emoji.Emoji("information_source") + " Pending: " + strconv.FormatFloat(unconfirmed_balance, 'f', 8, 64) + " BTC"
	if len(transactions) < 1 {
		unconfirmed_balance_text = ""
	}
	return confirmed_balance_text + unconfirmed_balance_text
}

func (message *Message) WithdrawHandler(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return emoji.Emoji("information_source") + " You must be registered to use the bot"
		} else {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	data := strings.Split(tmessage.Payload," ")
	if len(data) < 2 {
		return emoji.Emoji("information_source") + " Correct format: /withdraw address amount"
	}

	address, err := getAddress(data[0])
	if err != nil {
		return emoji.Emoji("information_source") + " Please enter a valid bitcoin address"
	}

	var amount float64

	if data[1] != "all" {
		amount, err = strconv.ParseFloat(data[1], 64)
		if err != nil {
			return emoji.Emoji("information_source") + " Please enter a valid amount"
		}
	} else {
		amount = user.Balance
	}

	minimum_withdrawal_amount, err := strconv.ParseFloat(os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW"), 64)
	if err != nil {
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	if amount < minimum_withdrawal_amount {
		return emoji.Emoji("information_source") + " Amount is less than the minimum amount required to withdraw " + os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW")
	}

	withdrawal_fee, err := strconv.ParseFloat(os.Getenv("WITHDRAWAL_FEE"),  64)
	if err != nil {
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	if user.Balance < amount  {
		return emoji.Emoji("no_entry_sign") + " Insufficient balance!"
	}

	withdrawal_amount, err := btcutil.NewAmount((amount - withdrawal_fee))
	if err != nil {
		log.Println(err)
		return emoji.Emoji("information_source") + " Please enter a valid amount"
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
		Confirmed:1,
		Address:address.String(),
	}
	if err := transaction.Create(); err != nil {
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	user.Balance = user.Balance - amount
	err = user.Update()
	if err != nil {
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	/*err = Client.SetTxFee(fee_amount)
	if err != nil {
		log.Println(err)
		return "Something went wrong"
	}*/

	tx, err := Client.SendToAddress(address, withdrawal_amount)
	if err != nil {
		user.Balance = user.Balance + amount
		user.Update()
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	transaction.TransactionId = tx.String()
	err = transaction.Update()
	if err != nil {
		log.Println(err)
	}

	return emoji.Emoji("ballot_box_with_check") + " Sent with tx id: " + tx.String()
}

func (message *Message) TipHandler(tmessage *telebot.Message) string {

	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return emoji.Emoji("information_source") + " You must be registered to use the bot"
		} else {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	if tmessage.ReplyTo == nil {
		return emoji.Emoji("information_source") + " You need to reply to the message you want to tip for"
	}

	otheruser, err := findUser(tmessage.ReplyTo.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return emoji.Emoji("information_source") + " The user must be registered to receive tips"
		} else {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	if user.Username == otheruser.Username {
		return emoji.Emoji("information_source") + " You cannot tip yourself"
	}

	data := strings.Split(tmessage.Text," ")
	if len(data) < 2 {
		return emoji.Emoji("information_source") + " Correct format : tip amount reason(optional)"
	}

	var amount float64

	if data[1] != "all" {
		amount, err = strconv.ParseFloat(data[1], 64)
		if err != nil {
			return emoji.Emoji("information_source") + " Correct format : tip amount reason(optional)"
		}

		if amount <= 0 {
			return emoji.Emoji("no_entry_sign") + " Cannot tip 0 or negative amount"
		}
	} else {
		amount = user.Balance
		if user.Balance <= 0 {
			return emoji.Emoji("no_entry_sign") + " No balance! Please deposit to tip!"
		}
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
		return emoji.Emoji("no_entry_sign") + " Insufficient balance!"
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

		return emoji.Emoji("information_source") + " Something went wrong"
	}

	BalanceMutexes[user.Username].Unlock()
	BalanceMutexes[otheruser.Username].Unlock()

	amount_text := strconv.FormatFloat(amount, 'f',8,64)

	return emoji.Emoji("ballot_box_with_check") + " @" + user.Username + " tipped " + amount_text + " btc to " + "@" + otheruser.Username
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

func withdrawalValidations(tmessage *telebot.Message) string {
	user, err := findUser(tmessage.Sender.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return emoji.Emoji("information_source") + " You must be registered to use the bot"
		} else {
			log.Println(err)
			return emoji.Emoji("no_entry_sign") + " Something went wrong."
		}
	}

	data := strings.Split(tmessage.Payload," ")
	if len(data) < 2 {
		return emoji.Emoji("information_source") + " Correct format: /withdraw address amount"
	}

	_ , err = getAddress(data[0])
	if err != nil {
		return emoji.Emoji("information_source") + " Please enter a valid bitcoin address"
	}

	var amount float64

	if data[1] != "all" {
		amount, err = strconv.ParseFloat(data[1], 64)
		if err != nil {
			return emoji.Emoji("information_source") + " Please enter a valid amount"
		}
	} else {
		amount = user.Balance
	}

	minimum_withdrawal_amount, err := strconv.ParseFloat(os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW"), 64)
	if err != nil {
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	if amount < minimum_withdrawal_amount {
		return emoji.Emoji("information_source") + " Amount is less than the minimum amount required to withdraw " + os.Getenv("MINIMUM_AMOUNT_TO_WITHDRAW")
	}

	withdrawal_fee, err := strconv.ParseFloat(os.Getenv("WITHDRAWAL_FEE"),  64)
	if err != nil {
		log.Println(err)
		return emoji.Emoji("no_entry_sign") + " Something went wrong"
	}

	if user.Balance < amount {
		return emoji.Emoji("no_entry_sign") + " Insufficient balance!"
	}

	_ , err = btcutil.NewAmount((amount - withdrawal_fee))
	if err != nil {
		log.Println(err)
		return emoji.Emoji("information_source") + " Please enter a valid amount"
	}

	return "ok"
}