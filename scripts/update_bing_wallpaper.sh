#!/bin/bash
# Bing 每日壁纸自动下载脚本
# 用法: ./update_bing_wallpaper.sh [分辨率]
# 分辨率可选: 1920x1080, UHD (默认 UHD)

set -e

# 配置
RESOLUTION="${1:-UHD}"
SAVE_DIR="/vol1/1000/embyhub/frontend/src/assets"
FILENAME="wallpaper.jpg"
BING_API="https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=zh-CN"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 确保保存目录存在
mkdir -p "$SAVE_DIR"

log_info "正在获取 Bing 每日壁纸信息..."

# 获取图片信息
RESPONSE=$(curl -s "$BING_API")

if [ -z "$RESPONSE" ]; then
    log_error "无法获取 Bing API 响应"
    exit 1
fi

# 解析 JSON 获取图片路径
IMAGE_PATH=$(echo "$RESPONSE" | grep -oP '"url":"\K[^"]+' | head -1)

if [ -z "$IMAGE_PATH" ]; then
    log_error "无法解析图片路径"
    exit 1
fi

# 提取图片基础名称 (OHR.XXX_ZH-CNXXX)
BASE_NAME=$(echo "$IMAGE_PATH" | grep -oP 'OHR\.[^_]+_[^_]+')

if [ -z "$BASE_NAME" ]; then
    log_error "无法提取图片名称"
    exit 1
fi

# 构建完整 URL（获取最高质量原图）
if [ "$RESOLUTION" = "UHD" ]; then
    IMAGE_URL="https://www.bing.com/th?id=${BASE_NAME}_UHD.jpg&qlt=100"
else
    IMAGE_URL="https://www.bing.com/th?id=${BASE_NAME}_1920x1080.jpg&qlt=100"
fi

log_info "图片地址: $IMAGE_URL"

# 获取图片标题（用于日志）
TITLE=$(echo "$RESPONSE" | grep -oP '"title":"\K[^"]+' | head -1)
COPYRIGHT=$(echo "$RESPONSE" | grep -oP '"copyright":"\K[^"]+' | head -1)

log_info "今日壁纸: $TITLE"
log_info "版权信息: $COPYRIGHT"

# 下载图片
SAVE_PATH="$SAVE_DIR/$FILENAME"
TEMP_PATH="$SAVE_DIR/.wallpaper_temp.jpg"

log_info "正在下载壁纸..."

if curl -s -o "$TEMP_PATH" "$IMAGE_URL"; then
    # 检查下载的文件是否有效
    if [ -s "$TEMP_PATH" ] && file "$TEMP_PATH" | grep -q "JPEG\|image"; then
        mv "$TEMP_PATH" "$SAVE_PATH"
        log_info "壁纸已保存到: $SAVE_PATH"
        
        # 显示文件信息
        FILE_SIZE=$(du -h "$SAVE_PATH" | cut -f1)
        log_info "文件大小: $FILE_SIZE"
        
        # 保存元信息
        cat > "$SAVE_DIR/wallpaper_info.json" << EOF
{
    "title": "$TITLE",
    "copyright": "$COPYRIGHT",
    "url": "$IMAGE_URL",
    "resolution": "$RESOLUTION",
    "updated_at": "$(date -Iseconds)",
    "date": "$(date +%Y-%m-%d)"
}
EOF
        log_info "元信息已保存到: $SAVE_DIR/wallpaper_info.json"
    else
        rm -f "$TEMP_PATH"
        log_error "下载的文件无效"
        exit 1
    fi
else
    rm -f "$TEMP_PATH"
    log_error "下载失败"
    exit 1
fi

log_info "完成！"
