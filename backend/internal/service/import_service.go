// Package service 批量导入服务
package service

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"time"

	"feiniu-user-system/internal/models"
	"feiniu-user-system/pkg/utils"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ImportService 导入服务
type ImportService struct {
	db *gorm.DB
}

// NewImportService 创建导入服务
func NewImportService(db *gorm.DB) *ImportService {
	return &ImportService{db: db}
}

// ImportResult 导入结果
type ImportResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// UserImportRow 用户导入行数据
type UserImportRow struct {
	Username string
	Email    string
	Password string
	Nickname string
}

// ImportFromCSV 从CSV导入用户
func (s *ImportService) ImportFromCSV(reader io.Reader, adminID uuid.UUID) (*ImportResult, error) {
	csvReader := csv.NewReader(reader)

	// 跳过标题行
	_, err := csvReader.Read()
	if err != nil {
		return nil, errors.New("读取CSV文件失败")
	}

	result := &ImportResult{}
	var rows []UserImportRow

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, "CSV格式错误")
			continue
		}

		if len(record) < 3 {
			result.Errors = append(result.Errors, "列数不足")
			continue
		}

		rows = append(rows, UserImportRow{
			Username: strings.TrimSpace(record[0]),
			Email:    strings.TrimSpace(record[1]),
			Password: strings.TrimSpace(record[2]),
			Nickname: func() string {
				if len(record) > 3 {
					return strings.TrimSpace(record[3])
				}
				return ""
			}(),
		})
	}

	return s.importUsers(rows, result)
}

// ImportFromExcel 从Excel导入用户
func (s *ImportService) ImportFromExcel(reader io.Reader, adminID uuid.UUID) (*ImportResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, errors.New("读取Excel文件失败")
	}
	defer f.Close()

	// 获取第一个工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("Excel文件中没有工作表")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, errors.New("读取工作表失败")
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel文件中没有数据")
	}

	result := &ImportResult{}
	var importRows []UserImportRow

	// 从第二行开始（跳过标题）
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			result.Errors = append(result.Errors, "第"+string(rune('0'+i+1))+"行: 列数不足")
			continue
		}

		importRows = append(importRows, UserImportRow{
			Username: strings.TrimSpace(row[0]),
			Email:    strings.TrimSpace(row[1]),
			Password: strings.TrimSpace(row[2]),
			Nickname: func() string {
				if len(row) > 3 {
					return strings.TrimSpace(row[3])
				}
				return ""
			}(),
		})
	}

	return s.importUsers(importRows, result)
}

// importUsers 执行用户导入
func (s *ImportService) importUsers(rows []UserImportRow, result *ImportResult) (*ImportResult, error) {
	result.Total = len(rows)

	for i, row := range rows {
		rowNum := i + 2 // Excel行号从1开始，加上标题行

		// 验证数据
		if row.Username == "" {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 用户名为空")
			continue
		}
		if row.Email == "" {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 邮箱为空")
			continue
		}
		if row.Password == "" {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 密码为空")
			continue
		}
		if len(row.Password) < 6 {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 密码长度至少6位")
			continue
		}

		// 检查用户名是否已存在
		var count int64
		s.db.Model(&models.User{}).Where("username = ?", row.Username).Count(&count)
		if count > 0 {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 用户名已存在")
			continue
		}

		// 检查邮箱是否已存在
		s.db.Model(&models.User{}).Where("email = ?", row.Email).Count(&count)
		if count > 0 {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 邮箱已存在")
			continue
		}

		// 加密密码
		hashedPassword, err := utils.HashPassword(row.Password)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 密码加密失败")
			continue
		}

		// 创建用户
		user := &models.User{
			Username:  row.Username,
			Email:     row.Email,
			Password:  hashedPassword,
			Nickname:  row.Nickname,
			Status:    1,
			Role:      models.RoleUser,
			CreatedAt: time.Now(),
		}
		if user.Nickname == "" {
			user.Nickname = user.Username
		}

		if err := s.db.Create(user).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, "第"+itoa(rowNum)+"行: 创建失败")
			continue
		}

		result.Success++
	}

	// 限制错误信息数量
	if len(result.Errors) > 20 {
		result.Errors = append(result.Errors[:20], "...更多错误已省略")
	}

	return result, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// GetImportTemplate 获取导入模板
func (s *ImportService) GetImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// 设置标题行
	f.SetCellValue("Sheet1", "A1", "用户名")
	f.SetCellValue("Sheet1", "B1", "邮箱")
	f.SetCellValue("Sheet1", "C1", "密码")
	f.SetCellValue("Sheet1", "D1", "昵称（可选）")

	// 添加示例数据
	f.SetCellValue("Sheet1", "A2", "user001")
	f.SetCellValue("Sheet1", "B2", "user001@example.com")
	f.SetCellValue("Sheet1", "C2", "password123")
	f.SetCellValue("Sheet1", "D2", "用户1")

	// 设置列宽
	f.SetColWidth("Sheet1", "A", "D", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
