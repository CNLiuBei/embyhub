package util

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword 验证密码强度
// 要求：至少6位，包含字母和数字
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("密码至少6个字符")
	}
	if len(password) > 50 {
		return errors.New("密码最多50个字符")
	}

	// 检查是否包含字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	if !hasLetter {
		return errors.New("密码需包含字母")
	}

	// 检查是否包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return errors.New("密码需包含数字")
	}

	return nil
}

// ValidateStrongPassword 验证强密码
// 要求：至少8位，包含大小写字母、数字和特殊字符
func ValidateStrongPassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码至少8个字符")
	}
	if len(password) > 50 {
		return errors.New("密码最多50个字符")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("密码需包含小写字母")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("密码需包含大写字母")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("密码需包含数字")
	}
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return errors.New("密码需包含特殊字符")
	}

	return nil
}
