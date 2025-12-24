#!/bin/bash

# 闲管家接口测试脚本
# 配置信息（需要与数据库中的配置一致）
APP_ID="1377283335341765"
APP_SECRET="XveEkF3KIzvFXeOQK1oFVlhkkYlJUIY1"
MCH_ID="1001"
MCH_SECRET="XveEkF3KIzvFXeOQK1oFVlhkkYlJUIY1"
BASE_URL="http://localhost:54680"

# 生成时间戳
TIMESTAMP=$(date +%s)

# 计算签名函数
# 签名规则: md5("app_id,app_secret,bodyMd5,timestamp,mch_id,mch_secret")
calc_sign() {
    local body="$1"
    local body_md5=$(echo -n "$body" | md5)
    local sign_str="${APP_ID},${APP_SECRET},${body_md5},${TIMESTAMP},${MCH_ID},${MCH_SECRET}"
    echo -n "$sign_str" | md5
}

echo "=========================================="
echo "闲管家接口测试"
echo "=========================================="
echo "时间戳: $TIMESTAMP"
echo ""

# 1. 测试平台信息接口
echo "1. 测试平台信息接口 POST /goofish/open/info"
echo "-------------------------------------------"
BODY='{}'
SIGN=$(calc_sign "$BODY")
echo "签名: $SIGN"
curl -s -X POST "${BASE_URL}/goofish/open/info?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 2. 测试商户信息接口
echo "2. 测试商户信息接口 POST /goofish/user/info"
echo "-------------------------------------------"
BODY='{}'
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/user/info?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 3. 测试商品列表接口
echo "3. 测试商品列表接口 POST /goofish/goods/list"
echo "-------------------------------------------"
BODY='{"goods_type":2,"page":1,"page_size":10}'
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/goods/list?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 4. 测试商品详情接口
echo "4. 测试商品详情接口 POST /goofish/goods/detail"
echo "-------------------------------------------"
BODY='{"goods_no":"CARD_MONTH"}'
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/goods/detail?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 5. 测试创建卡密订单接口
echo "5. 测试创建卡密订单接口 POST /goofish/order/purchase/create"
echo "-------------------------------------------"
ORDER_NO="TEST$(date +%Y%m%d%H%M%S)"
BODY="{\"order_no\":\"${ORDER_NO}\",\"goods_no\":\"CARD_MONTH\",\"buy_quantity\":1}"
SIGN=$(calc_sign "$BODY")
echo "订单号: $ORDER_NO"
curl -s -X POST "${BASE_URL}/goofish/order/purchase/create?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 6. 测试订单详情接口
echo "6. 测试订单详情接口 POST /goofish/order/detail"
echo "-------------------------------------------"
BODY="{\"order_no\":\"${ORDER_NO}\"}"
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/order/detail?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 7. 测试幂等性（重复订单）
echo "7. 测试幂等性（重复提交相同订单号）"
echo "-------------------------------------------"
BODY="{\"order_no\":\"${ORDER_NO}\",\"goods_no\":\"CARD_MONTH\",\"buy_quantity\":1}"
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/order/purchase/create?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="


# 8. 测试商品订阅接口
echo "8. 测试商品订阅接口 POST /goofish/goods/change/subscribe"
echo "-------------------------------------------"
BODY='{"goods_type":2,"goods_no":"CARD_MONTH","token":"test_token_001","notify_url":"http://example.com/notify"}'
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/goods/change/subscribe?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 9. 测试查询订阅列表接口
echo "9. 测试查询订阅列表接口 POST /goofish/goods/change/subscribe/list"
echo "-------------------------------------------"
BODY='{"goods_type":2,"page_no":1,"page_size":10}'
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/goods/change/subscribe/list?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""

# 10. 测试取消订阅接口
echo "10. 测试取消订阅接口 POST /goofish/goods/change/unsubscribe"
echo "-------------------------------------------"
BODY='{"goods_type":2,"goods_no":"CARD_MONTH","token":"test_token_001"}'
SIGN=$(calc_sign "$BODY")
curl -s -X POST "${BASE_URL}/goofish/goods/change/unsubscribe?mch_id=${MCH_ID}&timestamp=${TIMESTAMP}&sign=${SIGN}" \
  -H "Content-Type: application/json" \
  -d "$BODY" | python3 -m json.tool 2>/dev/null || echo "请求失败"
echo ""
