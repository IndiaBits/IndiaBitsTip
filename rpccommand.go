package main

func generateAddress() (string, error){
	address, err := Client.GetNewAddress("")
	if err != nil {
		return "", err
	}
	return address.String(), nil
}