// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    subredditResponse, err := UnmarshalSubredditResponse(bytes)
//    bytes, err = subredditResponse.Marshal()

package redditApi

import "encoding/json"

func UnmarshalSubredditResponse(data []byte) (SubredditResponse, error) {
	var r SubredditResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SubredditResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type SubredditResponse struct {
	Kind string                `json:"kind"`
	Data SubredditResponseData `json:"data"`
}

type SubredditResponseData struct {
	Modhash  string  `json:"modhash"`
	Dist     int64   `json:"dist"`
	Children []Child `json:"children"`
	After    string  `json:"after"`
	Before   string  `json:"before"`
}

type Child struct {
	Kind Kind      `json:"kind"`
	Data ChildData `json:"data"`
}

type ChildData struct {
	Subreddit     string `json:"subreddit"`
	Title         string `json:"title"`
	Name          string `json:"name"`
	SubredditType string `json:"subreddit_type"`
	PostHint      string `json:"post_hint,omitempty"`
	ID            string `json:"id"`
	Author        string `json:"author"`
	Permalink     string `json:"permalink"`
	URL           string `json:"url"`
}

type Preview struct {
	Images  []ImageElement `json:"images"`
	Enabled bool           `json:"enabled"`
}

type ImageElement struct {
	Source      Source   `json:"source"`
	Resolutions []Source `json:"resolutions"`
	ID          string   `json:"id"`
}

type Source struct {
	URL    string `json:"url"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

const (
	PostHintImage string = "image"
	PostHintLink  string = "link"
)

type Kind string

const (
	T3 Kind = "t3"
)
