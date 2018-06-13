package main

import (
	"log"
	"github.com/jinzhu/gorm"
	"time"
)

type Transaction struct {
	Id int
	Type int
	Amount int64
	UserId int
	Address string
	MessageId int
	TransactionId string
	Confirmed int
}

func (transaction *Transaction) Create() error {
	DB.Create(transaction)
	return DB.Error
}

func (transaction *Transaction) Update() error {
	DB.Save(transaction)
	return DB.Error
}

func (transaction *Transaction) Find() ([]Transaction,error) {
	var transactions []Transaction
	DB.Find(transactions, transaction)
	return transactions, DB.Error
}

func (transaction *Transaction) First() (error) {
	err := DB.First(transaction)
	return err.Error
}

func ProcessTransactions() {
	for {
		transaction := Transaction{
			Type:1,
			Confirmed:0,
		}

		transactions, err := transaction.Find()
		if err != nil {
			log.Println(err)
			continue
		}

		received_transactions, err := Client.ListTransactionsCountFrom("", 100, len(transactions))
		if err != nil {
			log.Println(err)
			continue
		}

		for _,tx := range received_transactions {
			user, err := findUserByAddress(tx.Address)
			if err != nil {
				log.Println(err)
				continue
			}

			new_tx := Transaction{
				Type:1,
				Address: tx.Address,
				Amount:int64(tx.Amount),
				UserId:user.Id,
			}

			 err = new_tx.First()
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					other_err := new_tx.Create()
					if other_err != nil {
						log.Println(err)
						continue
					}
				} else {
					log.Println(err)
					continue
				}
			}

			if tx.Confirmations > 2 {
				new_tx.Confirmed = 1
				user.Balance = user.Balance + new_tx.Amount
				err := user.Update()
				if err != nil {
					log.Println(err)
					continue
				}
				err = new_tx.Update()
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
		time.Sleep(2 * time.Minute)
	}
}