# Cookie 配置指南

## 为什么需要 Cookie？

某些视频平台（如抖音、TikTok等）需要登录才能访问和下载视频。使用Cookie可以让 yt-dlp 模拟已登录状态。

## 配置步骤

### 方法1：使用浏览器扩展（推荐）

#### 1. 安装浏览器扩展

**Chrome/Edge浏览器：**
1. 访问 [Chrome Web Store](https://chrome.google.com/webstore)
2. 搜索并安装 "Get cookies.txt LOCALLY" 扩展
3. 安装后，浏览器右上角会出现扩展图标

**Firefox浏览器：**
1. 访问 [Firefox Add-ons](https://addons.mozilla.org)
2. 搜索并安装 "Get cookies.txt LOCALLY" 扩展

#### 2. 导出抖音Cookies

1. 在浏览器中打开 [抖音网页版](https://www.douyin.com)
2. 登录你的抖音账号
3. 点击浏览器扩展图标
4. 选择 "Current Site" 或 "All Cookies"
5. 点击 "Export" 下载 `cookies.txt` 文件
6. 将文件保存到项目目录：`/Users/terry/Documents/trae_code/batch_download_videos/cookies.txt`

### 方法2：使用开发者工具（手动导出）

#### 1. 打开开发者工具

1. 在浏览器中打开 [抖音网页版](https://www.douyin.com)
2. 登录你的抖音账号
3. 按 `F12` 或右键选择"检查"打开开发者工具
4. 切换到 "Application" (Chrome) 或 "存储" (Firefox) 标签

#### 2. 导出Cookies

1. 在左侧找到 "Cookies" → "https://www.douyin.com"
2. 右键点击任意cookie
3. 选择 "Copy all cookies as JSON" 或类似选项
4. 将内容保存为 `cookies.txt` 文件

**cookies.txt 格式示例：**
```
# Netscape HTTP Cookie File
# This is a generated file! Do not edit.

.douyin.com	TRUE	/	FALSE	0	sessionid	你的sessionid值
.douyin.com	TRUE	/	FALSE	0	passport_csrf_token	你的token值
```

## 使用方法

### 1. 将 cookies.txt 放在项目根目录

确保 `cookies.txt` 文件位于：
```
/Users/terry/Documents/trae_code/batch_download_videos/cookies.txt
```

### 2. 运行下载工具

程序会自动检测 `cookies.txt` 文件，如果存在则自动使用：

```bash
# 下载抖音视频（自动使用cookies）
./batch_download resource_urls/douyin_test.txt 720
```

### 3. 查看日志

如果Cookie被正确使用，会看到：
```
2026/01/08 15:30:00 使用Cookie文件: cookies.txt
2026/01/08 15:30:01 开始下载: 视频标题 (ID: xxx, 网站: douyin, 分辨率: 720)
```

## 测试Cookie是否有效

```bash
# 测试是否能正常获取视频信息
yt-dlp --cookies cookies.txt --dump-json "https://www.douyin.com/video/视频ID"
```

## 注意事项

1. **Cookie有效期**：Cookie会过期，需要定期更新
2. **安全性**：不要分享你的cookies.txt文件，包含登录凭证
3. **平台限制**：某些平台可能限制Cookie的使用
4. **文件权限**：确保cookies.txt文件有读取权限

## 常见问题

### Q: Cookie过期了怎么办？

A: 重新导出Cookie文件，覆盖旧的cookies.txt

### Q: 不使用Cookie能下载吗？

A: 某些平台（如抖音）需要Cookie，其他平台（如YouTube、Bilibili）不需要

### Q: 如何保护Cookie安全？

A: 
- 不要将cookies.txt提交到Git
- 不要分享给他人
- 定期更新Cookie

### Q: 多个平台需要不同的Cookie吗？

A: 是的，每个平台的Cookie不同。目前工具使用统一的cookies.txt文件，如果需要支持多个平台，可以分别配置。

## 其他平台Cookie配置

### TikTok
- 访问：https://www.tiktok.com
- 导出Cookie方法相同

### 其他需要登录的平台
参考相同步骤，在对应平台登录后导出Cookie
