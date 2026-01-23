# 批量视频下载工具 v2.0 - 混合架构

## 概述

本项目是一个功能强大的批量视频下载工具，采用混合架构设计，支持多种下载模式和视频平台。

### 混合架构特点

- **YouTube 专用下载器**：使用原生 Go 库，性能优化，下载速度快
- **多平台下载器**：基于 yt-dlp，支持 9+ 视频平台
- **统一接口设计**：模块化架构，易于扩展和维护
- **灵活配置**：支持配置文件和命令行参数

## 功能特点

### 核心功能

- **混合下载架构**：YouTube 专用下载器 + 多平台下载器
- **智能去重**：只记录视频ID，文件移动不影响下载状态
- **并发下载**：可配置的并发数，提高下载效率
- **错误重试**：自动重试机制，支持指数退避延迟
- **进度显示**：YouTube 下载显示实时进度条
- **多级别日志**：DEBUG/INFO/WARN/ERROR，支持文件和控制台输出
- **配置管理**：支持 JSON 配置文件和命令行参数
- **自动清理**：自动清理 0 字节的失败文件
- **下载记录**：自动生成和更新下载记录文档
- **自定义输出文件名**：支持配置文件名模板，如 `%(upload_date)s_%(title)s.%(ext)s`
- **Meta 文件生成**：自动生成包含视频标题和标签的 `.txt` 文件
- **任务队列管理**：支持任务暂停、恢复和取消操作
- **代理支持**：可配置网络代理，提高在不同网络环境下的下载成功率
- **文件名长度限制**：可配置文件名最大长度，避免文件名过长导致的问题

### YouTube 专用下载器

- **性能优化**：使用原生 Go 库，无需外部命令
- **进度条**：实时显示下载进度、百分比和预计完成时间
- **智能格式选择**：自动选择最接近目标分辨率的视频格式
- **并发安全**：线程安全的下载管理

### 多平台下载器

- **广泛支持**：支持 YouTube、抖音、微博、Bilibili、TikTok、Vimeo、Instagram、Twitter、Facebook 等
- **基于 yt-dlp**：利用 yt-dlp 强大的下载能力
- **Cookie 支持**：支持 Cookie 文件，可下载需要登录的视频
- **自动分类**：根据 URL 自动识别视频来源网站

## 安装指南

### 前置要求

- **Go 1.16+**：编译和运行程序
- **yt-dlp**：多平台下载依赖（可选，仅使用多平台下载器时需要）
- **ffmpeg**：高分辨率视频下载（可选，720p 及以上需要）

### 安装 Go

#### macOS

```bash
brew install go
```

#### Linux

```bash
sudo apt-get install golang
# 或
sudo yum install golang
```

### 安装 yt-dlp

#### macOS

```bash
brew install yt-dlp
```

#### Linux

```bash
sudo apt-get install yt-dlp
# 或
sudo yum install yt-dlp
```

#### 更新 yt-dlp

```bash
yt-dlp -U
```

### 安装 ffmpeg

#### macOS

```bash
brew install ffmpeg
```

#### Linux

```bash
sudo apt-get install ffmpeg
# 或
sudo yum install ffmpeg
```

### 验证安装

```bash
# 验证 Go
go version

# 验证 yt-dlp
yt-dlp --version

# 验证 ffmpeg
ffmpeg -version
```

### 编译程序

```bash
# 进入项目目录
cd batch_download_videos

# 下载依赖
go mod download

# 编译程序
go build -o batch_download main.go

编译成功后，会生成 `batch_download` 可执行文件。

## 使用方法

### 基本用法

#### 方式一：使用默认配置

```bash
./batch_download
```

使用配置文件中的默认设置（分辨率：720p，下载器：multi）

#### 方式二：指定分辨率

```bash
# 下载 720p 视频（推荐）
./batch_download -r 720

# 下载 1080p 视频
./batch_download -r 1080

# 下载 480p 视频
./batch_download -r 480

# 下载 360p 视频
./batch_download -r 360
```

#### 方式三：指定下载器

```bash
# 使用 YouTube 专用下载器（性能更好）
./batch_download -d youtube

