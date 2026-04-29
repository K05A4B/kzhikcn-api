package machineresouces

import (
	"encoding/xml"
	"fmt"
	"io"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"time"

	"gorm.io/gorm"
)

type SitemapURL struct {
	Location        string `xml:"loc" json:"loc"`
	LastModify      string `xml:"lastmod,omitempty" json:"last_mod"`
	ChangeFrequency string `xml:"changefreq,omitempty" json:"change_freq"`
	Priority        int8   `xml:"priority,omitempty" json:"priority"`
}

func SitemapExtend2Url(c config.SitemapExtend) *SitemapURL {
	return &SitemapURL{
		Location:        c.Location,
		LastModify:      c.LastModify,
		ChangeFrequency: c.ChangeFrequency,
		Priority:        c.Priority,
	}
}

func parseTpl(conf *config.MachineReadableResources, tpl config.TemplateString, fn func(string) any) (string, error) {
	res, err := tpl.Parse(fn)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", conf.BaseUrl, res), nil
}

func GenerateSitemap(w io.Writer) error {
	conf := config.Conf().MRR

	urls := []SitemapURL{}

	for _, item := range conf.Sitemap.Extends {
		u := *SitemapExtend2Url(item)
		u.Location = conf.BaseUrl.String() + u.Location
		urls = append(urls, u)
	}

	articleUrls, err := articleURLs(&conf)
	if err != nil {
		return err
	}

	urls = append(urls, articleUrls...)

	d := struct {
		Urls []SitemapURL `xml:"url"`
	}{
		Urls: urls,
	}

	elem := xml.StartElement{
		Name: xml.Name{
			Local: "urlset",
		},
		Attr: []xml.Attr{
			{
				Name: xml.Name{
					Local: "xmlns",
				},
				Value: "http://www.sitemaps.org/schemas/sitemap/0.9",
			},
		},
	}

	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)

	enc.EncodeElement(d, elem)

	return nil
}

func articleURLs(conf *config.MachineReadableResources) ([]SitemapURL, error) {
	articles, err := data.GetArticles(func(tx *gorm.DB) *gorm.DB {
		return tx.Select("id, custom_id, updated_at").Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
	})

	if err != nil {
		return nil, err
	}

	result := []SitemapURL{}

	for _, article := range articles {
		lastMod := article.UpdatedAt

		loc, err := parseTpl(conf, conf.UrlTemplates.Article, func(s string) any {
			if s == "article.id" {
				return article.ID
			}

			if s == "article.custom_id" {
				return article.CustomID
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		result = append(result, SitemapURL{
			Location:   loc,
			LastModify: lastMod.Format(time.RFC3339),
		})
	}

	return result, nil
}
