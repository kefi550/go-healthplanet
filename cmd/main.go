package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kefi550/healthplanet"
)

var (
	loginId = os.Getenv("HEALTHPLANET_LOGIN_ID")
	loginPassword = os.Getenv("HEALTHPLANET_LOGIN_PASSWORD")
	clientId = os.Getenv("HEALTHPLANET_CLIENT_ID")
	clientSecret = os.Getenv("HEALTHPLANET_CLIENT_SECRET")

	influxdbUrl = os.Getenv("INFLUXDB_URL")
	influxdbToken = os.Getenv("INFLUXDB_TOKEN")
	influxdbOrg = os.Getenv("INFLUXDB_ORG")
	influxdbBucket = os.Getenv("INFLUXDB_BUCKET")
	influxdbMeasurement = os.Getenv("INFLUXDB_MEASUREMENT")
)

func main() {
	hp := healthplanet.NewClient(
		loginId,
		loginPassword,
		clientId,
		clientSecret,
	)

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}
	// 環境変数 STATUS_TO_DATETIME が設定されている場合はその値をtoとし、設定されていない場合は現在時刻をto, 1ヶ月前をfromとする
	now := time.Now()
	now = now.In(jst)
	to := os.Getenv("STATUS_TO_DATETIME")
	if to == "" {
		to = now.Format("20060102150405")
	}
	parsedTo, err := time.ParseInLocation("20060102150405", to, jst)
	from := parsedTo.AddDate(-1, 0, 0).Format("20060102150405")

	getInnerScanRequest := healthplanet.GetStatusRequest{
		DateMode:    healthplanet.DateMode_MeasuredDate,
		From:        from,
		To:          to,
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
		parsedTime, _ := time.ParseInLocation("200601021504", data.Date, jst)
		value, _ := strconv.ParseFloat(data.KeyData, 64)
		err = healthplanet.WriteInfluxDB(influxdbUrl, influxdbToken, influxdbOrg, influxdbBucket, influxdbMeasurement, tag, value, parsedTime)
		if err != nil {
			log.Fatal(err)
		}
	}
}