# 使用多平台下载器（支持更多平台）
./batch_download -d multi
```

#### 方式四：指定配置文件

```bash
# 使用自定义配置文件
./batch_download -c my_config.json
```

#### 方式五：指定日志级别

```bash
# 启用调试日志
./batch_download -log-level debug

# 只显示错误
./batch_download -log-level error
```

#### 方式六：下载指定文件

```bash
# 下载指定文件中的 URL
./batch_download -f resource_urls/example.txt

# 指定分辨率和下载器
./batch_download -f resource_urls/example.txt -r 1080 -d youtube
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-r` | 视频分辨率（360/480/720/1080） | 从配置文件读取 |
| `-f` | URL 文件路径 | 自动扫描 resource_urls 目录 |
| `-d` | 下载器类型（youtube/multi） | 从配置文件读取 |
| `-c` | 配置文件路径 | config.json |
| `-log` | 日志目录 | logs |
| `-log-level` | 日志级别（debug/info/warn/error） | info |
| `-help` | 显示帮助信息 | - |
| `-version` | 显示版本信息 | - |

### 配置文件

配置文件使用 JSON 格式，默认路径为 `config.json`。

```json
{
  "batch_size": 10,
  "max_concurrency": 3,
  "timeout_per_video": "1h0m0s",
  "max_retries": 3,
  "base_retry_delay": "2s",
  "default_output_dir": "Output",
  "resource_urls_dir": "resource_urls",
  "cookie_file": "cookies.txt",
  "index_file": ".video_downloaded.index",
  "record_file": "下载记录.md",
  "default_resolution": "720",
  "default_downloader": "multi",
  "output_template": "%(upload_date)s_%(title)s.%(ext)s",
  "filename_max_length": 200,
  "generate_meta_file": true,
  "recode_video": "mp4",
  "max_concurrent_downloads": 3,
  "proxy": "",
  "limit_rate": "",
  "ffmpeg_path": ""
}
```

#### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `batch_size` | 每批次处理的 URL 数量 | 10 |
| `max_concurrency` | 最大并发下载数 | 3 |
| `timeout_per_video` | 单个视频下载超时时间 | 1h0m0s |
| `max_retries` | 最大重试次数 | 3 |
| `base_retry_delay` | 基础重试延迟 | 2s |
| `default_output_dir` | 默认输出目录 | Output |
| `resource_urls_dir` | URL 资源目录 | resource_urls |
| `cookie_file` | Cookie 文件路径 | cookies.txt |
| `index_file` | 索引文件路径 | .video_downloaded.index |
| `record_file` | 下载记录文件路径 | 下载记录.md |
| `default_resolution` | 默认分辨率 | 720 |
| `default_downloader` | 默认下载器 | multi |
| `output_template` | 自定义输出文件名模板 | `%(upload_date)s_%(title)s.%(ext)s` |
| `generate_meta_file` | 是否生成 Meta 文件 | true |
| `recode_video` | 视频格式转换目标格式 | "" |
| `max_concurrent_downloads` | 最大并发下载数量（yt-dlp） | 3 |
| `proxy` | 网络代理设置 | "" |
| `limit_rate` | 下载速度限制 | "" |
| `filename_max_length` | 文件名最大长度限制 | 200 |
| `ffmpeg_path` | ffmpeg 可执行文件路径 | "" |

## 支持的平台

### YouTube 专用下载器

| 平台 | URL 模式 | 状态 |
|--------|-----------|------|
| YouTube | youtube.com, youtu.be | ✅ |

### 多平台下载器

| 平台 | URL 模式 | 状态 |
|--------|-----------|------|
| YouTube | youtube.com, youtu.be | ✅ |
| 抖音 | douyin.com | ✅ |
| 微博 | weibo.com | ✅ |
| Bilibili | bilibili.com | ✅ |
| TikTok | tiktok.com | ✅ |
| Vimeo | vimeo.com | ✅ |
| Instagram | instagram.com | ✅ |
| Twitter | twitter.com, x.com | ✅ |
| Facebook | facebook.com | ✅ |

## 下载器对比

| 特性 | YouTube 专用下载器 | 多平台下载器 |
|------|------------------|--------------|
| 支持平台 | 仅 YouTube | 9+ 平台 |
| 依赖 | Go 库 | yt-dlp |
| 下载速度 | 快 | 取决于 yt-dlp |
| 进度显示 | 实时进度条 | 文本日志 |
| Cookie 支持 | 不支持 | ✅ 支持 |
| 推荐场景 | 大量 YouTube 视频 | 多平台视频 |

### 选择建议

- **仅下载 YouTube 视频**：使用 YouTube 专用下载器，性能更好
- **需要下载多个平台视频**：使用多平台下载器
- **需要 Cookie 认证**：使用多平台下载器
- **不确定使用哪个**：先尝试 YouTube 专用下载器，不满足再使用多平台下载器

## 目录结构

```
batch_download_videos/
├── batch_download          # 编译后的可执行文件
├── main.go                # 主程序入口
├── go.mod                 # Go 模块文件
├── go.sum                 # Go 依赖锁定文件
├── config/                # 配置管理模块
│   └── config.go
├── downloader/            # 下载器模块
│   ├── downloader.go          # 下载器接口定义
│   ├── youtube_downloader.go   # YouTube 专用下载器
│   └── multi_downloader.go    # 多平台下载器
├── indexer/               # 索引管理模块
│   ├── indexer.go
│   └── indexer_test.go
├── logger/                # 日志系统模块
│   └── logger.go
├── utils/                 # 工具函数模块
│   ├── utils.go
│   └── utils_test.go
├── resource_urls/          # URL 资源目录
│   ├── example.txt
│   └── ...
├── Output/                 # 下载输出目录（自动创建）
│   ├── .video_downloaded.index  # 下载索引文件
│   ├── youtube/              # YouTube 视频目录
│   │   └── 2026-01/       # 按月分类
│   ├── douyin/              # 抖音视频目录
│   │   └── 2026-01/
│   ├── weibo/               # 微博视频目录
│   │   └── 2026-01/
│   └── ...
├── logs/                  # 日志目录（自动创建）
│   └── download_20260112_150000.log
├── config.json            # 配置文件（可选）
├── cookies.txt            # Cookie 文件（可选）
├── README.md             # 本文档
├── README_v1.0.md        # v1.0 版本文档（备份）
├── CODE_REVIEW.md        # 代码审查报告
├── PROJECT_STATUS.md      # 项目状态总结
└── COOKIE_GUIDE.md       # Cookie 配置指南
```

## 使用流程

### 1. 准备 URL 文件

将收集到的视频 URL 放入 `resource_urls` 目录中的任意文本文件。支持以下扩展名：
- `.txt`
- `.url`
- `.list`

每个 URL 占一行，例如：

```
https://www.youtube.com/watch?v=rFejpH_tAHM
https://www.douyin.com/video/123456
https://www.bilibili.com/video/BV1xx411c7mD
```

### 2. （可选）创建配置文件

系统会在首次运行时自动生成默认的 `config.json` 文件。如果需要自定义配置，可以在项目根目录创建或修改 `config.json` 文件：

```json
{
  "default_resolution": "1080",
  "default_downloader": "youtube",
  "max_concurrency": 5
}
```

**注意**：当配置文件不存在时，系统会自动生成包含所有默认值的配置文件，确保程序能够正常运行。

### 3. （可选）配置 Cookie（用于需要登录的视频）

如果需要下载需要登录的视频（如抖音），请参考 [COOKIE_GUIDE.md](file:///Users/terry/Documents/trae_code/batch_download_videos/COOKIE_GUIDE.md) 配置 Cookie 文件。

### 4. 运行程序

```bash
# 使用默认配置
./batch_download

