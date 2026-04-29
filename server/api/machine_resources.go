package api

import (
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/hdl"
	"kzhikcn/server/common/machineresouces"
	"net/http"

	"gorm.io/gorm"
)

var SitemapDotXMLHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	w := hdl.ResponseWriter(r)

	if !config.Conf().MRR.Sitemap.Enable {
		return hdl.Error(http.StatusNotFound, "Sitemap未启用", nil, "system.machine_resources.sitemap_not_enabled")
	}

	w.Header().Set("Content-Type", "application/xml")
	err := machineresouces.GenerateSitemap(w)
	if err != nil {
		return hdl.Error(http.StatusInternalServerError, "生成Sitemap失败", err, "system.machine_resources.sitemap_generate_failed")
	}
	return hdl.NoRespond
})

var RssDotXMLHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	if !config.Conf().MRR.Rss.Enable {
		return hdl.Error(http.StatusNotFound, "RSS未启用", nil, "system.machine_resources.rss_not_enabled")
	}

	max := config.Conf().MRR.Rss.MaxArticles

	if max <= 0 {
		max = 10
	}

	articles, err := data.GetArticles(func(tx *gorm.DB) *gorm.DB {
		return tx.Limit(max).Order("published_at DESC").Where("status=?", data.ARTICLE_STATUS_PUBLISHED)
	})

	if err != nil {
		return hdl.Error(http.StatusInternalServerError, "获取文章失败", err, "system.machine_resources.article_get_failed")
	}

	rss, err := machineresouces.GenerateRSS(articles)
	if err != nil {
		return hdl.Error(http.StatusInternalServerError, "生成RSS失败", err, "system.machine_resources.rss_generate_failed")
	}
	return hdl.WriteRawData(r, rss, "application/xml")
})
