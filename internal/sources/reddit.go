package sources

import (
	"bgfreshd/internal"
	"bgfreshd/internal/pipeline"
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

func init() {
	pipeline.AddSourceRegistration("reddit", NewRedditSource)
}

func NewRedditSource(config *source.Configuration, dbFactory source.DbFactoryFunc, sourceLog *logrus.Entry) (source.Source, error) {
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

	newSource := &redditSource{
		log:          sourceLog,
		db:           nil,
		opt:          options,
		api:          redditApi.NewRedditApi(sourceLog, 10),
		sortOptions:  sortOpt,
		currentAfter: "",
	}

	db, err := dbFactory(newSource.GetName())
	if  err != nil {
		return nil, err
	}

	newSource.db = db
	if exists, err := newSource.db.KeyExists("after"); exists && err == nil {
		newSource.currentAfter, err = newSource.db.GetString("after")
		if err != nil {
			return nil, err
		}
	}else if err != nil {
		return nil, err
	}

	return newSource, nil
}

type redditSource struct {
	log         *logrus.Entry
	db          source.Db
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

	// currently we just store our progress in db, never messing with it
	// todo: make this paging much smarter (IE if we're on hot for last day, reset this once a day, etc)
	if err := r.db.SetString("after", r.currentAfter); err != nil {
		return nil, err
	}

	return r.Next()
}

func (r *redditSource) GetName() string {
	return fmt.Sprintf("reddit-%s-%s", r.opt.Subreddit, r.sortOptions.SortBy)
}

func (r *redditSource) processPage(post redditApi.Child) background.Background {
	r.log.Debugf("process post \"%s\"", post.Data.Title)

	if post.Data.PostHint != redditApi.PostHintImage {
		r.log.Debug("post not image")
		return nil
	}

	img, err := r.downloadImage(post.Data.URL)
	if err != nil {
		r.log.Warnf("error downloading image %s for post %s : \"%s\"", post.Data.URL, post.Data.Title, err.Error())
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
	defer internal.Deferrer(r.log, resp.Body.Close)

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
