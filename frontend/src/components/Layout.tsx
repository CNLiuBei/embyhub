import React, { useState, useMemo, useEffect, useRef } from 'react';
import { Layout as AntLayout, Menu, Breadcrumb, Dropdown, Avatar, Space } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '@/store';
import { clearAuthInfo } from '@/store/slices/authSlice';
import { logout } from '@/api/auth';
import { usePageView } from '@/hooks/usePageView';
import { usePermission } from '@/hooks/usePermission';
import {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  HistoryOutlined,
  SettingOutlined,
  CloudSyncOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  KeyOutlined,
  CrownOutlined,
  PlayCircleFilled,
  VideoCameraOutlined,
} from '@ant-design/icons';
import VipUpgrade from './VipUpgrade';
import './Layout.css';

const { Header, Sider, Content } = AntLayout;

// 液态玻璃工具函数
function smoothStep(a: number, b: number, t: number): number {
  t = Math.max(0, Math.min(1, (t - a) / (b - a)));
  return t * t * (3 - 2 * t);
}

function lengthFn(x: number, y: number): number {
  return Math.sqrt(x * x + y * y);
}

function roundedRectSDF(x: number, y: number, width: number, height: number, radius: number): number {
  const qx = Math.abs(x) - width + radius;
  const qy = Math.abs(y) - height + radius;
  return Math.min(Math.max(qx, qy), 0) + lengthFn(Math.max(qx, 0), Math.max(qy, 0)) - radius;
}

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useDispatch();
  const [collapsed, setCollapsed] = useState(false);
  const [vipModalVisible, setVipModalVisible] = useState(false);
  const userInfo = useSelector((state: RootState) => state.auth.userInfo);
  const { hasPermission, isSuperAdmin } = usePermission();
  
  const siderCanvasRef = useRef<HTMLCanvasElement>(null);
  const siderFeImageRef = useRef<SVGFEImageElement>(null);
  const siderFeDisplacementRef = useRef<SVGFEDisplacementMapElement>(null);

  // 侧边栏液态玻璃效果
  useEffect(() => {
    const canvas = siderCanvasRef.current;
    const feImage = siderFeImageRef.current;
    const feDisplacement = siderFeDisplacementRef.current;
    
    if (!canvas || !feImage || !feDisplacement) return;

    const width = 200;
    const height = 800;
    const canvasDPI = 1;

    const updateShader = () => {
      const context = canvas.getContext('2d');
      if (!context) return;

      const w = width * canvasDPI;
      const h = height * canvasDPI;
      const data = new Uint8ClampedArray(w * h * 4);

      let maxScale = 0;
      const rawValues: number[] = [];

      for (let i = 0; i < data.length; i += 4) {
        const x = (i / 4) % w;
        const y = Math.floor(i / 4 / w);
        
        const uvX = x / w;
        const uvY = y / h;
        const ix = uvX - 0.5;
        const iy = uvY - 0.5;
        
        // 垂直方向的液态效果
        const distanceToEdge = roundedRectSDF(ix, iy, 0.4, 0.45, 0.1);
        const displacement = smoothStep(0.5, 0, distanceToEdge - 0.05);
        const scaled = smoothStep(0, 1, displacement);
        
        const posX = ix * scaled + 0.5;
        const posY = iy * scaled + 0.5;
        
        const dx = posX * w - x;
        const dy = posY * h - y;
        maxScale = Math.max(maxScale, Math.abs(dx), Math.abs(dy));
        rawValues.push(dx, dy);
      }

      maxScale *= 0.5;
      if (maxScale === 0) maxScale = 1;

      let index = 0;
      for (let i = 0; i < data.length; i += 4) {
        const r = rawValues[index++] / maxScale + 0.5;
        const g = rawValues[index++] / maxScale + 0.5;
        data[i] = r * 255;
        data[i + 1] = g * 255;
        data[i + 2] = 0;
        data[i + 3] = 255;
      }

      context.putImageData(new ImageData(data, w, h), 0, 0);
      feImage.setAttributeNS('http://www.w3.org/1999/xlink', 'href', canvas.toDataURL());
      feDisplacement.setAttribute('scale', (maxScale / canvasDPI).toString());
    };

    updateShader();
  }, []);
  
  // 自动记录页面访问
  usePageView();

  // 菜单项配置（根据权限动态生成）
  const menuItems = useMemo(() => {
    // 管理员菜单
    const adminMenuItems = [
      { key: '/', icon: <DashboardOutlined />, label: '仪表盘', permission: 'stats:view' },
      { key: '/media', icon: <VideoCameraOutlined />, label: '媒体库', permission: 'emby:view' },
      { key: '/users', icon: <UserOutlined />, label: '用户管理', permission: 'user:view' },
      { key: '/card-keys', icon: <KeyOutlined />, label: '卡密管理', permission: 'cardkey:view' },
      { 
        key: 'rbac', 
        icon: <SafetyOutlined />, 
        label: '权限配置',
        children: [
          { key: '/roles', icon: <TeamOutlined />, label: '角色管理', permission: 'role:view' },
          { key: '/permissions', icon: <SafetyOutlined />, label: '权限管理', permission: 'permission:view' },
        ]
      },
      { key: '/emby', icon: <CloudSyncOutlined />, label: 'Emby同步', permission: 'emby:view' },
      { key: '/access-records', icon: <HistoryOutlined />, label: '访问记录', permission: 'stats:view' },
      { key: '/system', icon: <SettingOutlined />, label: '系统设置', permission: 'system:view' },
    ];

    // 普通用户菜单
    const userMenuItems = [
      { key: '/', icon: <UserOutlined />, label: '个人中心' },
      { key: '/media', icon: <VideoCameraOutlined />, label: '媒体库' },
    ];

    // 超级管理员显示所有菜单
    if (isSuperAdmin()) {
      return adminMenuItems;
    }

    // 管理员根据权限过滤菜单
    if (userInfo?.role_id === 2) {
      return adminMenuItems.filter(item => !item.permission || hasPermission(item.permission));
    }

    // 普通用户只显示个人中心
    return userMenuItems;
  }, [hasPermission, isSuperAdmin, userInfo?.role_id]);

  // 面包屑映射
  const breadcrumbMap: Record<string, string> = {
    '/': '首页',
    '/media': '媒体库',
    '/users': '用户管理',
    '/roles': '角色管理',
    '/permissions': '权限管理',
    '/emby': 'Emby同步',
    '/card-keys': '卡密管理',
    '/access-records': '访问记录',
    '/system': '系统设置',
  };

  // 处理登出
  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('登出错误:', error);
    } finally {
      dispatch(clearAuthInfo());
      navigate('/login');
    }
  };

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'vip',
      icon: <CrownOutlined style={{ color: '#faad14' }} />,
      label: 'VIP升级',
      onClick: () => setVipModalVisible(true),
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  // 菜单点击处理
  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  return (
    <AntLayout className="layout-container">
      {/* 侧边栏液态玻璃 SVG 滤镜 */}
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="0"
        height="0"
        style={{ position: 'fixed', top: 0, left: 0, pointerEvents: 'none' }}
      >
        <defs>
          <filter
            id="sider-liquid-glass"
            filterUnits="userSpaceOnUse"
            colorInterpolationFilters="sRGB"
            x="0"
            y="0"
            width="200"
            height="800"
          >
            <feImage
              ref={siderFeImageRef}
              id="sider-liquid-map"
              width="200"
              height="800"
            />
            <feDisplacementMap
              ref={siderFeDisplacementRef}
              in="SourceGraphic"
              in2="sider-liquid-map"
              xChannelSelector="R"
              yChannelSelector="G"
            />
          </filter>
        </defs>
      </svg>
      <canvas
        ref={siderCanvasRef}
        width={200}
        height={800}
        style={{ display: 'none' }}
      />
      
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        className="layout-sider"
        style={{
          backdropFilter: 'url(#sider-liquid-glass) blur(20px) saturate(180%) brightness(1.1)',
          WebkitBackdropFilter: 'url(#sider-liquid-glass) blur(20px) saturate(180%) brightness(1.1)',
        }}
      >
        <div className="logo">
          <PlayCircleFilled className="logo-icon" />
          {!collapsed && <h2>Emby Hub</h2>}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      
      <AntLayout>
        <Header className="layout-header">
          <div className="header-left">
            {React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
              className: 'trigger',
              onClick: () => setCollapsed(!collapsed),
            })}
            <Breadcrumb 
              className="breadcrumb"
              items={[
                { title: '首页' },
                ...(location.pathname !== '/' ? [{ title: breadcrumbMap[location.pathname] }] : [])
              ]}
            />
          </div>
          
          <div className="header-right">
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Space className="user-info">
                <Avatar icon={<UserOutlined />} />
                <span>{userInfo?.username || '管理员'}</span>
              </Space>
            </Dropdown>
          </div>
        </Header>
        
        <Content className="layout-content">
          <div className="content-wrapper">
            <Outlet />
          </div>
        </Content>
      </AntLayout>

      {/* VIP升级弹窗 */}
      <VipUpgrade 
        visible={vipModalVisible} 
        onClose={() => setVipModalVisible(false)} 
      />
    </AntLayout>
  );
};

export default Layout;
