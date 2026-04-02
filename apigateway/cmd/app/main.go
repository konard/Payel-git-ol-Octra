package main

import (
	"apigateway/internal/fetcher"
	"log"

	"github.com/Payel-git-ol/azure"
)

func main() {
	a := azure.Defoult
	fetcher.HttpManager(a)

	if err := a.Run("3111"); err != nil {
		log.Fatal("HTTP server error:", err)
	}
}
