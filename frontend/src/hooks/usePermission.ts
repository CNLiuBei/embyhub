// 权限检查Hook
import { useSelector } from 'react-redux';
import type { RootState } from '@/store';

export const usePermission = () => {
  const userInfo = useSelector((state: RootState) => state.auth.userInfo);
  const role = userInfo?.role;

  // 检查是否有特定权限
  const hasPermission = (permissionKey: string): boolean => {
    if (!role || !role.permissions) {
      return false;
    }

    // 超级管理员拥有所有权限
    if (role.role_id === 1) {
      return true;
    }

    // 检查是否在权限列表中
    return role.permissions.some(p => p.permission_key === permissionKey);
  };

  // 检查是否有任一权限
  const hasAnyPermission = (permissionKeys: string[]): boolean => {
    return permissionKeys.some(key => hasPermission(key));
  };

  // 检查是否拥有所有权限
  const hasAllPermissions = (permissionKeys: string[]): boolean => {
    return permissionKeys.every(key => hasPermission(key));
  };

  // 检查是否是超级管理员
  const isSuperAdmin = (): boolean => {
    return role?.role_id === 1;
  };

  // 检查是否是管理员（超级管理员或普通管理员）
  const isAdmin = (): boolean => {
    return role?.role_id === 1 || role?.role_id === 2;
  };

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    isSuperAdmin,
    isAdmin,
    userInfo,
    role,
  };
};
