// Package models 闲管家虚拟货源对接数据模型
package models

import (
	"time"
)

// =====================================================
// 闲管家配置
// =====================================================

// GoofishConfig 闲管家配置
type GoofishConfig struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	AppID              int64     `gorm:"not null" json:"app_id"`                              // 闲管家应用ID
	AppSecretEncrypted string    `gorm:"column:app_secret_encrypted" json:"-"`               // 加密存储的app_secret
	MchID              string    `gorm:"size:64;not null" json:"mch_id"`                      // 货源商户ID
	MchSecretEncrypted string    `gorm:"column:mch_secret_encrypted" json:"-"`               // 加密存储的mch_secret
	Enabled            bool      `gorm:"default:false" json:"enabled"`                       // 是否启用
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GoofishConfig) TableName() string {
	return "goofish_configs"
}

// =====================================================
// 商品映射
// =====================================================

// GoofishGoods 商品映射（商品编码与卡密类型的对应关系）
type GoofishGoods struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	GoodsNo         string    `gorm:"uniqueIndex;size:64;not null" json:"goods_no"`          // 商品编码
	GoodsName       string    `gorm:"size:255;not null" json:"goods_name"`                   // 商品名称
	CardType        int8      `gorm:"not null" json:"card_type"`                             // 卡密类型 1月卡 2季卡 3半年卡 4年卡
	Price           int64     `gorm:"not null" json:"price"`                                 // 价格（分）
	Status          int8      `gorm:"default:1" json:"status"`                               // 状态 1在架 2下架
	AutoGenerate    bool      `gorm:"default:false" json:"auto_generate"`                    // 是否自动生成卡密
	CardPrefix      string    `gorm:"size:32" json:"card_prefix"`                            // 卡密前缀
	Duration        int       `gorm:"default:30" json:"duration"`                            // 有效天数
	MaxAutoGenerate int       `gorm:"default:10" json:"max_auto_generate"`                   // 单次最大自动生成数量
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GoofishGoods) TableName() string {
	return "goofish_goods"
}

// 商品状态常量
const (
	GoofishGoodsStatusOnSale  int8 = 1 // 在架
	GoofishGoodsStatusOffSale int8 = 2 // 下架
)

// 商品类型常量（固定为卡密商品）
const (
	GoofishGoodsTypeKami int = 2 // 卡密商品
)

// =====================================================
// 订单记录
// =====================================================

// GoofishOrder 闲管家订单记录
type GoofishOrder struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderNo     string    `gorm:"uniqueIndex;size:64;not null" json:"order_no"`   // 闲管家订单号
	OutOrderNo  string    `gorm:"size:64" json:"out_order_no"`                    // 本系统订单号
	BizOrderNo  string    `gorm:"size:64;index" json:"biz_order_no"`              // 业务订单号（闲鱼订单号）
	GoodsNo     string    `gorm:"size:64;not null;index" json:"goods_no"`         // 商品编码
	GoodsName   string    `gorm:"size:255" json:"goods_name"`                     // 商品名称
	BuyQuantity int       `gorm:"default:1" json:"buy_quantity"`                  // 购买数量
	OrderAmount int64     `gorm:"not null" json:"order_amount"`                   // 订单金额（分）
	OrderStatus int8      `gorm:"default:20" json:"order_status"`                 // 订单状态 10处理中 20已成功 30已失败
	CardCodes   string    `gorm:"type:text" json:"card_codes"`                    // 发货的卡密码（JSON数组）
	Remark      string    `gorm:"size:500" json:"remark"`                         // 备注
	ClientIP    string    `gorm:"size:64" json:"client_ip"`                       // 请求来源IP
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GoofishOrder) TableName() string {
	return "goofish_orders"
}

// 订单状态常量
const (
	GoofishOrderStatusProcessing int8 = 10 // 处理中
	GoofishOrderStatusSuccess    int8 = 20 // 已成功
	GoofishOrderStatusFailed     int8 = 30 // 已失败
)

// =====================================================
// 发货记录（防重复）
// =====================================================

// GoofishOrderCard 发货记录（用于幂等处理，防止重复发货）
type GoofishOrderCard struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderNo   string    `gorm:"index;size:64;not null" json:"order_no"`  // 闲管家订单号
	CardID    uint64    `gorm:"not null" json:"card_id"`                 // 卡密ID
	CardCode  string    `gorm:"size:64;not null" json:"card_code"`       // 卡密码
	GoodsNo   string    `gorm:"size:64;not null" json:"goods_no"`        // 商品编码
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (GoofishOrderCard) TableName() string {
	return "goofish_order_cards"
}

// =====================================================
// 商品订阅
// =====================================================

// GoofishGoodsSubscription 商品变更订阅
type GoofishGoodsSubscription struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	GoodsType     int       `gorm:"not null" json:"goods_type"`                           // 商品类型 1直充 2卡密
	GoodsNo       string    `gorm:"size:64;not null;index" json:"goods_no"`               // 商品编码
	Token         string    `gorm:"size:64" json:"token"`                                 // 订阅标识（固定回调地址模式）
	NotifyURL     string    `gorm:"size:512" json:"notify_url"`                           // 回调地址（动态回调地址模式）
	SubscribeTime time.Time `json:"subscribe_time"`                                       // 订阅时间
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (GoofishGoodsSubscription) TableName() string {
	return "goofish_goods_subscriptions"
}

// =====================================================
// API日志
// =====================================================

// GoofishLog API调用日志
type GoofishLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Endpoint     string    `gorm:"size:128;not null;index" json:"endpoint"`  // 接口路径
	Method       string    `gorm:"size:16" json:"method"`                    // 请求方法
	RequestBody  string    `gorm:"type:text" json:"request_body"`            // 请求内容
	ResponseBody string    `gorm:"type:text" json:"response_body"`           // 响应内容
	ResponseCode int       `gorm:"default:0" json:"response_code"`           // 响应码
	Duration     int64     `gorm:"default:0" json:"duration"`                // 耗时(ms)
	ClientIP     string    `gorm:"size:64" json:"client_ip"`                 // 客户端IP
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (GoofishLog) TableName() string {
	return "goofish_logs"
}
