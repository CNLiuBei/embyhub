// Package service 卡密导出服务
package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"feiniu-user-system/internal/models"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// CardExportService 卡密导出服务
type CardExportService struct {
	db *gorm.DB
}

// NewCardExportService 创建卡密导出服务
func NewCardExportService(db *gorm.DB) *CardExportService {
	return &CardExportService{db: db}
}

// ExportFilter 导出筛选条件
type ExportFilter struct {
	BatchNo  string `json:"batch_no"`
	CardType int8   `json:"card_type"`
	Status   *int8  `json:"status"`
	Limit    int    `json:"limit"` // 最大导出数量，0表示全部
}

// ExportToCSV 导出为CSV格式
func (s *CardExportService) ExportToCSV(filter *ExportFilter) ([]byte, string, error) {
	cards, err := s.queryCards(filter)
	if err != nil {
		return nil, "", err
	}

	if len(cards) == 0 {
		return nil, "", fmt.Errorf("没有符合条件的卡密数据")
	}

	// 创建CSV缓冲区
	buf := new(bytes.Buffer)
	// 添加UTF-8 BOM，让Excel正确识别中文
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)
	writer.Comma = ',' // 使用逗号作为分隔符

	// 写入表头
	headers := []string{"ID", "卡密码", "批次号", "卡密类型", "天数", "状态", "使用者ID", "使用时间", "过期时间", "创建时间", "备注"}
	if err := writer.Write(headers); err != nil {
		return nil, "", err
	}

	// 写入数据
	for _, card := range cards {
		usedBy := ""
		if card.UsedBy != nil {
			usedBy = card.UsedBy.String()
		}

		usedAt := ""
		if card.UsedAt != nil {
			usedAt = card.UsedAt.Format("2006-01-02 15:04:05")
		}

		expireAt := ""
		if card.ExpireAt != nil {
			expireAt = card.ExpireAt.Format("2006-01-02 15:04:05")
		}

		row := []string{
			fmt.Sprintf("%d", card.ID),
			card.Code,
			card.BatchNo,
			card.GetCardTypeName(),
			fmt.Sprintf("%d", card.Duration),
			card.GetStatusName(),
			usedBy,
			usedAt,
			expireAt,
			card.CreatedAt.Format("2006-01-02 15:04:05"),
			card.Remark,
		}

		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("cards_%s.csv", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// ExportToExcel 导出为Excel格式
func (s *CardExportService) ExportToExcel(filter *ExportFilter) ([]byte, string, error) {
	cards, err := s.queryCards(filter)
	if err != nil {
		return nil, "", err
	}

	if len(cards) == 0 {
		return nil, "", fmt.Errorf("没有符合条件的卡密数据")
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "卡密列表"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 写入表头
	headers := []string{"ID", "卡密码", "批次号", "卡密类型", "天数", "状态", "使用者ID", "使用时间", "过期时间", "创建时间", "备注"}
	for i, h := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 8)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 18)
	f.SetColWidth(sheetName, "D", "D", 10)
	f.SetColWidth(sheetName, "E", "E", 8)
	f.SetColWidth(sheetName, "F", "F", 10)
	f.SetColWidth(sheetName, "G", "G", 38)
	f.SetColWidth(sheetName, "H", "I", 20)
	f.SetColWidth(sheetName, "J", "K", 20)

	// 写入数据
	for i, card := range cards {
		row := i + 2

		usedBy := ""
		if card.UsedBy != nil {
			usedBy = card.UsedBy.String()
		}

		usedAt := ""
		if card.UsedAt != nil {
			usedAt = card.UsedAt.Format("2006-01-02 15:04:05")
		}

		expireAt := ""
		if card.ExpireAt != nil {
			expireAt = card.ExpireAt.Format("2006-01-02 15:04:05")
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), card.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), card.Code)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), card.BatchNo)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), card.GetCardTypeName())
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), card.Duration)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), card.GetStatusName())
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), usedBy)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), usedAt)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), expireAt)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), card.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), card.Remark)
	}

	// 生成文件
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("cards_%s.xlsx", time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// ExportCodesOnly 仅导出卡密码（用于发放）
func (s *CardExportService) ExportCodesOnly(batchNo string) (string, string, error) {
	var cards []models.Card
	query := s.db.Where("batch_no = ? AND status = ?", batchNo, models.CardStatusUnused)

	if err := query.Find(&cards).Error; err != nil {
		return "", "", err
	}

	if len(cards) == 0 {
		return "", "", fmt.Errorf("没有可用的卡密")
	}

	// 生成卡密列表文本
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("批次号: %s\n", batchNo))
	builder.WriteString(fmt.Sprintf("总数量: %d 张\n", len(cards)))
	builder.WriteString(fmt.Sprintf("导出时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString("\n========== 卡密列表 ==========\n\n")

	for i, card := range cards {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, card.Code))
	}

	builder.WriteString("\n========== 温馨提示 ==========\n")
	builder.WriteString("1. 请妥善保管卡密，避免泄露\n")
	builder.WriteString("2. 每个卡密仅可使用一次\n")
	builder.WriteString("3. 使用后会自动升级为会员\n")

	filename := fmt.Sprintf("codes_%s.txt", time.Now().Format("20060102_150405"))
	return builder.String(), filename, nil
}

