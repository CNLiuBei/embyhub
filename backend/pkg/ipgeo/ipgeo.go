// Package ipgeo IP地理位置解析
package ipgeo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Location 地理位置信息
type Location struct {
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省份
	City     string `json:"city"`     // 城市
}

// IPGeoService IP地理位置服务
type IPGeoService struct {
	cache      map[string]*Location
	cacheMutex sync.RWMutex
	client     *http.Client
}

var (
	instance *IPGeoService
	once     sync.Once
)

// GetInstance 获取单例实例
func GetInstance() *IPGeoService {
	once.Do(func() {
		instance = &IPGeoService{
			cache: make(map[string]*Location),
			client: &http.Client{
				Timeout: 3 * time.Second,
			},
		}
	})
	return instance
}

// GetLocation 获取IP地理位置
func (s *IPGeoService) GetLocation(ip string) string {
	// 处理本地IP - 不显示地区
	if s.isLocalIP(ip) {
		return ""
	}

	// 检查缓存
	s.cacheMutex.RLock()
	if loc, ok := s.cache[ip]; ok {
		s.cacheMutex.RUnlock()
		return s.formatLocation(loc)
	}
	s.cacheMutex.RUnlock()

	// 查询IP位置
	loc := s.queryIP(ip)
	if loc != nil {
		s.cacheMutex.Lock()
		s.cache[ip] = loc
		s.cacheMutex.Unlock()
		return s.formatLocation(loc)
	}

	return "未知"
}

// isLocalIP 判断是否是本地IP
func (s *IPGeoService) isLocalIP(ip string) bool {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 私有IP段
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	}

	for _, block := range privateBlocks {
		_, cidr, _ := net.ParseCIDR(block)
		if cidr.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// queryIP 查询IP地理位置（使用多个免费API）
func (s *IPGeoService) queryIP(ip string) *Location {
	// 尝试使用 ip-api.com（免费，每分钟45次）
	loc := s.queryIPAPI(ip)
	if loc != nil {
		return loc
	}

	// 备用：使用 ip.sb
	loc = s.queryIPSB(ip)
	if loc != nil {
		return loc
	}

	return nil
}

// queryIPAPI 使用 ip-api.com 查询
func (s *IPGeoService) queryIPAPI(ip string) *Location {
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN&fields=status,country,regionName,city", ip)
	
	resp, err := s.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Status     string `json:"status"`
		Country    string `json:"country"`
		RegionName string `json:"regionName"`
		City       string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	if result.Status != "success" {
		return nil
	}

	return &Location{
		Country:  result.Country,
		Province: result.RegionName,
		City:     result.City,
	}
}

// queryIPSB 使用 ip.sb 查询
func (s *IPGeoService) queryIPSB(ip string) *Location {
	url := fmt.Sprintf("https://api.ip.sb/geoip/%s", ip)
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Country  string `json:"country"`
		Region   string `json:"region"`
		City     string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	return &Location{
		Country:  result.Country,
		Province: result.Region,
		City:     result.City,
	}
}

// formatLocation 格式化地理位置显示
func (s *IPGeoService) formatLocation(loc *Location) string {
	if loc == nil {
		return "未知"
	}

	// 国内显示到市级
	if loc.Country == "中国" || loc.Country == "China" {
		if loc.City != "" && loc.City != loc.Province {
			// 去掉"省"、"市"等后缀，简化显示
			province := strings.TrimSuffix(loc.Province, "省")
			province = strings.TrimSuffix(province, "市")
			city := strings.TrimSuffix(loc.City, "市")
			
			// 直辖市特殊处理
			if province == city || loc.Province == loc.City {
				return city
			}
			return province + "·" + city
		}
		if loc.Province != "" {
			return strings.TrimSuffix(strings.TrimSuffix(loc.Province, "省"), "市")
		}
		return "中国"
	}

	// 国外只显示国家
	if loc.Country != "" {
		return loc.Country
	}

	return "未知"
}

// GetLocationAsync 异步获取IP地理位置（不阻塞）
func (s *IPGeoService) GetLocationAsync(ip string, callback func(string)) {
	go func() {
		location := s.GetLocation(ip)
		if callback != nil {
			callback(location)
		}
	}()
}
