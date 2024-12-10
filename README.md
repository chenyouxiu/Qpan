# Qpan 文件管理系统

Qpan 是一个基于 Go 和 Gin 框架的文件管理系统，允许用户上传、下载、删除和管理文件。

## 功能

- 用户身份验证
- 文件上传和下载
- 文件删除和重命名
- 文件搜索和分类
- 文件夹管理
- 文件统计信息

## 安装

1. 确保你已经安装了 Go 语言环境。
2. 克隆项目到本地：
   ```bash
   git clone https://github.com/yourusername/qpan.git
   cd qpan
   ```
3. 安装依赖：
   ```bash
   go mod tidy
   ```
4. 配置数据库连接（在 `config` 文件夹中）。
5. 运行数据库迁移：
   ```bash
   go run main.go migrate
   ```
6. 启动服务器：
   ```bash
   go run main.go
   ```

## 使用

- 启动服务器后，访问 `http://localhost:8080`。
- 使用 Postman 或其他 API 客户端进行 API 测试。

### API 端点

- **上传文件**: `POST /upload`
- **下载文件**: `GET /download/:id`
- **删除文件**: `DELETE /delete/:id`
- **搜索文件**: `GET /search`
- **获取文件列表**: `GET /files`
- **获取文件夹内容**: `GET /folders/:id`

## 贡献

欢迎任何形式的贡献！请遵循以下步骤：

1. Fork 本项目。
2. 创建你的特性分支 (`git checkout -b feature/YourFeature`)。
3. 提交你的更改 (`git commit -m 'Add some feature'`)。
4. 推送到分支 (`git push origin feature/YourFeature`)。
5. 创建一个新的 Pull Request。

## 许可证

本项目使用 MIT 许可证，详细信息请查看 [LICENSE](LICENSE) 文件。