package main

type Transaction struct {
	Id int
	Type int
	Amount int64
	UserId int
	MessageId int
	TransactionId string
	Confirmed int
}

func (transaction *Transaction) Create() error {
	DB.Create(transaction)
	return DB.Error
}

func (transaction *Transaction) Find() error {
	DB.Find(transaction)
	return DB.Error
}

func ProcessTransactions() {
}