import React, { useState, useEffect } from 'react';
import { Card, message, Button } from 'antd';
import { CheckCircleOutlined } from '@ant-design/icons';
import LogoIcon from '@/components/LogoIcon';
import { Link, useNavigate } from 'react-router-dom';
import { post } from '@/utils/request';
import ColorDots from '@/components/ColorDots';
import './Login.css';

const ForgotPassword: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [sendingCode, setSendingCode] = useState(false);
  const [success, setSuccess] = useState(false);
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [focused, setFocused] = useState<string | null>(null);

  // 倒计时
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [countdown]);

  // 发送验证码
  const handleSendCode = async () => {
    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      message.error('请输入有效的邮箱地址');
      return;
    }

    setSendingCode(true);
    try {
      const res = await post('/email/reset-code', { email });
      if (res.code === 200) {
        message.success('验证码已发送');
        setCountdown(60);
      }
    } catch (error: any) {
      message.error(error.message || '发送失败');
    } finally {
      setSendingCode(false);
    }
  };

  // 重置密码
  const handleReset = async () => {
    if (!email) { message.error('请输入邮箱'); return; }
    if (!code || code.length !== 6) { message.error('请输入6位验证码'); return; }
    if (!password || password.length < 6) { message.error('密码至少6个字符'); return; }
    if (password !== confirmPassword) { message.error('两次密码不一致'); return; }

    setLoading(true);
    try {
      const res = await post('/email/reset-password', { email, code, password });
      if (res.code === 200) {
        setSuccess(true);
        setTimeout(() => navigate('/login'), 2000);
      }
    } catch (error: any) {
      message.error(error.message || '重置失败');
    } finally {
      setLoading(false);
    }
  };

  // 成功页面
  if (success) {
    return (
      <div className="login-container">
        <div className="login-brand">Emby Hub</div>
        <div className="login-box">
          <Card className="login-card" variant="borderless">
            <div style={{ textAlign: 'center', padding: '60px 0' }}>
              <CheckCircleOutlined style={{ fontSize: 72, color: '#52c41a' }} />
              <h2 style={{ marginTop: 24, color: '#1d1d1f' }}>密码重置成功</h2>
              <p style={{ color: '#86868b', marginTop: 8 }}>正在跳转到登录页...</p>
            </div>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="login-container">
      <div className="login-brand">Emby Hub</div>

      <div className="login-box">
        <Card className="login-card" variant="borderless">
          {/* Logo */}
          <div className="login-logo-wrapper">
            <div className="login-logo">
              <ColorDots />
              <LogoIcon size={52} />
            </div>
          </div>

          {/* 标题 */}
          <div className="login-header">
            <h1>找回密码</h1>
          </div>
          
          {/* 输入框 */}
          <div className="login-input-box">
            {/* 邮箱 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'email' || email ? '' : 'login-input-placeholder'}`}>
                邮箱地址
              </span>
              <input
                type="email"
                placeholder={focused === 'email' || email ? '' : '注册时使用的邮箱'}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onFocus={() => setFocused('email')}
                onBlur={() => setFocused(null)}
                style={{ paddingRight: 100 }}
              />
              <button 
                className="login-code-btn"
                onClick={handleSendCode}
                disabled={sendingCode || countdown > 0}
              >
                {sendingCode ? '发送中...' : countdown > 0 ? `${countdown}s` : '获取验证码'}
              </button>
            </div>

            {/* 验证码 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'code' || code ? '' : 'login-input-placeholder'}`}>
                验证码
              </span>
              <input
                type="text"
                placeholder={focused === 'code' || code ? '' : '6位验证码'}
                value={code}
                maxLength={6}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
                onFocus={() => setFocused('code')}
                onBlur={() => setFocused(null)}
              />
            </div>

            {/* 新密码 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'password' || password ? '' : 'login-input-placeholder'}`}>
                新密码
              </span>
              <input
                type="password"
                placeholder={focused === 'password' || password ? '' : '新密码（至少6个字符）'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onFocus={() => setFocused('password')}
                onBlur={() => setFocused(null)}
              />
            </div>

            {/* 确认密码 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'confirm' || confirmPassword ? '' : 'login-input-placeholder'}`}>
                确认新密码
              </span>
              <input
                type="password"
                placeholder={focused === 'confirm' || confirmPassword ? '' : '确认新密码'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                onFocus={() => setFocused('confirm')}
                onBlur={() => setFocused(null)}
              />
            </div>
          </div>

          {/* 重置按钮 */}
          <Button
            type="primary"
            block
            size="large"
            loading={loading}
            onClick={handleReset}
            style={{
              marginTop: 24,
              height: 48,
              borderRadius: 12,
              fontSize: 16,
              fontWeight: 500,
            }}
          >
            重置密码
          </Button>

          {/* 底部链接 */}
          <div className="login-links" style={{ borderTop: 'none', marginTop: 20 }}>
            <div className="login-links-row">
              <Link to="/login">返回登录</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default ForgotPassword;
