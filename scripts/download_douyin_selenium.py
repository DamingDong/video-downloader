#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
使用Selenium下载抖音视频的脚本
通过模拟浏览器行为来绕过抖音的反爬虫机制
"""

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time
import re
import json
import requests
import os
import sys

def download_douyin_video_selenium(video_url, cookie_file, output_dir='downloads'):
    """
    使用Selenium下载抖音视频
    
    Args:
        video_url: 抖音视频URL
        cookie_file: Cookie文件路径
        output_dir: 输出目录
        
    Returns:
        bool: 下载是否成功
    """
    print(f"开始使用Selenium下载抖音视频: {video_url}")
    
    # 确保输出目录存在
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    # 配置Chrome选项
    chrome_options = Options()
    chrome_options.add_argument('--headless')  # 无头模式
    chrome_options.add_argument('--disable-gpu')
    chrome_options.add_argument('--no-sandbox')
    chrome_options.add_argument('--disable-dev-shm-usage')
    chrome_options.add_argument('--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36')
    
    try:
        # 启动Chrome浏览器
        driver = webdriver.Chrome(options=chrome_options)
        print("Chrome浏览器启动成功")
        
        # 加载Cookie
        driver.get('https://www.douyin.com')
        time.sleep(2)  # 等待页面加载
        
        # 从Cookie文件中加载cookies
        cookies = _load_cookies(cookie_file)
        for name, value in cookies.items():
            driver.add_cookie({
                'name': name,
                'value': value,
                'domain': '.douyin.com',
                'path': '/',
                'expires': None
            })
        print(f"成功加载 {len(cookies)} 个cookie")
        
        # 访问视频URL
        print(f"访问视频URL: {video_url}")
        driver.get(video_url)
        
        # 等待页面加载完成
        print("等待页面加载完成...")
        time.sleep(5)  # 等待5秒，确保页面完全加载
        
        # 获取页面内容
        page_source = driver.page_source
        print(f"页面大小: {len(page_source)} 字符")
        
        # 保存页面内容到文件，以便分析
        with open(os.path.join(output_dir, 'douyin_selenium_page.html'), 'w', encoding='utf-8') as f:
            f.write(page_source)
        print(f"页面内容已保存到 {output_dir}/douyin_selenium_page.html")
        
        # 尝试提取视频URL
        video_url = _extract_video_url_from_page(page_source)
        if video_url:
            print(f"找到视频URL: {video_url}")
            
            # 下载视频
            video_filename = f"douyin_{int(time.time())}.mp4"
            video_path = os.path.join(output_dir, video_filename)
            
            print(f"开始下载视频到: {video_path}")
            
            # 使用requests下载视频
            headers = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
                'Referer': 'https://www.douyin.com/',
            }
            
            response = requests.get(video_url, headers=headers, stream=True, timeout=60)
            if response.status_code == 200:
                total_size = int(response.headers.get('content-length', 0))
                downloaded_size = 0
                
                with open(video_path, 'wb') as f:
                    for chunk in response.iter_content(chunk_size=8192):
                        if chunk:
                            f.write(chunk)
                            downloaded_size += len(chunk)
                            # 显示下载进度
                            if total_size > 0:
                                progress = (downloaded_size / total_size) * 100
                                print(f"下载进度: {progress:.2f}%", end='\r')
                
                print(f"\n视频下载成功: {video_path}")
                return True
            else:
                print(f"下载视频失败，状态码: {response.status_code}")
                return False
        else:
            print("无法提取视频URL")
            return False
            
    except Exception as e:
        print(f"下载过程中出错: {e}")
        return False
    finally:
        # 关闭浏览器
        if 'driver' in locals():
            driver.quit()
            print("Chrome浏览器已关闭")

def _load_cookies(cookie_file):
    """
    从Cookie文件中加载cookies
    
    Args:
        cookie_file: Cookie文件路径
        
    Returns:
        dict: cookies字典
    """
    cookies = {}
    try:
        with open(cookie_file, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#'):
                    parts = line.split('\t')
                    if len(parts) >= 7:
                        domain, _, path, secure, expiry, name, value = parts[:7]
                        cookies[name] = value
    except Exception as e:
        print(f"加载Cookie文件失败: {e}")
    return cookies

def _extract_video_url_from_page(page_source):
    """
    从页面中提取视频URL
    
    Args:
        page_source: 页面源代码
        
    Returns:
        str: 视频URL
    """
    # 尝试不同的方法提取视频URL
    patterns = [
        # 方法1: 查找playAddr或play_addr
        r'playAddr[\s\S]*?"([^"]+\.mp4[^"]*)"',
        r'play_addr[\s\S]*?"([^"]+\.mp4[^"]*)"',
        # 方法2: 查找video或videoList
        r'video[\s\S]*?"([^"]+\.mp4[^"]*)"',
        r'videoList[\s\S]*?"([^"]+\.mp4[^"]*)"',
        # 方法3: 查找url_list
        r'url_list[\s\S]*?"([^"]+\.mp4[^"]*)"',
        # 方法4: 直接查找mp4 URL
        r'https?://[^"]+\.mp4[^"]*',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, page_source)
        if match:
            video_url = match.group(1)
            # 确保URL是完整的
            if not video_url.startswith('http'):
                video_url = 'https:' + video_url
            return video_url
    
    return None

if __name__ == '__main__':
    if len(sys.argv) > 1:
        video_url = sys.argv[1]
        print(f"从命令行获取视频URL: {video_url}")
    else:
        video_url = 'https://www.douyin.com/video/73075767586867398190'
        print(f"使用默认视频URL: {video_url}")
    cookie_file = 'www.douyin.com_cookies.txt'
    output_dir = 'downloads'
    
    download_douyin_video_selenium(video_url, cookie_file, output_dir)
