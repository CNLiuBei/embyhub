// 权限守卫组件
import React from 'react';
import { usePermission } from '@/hooks/usePermission';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';

interface PermissionGuardProps {
  permission?: string;
  permissions?: string[];
  requireAll?: boolean; // 是否需要所有权限
  fallback?: React.ReactNode; // 无权限时显示的内容
  children: React.ReactNode;
}

const PermissionGuard: React.FC<PermissionGuardProps> = ({
  permission,
  permissions,
  requireAll = false,
  fallback,
  children,
}) => {
  const { hasPermission, hasAnyPermission, hasAllPermissions } = usePermission();
  const navigate = useNavigate();

  let hasAccess = false;

  if (permission) {
    hasAccess = hasPermission(permission);
  } else if (permissions && permissions.length > 0) {
    hasAccess = requireAll
      ? hasAllPermissions(permissions)
      : hasAnyPermission(permissions);
  } else {
    // 没有指定权限，默认允许访问
    hasAccess = true;
  }

  if (!hasAccess) {
    if (fallback) {
      return <>{fallback}</>;
    }

    return (
      <Result
        status="403"
        title="403"
        subTitle="抱歉，您没有权限访问此页面"
        extra={
          <Button type="primary" onClick={() => navigate('/')}>
            返回首页
          </Button>
        }
      />
    );
  }

  return <>{children}</>;
};

export default PermissionGuard;
