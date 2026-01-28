#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
直接下载抖音视频的脚本
通过从HTML响应中提取视频URL来绕过yt-dlp的提取器
"""

import requests
import re
import json
import os
import urllib.parse

def load_cookies(cookie_file):
    """
    从Netscape格式的cookie文件中加载cookies
    """
    cookies = {}
    with open(cookie_file, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#'):
                parts = line.split('\t')
                if len(parts) >= 7:
                    domain, _, path, secure, expiry, name, value = parts[:7]
                    cookies[name] = value
    return cookies

def download_douyin_video(video_url, cookie_file, output_dir='downloads'):
    """
    下载抖音视频
    """
    print(f"正在下载抖音视频: {video_url}")
    print(f"使用Cookie文件: {cookie_file}")
    
    # 加载cookies
    cookies = load_cookies(cookie_file)
    print(f"加载了 {len(cookies)} 个cookie")
    
    # 设置请求头
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Referer': 'https://www.douyin.com/',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8',
        'Accept-Language': 'zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3',
        'Accept-Encoding': 'gzip, deflate, br',
        'DNT': '1',
        'Connection': 'keep-alive',
        'Upgrade-Insecure-Requests': '1'
    }
    
    try:
        # 发送请求获取页面
        response = requests.get(video_url, headers=headers, cookies=cookies, timeout=30)
        print(f"响应状态码: {response.status_code}")
        
        # 检查响应内容
        if len(response.text) == 0:
            print("响应内容为空")
            return False
        
        # 保存响应内容到文件，以便分析
        with open('douyin_response_new.html', 'w', encoding='utf-8') as f:
            f.write(response.text)
        print("响应内容已保存到 douyin_response_new.html")
        print(f"响应长度: {len(response.text)} 字符")
        
        # 尝试提取视频信息
        # 方法1: 查找__INITIAL_STATE__
        initial_state_match = re.search(r'window\.__INITIAL_STATE__\s*=\s*({[\s\S]*?});', response.text)
        if initial_state_match:
            print("发现__INITIAL_STATE__数据")
            try:
                initial_state = json.loads(initial_state_match.group(1))
                # 遍历查找视频信息
                if 'aweme' in initial_state:
                    print("发现aweme数据")
                    if isinstance(initial_state['aweme'], dict) and 'detail' in initial_state['aweme']:
                        detail = initial_state['aweme']['detail']
                        if 'video' in detail and 'play_addr' in detail['video']:
                            play_addr = detail['video']['play_addr']
                            if 'url_list' in play_addr:
                                video_urls = play_addr['url_list']
                                print(f"找到 {len(video_urls)} 个视频URL")
                                for i, url in enumerate(video_urls):
                                    print(f"视频URL {i+1}: {url}")
            except Exception as e:
                print(f"解析__INITIAL_STATE__失败: {e}")
        
        # 方法2: 查找aweme数据
        aweme_match = re.search(r'aweme\s*=\s*({[\s\S]*?});', response.text)
        if aweme_match:
            print("发现aweme数据")
            try:
                aweme_data = json.loads(aweme_match.group(1))
                print(f"aweme数据类型: {type(aweme_data)}")
            except Exception as e:
                print(f"解析aweme数据失败: {e}")
        
        # 方法3: 查找视频播放器配置
        player_match = re.search(r'playerConfig\s*=\s*({[\s\S]*?});', response.text)
        if player_match:
            print("发现playerConfig数据")
            try:
                player_config = json.loads(player_match.group(1))
                print(f"playerConfig数据类型: {type(player_config)}")
            except Exception as e:
                print(f"解析playerConfig失败: {e}")
        
        # 方法4: 直接查找视频URL
        video_url_match = re.findall(r'https?://[^"\'\s]+\.mp4[^"\'\s]*', response.text)
        if video_url_match:
            print(f"找到 {len(video_url_match)} 个视频URL")
            for i, url in enumerate(video_url_match[:5]):  # 只显示前5个
                print(f"视频URL {i+1}: {url}")
        
        # 方法5: 查找包含video的JSON数据
        print("\n=== 尝试使用更强大的提取方法 ===")
        # 查找所有可能的JSON数据块
        json_patterns = [
            r'\{[\s\S]*?"video"[\s\S]*?\}',
            r'\{[\s\S]*?"play_addr"[\s\S]*?\}',
            r'\{[\s\S]*?"url_list"[\s\S]*?\}',
        ]
        
        for pattern in json_patterns:
            matches = re.findall(pattern, response.text)
            if matches:
                print(f"使用模式 '{pattern}' 找到 {len(matches)} 个匹配")
                for i, match in enumerate(matches[:2]):  # 只显示前2个
                    try:
                        data = json.loads(match)
                        print(f"匹配 {i+1} 解析成功，包含键: {list(data.keys())}")
                    except Exception as e:
                        print(f"匹配 {i+1} 解析失败: {e}")
        
    except Exception as e:
        print(f"请求失败: {e}")
        return False
    
    return True

if __name__ == '__main__':
    import sys
    if len(sys.argv) > 1:
        video_url = sys.argv[1]
        print(f"从命令行获取视频URL: {video_url}")
    else:
        video_url = 'https://www.douyin.com/video/73075767586867398190'
        print(f"使用默认视频URL: {video_url}")
    cookie_file = 'www.douyin.com_cookies.txt'
    # 保存响应内容到文件
    download_douyin_video(video_url, cookie_file)
