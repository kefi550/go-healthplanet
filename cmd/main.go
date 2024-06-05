package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kefi550/healthplanet"
)

func main() {
	loginId := os.Getenv("HEALTHPLANET_LOGIN_ID")
	loginPassword := os.Getenv("HEALTHPLANET_LOGIN_PASSWORD")
	clientId := os.Getenv("HEALTHPLANET_CLIENT_ID")
	clientSecret := os.Getenv("HEALTHPLANET_CLIENT_SECRET")


	hp := healthplanet.NewClient(
		loginId,
		loginPassword,
		clientId,
		clientSecret,
	)
	getInnerScanRequest := healthplanet.GetStatusRequest{
		DateMode:    "0",
		From:        "20240501000000",
		To:          "20240520000000",
		Tag:         healthplanet.Weight,
	}
	status, err := hp.GetInnerscan(getInnerScanRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, data := range status.Data {
		fmt.Println(data.Date)
		fmt.Println(data.KeyData)
		fmt.Println(data.Tag)
	}
}
