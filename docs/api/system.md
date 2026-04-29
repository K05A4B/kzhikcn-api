# 系统资源 API

> 这些接口不在 `/api/v1` 下。

## 站点地图（Sitemap）

```
GET /sitemap.xml
```

**认证**：无需认证

**响应**：XML 内容

**接口错误代码**：

- `system.machine_resources.sitemap_not_enabled` 未启用 Sitemap
- `system.machine_resources.sitemap_generate_failed` 生成 Sitemap 失败

---

## RSS

```
GET /rss.xml
```

**认证**：无需认证

**响应**：XML 内容

**接口错误代码**：

- `system.machine_resources.rss_not_enabled` 未启用 RSS
- `system.machine_resources.article_get_failed` 获取文章失败
- `system.machine_resources.rss_generate_failed` 生成 RSS 失败

---

## 配置提示

这些接口由 `machine_readable_resources` 配置控制，详情见 `docs/configuration.md`。
