package main

import (
	"os"

	healthplanet "github.com/kefi550/health-planet-monitoring"
)

func main() {
	loginId := os.Getenv("HEALTHPLANET_LOGIN_ID")
	loginPassword := os.Getenv("HEALTHPLANET_LOGIN_PASSWORD")
	clientId := os.Getenv("HEALTHPLANET_CLIENT_ID")
	clientSecret := os.Getenv("HEALTHPLANET_CLIENT_SECRET")


	hp := healthplanet.NewHealthPlanet(
		loginId,
		loginPassword,
		clientId,
		clientSecret,
	)
	hp.Run()
}
