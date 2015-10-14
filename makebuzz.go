package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
	"strings"

    "github.com/Sirupsen/logrus"
)

const (
	prefixLength           = 1
	savedHeadlinesFileName = "headlines.json"
)

var (
	ChainLock     sync.RWMutex
	HeadlineChain *Chain
	Logger *logrus.Logger
)

func init() {
	Logger = logrus.New()
}

func main() {
	// Catch interrupts
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			SaveHeadlines()
			os.Exit(0)
		}
	}()

	ParseConfig()

	rand.Seed(time.Now().UnixNano())

	// Create the markov chain
	HeadlineChain = NewChain(prefixLength)

	// Start fetching new headlines
	feeds := CreateFeeds()
	pollChan := feeds.Poll(10 * time.Second)
	var nextTweetChan <-chan time.Time

	LoadExistingHeadlines()

	for {
		<-pollChan

		avg := Headlines(FetchedHeadlines).AverageWords()
		fmt.Println(avg)

//		for i := 0; i < 10; i++ {
//			fmt.Println(HeadlineChain.Generate(avg))
//		}

		if nextTweetChan == nil {
			nextTweetChan = time.After(duration())
//			PostTweet(HeadlineChain.Generate(avg))
			continue
		}

		select {
		case <-time.After(1 * time.Second):
			continue
		case <-nextTweetChan:
//			PostTweet(HeadlineChain.Generate(avg))
			nextTweetChan = time.After(duration())
			break
		}
	}

	// Create a new markov
}

func duration() time.Duration {
	return 1 * time.Hour
}

func LoadExistingHeadlines() {
	var headlines []string

	file, err := os.Open(savedHeadlinesFileName)
	if err != nil {
		return
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&headlines)

	if err != nil {
		Logger.Errorln(err)
		return
	}


	ChainLock.Lock()
	FetchedHeadlinesLock.Lock()
	defer ChainLock.Unlock()
	defer FetchedHeadlinesLock.Unlock()

	for _, line := range headlines {
		reader := strings.NewReader(line)
		FetchedHeadlines = append(FetchedHeadlines, line)
		HeadlineChain.Build(reader)
	}
}

func SaveHeadlines() {
	file, err := os.OpenFile(savedHeadlinesFileName, os.O_CREATE | os.O_WRONLY, 0775)
	if err != nil {
		Logger.Errorln(err)
		return
	}

	defer file.Close()

	encoder := json.NewEncoder(file)

	FetchedHeadlinesLock.RLock()
	defer FetchedHeadlinesLock.RUnlock()

	err = encoder.Encode(FetchedHeadlines)

	if err != nil {
		Logger.Errorln(err)
	}
}
