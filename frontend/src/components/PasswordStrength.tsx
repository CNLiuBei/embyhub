import React from 'react';
import { Progress } from 'antd';

interface PasswordStrengthProps {
  password: string;
}

// 密码强度等级
const getPasswordStrength = (password: string): { level: number; text: string; color: string } => {
  if (!password) return { level: 0, text: '', color: '#d9d9d9' };
  
  let score = 0;
  
  // 长度检查
  if (password.length >= 6) score += 20;
  if (password.length >= 8) score += 10;
  if (password.length >= 12) score += 10;
  
  // 包含小写字母
  if (/[a-z]/.test(password)) score += 15;
  
  // 包含大写字母
  if (/[A-Z]/.test(password)) score += 15;
  
  // 包含数字
  if (/[0-9]/.test(password)) score += 15;
  
  // 包含特殊字符
  if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) score += 15;
  
  if (score < 30) return { level: score, text: '弱', color: '#ff4d4f' };
  if (score < 60) return { level: score, text: '中', color: '#faad14' };
  if (score < 80) return { level: score, text: '强', color: '#52c41a' };
  return { level: score, text: '很强', color: '#1890ff' };
};

// 密码强度提示
const getPasswordTips = (password: string): string[] => {
  const tips: string[] = [];
  
  if (!password) return tips;
  
  if (password.length < 8) tips.push('建议至少8个字符');
  if (!/[a-z]/.test(password)) tips.push('添加小写字母');
  if (!/[A-Z]/.test(password)) tips.push('添加大写字母');
  if (!/[0-9]/.test(password)) tips.push('添加数字');
  if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) tips.push('添加特殊字符');
  
  return tips;
};

const PasswordStrength: React.FC<PasswordStrengthProps> = ({ password }) => {
  const strength = getPasswordStrength(password);
  const tips = getPasswordTips(password);
  
  if (!password) return null;
  
  return (
    <div style={{ marginTop: -16, marginBottom: 8 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <Progress 
          percent={strength.level} 
          size="small" 
          showInfo={false}
          strokeColor={strength.color}
          style={{ flex: 1 }}
        />
        <span style={{ color: strength.color, fontSize: 12, minWidth: 32 }}>{strength.text}</span>
      </div>
      {tips.length > 0 && (
        <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
          提示: {tips.join('、')}
        </div>
      )}
    </div>
  );
};

export default PasswordStrength;

// 密码验证规则（用于表单）
export const passwordRules = [
  { required: true, message: '请输入密码' },
  { min: 6, message: '密码至少6个字符' },
  { max: 50, message: '密码最多50个字符' },
  {
    pattern: /^(?=.*[a-zA-Z])(?=.*[0-9])/,
    message: '密码需包含字母和数字',
  },
];

// 强密码验证规则
export const strongPasswordRules = [
  { required: true, message: '请输入密码' },
  { min: 8, message: '密码至少8个字符' },
  { max: 50, message: '密码最多50个字符' },
  {
    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?])/,
    message: '密码需包含大小写字母、数字和特殊字符',
  },
];
