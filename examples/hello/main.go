package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	file, err := os.OpenFile("./logs/hellod.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	backd := geamon.Geamon{
		Stdout: file,
		Stderr: file,
		// PidFile:      "/var/run/hello.pid",
		PidFile:      "./logs/pid/hello.pid",
		ProcessTitle: "hellod",
	}
	if err := backd.Run(); err != nil {
		panic(err)
	}
	defer backd.ReleasePidFile()
	mylog := log.New(file, fmt.Sprintf("[%d]", os.Getpid()), log.Ldate|log.Lmicroseconds|log.Lshortfile)
	go RunWorker(mylog)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
