package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mengdu/geamon"
)

func Interval(fn func(chan bool), d time.Duration) {
	t := time.NewTicker(d)
	c := make(chan bool)
	for {
		select {
		case <-t.C:
			go fn(c)
		case <-c:
			t.Stop()
			return
		}
	}
}

func RunWorker(log *log.Logger) {
	Interval(func(c chan bool) {
		log.Println(time.Now().Unix())
	}, 2*time.Second)
}

func main() {
	if err := os.MkdirAll("./logs", 0777); err != nil {
		panic(err)
	}
	file, err := os.OpenFile("./logs/deamon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	deamon := geamon.Geamon{
		Stdout:     file,
		Stderr:     file,
		MaxRestart: 3,
		// PidFile:    "/var/run/hello.pid",
		DeamonName: "hellod",
	}
	if err := deamon.Run(); err != nil {
		panic(err)
	}

	mylog := log.New(file, fmt.Sprintf("[%d]", os.Getpid()), log.Ldate|log.Lmicroseconds|log.Lshortfile)
	RunWorker(mylog)
}
