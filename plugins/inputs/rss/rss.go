package rss

import (
	"strings"
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
				fields["title"] = strings.Replace(item.Title, "\n", "", -1)
			case "description":
				fields["description"] = strings.Replace(item.Description, "\n", "", -1)
			case "content":
				fields["content"] = strings.Replace(item.Content, "\n", "", -1)
			case "link":
				fields["link"] = strings.Replace(item.Link, "\n", "", -1)
			case "updated":
				fields["updated"] = strings.Replace(item.Updated, "\n", "", -1)
			case "published":
				fields["published"] = strings.Replace(item.Published, "\n", "", -1)
			case "author":
				if item.Author != nil {
					fields["author"] = item.Author.Name
				}
			case "guid":
				fields["guid"] = strings.ReplaceAll(item.GUID, "\n", "")
			}
		}

		acc.AddFields("rss", fields, tags)

	}
	return nils
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
