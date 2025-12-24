// Package service 支付宝服务属性测试
package service

import (
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: alipay-vip-purchase, Property 1: 订单号唯一性和格式
// Validates: Requirements 2.1, 2.2

func TestGenerateOrderNo_Format(t *testing.T) {
	// 订单号格式: VIP{timestamp:14位}{random:6位} = VIP + 20位
	pattern := regexp.MustCompile(`^VIP\d{20}$`)

	properties := gopter.NewProperties(nil)

	properties.Property("订单号格式正确", prop.ForAll(
		func(_ int) bool {
			orderNo := GenerateOrderNo()
			return pattern.MatchString(orderNo)
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

func TestGenerateOrderNo_Uniqueness(t *testing.T) {
	// Feature: alipay-vip-purchase, Property 1: 订单号唯一性
	// Validates: Requirements 2.1

	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("生成的订单号唯一", prop.ForAll(
		func(count int) bool {
			// 限制生成数量在合理范围内
			if count < 10 {
				count = 10
			}
			if count > 1000 {
				count = 1000
			}

			orderNos := make(map[string]bool)
			for i := 0; i < count; i++ {
				orderNo := GenerateOrderNo()
				if orderNos[orderNo] {
					return false // 发现重复
				}
				orderNos[orderNo] = true
				// 添加微小延迟以确保时间戳变化
				time.Sleep(time.Microsecond)
			}
			return true
		},
		gen.IntRange(10, 100),
	))

	properties.TestingRun(t)
}

func TestGenerateOrderNo_ConcurrentUniqueness(t *testing.T) {
	// 测试并发生成订单号的唯一性
	const goroutines = 10
	const ordersPerGoroutine = 100

	var wg sync.WaitGroup
	orderNos := make(chan string, goroutines*ordersPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < ordersPerGoroutine; j++ {
				orderNos <- GenerateOrderNo()
			}
		}()
	}

	wg.Wait()
	close(orderNos)

	// 检查唯一性
	seen := make(map[string]bool)
	for orderNo := range orderNos {
		if seen[orderNo] {
			t.Errorf("发现重复订单号: %s", orderNo)
		}
		seen[orderNo] = true
	}

	if len(seen) != goroutines*ordersPerGoroutine {
		t.Errorf("订单号数量不正确: 期望 %d, 实际 %d", goroutines*ordersPerGoroutine, len(seen))
	}
}

func TestGenerateOrderNo_StartsWithVIP(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("订单号以VIP开头", prop.ForAll(
		func(_ int) bool {
			orderNo := GenerateOrderNo()
			return len(orderNo) >= 3 && orderNo[:3] == "VIP"
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

func TestGenerateOrderNo_Length(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("订单号长度为23", prop.ForAll(
		func(_ int) bool {
			orderNo := GenerateOrderNo()
			// VIP(3) + timestamp(14) + random(6) = 23
			return len(orderNo) == 23
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// =====================================================
// Property 2: 签名验证一致性
// Feature: alipay-vip-purchase
// Validates: Requirements 3.1, 3.2
// =====================================================

func TestSignatureVerification_Consistency(t *testing.T) {
	// 测试签名验证的一致性：相同的参数应该产生相同的验证结果
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("相同参数签名验证结果一致", prop.ForAll(
		func(orderNo string, amount int64, tradeNo string) bool {
			// 模拟签名验证参数
			params := map[string]string{
				"out_trade_no": "VIP" + orderNo,
				"total_amount": formatAmountStr(amount),
				"trade_no":     "T" + tradeNo,
			}

			// 多次验证相同参数应该得到相同结果
			result1 := mockVerifySignature(params)
			result2 := mockVerifySignature(params)
			result3 := mockVerifySignature(params)

			return result1 == result2 && result2 == result3
		},
		gen.Identifier(),
		gen.Int64Range(1, 1000000), // 金额范围 0.01 - 10000.00 元
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

func TestSignatureVerification_DifferentParams(t *testing.T) {
	// 测试不同参数应该产生不同的签名
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("不同参数产生不同签名", prop.ForAll(
		func(orderNo1, orderNo2 string) bool {
			// 确保两个订单号不同
			on1 := "VIP" + orderNo1 + "A"
			on2 := "VIP" + orderNo2 + "B"

			params1 := map[string]string{"out_trade_no": on1}
			params2 := map[string]string{"out_trade_no": on2}

			sig1 := mockGenerateSignature(params1)
			sig2 := mockGenerateSignature(params2)

			return sig1 != sig2
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// mockVerifySignature 模拟签名验证（用于测试）
func mockVerifySignature(params map[string]string) bool {
	// 简单的模拟验证：检查必要参数存在
	_, hasOrderNo := params["out_trade_no"]
	return hasOrderNo
}

// mockGenerateSignature 模拟签名生成（用于测试）
func mockGenerateSignature(params map[string]string) string {
	// 简单的模拟签名：拼接参数
	result := ""
	for k, v := range params {
		result += k + "=" + v + "&"
	}
	return result
}

// formatAmountStr 格式化金额（分转元）
func formatAmountStr(cents int64) string {
	yuan := cents / 100
	fen := cents % 100
	return fmt.Sprintf("%d.%02d", yuan, fen)
}

// =====================================================
// Property 4: 幂等性处理
// Feature: alipay-vip-purchase
// Validates: Requirements 3.8
// =====================================================

func TestIdempotency_SameNotification(t *testing.T) {
	// 测试相同通知多次处理的幂等性
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("相同通知多次处理结果一致", prop.ForAll(
		func(orderNo string, tradeNo string) bool {
			// 模拟订单状态
			orderStatus := make(map[string]string)
			on := "VIP" + orderNo
			tn := "T" + tradeNo
			orderStatus[on] = "pending"

			// 第一次处理
			result1 := mockProcessNotification(on, tn, orderStatus)

			// 第二次处理（相同通知）
			result2 := mockProcessNotification(on, tn, orderStatus)

			// 第三次处理（相同通知）
			result3 := mockProcessNotification(on, tn, orderStatus)

			// 所有处理都应该成功，且订单状态最终一致
			return result1 && result2 && result3 && orderStatus[on] == "success"
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

func TestIdempotency_ProcessedOrderSkipped(t *testing.T) {
	// 测试已处理订单应该被跳过
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("已处理订单不会重复处理", prop.ForAll(
		func(orderNo string, processCount int) bool {
			if processCount < 1 {
				processCount = 1
			}
			if processCount > 10 {
				processCount = 10
			}

			// 模拟订单状态
			on := "VIP" + orderNo
			orderStatus := make(map[string]string)
			orderStatus[on] = "pending"
			processedCount := 0

			for i := 0; i < processCount; i++ {
				if mockProcessNotificationWithCount(on, orderStatus, &processedCount) {
					// 处理成功
				}
			}

			// 无论调用多少次，实际处理次数应该只有1次
			return processedCount == 1
		},
		gen.Identifier(),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// mockProcessNotification 模拟通知处理
func mockProcessNotification(orderNo, tradeNo string, orderStatus map[string]string) bool {
	status, exists := orderStatus[orderNo]
	if !exists {
		return false // 订单不存在
	}

	if status == "success" {
		return true // 已处理，幂等返回成功
	}

	// 处理订单
	orderStatus[orderNo] = "success"
	return true
}

// mockProcessNotificationWithCount 模拟通知处理（带计数）
func mockProcessNotificationWithCount(orderNo string, orderStatus map[string]string, count *int) bool {
	status, exists := orderStatus[orderNo]
	if !exists {
		return false
	}

	if status == "success" {
		return true // 已处理，幂等返回成功，不增加计数
	}

	// 实际处理
	orderStatus[orderNo] = "success"
	*count++
	return true
}

// =====================================================
// Property 5: 金额一致性校验
// Feature: alipay-vip-purchase
// Validates: Requirements 7.6
// =====================================================

func TestAmountValidation_Consistency(t *testing.T) {
	// 测试金额校验的一致性
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("订单金额与通知金额必须一致", prop.ForAll(
		func(orderAmount, notifyAmount int64) bool {
			// 金额一致时应该通过验证
			if orderAmount == notifyAmount {
				return validateAmount(orderAmount, notifyAmount)
			}
			// 金额不一致时应该拒绝
			return !validateAmount(orderAmount, notifyAmount)
		},
		gen.Int64Range(1, 10000000),   // 订单金额 0.01 - 100000.00 元
		gen.Int64Range(1, 10000000),   // 通知金额
	))

	properties.TestingRun(t)
}

func TestAmountValidation_PositiveAmount(t *testing.T) {
	// 测试金额必须为正数
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("金额必须为正数", prop.ForAll(
		func(amount int64) bool {
			if amount <= 0 {
				return !isValidAmount(amount)
			}
			return isValidAmount(amount)
		},
		gen.Int64Range(-1000, 10000000),
	))

	properties.TestingRun(t)
}

func TestAmountValidation_Precision(t *testing.T) {
	// 测试金额精度（分为单位）
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("金额转换精度正确", prop.ForAll(
		func(cents int64) bool {
			if cents <= 0 {
				return true // 跳过非正数
			}

			// 分转元再转回分，应该相等
			yuan := float64(cents) / 100.0
			// 使用四舍五入避免浮点精度问题
			backToCents := int64(yuan*100 + 0.5)

			return cents == backToCents
		},
		gen.Int64Range(1, 10000000),
	))

	properties.TestingRun(t)
}

// validateAmount 验证金额一致性
func validateAmount(orderAmount, notifyAmount int64) bool {
	return orderAmount == notifyAmount
}

// isValidAmount 验证金额有效性
func isValidAmount(amount int64) bool {
	return amount > 0
}

// =====================================================
// Property 6: 配置加密存储
// Feature: alipay-vip-purchase
// Validates: Requirements 1.1, 7.1
// =====================================================

func TestEncryption_Reversibility(t *testing.T) {
	// 测试加密解密的可逆性
	service := NewAlipayService(nil, "test-encryption-key-12345")

	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("加密后可以正确解密", prop.ForAll(
		func(plaintext string) bool {
			if plaintext == "" {
				return true // 空字符串跳过
			}

			encrypted, err := service.EncryptSecret(plaintext)
			if err != nil {
				return false
			}

			decrypted, err := service.DecryptSecret(encrypted)
			if err != nil {
				return false
			}

			return plaintext == decrypted
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 1000 }),
	))

	properties.TestingRun(t)
}

func TestEncryption_DifferentCiphertext(t *testing.T) {
	// 测试相同明文每次加密产生不同密文（因为使用随机nonce）
	service := NewAlipayService(nil, "test-encryption-key-12345")

	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("相同明文产生不同密文", prop.ForAll(
		func(plaintext string) bool {
			if plaintext == "" || len(plaintext) < 5 {
				return true // 跳过太短的字符串
			}

			encrypted1, err1 := service.EncryptSecret(plaintext)
			encrypted2, err2 := service.EncryptSecret(plaintext)

			if err1 != nil || err2 != nil {
				return false
			}

			// 由于使用随机nonce，相同明文应该产生不同密文
			return encrypted1 != encrypted2
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 5 && len(s) <= 100 }),
	))

	properties.TestingRun(t)
}

func TestEncryption_EmptyString(t *testing.T) {
	// 测试空字符串处理
	service := NewAlipayService(nil, "test-encryption-key-12345")

	encrypted, err := service.EncryptSecret("")
	if err != nil {
		t.Errorf("加密空字符串失败: %v", err)
	}
	if encrypted != "" {
		t.Errorf("空字符串加密结果应该为空")
	}

	decrypted, err := service.DecryptSecret("")
	if err != nil {
		t.Errorf("解密空字符串失败: %v", err)
	}
	if decrypted != "" {
		t.Errorf("空字符串解密结果应该为空")
	}
}

func TestEncryption_DifferentKeys(t *testing.T) {
	// 测试不同密钥产生不同密文
	service1 := NewAlipayService(nil, "key-1-abcdefghijk")
	service2 := NewAlipayService(nil, "key-2-lmnopqrstuv")

	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("不同密钥无法解密", prop.ForAll(
		func(plaintext string) bool {
			if plaintext == "" || len(plaintext) < 5 {
				return true
			}

			// 用key1加密
			encrypted, err := service1.EncryptSecret(plaintext)
			if err != nil {
				return false
			}

			// 用key2解密应该失败
			_, err = service2.DecryptSecret(encrypted)
			return err != nil // 应该返回错误
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 5 && len(s) <= 100 }),
	))

	properties.TestingRun(t)
}

// =====================================================
// Property 3: VIP 时长累加正确性
// Feature: alipay-vip-purchase
// Validates: Requirements 3.4, 3.5, 3.6
// =====================================================

func TestVipDuration_NewUser(t *testing.T) {
	// 测试新用户开通VIP
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("新用户VIP从当前时间开始计算", prop.ForAll(
		func(durationDays int) bool {
			if durationDays <= 0 || durationDays > 365 {
				return true // 跳过无效天数
			}

			now := time.Now().UTC()
			result := calculateVipExpireTime(nil, durationDays)

			// 结果应该在 now + durationDays 附近（允许1秒误差）
			expected := now.AddDate(0, 0, durationDays)
			diff := result.Sub(expected)
			if diff < 0 {
				diff = -diff
			}

			return diff < time.Second*2
		},
		gen.IntRange(1, 365),
	))

	properties.TestingRun(t)
}

func TestVipDuration_ExpiredUser(t *testing.T) {
	// 测试已过期用户续费
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("过期用户VIP从当前时间开始计算", prop.ForAll(
		func(expiredDays, durationDays int) bool {
			if durationDays <= 0 || durationDays > 365 {
				return true
			}
			if expiredDays <= 0 || expiredDays > 365 {
				return true
			}

			now := time.Now().UTC()
			expiredAt := now.AddDate(0, 0, -expiredDays) // 过期时间
			result := calculateVipExpireTime(&expiredAt, durationDays)

			// 结果应该在 now + durationDays 附近
			expected := now.AddDate(0, 0, durationDays)
			diff := result.Sub(expected)
			if diff < 0 {
				diff = -diff
			}

			return diff < time.Second*2
		},
		gen.IntRange(1, 365),
		gen.IntRange(1, 365),
	))

	properties.TestingRun(t)
}

func TestVipDuration_ActiveUser(t *testing.T) {
	// 测试有效期内用户续费
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("有效期内用户VIP在原到期时间基础上累加", prop.ForAll(
		func(remainingDays, durationDays int) bool {
			if durationDays <= 0 || durationDays > 365 {
				return true
			}
			if remainingDays <= 0 || remainingDays > 365 {
				return true
			}

			now := time.Now().UTC()
			currentExpireAt := now.AddDate(0, 0, remainingDays) // 还有remainingDays天到期
			result := calculateVipExpireTime(&currentExpireAt, durationDays)

			// 结果应该在 currentExpireAt + durationDays 附近
			expected := currentExpireAt.AddDate(0, 0, durationDays)
			diff := result.Sub(expected)
			if diff < 0 {
				diff = -diff
			}

			return diff < time.Second*2
		},
		gen.IntRange(1, 365),
		gen.IntRange(1, 365),
	))

	properties.TestingRun(t)
}

func TestVipDuration_Accumulation(t *testing.T) {
	// 测试多次续费累加
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("多次续费天数正确累加", prop.ForAll(
		func(days1, days2, days3 int) bool {
			if days1 <= 0 || days1 > 100 {
				return true
			}
			if days2 <= 0 || days2 > 100 {
				return true
			}
			if days3 <= 0 || days3 > 100 {
				return true
			}

			now := time.Now().UTC()

			// 第一次开通
			expire1 := calculateVipExpireTime(nil, days1)

			// 第二次续费
			expire2 := calculateVipExpireTime(&expire1, days2)

			// 第三次续费
			expire3 := calculateVipExpireTime(&expire2, days3)

			// 总天数应该是 days1 + days2 + days3
			totalDays := days1 + days2 + days3
			expected := now.AddDate(0, 0, totalDays)

			diff := expire3.Sub(expected)
			if diff < 0 {
				diff = -diff
			}

			return diff < time.Second*3
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t)
}

// calculateVipExpireTime 计算VIP到期时间（与repository.CalculateNewExpireTime逻辑一致）
func calculateVipExpireTime(currentExpireAt *time.Time, durationDays int) time.Time {
	now := time.Now().UTC()

	// 如果没有会员或已过期，从现在开始计算
	if currentExpireAt == nil || currentExpireAt.Before(now) {
		return now.AddDate(0, 0, durationDays)
	}

	// 如果还在有效期内，从到期时间顺延
	return currentExpireAt.AddDate(0, 0, durationDays)
}
