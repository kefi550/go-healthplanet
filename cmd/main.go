package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kefi550/healthplanet"
)

func main() {
	loginId := os.Getenv("HEALTHPLANET_LOGIN_ID")
	loginPassword := os.Getenv("HEALTHPLANET_LOGIN_PASSWORD")
	clientId := os.Getenv("HEALTHPLANET_CLIENT_ID")
	clientSecret := os.Getenv("HEALTHPLANET_CLIENT_SECRET")

	influxdbUrl := os.Getenv("INFLUXDB_URL")
	influxdbToken := os.Getenv("INFLUXDB_TOKEN")
	influxdbOrg := os.Getenv("INFLUXDB_ORG")
	influxdbBucket := os.Getenv("INFLUXDB_BUCKET")

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
		tag, err := hp.GetTagValue(data.Tag)
		if err != nil {
			log.Fatal(err)
		}
		parsedTime, _ := time.Parse("200601021504", data.Date)
		value, _ := strconv.ParseFloat(data.KeyData, 64)
		err = healthplanet.WriteInfluxDB(influxdbUrl, influxdbToken, influxdbOrg, influxdbBucket, tag, value, parsedTime)
		if err != nil {
			log.Fatal(err)
		}
	}
}
