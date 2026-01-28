#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
搜索HTML文件中的视频URL
"""

import re
import json

def search_video_urls(html_file):
    """
    搜索HTML文件中的视频URL
    """
    print(f"正在搜索 {html_file} 中的视频URL")
    
    try:
        # 读取HTML文件
        with open(html_file, 'r', encoding='utf-8') as f:
            content = f.read()
        
        print(f"文件大小: {len(content)} 字符")
        
        # 搜索模式1: 直接搜索视频URL
        video_url_patterns = [
            r'https?://[^"]+\.mp4[^"]*',
            r"https?://[^']+\.mp4[^']*",
            r'https?://[^\s]+\.mp4[^\s]*',
        ]
        
        print("\n=== 搜索直接视频URL ===")
        for pattern in video_url_patterns:
            matches = re.findall(pattern, content)
            if matches:
                print(f"使用模式 '{pattern}' 找到 {len(matches)} 个匹配")
                for i, match in enumerate(matches[:10]):  # 只显示前10个
                    if isinstance(match, tuple):
                        url = match[0]
                    else:
                        url = match
                    print(f"  {i+1}. {url}")
                if len(matches) > 10:
                    print(f"  ... 还有 {len(matches) - 10} 个匹配")
                break
        else:
            print("未找到直接视频URL")
        
        # 搜索模式2: 搜索JSON数据中的视频信息
        print("\n=== 搜索JSON数据中的视频信息 ===")
        
        # 查找__INITIAL_STATE__
        initial_state_match = re.search(r'window\.__INITIAL_STATE__\s*=\s*({[\s\S]*?});', content)
        if initial_state_match:
            print("找到__INITIAL_STATE__")
            try:
                initial_state = json.loads(initial_state_match.group(1))
                # 遍历查找视频信息
                def find_video_info(obj, path=""):
                    if isinstance(obj, dict):
                        for key, value in obj.items():
                            new_path = f"{path}.{key}" if path else key
                            if key == 'video' and isinstance(value, dict):
                                print(f"找到视频信息: {new_path}")
                                if 'play_addr' in value:
                                    print(f"  找到播放地址: {new_path}.play_addr")
                                    if 'url_list' in value['play_addr']:
                                        urls = value['play_addr']['url_list']
                                        print(f"  找到 {len(urls)} 个视频URL")
                                        for i, url in enumerate(urls):
                                            print(f"    {i+1}. {url}")
                            find_video_info(value, new_path)
                    elif isinstance(obj, list):
                        for i, item in enumerate(obj):
                            new_path = f"{path}[{i}]"
                            find_video_info(item, new_path)
                
                find_video_info(initial_state)
            except Exception as e:
                print(f"解析__INITIAL_STATE__失败: {e}")
        else:
            print("未找到__INITIAL_STATE__")
        
        # 搜索模式3: 搜索aweme数据
        aweme_match = re.search(r'aweme\s*=\s*({[\s\S]*?});', content)
        if aweme_match:
            print("\n找到aweme数据")
            try:
                aweme_data = json.loads(aweme_match.group(1))
                if isinstance(aweme_data, dict):
                    print(f"aweme数据包含 {len(aweme_data.keys())} 个键")
                    if 'video' in aweme_data:
                        print("找到video键")
                        video = aweme_data['video']
                        if 'play_addr' in video:
                            print("找到play_addr键")
                            play_addr = video['play_addr']
                            if 'url_list' in play_addr:
                                urls = play_addr['url_list']
                                print(f"找到 {len(urls)} 个视频URL")
                                for i, url in enumerate(urls):
                                    print(f"  {i+1}. {url}")
            except Exception as e:
                print(f"解析aweme数据失败: {e}")
        else:
            print("\n未找到aweme数据")
        
    except Exception as e:
        print(f"搜索失败: {e}")

if __name__ == '__main__':
    search_video_urls('douyin_response.html')