# 或指定参数
./batch_download -r 1080 -d youtube
```

### 5. 查看下载结果

下载的视频会保存在 `Output` 目录中，按网站和月份分类。

## 日志说明

### 日志级别

| 级别 | 说明 | 使用场景 |
|--------|------|----------|
| DEBUG | 详细调试信息 | 开发调试 |
| INFO | 正常信息 | 日常使用（默认） |
| WARN | 警告信息 | 需要注意的问题 |
| ERROR | 错误信息 | 下载失败 |

### 日志输出

程序会同时输出到：
- **控制台**：实时显示日志
- **日志文件**：保存在 `logs` 目录，文件名格式为 `download_YYYYMMDD_HHMMSS.log`

### 日志示例

```
[INFO] 使用分辨率: 720p
[INFO] 使用下载器: multi
[INFO] 开始扫描 resource_urls 目录...
[INFO] 找到 3 个 URL 文件
[INFO] 开始处理文件: resource_urls/example.txt (共 5 个URL)
[INFO] 开始处理 5 个URL，最大并发数: 3
[DEBUG] [1/5] 开始下载: https://www.youtube.com/watch?v=xxx
[INFO] 下载成功: 视频标题 (ID: xxx, 重试: 0, 大小: 125.3 MB)
[INFO] 下载完成统计: 成功=3, 失败=1, 跳过=1, 总计=5
[INFO] 所有任务完成！
```

## 分辨率说明

| 分辨率 | 参数 | 文件大小 | 下载速度 | 适用场景 |
|--------|------|----------|----------|----------|
| 1080p | `1080` | 1GB+ | 较慢 | 高清观看 |
| 720p | `720` | 600-700MB | 中等 | 平衡选择 ⭐ |
| 480p | `480` | 300-400MB | 较快 | 节省空间 |
| 360p | `360` | 150-200MB | 最快 | 快速预览 |

### 注意事项

- **720p 及以上**：需要安装 ffmpeg 进行音视频合并
- **480p 和 360p**：可以直接下载，不需要 ffmpeg
- **YouTube 专用下载器**：自动选择最佳格式，可能不需要 ffmpeg
- **多平台下载器**：高分辨率视频需要 ffmpeg

## 常见问题

### Q: 如何重新下载已下载的视频？

A: 删除 `.video_downloaded.index` 文件，或删除其中的对应记录。

### Q: 下载失败怎么办？

A: 程序会自动重试最多 3 次。如果仍然失败，可以：
- 检查网络连接
- 降低分辨率（如使用 720p）
- 检查是否被平台限流
- 使用多平台下载器（如果使用的是 YouTube 专用下载器）

### Q: 如何选择合适的下载器？

A: 根据需求选择：
- **大量 YouTube 视频**：使用 YouTube 专用下载器，性能更好
- **多平台视频**：使用多平台下载器
- **需要 Cookie 认证**：使用多平台下载器

### Q: 如何下载需要登录的视频（如抖音）？

A: 需要配置 Cookie 文件：
1. 安装浏览器 Cookie 导出插件
2. 导出 Cookie 到 `cookies.txt` 文件
3. 运行程序（多平台下载器会自动使用 Cookie）

详细步骤请参考 [COOKIE_GUIDE.md](file:///Users/terry/Documents/trae_code/batch_download_videos/COOKIE_GUIDE.md)。

### Q: 下载超时怎么办？

A: 可以在配置文件中调整 `timeout_per_video` 参数，默认为 1 小时。

### Q: 如何限制下载速度？

A: 可以在配置文件中降低 `max_concurrency` 参数，或使用系统工具限制带宽。

### Q: 文件移动会影响下载状态吗？

A: 不会。索引只记录视频 ID，不依赖文件路径。文件移动或重命名不会影响。

### Q: 可以同时使用两个下载器吗？

A: 不可以。每次运行只能选择一个下载器。可以分别运行两次程序，使用不同的下载器。

### Q: 如何查看下载记录？

A: 程序会自动生成 `Output/下载记录.md` 文件，包含所有已下载视频的详细信息。

### Q: 日志文件在哪里？

A: 日志文件保存在 `logs` 目录，文件名格式为 `download_YYYYMMDD_HHMMSS.log`。

### Q: 如何启用调试日志？

A: 使用 `-log-level debug` 参数：

```bash
./batch_download -log-level debug
```

## 性能优化建议

### 提高下载速度

1. **增加并发数**：在配置文件中设置 `max_concurrency` 为 5-10
2. **使用 YouTube 专用下载器**：仅下载 YouTube 时性能更好
3. **降低分辨率**：480p 或 360p 下载更快

### 节省磁盘空间

1. **使用较低分辨率**：720p 平衡画质和文件大小
2. **定期清理**：删除不需要的视频文件
3. **压缩存储**：使用压缩工具归档旧视频

### 避免被限流

1. **控制并发数**：不要设置过高的并发数（建议 3-5）
2. **使用合适的下载器**：YouTube 专用下载器更稳定
3. **分批下载**：不要一次性下载大量视频

## 技术文档

- **[CODE_REVIEW.md](file:///Users/terry/Documents/trae_code/batch_download_videos/CODE_REVIEW.md)** - 详细的代码审查报告
- **[PROJECT_STATUS.md](file:///Users/terry/Documents/trae_code/batch_download_videos/PROJECT_STATUS.md)** - 项目状态总结
- **[COOKIE_GUIDE.md](file:///Users/terry/Documents/trae_code/batch_download_videos/COOKIE_GUIDE.md)** - Cookie 配置指南

## 开发信息

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test ./... -cover

# 运行特定包的测试
go test ./utils/...
```

