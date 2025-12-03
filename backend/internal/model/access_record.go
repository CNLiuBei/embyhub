package model

import "time"

// AccessRecord 访问记录模型
type AccessRecord struct {
	RecordID   int64     `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`
	UserID     int       `gorm:"column:user_id;not null;index" json:"user_id"`
	AccessTime time.Time `gorm:"column:access_time;not null;default:CURRENT_TIMESTAMP;index" json:"access_time"`
	Resource   string    `gorm:"column:resource;type:varchar(200)" json:"resource"`
	IPAddress  string    `gorm:"column:ip_address;type:varchar(50)" json:"ip_address"`
	DeviceInfo string    `gorm:"column:device_info;type:varchar(100)" json:"device_info"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (AccessRecord) TableName() string {
	return "access_records"
}

// AccessRecordListRequest 访问记录查询请求
type AccessRecordListRequest struct {
	Page      int       `form:"page" binding:"omitempty,gt=0"`
	PageSize  int       `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	UserID    int       `form:"user_id" binding:"omitempty,gt=0"`
	StartTime time.Time `form:"start_time" binding:"omitempty"`
	EndTime   time.Time `form:"end_time" binding:"omitempty"`
	Resource  string    `form:"resource"`
}

// AccessRecordListResponse 访问记录列表响应
type AccessRecordListResponse struct {
	Total int             `json:"total"`
	List  []*AccessRecord `json:"list"`
}

// AccessRecordCreateRequest 创建访问记录请求
type AccessRecordCreateRequest struct {
	UserID     int    `json:"user_id" binding:"required,gt=0"`
	Resource   string `json:"resource" binding:"omitempty,max=200"`
	IPAddress  string `json:"ip_address" binding:"omitempty,max=50"`
	DeviceInfo string `json:"device_info" binding:"omitempty,max=100"`
}
