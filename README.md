# 批量视频下载工具使用说明

## 概述

本项目提供两个视频下载工具：

1. **batch_dl** - YouTube 专用下载工具（基于 Go 库）
2. **batch_download** - 多平台下载工具（基于 yt-dlp）

## 功能特点

### batch_dl（YouTube 专用）

- **自动批量下载**：自动扫描 `resource_urls` 目录中的所有 URL 文件并批量下载
- **智能去重**：自动跳过已下载的视频，避免重复下载
- **并发下载**：支持并发下载，提高下载效率
- **断点续传**：每批次结束后保存索引，程序崩溃不会丢失进度
- **防限流机制**：内置防限流机制，避免被 YouTube 限制
- **多分辨率支持**：支持 1080p、720p、480p、360p 等多种分辨率下载
- **灵活参数**：支持通过命令行参数指定分辨率和文件
- **智能目录结构**：按网站和月份自动组织下载文件
- **自动清理**：自动清理下载失败的0字节文件
- **自动记录**：自动生成和更新下载记录文档

### batch_download（多平台支持）

- **多平台支持**：支持 YouTube、Bilibili、TikTok、Vimeo、Instagram、Twitter、Facebook 等多个视频平台
- **基于 yt-dlp**：利用 yt-dlp 强大的下载能力，支持更多网站
- **智能分类**：自动识别视频来源网站并分类存储
- **去重机制**：共享索引文件，避免重复下载
- **并发控制**：支持并发下载，提高效率
- **自动记录**：自动生成和更新下载记录文档
- **灵活参数**：支持指定文件和分辨率

## 安装指南

### 前置要求

- Go 1.16 或更高版本
- ffmpeg（可选，用于高分辨率视频下载）
- yt-dlp（用于 batch_download 多平台下载）

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

### 验证安装

```bash
# 验证 Go
go version

# 验证 ffmpeg
ffmpeg -version

# 验证 yt-dlp
yt-dlp --version
```

### 编译程序

```bash
# 进入项目目录
cd batch_download_videos

# 下载依赖
go mod download

# 编译 YouTube 专用工具
go build -o batch_dl batch_dl.go

# 编译多平台下载工具
go build -o batch_download batch_download.go
```

编译成功后，会生成以下可执行文件：
- `batch_dl` - YouTube 专用下载工具
- `batch_download` - 多平台下载工具

## 使用方法

### batch_dl（YouTube 专用）

#### 方式一：自动扫描（推荐）

直接运行脚本，使用默认分辨率（1080p）：

```bash
./batch_dl
```

脚本会自动扫描 `resource_urls` 目录中的所有 URL 文件并批量下载。

#### 方式二：指定分辨率

通过命令行参数指定下载分辨率：

```bash
# 下载 720p 视频（推荐，文件大小适中）
./batch_dl 720

# 下载 1080p 视频
./batch_dl 1080

# 下载 480p 视频
./batch_dl 480

# 下载 360p 视频
./batch_dl 360
```

#### 方式三：指定单个文件

```bash
# 使用默认分辨率下载指定文件
./batch_dl resource_urls/example.txt

# 指定分辨率下载指定文件
./batch_dl resource_urls/example.txt 720
```

### batch_download（多平台支持）

#### 支持的平台

- YouTube（youtube.com, youtu.be）
- 抖音（douyin.com）
- 微博（weibo.com）
- Bilibili（bilibili.com）
- TikTok（tiktok.com）
- Vimeo（vimeo.com）
- Instagram（instagram.com）
- Twitter（twitter.com, x.com）
- Facebook（facebook.com）
- 其他 yt-dlp 支持的网站

#### 方式一：自动扫描（推荐）

直接运行脚本，使用默认分辨率（720p）：

```bash
./batch_download
```

脚本会自动扫描 `resource_urls` 目录中的所有 URL 文件并批量下载。

#### 方式二：指定分辨率

```bash
# 下载 720p 视频（推荐）
./batch_download 720

# 下载 1080p 视频
./batch_download 1080

# 下载 480p 视频
./batch_download 480

# 下载 360p 视频
./batch_download 360
```

#### 方式三：指定单个文件

```bash
# 使用默认分辨率下载指定文件
./batch_download resource_urls/example.txt

# 指定分辨率下载指定文件
./batch_download resource_urls/example.txt 720
```