### 代码检查

```bash
# 代码格式化
gofmt -w .

# 静态检查
go vet ./...

# 编译检查
go build -o batch_download main.go
```

### 项目结构

```
batch_download_videos/
├── config/          # 配置管理
├── downloader/       # 下载器实现
├── indexer/          # 索引管理
├── logger/           # 日志系统
└── utils/           # 工具函数
```

## 版本历史

### v2.4.0 (2026-01-23)

**功能更新**：配置文件自动生成和错误处理优化

- ✅ 添加配置文件自动生成功能，当配置文件不存在时自动生成默认配置
- ✅ 优化配置加载逻辑，确保配置文件缺失时的平滑处理
- ✅ 提高系统的容错能力，减少因配置问题导致的运行失败
- ✅ 完善测试覆盖，验证核心功能（断点续传、格式转换、超时、重试、代理、限速）

### v2.3.0 (2026-01-22)

**功能更新**：ffmpeg 配置支持和代码优化

- ✅ 添加 `ffmpeg_path` 配置项，支持自定义 ffmpeg 路径
- ✅ 优化 YouTube 下载器的 ffmpeg 路径检测逻辑
- ✅ 修复配置文件中 `ResourceUrlsDir` 字段的拼写错误
- ✅ 确保所有下载器都使用一致的配置管理
- ✅ 优化代码结构，提高可维护性

