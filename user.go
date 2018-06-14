package main

type User struct {
	Id int
	Username string
	Address string
	Balance	float64
}

func (u *User) Register() (error) {
	err := DB.Create(&u)
	return err.Error
}

func (u *User) Update() (error) {
	err := DB.Save(&u)
	return err.Error
}

func (u *User) First() (*User, error) {
	err := DB.Where(u).First(u)
	return u, err.Error
}