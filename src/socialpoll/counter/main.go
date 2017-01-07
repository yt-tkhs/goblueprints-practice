package main

import (
	"fmt"
	"flag"
	"os"
	"log"
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/bitly/go-nsq"
	"time"
	"gopkg.in/mgo.v2/bson"
	"os/signal"
	"syscall"
)

const updateDuration = 1 * time.Second

var (
	fatalErr error
	countsLock sync.Mutex
	counts map[string]int
)

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func main() {
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	log.Println("Connecting to Database...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
		return
	}
	defer func() {
		log.Println("Disconnecting Database...")
		db.Close()
	}()

	pollData := db.DB("ballots").C("polls")

	log.Println("Connecting to NSQ...")
	consumer, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}

	// メッセージ受信時のイベント
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(message.Body)
		counts[vote]++
		return nil
	}))

	// nsqlookupd に接続. 直接 twittervotes (NSQのインスタンス) に接続しない.
	// これにより, どこからメッセージが送られてくるかを抽象化できる
	if err := consumer.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return
	}


	// 投票の結果を定期的にDBに登録する
	log.Println("Waiting for voting on NSQ...")
	var updater *time.Timer
	updater = time.AfterFunc(updateDuration, func() {
		countsLock.Lock()
		defer countsLock.Unlock()

		if len(counts) == 0 {
			log.Println("There are no new votes. skipped updating database.")
		} else {
			log.Println("Updating database...")
			log.Println(counts)

			ok := true
			for option, count := range counts {
				sel := bson.M{"options": bson.M{"$in": []string{option}}}
				up := bson.M{"$inc": bson.M{"results." + option: count}}

				if _, err := pollData.UpdateAll(sel, up); err != nil {
					log.Println("Failed to update:", err)
					ok = false
					continue
				}

				counts[option] = 0
			}

			if ok {
				log.Println("Complete to update database.")
				counts = nil
			}
		}

		updater.Reset(updateDuration)
	})

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-termChan:
			updater.Stop()
			consumer.Stop()
		case <-consumer.StopChan:
			// Complete stopping
			return
		}
	}
}
