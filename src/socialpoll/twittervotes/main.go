package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"github.com/bitly/go-nsq"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var db *mgo.Session

type poll struct {
	Options []string
}

func main() {
	var stoplock sync.Mutex
	stop := false
	stopchan := make(chan struct{}, 1)
	signalchan := make(chan os.Signal, 1)

	// L49 からの goroutine で終了処理を行なっているので, 終了するためのシグナルが発生したことを伝える(stop = true)
	// こうすることで, Ctrl + C による強制終了でも各コネクションを閉じたりといった綺麗な終了処理を実行させることができる
	go func() {
		<-signalchan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Finishing")
		stopchan <- struct {}{}
		closeConn()
	}()

	// ここに登録したシグナルが発生すると, signalchan に os.Signal を流す
	signal.Notify(signalchan, syscall.SIGINT, syscall.SIGTERM)

	if err := dialDB(); err != nil {
		log.Fatalln("Failed to dial to MongoDB:", err)
	}
	defer closeDB()

	// 処理を開始
	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopchan, votes)

	// stop = true(os.Signal を受信している)なら, 処理を終了する.
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				break
			}
			stoplock.Unlock()
		}
	}()

	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}

func loadOptions() ([]string, error) {
	var options []string
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	var p poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()


	return options, iter.Err()
}

func dialDB() error {
	url := "localhost"
	var err error
	log.Printf("Dial to MongoDB: %s\n", url)
	db, err = mgo.Dial(url)
	return err
}

func closeDB() {
	db.Close()
	log.Println("Closed MongoDB connection.")
}

func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{})

	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())

	go func() {
		defer func() {
			stopchan <- struct {}{}
		}()

		for vote := range votes {
			pub.Publish("votes", []byte(vote))
		}

		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
	}()

	return stopchan
}