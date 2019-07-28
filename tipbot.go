package main

import (
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
	"sync"
)

var DB *gorm.DB
var Client *rpcclient.Client
var BalanceMutexes = make(map[string]*sync.Mutex)
var withdrawal_confirmations = make(map[string]*telebot.Message)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initDB()
	defer DB.Close()
	initRPC()
	defer Client.Shutdown()

	/*go ProcessTransactions()

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
	}

	InitTelegramCommands(bot)
	log.Println("Server started..")
	bot.Start()*/

	for {
		address, err := generateAddress()
		if err != nil {
			fmt.Print(err)
		}
		fmt.Println(address)
	}

}

func initRPC() {
	connCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv("RPC_HOST") + ":" + os.Getenv("RPC_PORT"),
		User:         os.Getenv("RPC_USER"),
		Pass:         os.Getenv("RPC_PASS"),
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	var err error
	Client, err = rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func initDB() {
	log.Println("Connecting to DB...")
	var err error
	DB, err = gorm.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@/"+os.Getenv("DB_NAME")+"?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to DB")
}
