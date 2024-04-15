package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yardz/test-apybara-go-bot/src/app"
)

func runBot() {
	// ticker := time.(24 * time.Hour)
	ticker := time.NewTicker(5 * time.Second)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Cleanup before exit")
		ticker.Stop()
		os.Exit(1)
	}()

	app := app.NewApp()

	for {
		select {
		case <-ticker.C:
			currentTime := time.Now()
			fmt.Printf("Executing at: %v/%v/%v %v:%v:%v\n", currentTime.Day(), currentTime.Month(), currentTime.Year(), currentTime.Hour(), currentTime.Minute(), currentTime.Second())
			app.RunStakeBot()
		}
	}
}

func main() {
	runBot()
	fmt.Printf("Finished at %v\n", time.Now())
}
