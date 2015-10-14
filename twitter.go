package main

import (
	"fmt"

	"github.com/ChimeraCoder/anaconda"
)

const (
	TweetMaxLength = 140
)

var (
	twitterClient *anaconda.TwitterApi
)

func init() {
}

func PostTweet(t string) {
	if twitterClient == nil {
		twitterConfig := GlobalConfig.Twitter

		anaconda.SetConsumerKey(twitterConfig.ConsumerToken)
		anaconda.SetConsumerSecret(twitterConfig.ConsumerSecret)

		twitterClient = anaconda.NewTwitterApi(twitterConfig.AccessToken, twitterConfig.AccessSecret)
	}

	_, err := twitterClient.PostTweet(t, nil)
	if err != nil {
		Logger.Errorln(err)
		return
	}

	fmt.Println("Tweet posted")
}
