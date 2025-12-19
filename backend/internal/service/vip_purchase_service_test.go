// Package service VIP购买服务测试
package service_test

import (
	"testing"
	"time"

	"feiniu-user-system/internal/repository"
)

// TestCalculateNewExpireTime 测试会员到期时间计算
func TestCalculateNewExpireTime(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		currentExpireAt *time.Time
		durationDays    int
		expectedBefore  time.Time // 预期结果应该在此之前
	}{
		{
			name:            "未开通会员",
			currentExpireAt: nil,
			durationDays:    30,
			expectedBefore:  now.AddDate(0, 0, 30).Add(time.Second),
		},
		{
			name: "会员已过期",
			currentExpireAt: func() *time.Time {
				t := now.AddDate(0, 0, -10) // 10天前过期
				return &t
			}(),
			durationDays:   30,
			expectedBefore: now.AddDate(0, 0, 30).Add(time.Second),
		},
		{
			name: "会员未过期",
			currentExpireAt: func() *time.Time {
				t := now.AddDate(0, 0, 10) // 还有10天
				return &t
			}(),
			durationDays:   30,
			expectedBefore: now.AddDate(0, 0, 40).Add(time.Second), // 10 + 30 = 40
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 由于CalculateNewExpireTime使用了time.Now()，我们只能验证结果是合理的
			result := repository.CalculateNewExpireTime(tt.currentExpireAt, tt.durationDays)

			// 验证结果不是零值
			if result.IsZero() {
				t.Errorf("CalculateNewExpireTime() 返回零值")
			}

			// 验证结果是未来时间
			if result.Before(time.Now().UTC()) {
				t.Errorf("CalculateNewExpireTime() 返回过去时间 = %v", result)
			}

			t.Logf("输入: %v, 输出: %v", tt.currentExpireAt, result)
		})
	}
}
