import React from 'react';

/**
 * Emby Hub Logo 图标 - 渐变色 E 字母
 */
const LogoIcon: React.FC<{ size?: number }> = ({ size = 48 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    style={{ filter: 'drop-shadow(0 2px 6px rgba(0, 122, 255, 0.4))' }}
  >
    <defs>
      <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#007AFF" />
        <stop offset="50%" stopColor="#5856D6" />
        <stop offset="100%" stopColor="#AF52DE" />
      </linearGradient>
    </defs>
    {/* 圆形背景 */}
    <circle cx="24" cy="24" r="24" fill="url(#logoGradient)" />
    {/* E 字母 */}
    <path
      d="M16 14H32V18H20V22H30V26H20V30H32V34H16V14Z"
      fill="white"
    />
  </svg>
);

export default LogoIcon;