#### 示例

```bash
# 下载多平台视频
./batch_download resource_urls/multi_platform_test.txt 720
```

### 工具对比

| 特性 | batch_dl | batch_download |
|------|----------|----------------|
| 支持平台 | 仅 YouTube | 多平台（YouTube、抖音、微博、Bilibili、TikTok等） |
| 依赖 | Go + ffmpeg（可选） | Go + yt-dlp + ffmpeg（可选） |
| 下载速度 | 较快 | 取决于 yt-dlp |
| 稳定性 | 高 | 高 |
| 功能完整度 | YouTube 专用 | 多平台支持 |
| 推荐场景 | 仅下载 YouTube | 需要下载多个平台视频 |

### 选择建议

- **仅下载 YouTube 视频**：使用 `batch_dl`，速度更快
- **需要下载多个平台视频**：使用 `batch_download`，支持更多网站
- **不确定使用哪个**：先尝试 `batch_download`，如果不满意再使用 `batch_dl`

## 日志说明

### 日志输出

脚本运行时会实时输出日志信息，包括：

- **下载进度**：显示当前下载的视频和进度
- **错误信息**：显示下载失败的原因
- **警告信息**：显示需要注意的问题
- **完成信息**：显示下载完成的统计

### 查看日志

#### 实时查看

直接运行脚本，日志会实时显示在终端：

```bash
./batch_dl 720
```

#### 保存日志到文件

如果需要保存日志，可以重定向输出：

```bash
./batch_dl 720 2>&1 | tee download.log
```

### 常见日志信息

#### 正常下载日志

```
2026/01/07 12:18:08 使用分辨率: 720p
2026/01/07 12:18:08 开始处理第 1 批次 (共 3 个视频，已加载10个已下载索引)
2026/01/07 12:18:08 处理第 1/3 个URL: https://www.youtube.com/watch?v=xxx
2026/01/07 12:18:41 开始下载: 视频标题 (ID: xxx, 分辨率: hd720)
2026/01/07 12:18:45 下载完成: 视频标题 (ID: xxx)
2026/01/07 12:18:45 下载记录已更新: resource_urls/Output/下载记录.md
```

#### 跳过已下载日志

```
2026/01/07 12:18:08 跳过: 视频已下载: xxx (路径: resource_urls/Output/youtube/2026-01/视频.mp4)
```

#### 文件不存在警告

```
2026/01/07 12:18:38 警告: 索引中的文件不存在，已移除记录: resource_urls/Output/youtube/2026-01/视频.mp4 (ID: xxx)
```

#### 0字节文件清理

```
2026/01/07 11:57:21 发现0字节文件，删除: resource_urls/Output/youtube/2026-01/视频.mp4
```

#### 下载失败日志

```
2026/01/07 12:18:08 下载失败 https://www.youtube.com/watch?v=xxx: 下载失败: timeout
```

#### 重试日志

```
2026/01/07 12:18:08 首次下载失败，重试一次: 视频标题, 错误: timeout
```

### 日志级别说明

- **INFO**：正常信息（下载进度、完成等）
- **警告**：需要注意但不影响运行（文件不存在、分辨率降级等）
- **错误**：下载失败，需要重试

## 使用方法

### 方式一：自动扫描（推荐）

直接运行脚本，使用默认分辨率（1080p）：

```bash
./batch_dl
```

脚本会自动扫描 `resource_urls` 目录中的所有 URL 文件并批量下载。

### 方式二：指定分辨率

通过命令行参数指定下载分辨率：

```bash
# 下载 720p 视频（推荐，文件大小适中）
./batch_dl 720

# 下载 1080p 视频
./batch_dl 1080

# 下载 480p 视频
./batch_dl 480

# 下载 360p 视频
./batch_dl 360
```

### 方式三：指定单个文件

```bash
# 使用默认分辨率下载指定文件
./batch_dl resource_urls/example.txt

# 指定分辨率下载指定文件
./batch_dl resource_urls/example.txt 720
```

### 分辨率说明

