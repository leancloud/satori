package g

import (
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"time"
)

const LAST_MESSAGE_PATH = "/tmp/satori-last-message"

func LastMessage(name string) {
	msg := "Goroutine:" + name + "\n" +
		"Time: " + time.Now().Format("2006-01-02 15:04:05.000000") + "\n" +
		string(debug.Stack())
	log.Println("======== PANICKED! ========")
	log.Println(msg)
	log.Println("===========================")

	var f *os.File
	var err error

	if f, err = os.OpenFile(LAST_MESSAGE_PATH, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		log.Fatalln(err)
	}
	if _, err = f.WriteString(msg); err != nil {
		log.Fatalln(err)
	}
	_ = f.Close()
	log.Fatalln("Successfully left last message, die in peace.")
}

func ReportLastMessage() {
	time.Sleep(5 * time.Second)
	if _, err := os.Stat(LAST_MESSAGE_PATH); err == nil {
		log.Println("Found last message, reporting")
		if msg, err := ioutil.ReadFile(LAST_MESSAGE_PATH); err == nil {
			ReportFailure(".satori.agent.panic", string(msg), nil)
			_ = os.Remove(LAST_MESSAGE_PATH)
		} else {
			log.Println("Error reading last message:", err)
		}
	}
}
