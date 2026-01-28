#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
直接调用抖音API的脚本
"""

import requests
import json

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

def test_douyin_api_direct(video_id, cookie_file):
    """
    直接调用抖音API获取视频信息
    """
    print(f"测试抖音API，视频ID: {video_id}")
    print(f"使用Cookie文件: {cookie_file}")
    
    # 加载cookies
    cookies = load_cookies(cookie_file)
    print(f"加载了 {len(cookies)} 个cookie")
    
    # 设置请求头
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Referer': 'https://www.douyin.com/',
        'Accept': 'application/json, text/plain, */*',
        'Accept-Language': 'zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3',
        'Accept-Encoding': 'gzip, deflate, br',
        'DNT': '1',
        'Connection': 'keep-alive',
        'Content-Type': 'application/json'
    }
    
    # 抖音API端点
    api_url = f"https://www.douyin.com/aweme/v1/web/aweme/detail/?aweme_id={video_id}"
    print(f"API URL: {api_url}")
    
    try:
        # 发送请求
        response = requests.get(api_url, headers=headers, cookies=cookies, timeout=30)
        print(f"响应状态码: {response.status_code}")
        print(f"响应长度: {len(response.text)} 字符")
        
        # 检查响应内容
        if len(response.text) > 0:
            print("\n响应内容前500个字符:")
            print(response.text[:500])
            
            # 尝试解析JSON
            try:
                data = json.loads(response.text)
                print(f"\nJSON数据类型: {type(data)}")
                if 'aweme_detail' in data:
                    print("发现aweme_detail数据")
                    aweme_detail = data['aweme_detail']
                    if 'video' in aweme_detail:
                        print("发现video数据")
                        video = aweme_detail['video']
                        if 'play_addr' in video:
                            print("发现play_addr数据")
                            play_addr = video['play_addr']
                            if 'url_list' in play_addr:
                                print(f"找到 {len(play_addr['url_list'])} 个视频URL")
                                for i, url in enumerate(play_addr['url_list']):
                                    print(f"视频URL {i+1}: {url}")
            except Exception as e:
                print(f"解析JSON失败: {e}")
        else:
            print("\n响应内容为空")
            
    except Exception as e:
        print(f"请求失败: {e}")

if __name__ == '__main__':
    # 从URL中提取视频ID
    video_url = 'https://www.douyin.com/video/73075767586867398190'
    video_id = video_url.split('/')[-1]
    cookie_file = 'www.douyin.com_cookies.txt'
    test_douyin_api_direct(video_id, cookie_file)
