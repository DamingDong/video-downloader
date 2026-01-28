#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试抖音API响应的脚本
"""

import requests
import re

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

def test_douyin_api(video_url, cookie_file):
    """
    测试抖音API响应
    """
    print(f"测试URL: {video_url}")
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
        # 发送请求
        response = requests.get(video_url, headers=headers, cookies=cookies, timeout=30)
        print(f"响应状态码: {response.status_code}")
        print(f"响应长度: {len(response.text)} 字符")
        
        # 检查响应内容
        if len(response.text) > 0:
            print("\n响应内容前500个字符:")
            print(response.text[:500])
            
            # 检查是否有JSON数据
            if 'window.__INITIAL_STATE__' in response.text:
                print("\n发现__INITIAL_STATE__数据")
            elif 'aweme' in response.text:
                print("\n发现aweme数据")
            elif 'error' in response.text.lower():
                print("\n发现错误信息")
        else:
            print("\n响应内容为空")
            
    except Exception as e:
        print(f"请求失败: {e}")

if __name__ == '__main__':
    video_url = 'https://www.douyin.com/video/73075767586867398190'
    cookie_file = 'www.douyin.com_cookies.txt'
    test_douyin_api(video_url, cookie_file)