### v2.2.0 (2026-01-22)

**功能更新**：任务队列管理、代理支持和完善文档

- ✅ 实现任务队列管理功能，支持任务暂停、恢复和取消
- ✅ 添加代理支持，提高在不同网络环境下的下载成功率
- ✅ 实现文件名长度限制，避免文件名过长导致的问题
- ✅ 完善智能重试机制，采用指数退避策略
- ✅ 编写详细的 API 参考文档
- ✅ 全面更新 README.md 文档，包含所有新功能说明
- ✅ 修复 YouTube 下载器的代理设置问题
- ✅ 确保多平台下载器的代理设置正确传递

### v2.1.0 (2026-01-21)

**功能更新**：自定义输出文件名和 Meta 文件生成

- ✅ 实现自定义输出文件名功能，支持模板变量
- ✅ 实现 Meta 文件生成功能，包含视频标题和标签
- ✅ 修复 yt-dlp 版本兼容性问题
- ✅ 优化多平台下载器的错误处理
- ✅ 添加新的配置项：output_template、generate_meta_file、recode_video 等

### v2.0.0 (2026-01-12)

**重大更新**：混合架构

- ✅ 实现 YouTube 专用下载器（使用 Go 库）
- ✅ 实现多平台下载器（基于 yt-dlp）
- ✅ 添加配置文件支持（JSON 格式）
- ✅ 添加多级别日志系统
- ✅ 添加下载进度条（YouTube 专用下载器）
- ✅ 优化并发控制（信号量模式）
- ✅ 优化索引管理（只记录视频 ID）
- ✅ 添加单元测试（utils 和 indexer 包）
- ✅ 改进错误处理和重试机制
- ✅ 添加 0 字节文件清理

### v1.0.0

**初始版本**：分离的下载工具

- batch_dl：YouTube 专用下载工具
- batch_download：多平台下载工具

## 注意事项

- 请遵守各视频平台的服务条款和版权法律
- 仅下载您有权下载的内容
- 避免高频请求，以免被平台限流
- 定期清理不需要的视频文件，节省磁盘空间
- Cookie 文件包含敏感信息，请妥善保管

## 高级功能

### 任务队列管理