| 分辨率 | 参数 | 文件大小 | 下载速度 | 适用场景 |
|--------|------|----------|----------|----------|
| 1080p | `1080` 或 `hd1080` | 1GB+ | 较慢 | 高清观看 |
| 720p | `720` 或 `hd720` | 600-700MB | 中等 | 平衡选择 ⭐ |
| 480p | `480` 或 `medium` | 300-400MB | 较快 | 节省空间 |
| 360p | `360` 或 `small` | 150-200MB | 最快 | 快速预览 |

## 目录结构

```
batch_download_videos/
├── batch_dl.go              # 主程序
├── batch_dl                 # 编译后的可执行文件
├── go.mod                   # Go 模块文件
├── go.sum                   # Go 依赖锁定文件
├── resource_urls/           # URL 资源目录
│   └── example.txt          # 示例 URL 文件
└── Output/                  # 下载输出目录（自动创建）
    ├── .youtube_downloaded.index  # 去重索引文件
    ├── youtube/              # YouTube 视频目录
    │   └── 2026-01/       # 按月分类
    │       ├── Me at the zoo.mp4
    │       └── Rick Astley - Never Gonna Give You Up.mp4
    └── tiktok/              # TikTok 视频目录（如支持）
        └── 2026-01/
            └── video1.mp4
```

### 目录结构说明

- **按网站分类**：自动识别视频来源网站（youtube、tiktok、bilibili等）
- **按月分类**：每个网站下按月份（格式：YYYY-MM）创建子目录
- **便于管理**：结构清晰，方便查找和归档

## 使用流程

### 1. 准备 URL 文件

将收集到的 YouTube URL 放入 `resource_urls` 目录中的任意文本文件。支持以下扩展名：
- `.txt`
- `.url`
- `.list`

每个 URL 占一行，例如：

```
https://www.youtube.com/watch?v=rFejpH_tAHM
https://www.youtube.com/watch?v=dQw4w9WgXcQ
https://www.youtube.com/watch?v=jNQXAC9IVRw
```

### 2. 运行脚本

```bash
./batch_dl
```

### 3. 查看下载结果

下载的视频会保存在 `resource_urls/Output/batch_*` 目录中。

## 工作原理

1. **扫描阶段**：扫描 `resource_urls` 目录中的所有 URL 文件
2. **去重检查**：读取 `.youtube_downloaded.index` 索引文件，检查哪些视频已下载
3. **批量下载**：只下载未下载的视频，每批 10 个视频
4. **智能分类**：根据 URL 自动识别网站并创建对应目录（youtube、tiktok等）
5. **按月归档**：每个网站下按月份（YYYY-MM）创建子目录
6. **错误重试**：每个视频下载失败后会自动重试一次
7. **自动清理**：下载前自动清理0字节的失败文件
8. **进度保存**：每批次结束后保存索引，防止程序崩溃丢失进度
9. **防限流**：每 5 批次后休眠 30 秒，避免被 YouTube 限流

## 配置说明

可以在代码中修改以下配置：

```go
const (
	batchSize         = 10                    // 每批次下载的视频数量
	defaultOutputDir  = "Output"              // 基础输出目录
	resourceURLsDir   = "resource_urls"       // URL资源目录
	qualityHighest    = "hd1080"              // 优先下载的最高分辨率
	fallbackQuality   = "hd720"               // 降级分辨率
	maxConcurrency    = 3                     // 单批次内并发下载数（避免被限流）
	timeoutPerVideo   = 60 * time.Minute      // 单个视频下载超时（大文件需要更长时间）
)
```

### HTTP 客户端配置

```go
client := &youtube.Client{
	HTTPClient: &http.Client{
		Timeout: 120 * time.Second,              // HTTP 客户端超时
		Transport: &http.Transport{
			MaxIdleConns:        10,             // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时
			DisableCompression:  false,           // 是否禁用压缩
			TLSHandshakeTimeout: 20 * time.Second, // TLS 握手超时
			ResponseHeaderTimeout: 60 * time.Second, // 响应头超时
		},
	},
}
```

## 分辨率下载说明

下载高分辨率视频（hd1080/hd720）需要安装 ffmpeg：

### macOS

```bash
brew install ffmpeg
```

### Linux

```bash
sudo apt-get install ffmpeg
# 或
sudo yum install ffmpeg
```

### 验证安装

```bash
ffmpeg -version
```

### 注意事项

