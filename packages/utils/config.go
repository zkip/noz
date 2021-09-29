package utils

import (
	"os"
)

func GetServeHost(defaultPort int16) string {
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = ToString(defaultPort)
	}
	return ":" + portString
}
