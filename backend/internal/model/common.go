package model

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	UserInfo *User  `json:"user_info"`
}

// StatisticsResponse 统计数据响应
type StatisticsResponse struct {
	TotalUsers   int64              `json:"total_users"`
	ActiveUsers  int64              `json:"active_users"`
	TodayAccess  int64              `json:"today_access"`
	TopUsers     []*TopUserItem     `json:"top_users"`
	AccessTrend  []*AccessTrendItem `json:"access_trend"`
	UserGrowth   []*GrowthTrendItem `json:"user_growth,omitempty"`   // 用户增长趋势
	VipStats     *VipStatistics     `json:"vip_stats,omitempty"`     // VIP统计
	CardKeyStats *CardKeyStatistics `json:"cardkey_stats,omitempty"` // 卡密统计
}

// GrowthTrendItem 增长趋势项
type GrowthTrendItem struct {
	Date       string `json:"date"`
	NewUsers   int64  `json:"new_users"`   // 新增用户
	TotalUsers int64  `json:"total_users"` // 累计用户
}

// VipStatistics VIP统计
type VipStatistics struct {
	TotalVip     int64 `json:"total_vip"`      // 有效VIP总数
	ExpiredVip   int64 `json:"expired_vip"`    // 已过期VIP
	Expiring3Day int64 `json:"expiring_3_day"` // 3天内到期
	Expiring7Day int64 `json:"expiring_7_day"` // 7天内到期
}

// CardKeyStatistics 卡密统计
type CardKeyStatistics struct {
	TotalCards    int64 `json:"total_cards"`    // 卡密总数
	UnusedCards   int64 `json:"unused_cards"`   // 未使用
	UsedCards     int64 `json:"used_cards"`     // 已使用
	DisabledCards int64 `json:"disabled_cards"` // 已禁用
}

// TopUserItem Top用户项
type TopUserItem struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	AccessCount int64  `json:"access_count"`
}

// AccessTrendItem 访问趋势项
type AccessTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// EmbyUser Emby用户信息
type EmbyUser struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	HasPassword           bool   `json:"has_password"`
	HasConfiguredPassword bool   `json:"has_configured_password"`
	LastLoginDate         string `json:"last_login_date"`
	LastActivityDate      string `json:"last_activity_date"`
}

// EmbyUsersResponse Emby用户列表响应
type EmbyUsersResponse struct {
	Items            []*EmbyUser `json:"items"`
	TotalRecordCount int         `json:"total_record_count"`
}
