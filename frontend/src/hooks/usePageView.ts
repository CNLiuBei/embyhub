// 页面访问记录Hook
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useSelector } from 'react-redux';
import type { RootState } from '@/store';
import { createAccessRecord } from '@/api/accessRecord';

// 页面路径映射
const pageNames: Record<string, string> = {
  '/': '仪表盘',
  '/users': '用户管理',
  '/roles': '角色管理',
  '/permissions': '权限管理',
  '/access-records': '访问记录',
  '/system': '系统配置',
  '/emby': 'Emby同步',
};

export const usePageView = () => {
  const location = useLocation();
  const userInfo = useSelector((state: RootState) => state.auth.userInfo);

  useEffect(() => {
    const recordPageView = async () => {
      // 如果用户未登录，不记录
      if (!userInfo?.user_id) {
        return;
      }

      const pageName = pageNames[location.pathname] || '未知页面';
      
      try {
        await createAccessRecord({
          user_id: userInfo.user_id,
          resource: pageName,
          ip_address: '127.0.0.1', // 实际应用中应从客户端获取
          device_info: navigator.userAgent.substring(0, 100),
        });
      } catch (error) {
        // 静默失败，不影响用户体验
        console.debug('记录页面访问失败:', error);
      }
    };

    recordPageView();
  }, [location.pathname, userInfo]);
};
