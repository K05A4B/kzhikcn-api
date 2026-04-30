# kzhikcn-api

kzhikcn-api 是一个基于 Go 语言开发的简单的无头内容管理系统后端

## 核心功能

### 文章管理
- 创建、查询、更新、删除文章
- 文章内容（Markdown）管理
- 文章资源（附件）上传与管理
- 文章浏览量和点赞量统计
- 文章软删除与恢复

### 用户认证
- 账号密码登录
- TOTP 多因素认证
- 密码修改

### 分类与标签
- 分类管理（创建、更新、删除）
- 标签管理（创建、更新、删除）
- 按分类/标签查询文章

## API 文档

[API 文档](./docs/introduct.md)

## 许可协议 / LICENSE

Copyright (C) 2026 K05A4B (kzhik)

这个项目基于 GNU Affero General Public License v3.0 许可协议开源。
查看 [LICENSE](LICENSE) 文件以获取完整的许可证文本。

This project is licensed under the GNU Affero General Public License v3.0.
See [LICENSE](LICENSE) for full text.