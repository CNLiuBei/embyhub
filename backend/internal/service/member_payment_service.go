// Package service 会员支付服务
package service

import (
	"gorm.io/gorm"
)

// MemberPaymentService 会员支付服务
type MemberPaymentService struct {
	db *gorm.DB
}

// NewMemberPaymentService 创建会员支付服务
func NewMemberPaymentService(db *gorm.DB) *MemberPaymentService {
	return &MemberPaymentService{db: db}
}

// MemberPackage 会员套餐
type MemberPackage struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Days     int     `json:"days"`
	Price    float64 `json:"price"`
	Original float64 `json:"original"`
	Discount string  `json:"discount"`
}

// GetMemberPackages 获取会员套餐列表
func (s *MemberPaymentService) GetMemberPackages() []MemberPackage {
	return []MemberPackage{
		{ID: 1, Name: "月卡会员", Days: 30, Price: 15.00, Original: 20.00, Discount: "限时优惠"},
		{ID: 2, Name: "季卡会员", Days: 90, Price: 40.00, Original: 60.00, Discount: "最划算"},
		{ID: 3, Name: "年卡会员", Days: 365, Price: 120.00, Original: 240.00, Discount: "5折优惠"},
	}
}

// PurchaseMemberRequest 购买会员请求
type PurchaseMemberRequest struct {
	PackageID int `json:"package_id" binding:"required"`
}
