package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	prefixLength           = 2
	savedHeadlinesFileName = "headlines.json"
)

var (
	ChainLock     sync.RWMutex
	HeadlineChain *Chain
	Logger        *logrus.Logger
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
	nextTweetChan := time.After(1 * time.Millisecond)

	LoadExistingHeadlines()

	for {
		// Wait until RSS feeds have been updated
		<-pollChan

		avg := Headlines(FetchedHeadlines).AverageWords()

		// dump out some headlines for sample use
		for i := 0; i < 10; i++ {
			fmt.Println(HeadlineChain.Generate(avg))
		}

		if GlobalConfig.Twitter.PostTweet {
			// If we can't tweet within one second, continue
			select {
			case <-time.After(1 * time.Second):
				break
			case <-nextTweetChan:
				var result string
				for {
					result = HeadlineChain.Generate(avg)
					if strings.HasSuffix(result, "?") || strings.HasSuffix(result, ".") ||
						strings.HasSuffix(result, "!") || strings.HasSuffix(result, ":") {
						break
					}
				}
				PostTweet(result)
				nextTweetChan = time.After(timeBetweenTweets())
				break
			}
		}
	}
}

// Time to wait between tweets
func timeBetweenTweets() time.Duration {
	return 1 * time.Hour
}

// Loads saved BuzzFeed headlines
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

// Saves grabbed headlines to the output file for use at a later date (like after a restart)
func SaveHeadlines() {
	file, err := os.OpenFile(savedHeadlinesFileName, os.O_CREATE|os.O_WRONLY, 0775)
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
