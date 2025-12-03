import React, { useState, useRef, useEffect } from 'react';
import { Card, message, Checkbox } from 'antd';
import { ArrowRightOutlined } from '@ant-design/icons';
import LogoIcon from '@/components/LogoIcon';
import { useNavigate, Link } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { login } from '@/api/auth';
import { setAuthInfo } from '@/store/slices/authSlice';
import ColorDots from '@/components/ColorDots';
import './Login.css';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const [loading, setLoading] = useState(false);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [focused, setFocused] = useState<'username' | 'password' | null>(null);
  const passwordRef = useRef<HTMLInputElement>(null);

  // 处理登录
  const handleLogin = async () => {
    if (!username) {
      message.error('请输入用户名');
      return;
    }
    if (showPassword && !password) {
      message.error('请输入密码');
      return;
    }
    
    // 如果还没显示密码框，先显示
    if (!showPassword) {
      setShowPassword(true);
      return;
    }

    setLoading(true);
    try {
      const response = await login({ username, password });
      
      if (response.code === 200 && response.data) {
        dispatch(setAuthInfo({
          token: response.data.token,
          userInfo: response.data.user_info
        }));
        message.success('登录成功');
        navigate('/');
      } else {
        message.error(response.message || '登录失败');
      }
    } catch (error: any) {
      message.error(error.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  // 显示密码框后自动聚焦
  useEffect(() => {
    if (showPassword && passwordRef.current) {
      passwordRef.current.focus();
    }
  }, [showPassword]);

  // 回车提交
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleLogin();
    }
  };

  return (
    <div className="login-container">
      <div className="login-brand">Emby Hub</div>

      <div className="login-box">
        <Card className="login-card" variant="borderless">
          {/* 彩色圆点Logo */}
          <div className="login-logo-wrapper">
            <div className="login-logo">
              <ColorDots />
              <LogoIcon size={52} />
            </div>
          </div>

          {/* 标题 */}
          <div className="login-header">
            <h1>登录 Emby Hub</h1>
          </div>
          
          {/* 输入框容器 */}
          <div className={`login-input-box ${showPassword ? 'has-password' : ''}`}>
            {/* 用户名输入 */}
            <div className="login-input-item">
              <span className={`login-input-label ${focused === 'username' || username ? '' : 'login-input-placeholder'}`}>
                用户名或邮箱
              </span>
              <input
                type="text"
                placeholder={focused === 'username' || username ? '' : '用户名或邮箱'}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onFocus={() => setFocused('username')}
                onBlur={() => setFocused(null)}
                onKeyDown={handleKeyDown}
              />
              {!showPassword && (
                <button 
                  className={`login-submit-btn ${username ? 'active' : ''}`}
                  onClick={handleLogin}
                  disabled={loading}
                >
                  <ArrowRightOutlined />
                </button>
              )}
            </div>
            
            {/* 密码输入 */}
            {showPassword && (
              <div className="login-input-item">
                <span className={`login-input-label ${focused === 'password' || password ? '' : 'login-input-placeholder'}`}>
                  密码
                </span>
                <input
                  ref={passwordRef}
                  type="password"
                  placeholder={focused === 'password' || password ? '' : '密码'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  onFocus={() => setFocused('password')}
                  onBlur={() => setFocused(null)}
                  onKeyDown={handleKeyDown}
                />
                <button 
                  className={`login-submit-btn ${password ? 'active' : ''}`}
                  onClick={handleLogin}
                  disabled={loading}
                >
                  <ArrowRightOutlined />
                </button>
              </div>
            )}
          </div>

          {/* 记住登录 */}
          <div className="login-remember">
            <Checkbox defaultChecked>始终保持登录状态</Checkbox>
          </div>

          {/* 底部链接 */}
          <div className="login-links">
            <div className="login-links-row">
              <Link to="/forgot-password">
                忘记密码？<span className="link-arrow">↗</span>
              </Link>
            </div>
            <div className="login-links-row">
              <Link to="/register">创建账户</Link>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default Login;
