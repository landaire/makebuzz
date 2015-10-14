package main

import (
	"fmt"
	"time"

	"bufio"
	"strings"
	"sync"

	"math"
	"sort"

	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/jteeuwen/go-pkg-xmlx"
)

const (
	baseFeedUrl = "http://www.buzzfeed.com/%s.xml"
	timeout     = 5
)

type Headlines []string
type BuzzFeeds []*BuzzFeedRss

type BuzzFeedRss struct {
	*rss.Feed
}

var (
	FetchedHeadlines     Headlines
	FetchedHeadlinesLock sync.RWMutex
	uniquesChecked       = false
	buzzFeedConfig       *BuzzFeedConfig
)

func CreateFeeds() BuzzFeeds {
	buzzFeedConfig = &GlobalConfig.BuzzFeed

	var feeds BuzzFeeds

	// initialize the buzzfeed RSS with a timeout of 5 seconds
	for i := 0; i < len(buzzFeedConfig.Feeds); i++ {
		feeds = append(feeds, &BuzzFeedRss{rss.New(timeout, true, feedChannelHandler, feedItemHandler)})
	}

	fmt.Println(feeds)

	return feeds
}

// Polls a series of feeds for updates
func (this BuzzFeeds) Poll(timeout time.Duration) <-chan bool {
	c := make(chan bool)
	go func() {
		for {
			for i, feed := range this {
				// Wait until the next update time
				<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))

				feedUrl := fmt.Sprintf(baseFeedUrl, buzzFeedConfig.Feeds[i])
				feed.Poll(feedUrl, nil)
			}

			fmt.Println("done")

			// Write to the output file
			SaveHeadlines()

			uniquesChecked = true
			c <- true
		}
	}()

	return c
}

// Polls the RSS feed for new channels and items
func (feed *BuzzFeedRss) Poll(feedUrl string, cr xmlx.CharsetFunc) {
	fmt.Println("Fetching feed ", feedUrl)
	if err := feed.Fetch(feedUrl, cr); err != nil {
		Logger.Errorf("%s: %s\n", feedUrl, err)
	}
}

// Rounding utility function
// Courtesy of https://gist.github.com/DavidVaini/10308388
func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// Calculates the average length of each headline
func (h Headlines) AverageWords() int {
	if len(h) == 0 {
		return 0
	}

	length := 0
	for _, str := range h {
		length += WordCount(str)
	}

	return int(Round(float64(length)/float64(len(h)), 1, 0))
}

func WordCount(s string) int {

	length := 0

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		length++
	}

	return length
}

func feedChannelHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	// really do nothing with these -- they're not useful
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func feedItemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	ChainLock.Lock()
	FetchedHeadlinesLock.Lock()
	defer FetchedHeadlinesLock.Unlock()
	defer ChainLock.Unlock()

	for _, item := range newitems {
		if !uniquesChecked && len(FetchedHeadlines) > 0 {
			sort.Strings(FetchedHeadlines)
			if index := sort.SearchStrings(FetchedHeadlines, item.Title); index < len(FetchedHeadlines) && FetchedHeadlines[index] == item.Title {
				continue
			}
		}

		fmt.Println("Adding headline", item.Title)
		HeadlineChain.Build(strings.NewReader(item.Title))
		FetchedHeadlines = append(FetchedHeadlines, item.Title)
	}

	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)
}
