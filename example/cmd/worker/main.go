package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	intervalRaw := os.Getenv("INTERVAL")
	if intervalRaw == "" {
		intervalRaw = "10s"
	}
	interval, err := time.ParseDuration(intervalRaw)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	for {
		fmt.Println(time.Now(), "TASK", i+1)
		i++
		time.Sleep(interval)
	}
}
