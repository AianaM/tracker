package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var trackerUrl = ""

func init() {
	c = makeClouds()

	if loc, err := time.LoadLocation("Europe/Moscow"); err != nil {
		fmt.Println(err)
	} else {
		time.Local = loc
	}
}

func main() {
	trackerUrl = getTrackerUrl()
	newWebClient()
}

func getTrackerUrl() string {
	if len(os.Args) < 2 {
		fmt.Println("Если хотим ссылку на задачи, надо указать аргумент")
		fmt.Println("например, tracker.exe http://tracker.company.com")
		return ""
	}

	if strings.HasPrefix(os.Args[1], "http://") || strings.HasPrefix(os.Args[1], "https://") {
		return os.Args[1] // Возвращаем URL, если он начинается с http:// или https://
	}

	fmt.Println("Не понял ваши аргументы: некорректный URL")
	log.Printf("Unexpected argument '%s'", os.Args[1])
	return ""

}