- 720p 及以上分辨率需要 ffmpeg 进行音视频合并
- 480p 和 360p 分辨率可以直接下载，不需要 ffmpeg
- 如果未安装 ffmpeg，脚本会自动降级到不需要合并的格式

## 日常使用建议

1. **定期运行**：每天运行一次脚本，自动下载新视频
2. **批量收集**：将收集到的 URL 放入 `resource_urls` 目录
3. **分类管理**：可以创建不同的 URL 文件进行分类管理
4. **备份索引**：定期备份 `.youtube_downloaded.index` 文件
5. **选择合适分辨率**：
   - **日常推荐**：使用 `./batch_dl 720`，平衡文件大小和画质
   - **节省空间**：使用 `./batch_dl 480` 或 `./batch_dl 360`
   - **高清需求**：使用 `./batch_dl 1080`，但下载时间较长

## 示例

### 创建多个 URL 文件

```
resource_urls/
├── tech_videos.txt      # 技术类视频
├── music_videos.txt     # 音乐类视频
└── tutorial_videos.txt  # 教程类视频
```

### 每天运行脚本

```bash
# 使用 720p 分辨率下载所有新视频（推荐）
./batch_dl 720

# 可以设置 cron 任务每天自动运行
# 0 2 * * * cd /path/to/batch_download_videos && ./batch_dl 720
```

### 指定文件下载

```bash
# 下载特定文件中的视频
./batch_dl resource_urls/tech_videos.txt

# 指定分辨率下载特定文件
./batch_dl resource_urls/tech_videos.txt 720
```

### 不同场景的推荐用法

```bash
# 场景1：日常使用，平衡画质和文件大小
./batch_dl 720

# 场景2：高清需求，不介意下载时间长
./batch_dl 1080

# 场景3：快速预览，节省磁盘空间
./batch_dl 480

# 场景4：网络较慢，快速下载
./batch_dl 360
```

## 常见问题

### Q: 如何重新下载已下载的视频？

A: 删除 `.youtube_downloaded.index` 文件中的对应记录，或删除整个索引文件。

### Q: 下载失败怎么办？

A: 脚本会自动重试一次，如果仍然失败，可以：
- 检查网络连接
- 稍后重试
- 降低分辨率（如使用 `./batch_dl 720`）
- 检查是否被 YouTube 限流

### Q: 如何选择合适的分辨率？

A: 根据需求选择：
- **720p**：推荐，文件大小适中，画质良好
- **1080p**：高清画质，但文件大、下载慢
- **480p/360p**：节省空间，下载快，适合快速预览

### Q: 下载超时怎么办？

A: 如果下载大文件时出现超时：
- 脚本已将超时时间设置为 60 分钟
- 可以在代码中修改 `timeoutPerVideo` 配置
- 降低分辨率可以减少文件大小和下载时间

### Q: 如何限制下载速度？

A: 可以在代码中添加下载速度限制配置，或使用系统工具限制带宽。

### Q: 支持其他视频网站吗？

A: 当前仅支持 YouTube，如需支持其他网站，需要修改代码使用其他下载库。

### Q: 批次下载失败会影响其他批次吗？

A: 不会。每个批次独立处理，即使某个批次失败，其他批次仍会继续。每批次结束后都会保存索引，确保进度不丢失。

### Q: 目录结构是什么样的？

A: 新的目录结构按网站和月份分类：
```
Output/
├── youtube/
│   └── 2026-01/
│       ├── 视频1.mp4
│       └── 视频2.mp4
└── tiktok/
    └── 2026-01/
        └── 视频3.mp4
```
这样的结构便于查找和管理，不会因为文件太多而混乱。

### Q: 如何处理0字节的失败文件？

A: 脚本会自动处理：
- 下载前自动检测并删除0字节文件
- 重新下载该视频
- 无需手动清理

### Q: 可以修改目录结构吗？

A: 可以。在代码中修改 `getWebsiteName()` 和 `getCurrentMonthDir()` 函数即可自定义目录结构。

## 注意事项

- 请遵守 YouTube 的服务条款和版权法律
- 仅下载您有权下载的内容
- 避免高频请求，以免被 YouTube 限流
- 定期清理不需要的视频文件，节省磁盘空间

## 技术支持

如有问题或建议，请检查日志输出或联系开发者。