// GenerateUsageReport 生成使用情况报告
func (s *CardExportService) GenerateUsageReport(batchNo string) ([]byte, string, error) {
	var batch models.CardBatch
	if err := s.db.Where("batch_no = ?", batchNo).First(&batch).Error; err != nil {
		return nil, "", fmt.Errorf("批次不存在")
	}

	// 统计各状态数量
	var stats struct {
		Unused   int64
		Used     int64
		Expired  int64
		Disabled int64
	}

	s.db.Model(&models.Card{}).Where("batch_no = ? AND status = ?", batchNo, models.CardStatusUnused).Count(&stats.Unused)
	s.db.Model(&models.Card{}).Where("batch_no = ? AND status = ?", batchNo, models.CardStatusUsed).Count(&stats.Used)
	s.db.Model(&models.Card{}).Where("batch_no = ? AND status = ?", batchNo, models.CardStatusExpired).Count(&stats.Expired)
	s.db.Model(&models.Card{}).Where("batch_no = ? AND status = ?", batchNo, models.CardStatusDisable).Count(&stats.Disabled)

	// 创建报告Excel
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "使用报告"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// 写入报告内容
	row := 1
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "卡密使用情况报告")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row))
	row += 2

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "批次号:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), batch.BatchNo)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "卡密类型:")
	typeName := "月卡"
	if batch.CardType == 2 {
		typeName = "年卡"
	}
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), typeName)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "生成数量:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), batch.Quantity)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "创建时间:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), batch.CreatedAt.Format("2006-01-02 15:04:05"))
	row += 2

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "状态统计")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row))
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "未使用:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats.Unused)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "已使用:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats.Used)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "已过期:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats.Expired)
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "已禁用:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), stats.Disabled)
	row++

	usageRate := float64(0)
	if batch.Quantity > 0 {
		usageRate = float64(stats.Used) / float64(batch.Quantity) * 100
	}
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "使用率:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f%%", usageRate))

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 25)

	// 生成文件
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("report_%s_%s.xlsx", batchNo, time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// queryCards 查询卡密列表
func (s *CardExportService) queryCards(filter *ExportFilter) ([]models.Card, error) {
	var cards []models.Card
	query := s.db.Model(&models.Card{})

	if filter != nil {
		if filter.BatchNo != "" {
			query = query.Where("batch_no = ?", filter.BatchNo)
		}
		if filter.CardType > 0 {
			query = query.Where("card_type = ?", filter.CardType)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
	}

	if err := query.Order("id ASC").Find(&cards).Error; err != nil {
		return nil, err
	}

	return cards, nil
}
