package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	baseURL   = "http://localhost:54680"
	appID     = 1377283335341765
	appSecret = "XveEkF3KIzvFXeOQK1oFVlhkkYlJUIY1"
	mchID     = "1001"
	mchSecret = "XveEkF3KIzvFXeOQK1oFVlhkkYlJUIY1" // 与数据库中保存的一致
)

func md5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func calcSign(body string, timestamp int64) string {
	bodyMd5 := md5Hash(body)
	signStr := fmt.Sprintf("%d,%s,%s,%d,%s,%s", appID, appSecret, bodyMd5, timestamp, mchID, mchSecret)
	fmt.Printf("签名字符串: %s\n", signStr)
	return md5Hash(signStr)
}

func doRequest(method, path string, body interface{}) {
	timestamp := time.Now().Unix()
	
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			fmt.Printf("序列化失败: %v\n", err)
			return
		}
	} else {
		bodyBytes = []byte("{}")
	}
	
	sign := calcSign(string(bodyBytes), timestamp)
	
	url := fmt.Sprintf("%s%s?mch_id=%s&timestamp=%d&sign=%s", baseURL, path, mchID, timestamp, sign)
	fmt.Printf("请求URL: %s\n", url)
	fmt.Printf("请求Body: %s\n", string(bodyBytes))
	
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("响应状态: %d\n", resp.StatusCode)
	
	var result map[string]interface{}
	if json.Unmarshal(respBody, &result) == nil {
		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("响应内容:\n%s\n", string(prettyJSON))
	} else {
		fmt.Printf("响应内容: %s\n", string(respBody))
	}
}

func main() {
	fmt.Println("==========================================")
	fmt.Println("闲管家接口测试")
	fmt.Println("==========================================")
	fmt.Printf("APP_ID: %d\n", appID)
	fmt.Printf("MCH_ID: %s\n", mchID)
	fmt.Println()

	// 1. 测试平台信息接口
	fmt.Println("1. 测试平台信息接口 POST /goofish/open/info")
	fmt.Println("-------------------------------------------")
	doRequest("POST", "/goofish/open/info", nil)
	fmt.Println()

	// 2. 测试商户信息接口
	fmt.Println("2. 测试商户信息接口 POST /goofish/user/info")
	fmt.Println("-------------------------------------------")
	doRequest("POST", "/goofish/user/info", nil)
	fmt.Println()

	// 3. 测试商品列表接口
	fmt.Println("3. 测试商品列表接口 POST /goofish/goods/list")
	fmt.Println("-------------------------------------------")
	doRequest("POST", "/goofish/goods/list", map[string]interface{}{
		"goods_type": 2,
		"page":       1,
		"page_size":  10,
	})
	fmt.Println()

	// 4. 测试商品详情接口
	fmt.Println("4. 测试商品详情接口 POST /goofish/goods/detail")
	fmt.Println("-------------------------------------------")
	doRequest("POST", "/goofish/goods/detail", map[string]interface{}{
		"goods_no": "GOODS001",
	})
	fmt.Println()

	// 5. 测试创建卡密订单接口
	fmt.Println("5. 测试创建卡密订单接口 POST /goofish/order/purchase/create")
	fmt.Println("-------------------------------------------")
	orderNo := "TEST" + strconv.FormatInt(time.Now().Unix(), 10)
	fmt.Printf("订单号: %s\n", orderNo)
	doRequest("POST", "/goofish/order/purchase/create", map[string]interface{}{
		"order_no":     orderNo,
		"goods_no":     "GOODS001",
		"buy_quantity": 1,
	})
	fmt.Println()

	// 6. 测试订单详情接口
	fmt.Println("6. 测试订单详情接口 POST /goofish/order/detail")
	fmt.Println("-------------------------------------------")
	doRequest("POST", "/goofish/order/detail", map[string]interface{}{
		"order_no": orderNo,
	})
	fmt.Println()

	// 7. 测试幂等性
	fmt.Println("7. 测试幂等性（重复提交相同订单号）")
	fmt.Println("-------------------------------------------")
	doRequest("POST", "/goofish/order/purchase/create", map[string]interface{}{
		"order_no":     orderNo,
		"goods_no":     "GOODS001",
		"buy_quantity": 1,
	})
	fmt.Println()

	fmt.Println("==========================================")
	fmt.Println("测试完成")
	fmt.Println("==========================================")
}
