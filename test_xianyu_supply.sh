#!/bin/bash
# 闲管家货源授权接口测试脚本
# 用法: ./test_xianyu_supply.sh [服务器地址]

BASE_URL="${1:-http://localhost:54680}"
APP_KEY="1766317072846383"
APP_SECRET="a2umeee66yyyqqiiiaaa22uuummmee66"

echo "=========================================="
echo "闲管家货源授权接口测试"
echo "=========================================="
echo "服务器地址: $BASE_URL"
echo "AppKey: $APP_KEY"
echo "AppSecret: $APP_SECRET"
echo ""

# 生成时间戳
TIMESTAMP=$(date +%s)
BODY="{}"

# 计算 bodyMd5
BODY_MD5=$(echo -n "$BODY" | md5 2>/dev/null || echo -n "$BODY" | md5sum | cut -d' ' -f1)

# 计算签名: md5(appKey,bodyMd5,timestamp,appSecret)
SIGN_STR="${APP_KEY},${BODY_MD5},${TIMESTAMP},${APP_SECRET}"
SIGN=$(echo -n "$SIGN_STR" | md5 2>/dev/null || echo -n "$SIGN_STR" | md5sum | cut -d' ' -f1)

echo "=== 签名参数 ==="
echo "Timestamp: $TIMESTAMP"
echo "Body: $BODY"
echo "BodyMD5: $BODY_MD5"
echo "SignStr: $SIGN_STR"
echo "Sign: $SIGN"
echo ""

echo "=== 测试1: 健康检查 ==="
curl -s "${BASE_URL}/health" && echo ""
echo ""

echo "=== 测试2: /goofish/open/info (获取货源方信息) ==="
echo "URL: ${BASE_URL}/api/supply/xianyu/fetch/goofish/open/info?mch_id=${APP_KEY}&timestamp=${TIMESTAMP}&sign=${SIGN}"
curl -s -X POST "${BASE_URL}/api/supply/xianyu/fetch/goofish/open/info?mch_id=${APP_KEY}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY"
echo ""
echo ""

# 重新生成时间戳和签名（用于第二个请求）
TIMESTAMP=$(date +%s)
BODY='{"order_no":"TEST001","quantity":1,"card_type":"月卡"}'
BODY_MD5=$(echo -n "$BODY" | md5 2>/dev/null || echo -n "$BODY" | md5sum | cut -d' ' -f1)
SIGN_STR="${APP_KEY},${BODY_MD5},${TIMESTAMP},${APP_SECRET}"
SIGN=$(echo -n "$SIGN_STR" | md5 2>/dev/null || echo -n "$SIGN_STR" | md5sum | cut -d' ' -f1)

echo "=== 测试3: /goofish/open/kam/get (获取卡密) ==="
echo "URL: ${BASE_URL}/api/supply/xianyu/fetch/goofish/open/kam/get?mch_id=${APP_KEY}&timestamp=${TIMESTAMP}&sign=${SIGN}"
echo "Body: $BODY"
curl -s -X POST "${BASE_URL}/api/supply/xianyu/fetch/goofish/open/kam/get?mch_id=${APP_KEY}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY"
echo ""
echo ""

echo "=== 测试完成 ==="
