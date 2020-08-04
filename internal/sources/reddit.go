package sources

import (
	"bgfreshd/internal"
	"bgfreshd/internal/config"
	"bgfreshd/internal/sources/redditApi"
	"bgfreshd/pkg/background"
	"bgfreshd/pkg/source"
	"fmt"
	"github.com/sirupsen/logrus"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
)

type RedditOptions struct {
	Subreddit   string `yaml:"subreddit"`
	SortBy      string `yaml:"sort"`
	TopTimespan string `yaml:"topTimespan"`
}

func NewRedditSource(config *config.BgSource, sourceLog *logrus.Entry) (source.Source, error) {
	var options RedditOptions
	if err := internal.CastDecodedYamlToType(config.Options, &options); err != nil {
		return nil, err
	}

	sourceLog = sourceLog.WithFields(logrus.Fields{
		"sub": options.Subreddit,
	})

	sortOpt := redditApi.SortOptions{
		SortBy:   options.SortBy,
		Timespan: options.TopTimespan,
	}
	return &redditSource{
		sourceLog:    sourceLog,
		opt:          options,
		api:          redditApi.NewRedditApi(sourceLog, 10),
		sortOptions:  sortOpt,
		currentAfter: "",
	}, nil
}

type redditSource struct {
	sourceLog   *logrus.Entry
	opt         RedditOptions
	api         redditApi.Api
	sortOptions redditApi.SortOptions

	currentAfter string
}

func (r *redditSource) Next() (background.Background, error) {
	page, err := r.api.GetSubredditPosts(r.opt.Subreddit, r.sortOptions, r.currentAfter)
	if err != nil {
		return nil, err
	}

	for _, post := range page.Data.Children {
		bg := r.processPage(post)
		if bg != nil {
			r.currentAfter = post.Data.Name
			return bg, nil
		}
	}

	// no posts found, keep paging
	r.currentAfter = page.Data.After
	return r.Next()
}

func (r *redditSource) GetName() string {
	return fmt.Sprintf("reddit-%s-%s", r.opt.Subreddit, r.sortOptions.SortBy)
}

func (r *redditSource) processPage(post redditApi.Child) background.Background {
	r.sourceLog.Debugf("process post \"%s\"", post.Data.Title)

	if post.Data.PostHint != redditApi.PostHintImage {
		r.sourceLog.Debug("post not image")
		return nil
	}

	img, err := r.downloadImage(post.Data.URL)
	if err != nil {
		r.sourceLog.Warnf("error downloading image %s for post %s : \"%s\"", post.Data.URL, post.Data.Title, err.Error())
		return nil
	}

	bg := background.FromImage(img, post.Data.Name)
	bg.AddMetadata("title", post.Data.Title)
	bg.AddMetadata("permalink", fmt.Sprintf("https://reddit.com%s", post.Data.Permalink))
	bg.AddMetadata("source-name", r.GetName())

	return bg
}

func (r *redditSource) downloadImage(imageUrl string) (image.Image, error) {
	resp, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}
	defer internal.Deferrer(r.sourceLog, resp.Body.Close)

	contentType := resp.Header.Get("content-type")
	switch contentType {
	case "image/jpeg":
		return jpeg.Decode(resp.Body)
	case "image/jpg":
		return jpeg.Decode(resp.Body)
	case "image/png":
		return png.Decode(resp.Body)
	}

	return nil, &UnsupportedImageError{ContentType: contentType}
}

type UnsupportedImageError struct {
	ContentType string
}

func (u UnsupportedImageError) Error() string {
	return fmt.Sprintf("unsupported image type: \"%s\"", u.ContentType)
}
