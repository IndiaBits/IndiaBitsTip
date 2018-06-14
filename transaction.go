package main

import (
	"log"
	"github.com/jinzhu/gorm"
	"time"
	"sync"
)

type Transaction struct {
	Id int
	Type int
	Amount float64
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
	DB.Find(&transactions, transaction)
	return transactions, DB.Error
}

func (transaction *Transaction) First() (error) {
	err := DB.Where(transaction).First(transaction)
	return err.Error
}

func ProcessTransactions() {
	time.Sleep(10 * time.Second)
	for {
		transactions, err := Client.ListTransactions("")
		if err != nil {
			log.Println(err)
			continue
		}

		for _,tx := range transactions {

			if tx.Category == "send" {
				continue
			}

			user, err := findUserByAddress(tx.Address)
			if err != nil {
				log.Println(err)
				continue
			}

			new_tx := Transaction{
				Type:1,
				Address: tx.Address,
				Amount:tx.Amount,
				UserId:user.Id,
				TransactionId:tx.TxID,
			}

			err = new_tx.First()
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Println("Found new transaction")
					new_tx.Confirmed = 2
					err2 := new_tx.Create()
					if err2 != nil {
						log.Println(err)
						continue
					}
				} else {
					log.Println(err)
					continue
				}
			}

			if new_tx.Confirmed == 2 {
				if tx.Confirmations > 0 {
					new_tx.Confirmed = 1
					err = new_tx.Update()
					if err != nil {
						log.Println(err)
						continue
					}

					if(BalanceMutexes[user.Username] == nil) {
						BalanceMutexes[user.Username] = &sync.Mutex{}
					}

					BalanceMutexes[user.Username].Lock()

					user.Balance = user.Balance + new_tx.Amount
					err := user.Update()
					if err != nil {
						log.Println(err)
						continue
					}

					BalanceMutexes[user.Username].Unlock()
				}
			}
		}
	}
}