任务队列管理功能允许用户控制下载任务的生命周期，支持暂停、恢复和取消操作。

#### 任务状态

| 状态 | 描述 |
|------|------|
| `pending` | 任务已创建，等待执行 |
| `downloading` | 任务正在执行中 |
| `paused` | 任务已暂停，可恢复 |
| `completed` | 任务执行完成 |
| `canceled` | 任务已取消 |

#### 使用方法

任务队列管理功能通过 `TaskManager` 实现，在程序运行过程中，用户可以通过以下方式控制任务：

1. **暂停任务**：当任务正在下载时，可以暂停该任务
2. **恢复任务**：当任务被暂停后，可以恢复该任务的执行
3. **取消任务**：可以取消正在执行或等待执行的任务

#### 状态持久化

任务状态会在执行过程中自动持久化到文件中，程序重启后会恢复之前的任务状态。

### 文件命名模板选项

文件命名模板功能允许用户自定义下载文件的命名规则，支持多种变量占位符。

#### 支持的变量

| 变量 | 描述 | 示例 |
|------|------|------|
| `%(title)s` | 视频标题 | `Example Video` |
| `%(id)s` | 视频ID | `dQw4w9WgXcQ` |
| `%(upload_date)s` | 上传日期 | `20230101` |
| `%(ext)s` | 文件扩展名 | `mp4` |
| `%(resolution)s` | 视频分辨率 | `720p` |
| `%(author)s` | 视频作者 | `Example Channel` |
| `%(date)s` | 当前日期 | `20240101` |

#### 示例模板

```
# 日期 + 标题
%(upload_date)s_%(title)s.%(ext)s

# ID + 标题
%(id)s_%(title)s.%(ext)s

# 作者 + 标题
%(author)s_%(title)s.%(ext)s

# 分辨率 + 标题
%(resolution)s_%(title)s.%(ext)s
```

#### 长度限制

文件名长度会受到 `filename_max_length` 配置项的限制，默认为 200 个字符。超过限制的文件名会被自动截断。

### 代理支持

代理支持功能允许用户在不同网络环境下通过代理服务器进行下载，提高下载成功率。

#### 配置方法

在配置文件中设置 `proxy` 字段：

```json
{
  "proxy": "http://127.0.0.1:7890"
}
```

#### 支持的代理类型

- **HTTP 代理**：`http://proxy.example.com:8080`
- **HTTPS 代理**：`https://proxy.example.com:8443`

#### 工作原理

- **YouTube 专用下载器**：通过设置 `http.Client` 的 `Transport` 字段来使用代理
- **多平台下载器**：通过传递 `--proxy` 参数给 yt-dlp 来使用代理

### 智能重试机制

智能重试机制采用指数退避策略，在下载失败时自动重试，提高下载成功率。

#### 重试策略

1. **基础延迟**：由 `base_retry_delay` 配置项指定，默认为 2 秒
2. **指数退避**：每次重试的延迟时间会翻倍
3. **最大重试次数**：由 `max_retries` 配置项指定，默认为 3 次

#### 重试场景

- 网络连接失败
- 服务器响应错误
- 下载超时
- 其他临时错误

## API 参考

### 核心接口

#### Downloader 接口

```go
// Downloader 定义了下载器的通用接口
type Downloader interface {
    // Download 下载指定 URL 的视频
    // url: 视频 URL
    // resolution: 目标分辨率
    // outputDir: 输出目录
    // 返回值: 下载结果和错误
    Download(url string, resolution string, outputDir string) (*DownloadResult, error)
    
    // GetName 获取下载器名称
    // 返回值: 下载器名称
    GetName() string
}
```

#### TaskManager 接口

```go
// TaskManager 定义了任务管理的接口
type TaskManager interface {
    // AddTask 添加下载任务
    // url: 视频 URL
    // resolution: 目标分辨率
    // outputDir: 输出目录
    // 返回值: 任务 ID 和错误
    AddTask(url string, resolution string, outputDir string) (string, error)
    
    // PauseTask 暂停指定任务
    // taskID: 任务 ID
    // 返回值: 错误
    PauseTask(taskID string) error
    
    // ResumeTask 恢复指定任务
    // taskID: 任务 ID
    // 返回值: 错误
    ResumeTask(taskID string) error
    
    // CancelTask 取消指定任务
    // taskID: 任务 ID
    // 返回值: 错误
    CancelTask(taskID string) error
    
    // GetTaskStatus 获取指定任务的状态
    // taskID: 任务 ID
    // 返回值: 任务状态和错误
    GetTaskStatus(taskID string) (TaskStatus, error)
    
    // Start 开始执行任务队列
    // 返回值: 错误
    Start() error
}
```

