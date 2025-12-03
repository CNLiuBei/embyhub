import React from 'react';

/**
 * 彩色圆点组件 - Apple风格彩虹圆点动画
 */
const ColorDots: React.FC = () => {
  const colors = [
    // 外圈 - 大圆点
    { color: '#fb3b2e', size: 8, r: 54, angle: 0 },
    { color: '#fb5c2e', size: 8, r: 54, angle: 20 },
    { color: '#fb8c2e', size: 8, r: 54, angle: 40 },
    { color: '#f5af2e', size: 8, r: 54, angle: 60 },
    { color: '#e8d62e', size: 8, r: 54, angle: 80 },
    { color: '#a8d82e', size: 8, r: 54, angle: 100 },
    { color: '#4cd964', size: 8, r: 54, angle: 120 },
    { color: '#34c8b4', size: 8, r: 54, angle: 140 },
    { color: '#34aadc', size: 8, r: 54, angle: 160 },
    { color: '#5856d6', size: 8, r: 54, angle: 180 },
    { color: '#9b59b6', size: 8, r: 54, angle: 200 },
    { color: '#c74bbe', size: 8, r: 54, angle: 220 },
    { color: '#e74c8f', size: 8, r: 54, angle: 240 },
    { color: '#fb4c6a', size: 8, r: 54, angle: 260 },
    { color: '#fb3b4a', size: 8, r: 54, angle: 280 },
    { color: '#fb3b38', size: 8, r: 54, angle: 300 },
    { color: '#fb3b30', size: 8, r: 54, angle: 320 },
    { color: '#fb3b2e', size: 8, r: 54, angle: 340 },
    // 中圈 - 中圆点
    { color: '#fb6040', size: 6, r: 44, angle: 10 },
    { color: '#fbaa40', size: 6, r: 44, angle: 50 },
    { color: '#c8dc40', size: 6, r: 44, angle: 90 },
    { color: '#40d8a0', size: 6, r: 44, angle: 130 },
    { color: '#4090dc', size: 6, r: 44, angle: 170 },
    { color: '#8050d8', size: 6, r: 44, angle: 210 },
    { color: '#d050a0', size: 6, r: 44, angle: 250 },
    { color: '#fb5060', size: 6, r: 44, angle: 290 },
    { color: '#fb5040', size: 6, r: 44, angle: 330 },
    // 内圈 - 小圆点
    { color: '#fb8060', size: 4, r: 36, angle: 30 },
    { color: '#e0e060', size: 4, r: 36, angle: 70 },
    { color: '#60e090', size: 4, r: 36, angle: 110 },
    { color: '#60b0e0', size: 4, r: 36, angle: 150 },
    { color: '#a060d0', size: 4, r: 36, angle: 190 },
    { color: '#e060a0', size: 4, r: 36, angle: 230 },
    { color: '#fb6070', size: 4, r: 36, angle: 270 },
    { color: '#fb7060', size: 4, r: 36, angle: 310 },
  ];

  return (
    <div className="login-dots">
      {colors.map((dot, i) => {
        const x = 60 + dot.r * Math.cos((dot.angle - 90) * Math.PI / 180);
        const y = 60 + dot.r * Math.sin((dot.angle - 90) * Math.PI / 180);
        return (
          <div
            key={i}
            style={{
              position: 'absolute',
              width: dot.size,
              height: dot.size,
              borderRadius: '50%',
              background: dot.color,
              left: x - dot.size / 2,
              top: y - dot.size / 2,
            }}
          />
        );
      })}
    </div>
  );
};

export default ColorDots;
