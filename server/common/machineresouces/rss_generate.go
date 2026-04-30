package machineresouces

import (
	"bytes"
	"encoding/xml"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"time"
)

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Copyright   string `xml:"copyright,omitempty"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Category    string `xml:"category,omitempty"`
	PubDate     string `xml:"pubDate,omitempty"`
}

func GenerateRSS(articles []data.Article) ([]byte, error) {
	conf := config.Conf().MRR

	items := []Item{}
	tpl := conf.UrlTemplates

	for _, article := range articles {
		link, err := tpl.Article.Parse(func(s string) any {
			if s == "article.id" {
				return article.ID.String()
			}

			if s == "article.custom_id" {
				return article.CustomID
			}

			return nil
		})

		link = conf.BaseUrl.String() + link

		if err != nil {
			return nil, err
		}

		publishedAt := article.PublishedAt
		pubDate := ""

		if publishedAt != nil {
			pubDate = publishedAt.Format(time.RFC3339)
		}

		item := Item{
			Title:       article.Title,
			Category:    article.Category.CategoryName,
			PubDate:     pubDate,
			Description: article.Description,
			Link:        link,
		}

		items = append(items, item)
	}

	result := bytes.NewBuffer([]byte{})

	d := struct {
		Channel Channel `xml:"channel"`
	}{
		Channel: Channel{
			Title:       conf.Rss.Title.String(),
			Description: conf.Rss.Description.String(),
			Link:        conf.BaseUrl.String(),
			Items:       items,
		},
	}

	enc := xml.NewEncoder(result)
	enc.EncodeElement(d, xml.StartElement{
		Name: xml.Name{
			Local: "rss",
		},
		Attr: []xml.Attr{
			{
				Name: xml.Name{
					Local: "version",
				},
				Value: "2",
			},
		},
	})

	return result.Bytes(), nil
}
