package rss

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/mmcdole/gofeed"
)

// Rss rss plugin
type Rss struct {
	Feeds   []string
	Filters []string

	reported map[string]struct{}
	mu       sync.Mutex
}

// Description the rss feed plugin
func (m *Rss) Description() string {
	return `Get all the hot feed action`
}

// SampleConfig doo dad
func (m *Rss) SampleConfig() string {
	return `
  # Specify a list of one or more riak http servers
  feeds = ["http://feed.feed/feediemcfeed"]
  filter = ["title", "description", "content", "link", "updated", "published", "author", "guid"]`
}

// Gather defines what data the plugin will gather.
func (m *Rss) Gather(acc telegraf.Accumulator) error {
	if len(m.Feeds) == 0 {
		return nil
	}
	for _, feed := range m.Feeds {
		acc.AddError(m.gatherFeed(feed, acc))

	}

	return nil
}

func (m *Rss) gatherFeed(feed string, acc telegraf.Accumulator) error {
	fp := gofeed.NewParser()
	f, err := fp.ParseURL(feed)
	if err != nil {
		panic(err)
	}

	for _, item := range f.Items {

		tags := map[string]string{
			"feed": feed,
		}
		for _, cat := range item.Categories {
			tags[cat] = "true"
		}

		fields := map[string]interface{}{}
		for _, t := range m.Filters {
			switch t {
			case "title":
				fields["title"] = item.Title
			case "description":
				fields["description"] = item.Description
			case "content":
				fields["content"] = item.Content
			case "link":
				fields["link"] = item.Link
			case "updated":
				fields["updated"] = item.Updated
			case "published":
				fields["published"] = item.Published
			case "author":
				fields["author"] = item.Author
			case "guid":
				fields["guid"] = item.GUID
			}
		}

		acc.AddFields("rss", fields, tags)

	}

}

func (m Rss) Report(str string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.reported[str]
	m.reported[str] = struct{}{}
	return !ok

}

func init() {
	inputs.Add("rss", func() telegraf.Input {
		return &Rss{
			reported: map[string]struct{}{},
		}
	})
}
