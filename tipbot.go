package main

import (
	"github.com/joho/godotenv"
	"log"
	"github.com/jinzhu/gorm"
	"os"
	"github.com/btcsuite/btcd/rpcclient"
	"gopkg.in/tucnak/telebot.v2"
	"time"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB
var Client *rpcclient.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initDB()
	defer DB.Close()
	initRPC()
	defer Client.Shutdown()

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
	}

	InitTelegramCommands(bot)
	log.Println("Server started..")
	bot.Start()

}

func initRPC() {
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18332",
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