#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
模拟浏览器行为测试抖音视频下载
"""

import requests
import time
import random

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

def test_douyin_browser(video_url, cookie_file):
    """
    模拟浏览器行为测试抖音视频下载
    """
    print(f"测试URL: {video_url}")
    print(f"使用Cookie文件: {cookie_file}")
    
    # 加载cookies
    cookies = load_cookies(cookie_file)
    print(f"加载了 {len(cookies)} 个cookie")
    
    # 生成随机数用于请求头
    def generate_random():
        return str(random.random())[2:17]
    
    # 生成时间戳
    timestamp = int(time.time() * 1000)
    
    # 设置完整的请求头，模拟Chrome浏览器
    headers = {
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8',
        'Accept-Encoding': 'gzip, deflate, br',
        'Accept-Language': 'zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3',
        'Cache-Control': 'max-age=0',
        'Connection': 'keep-alive',
        'DNT': '1',
        'Host': 'www.douyin.com',
        'Sec-Fetch-Dest': 'document',
        'Sec-Fetch-Mode': 'navigate',
        'Sec-Fetch-Site': 'none',
        'Sec-Fetch-User': '?1',
        'Upgrade-Insecure-Requests': '1',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'X-Requested-With': 'XMLHttpRequest',
        # 添加抖音特定的请求头
        'X-Tt-Token': cookies.get('ttwid', ''),
        'X-Argus': '',
        'X-Gorgon': '',
        'X-Khronos': str(timestamp),
    }
    
    try:
        # 发送请求
        response = requests.get(video_url, headers=headers, cookies=cookies, timeout=30, allow_redirects=True)
        print(f"响应状态码: {response.status_code}")
        print(f"响应URL: {response.url}")
        print(f"响应长度: {len(response.text)} 字符")
        
        # 检查响应内容
        if len(response.text) > 0:
            # 检查是否有重定向
            if response.history:
                print(f"\n重定向历史: {len(response.history)} 次")
                for i, hist in enumerate(response.history):
                    print(f"重定向 {i+1}: {hist.status_code} -> {hist.url}")
            
            # 检查是否有视频相关的关键词
            video_keywords = ['video', 'play', 'url', 'mp4', 'aweme']
            for keyword in video_keywords:
                count = response.text.count(keyword)
                if count > 0:
                    print(f"\n关键词 '{keyword}' 出现次数: {count}")
            
            # 保存响应内容到文件，以便分析
            with open('douyin_response.html', 'w', encoding='utf-8') as f:
                f.write(response.text)
            print("\n响应内容已保存到 douyin_response.html")
        else:
            print("\n响应内容为空")
            
    except Exception as e:
        print(f"请求失败: {e}")

if __name__ == '__main__':
    video_url = 'https://www.douyin.com/video/73075767586867398190'
    cookie_file = 'www.douyin.com_cookies.txt'
    test_douyin_browser(video_url, cookie_file)