### 具体实现

#### YouTubeDownloader

```go
// YouTubeDownloader YouTube 专用下载器
type YouTubeDownloader struct {
    client     *youtube.Client
    config     *config.Config
    indexer    *indexer.Indexer
    outputDir  string
}

// NewYouTubeDownloader 创建新的 YouTube 下载器
// cfg: 配置对象
// idx: 索引器对象
// 返回值: YouTubeDownloader 实例
func NewYouTubeDownloader(cfg *config.Config, idx *indexer.Indexer) *YouTubeDownloader

// Download 下载 YouTube 视频
// url: 视频 URL
// resolution: 目标分辨率
// outputDir: 输出目录
// 返回值: 下载结果和错误
func (yd *YouTubeDownloader) Download(url string, resolution string, outputDir string) (*DownloadResult, error)
```

#### MultiPlatformDownloader

```go
// MultiPlatformDownloader 多平台下载器
type MultiPlatformDownloader struct {
    config    *config.Config
    indexer   *indexer.Indexer
    outputDir string
}

// NewMultiPlatformDownloader 创建新的多平台下载器
// cfg: 配置对象
// idx: 索引器对象
// 返回值: MultiPlatformDownloader 实例
func NewMultiPlatformDownloader(cfg *config.Config, idx *indexer.Indexer) *MultiPlatformDownloader

// Download 下载多平台视频
// url: 视频 URL
// resolution: 目标分辨率
// outputDir: 输出目录
// 返回值: 下载结果和错误
func (mpd *MultiPlatformDownloader) Download(url string, resolution string, outputDir string) (*DownloadResult, error)
```

### 配置管理

#### Config 结构体

```go
// Config 应用配置结构体
type Config struct {
    BatchSize             int           `json:"batch_size"`
    MaxConcurrency        int           `json:"max_concurrency"`
    TimeoutPerVideo       time.Duration `json:"timeout_per_video"`
    MaxRetries            int           `json:"max_retries"`
    BaseRetryDelay        time.Duration `json:"base_retry_delay"`
    DefaultOutputDir      string        `json:"default_output_dir"`
    ResourceUrlsDir       string        `json:"resource_urls_dir"`
    CookieFile            string        `json:"cookie_file"`
    IndexFile             string        `json:"index_file"`
    RecordFile            string        `json:"record_file"`
    DefaultResolution     string        `json:"default_resolution"`
    DefaultDownloader     string        `json:"default_downloader"`
    OutputTemplate        string        `json:"output_template"`
    FilenameMaxLength     int           `json:"filename_max_length"`
    GenerateMetaFile      bool          `json:"generate_meta_file"`
    RecodeVideo           string        `json:"recode_video"`
    MaxConcurrentDownloads int          `json:"max_concurrent_downloads"`
    Proxy                 string        `json:"proxy"`
    LimitRate             string        `json:"limit_rate"`
}

// LoadConfig 从文件加载配置
// filePath: 配置文件路径
// 返回值: 配置对象和错误
func LoadConfig(filePath string) (*Config, error)

// SaveConfig 保存配置到文件
// filePath: 配置文件路径
// 返回值: 错误
func (c *Config) SaveConfig(filePath string) error
```

## 技术支持

如有问题或建议，请：
1. 查看 [CODE_REVIEW.md](file:///Users/terry/Documents/trae_code/batch_download_videos/CODE_REVIEW.md) 了解项目详情
2. 查看 [PROJECT_STATUS.md](file:///Users/terry/Documents/trae_code/batch_download_videos/PROJECT_STATUS.md) 了解项目状态
3. 检查日志文件获取详细错误信息
4. 提交 Issue 或 Pull Request

---

**项目版本**：v2.4.0（完整功能）
**最后更新**：2026-01-23
**Go 版本**：1.25.0
**许可证**：MIT
