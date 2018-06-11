package main

func generateAddress(username string) (string, error){
	address, err := Client.GetNewAddress("")
	if err != nil {
		return "", err
	}
	return address.String(), nil
}