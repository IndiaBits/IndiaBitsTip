package main

import "github.com/btcsuite/btcutil"

func generateAddress() (string, error) {
	address, err := Client.GetNewAddress("")
	if err != nil {
		return "", err
	}
	return address.String(), nil
}

func GetAllAddresses() ([]btcutil.Address, error) {
	addresses, err := Client.GetAddressesByAccount("")
	if err != nil {
		return nil, err
	}
	return addresses, nil
}
