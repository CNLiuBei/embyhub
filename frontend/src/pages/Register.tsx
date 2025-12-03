import React, { useState, useEffect } from 'react';
import { Card, message, Button } from 'antd';
import LogoIcon from '@/components/LogoIcon';
import { useNavigate, Link } from 'react-router-dom';
import { register, sendEmailCode } from '@/api/auth';
import ColorDots from '@/components/ColorDots';
import './Login.css';

const Register: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [sendingCode, setSendingCode] = useState(false);
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [username, setUsername] = useState('');
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
      const response = await sendEmailCode({ email, type: 'register' });
      if (response.code === 200) {
        message.success('验证码已发送');
        setCountdown(60);
      } else {
        message.error(response.message || '发送失败');
      }
    } catch (error: any) {
      message.error(error.message || '发送失败');
    } finally {
      setSendingCode(false);
    }
  };

  // 提交注册
  const handleRegister = async () => {
    if (!email) { message.error('请输入邮箱'); return; }
    if (!code || code.length !== 6) { message.error('请输入6位验证码'); return; }
    if (!username || username.length < 3) { message.error('用户名至少3个字符'); return; }
    if (!password || password.length < 6) { message.error('密码至少6个字符'); return; }
    if (password !== confirmPassword) { message.error('两次密码不一致'); return; }

    setLoading(true);
    try {
      const response = await register({ email, code, username, password });
      if (response.code === 200) {
        message.success('注册成功！');
        setTimeout(() => navigate('/login'), 1500);
      } else {
        message.error(response.message || '注册失败');
      }
    } catch (error: any) {
      message.error(error.message || '注册失败');
    } finally {
      setLoading(false);
    }
  };

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
            <h1>创建账户</h1>
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
                placeholder={focused === 'email' || email ? '' : '邮箱地址'}
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

            {/* 用户名 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'username' || username ? '' : 'login-input-placeholder'}`}>
                用户名
              </span>
              <input
                type="text"
                placeholder={focused === 'username' || username ? '' : '用户名（至少3个字符）'}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onFocus={() => setFocused('username')}
                onBlur={() => setFocused(null)}
              />
            </div>

            {/* 密码 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'password' || password ? '' : 'login-input-placeholder'}`}>
                密码
              </span>
              <input
                type="password"
                placeholder={focused === 'password' || password ? '' : '密码（至少6个字符）'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onFocus={() => setFocused('password')}
                onBlur={() => setFocused(null)}
              />
            </div>

            {/* 确认密码 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'confirm' || confirmPassword ? '' : 'login-input-placeholder'}`}>
                确认密码
              </span>
              <input
                type="password"
                placeholder={focused === 'confirm' || confirmPassword ? '' : '确认密码'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                onFocus={() => setFocused('confirm')}
                onBlur={() => setFocused(null)}
              />
            </div>
          </div>

          {/* 注册按钮 */}
          <Button
            type="primary"
            block
            size="large"
            loading={loading}
            onClick={handleRegister}
            style={{
              marginTop: 24,
              height: 48,
              borderRadius: 12,
              fontSize: 16,
              fontWeight: 500,
            }}
          >
            注册
          </Button>

          {/* 底部链接 */}
          <div className="login-links" style={{ borderTop: 'none', marginTop: 20 }}>
            <div className="login-links-row">
              <Link to="/login">已有账户？立即登录</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default Register;
