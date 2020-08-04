package redditApi

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	Top   string = "top"
	New   string = "new"
	Hot   string = "hot"
	Hour  string = "hour"
	Day   string = "day"
	Week  string = "week"
	Month string = "month"
	Year  string = "year"
	All   string = "all"
)

type SortOptions struct {
	SortBy   string
	Timespan string
}

type Api interface {
	GetSubredditPosts(subreddit string, sort SortOptions, after string) (*SubredditResponse, error)
}

type api struct {
	logger    *logrus.Entry
	redditUrl string
	pageSize  int
	client    *http.Client
}

func NewRedditApi(logger *logrus.Entry, pageSize int) Api {
	return &api{
		logger:    logger,
		redditUrl: "https://reddit.com",
		pageSize:  pageSize,
		client:    &http.Client{},
	}
}

func (a *api) GetSubredditPosts(subreddit string, sort SortOptions, after string) (*SubredditResponse, error) {
	if sort.SortBy == "" {
		sort.SortBy = Top
	}

	if sort.SortBy == Top && sort.Timespan == "" {
		sort.Timespan = All
	}

	requestUrl, err := url.Parse(fmt.Sprintf("%s/r/%s/%s.json", a.redditUrl, subreddit, sort.SortBy))
	if err != nil {
		return nil, err
	}
	query := &url.Values{}
	query.Add("limit", fmt.Sprintf("%d", a.pageSize))
	if len(after) != 0 {
		query.Add("after", after)
	}

	if sort.SortBy == Top {
		query.Add("t", sort.Timespan)
	}

	requestUrl.RawQuery = query.Encode()
	reqStr := requestUrl.String()
	a.logger.Debugf("sending request %s", reqStr)

	req, err := a.buildRequest(reqStr)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Errorf("reddit request failed %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	a.logger.Debugf("Request responded with status %d", resp.StatusCode)
	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case 429:
			a.logger.Warnf("Reddit throttling detected, sleeping...")
			time.Sleep(time.Second * 5)

			return a.GetSubredditPosts(subreddit, sort, after)
		}

		return nil, fmt.Errorf("reddit responded with status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	deserialized, err := UnmarshalSubredditResponse(body)
	return &deserialized, err
}

func (a *api) buildRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-agent", "bgfreshd/1.0")
	return req, nil
}
