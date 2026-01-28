#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
更新yt-dlp到最新版本的脚本
"""

import subprocess
import sys

def update_yt_dlp():
    """
    更新yt-dlp到最新版本
    """
    print("正在更新yt-dlp到最新版本...")
    
    try:
        # 使用yt-dlp的内置更新功能
        result = subprocess.run(
            [sys.executable, '-m', 'pip', 'install', '--upgrade', 'yt-dlp'],
            capture_output=True,
            text=True
        )
        
        if result.returncode == 0:
            print("yt-dlp更新成功!")
            # 检查版本
            version_result = subprocess.run(
                ['yt-dlp', '--version'],
                capture_output=True,
                text=True
            )
            if version_result.returncode == 0:
                print(f"当前版本: {version_result.stdout.strip()}")
            else:
                print("检查版本失败:", version_result.stderr)
        else:
            print("yt-dlp更新失败:", result.stderr)
            
        # 也尝试直接更新本地的yt-dlp.exe
        print("\n尝试更新本地的yt-dlp.exe...")
        result = subprocess.run(
            ['./yt-dlp.exe', '-U'],
            capture_output=True,
            text=True
        )
        print(f"本地yt-dlp.exe更新结果: {result.stdout}")
        if result.stderr:
            print(f"错误: {result.stderr}")
            
    except Exception as e:
        print(f"更新过程中出错: {e}")

if __name__ == '__main__':
    update_yt_dlp()
