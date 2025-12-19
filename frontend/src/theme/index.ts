// 主题配置文件 - 修改这里即可改变全站风格

export interface ThemeConfig {
  // 品牌信息
  brand: {
    name: string
    shortName: string
    logo?: string
  }
  
  // 背景配置
  background: {
    image: string
    overlay: string  // 背景遮罩颜色
  }
  
  // 颜色配置
  colors: {
    primary: string
    secondary: string
    success: string
    warning: string
    error: string
    info: string
  }
  
  // 毛玻璃效果
  glass: {
    background: string
    blur: string
    border: string
    shadow: string
  }
  
  // 侧边栏
  sidebar: {
    background: string
    textColor: string
    activeBackground: string
    activeTextColor: string
  }
  
  // 统计卡片图标背景色
  statIcons: {
    users: string
    active: string
    visits: string
    vip: string
  }
  
  // 圆角
  borderRadius: {
    sm: string
    md: string
    lg: string
    xl: string
  }
}

// 默认主题 - Emby主题
export const defaultTheme: ThemeConfig = {
  brand: {
    name: 'Emby用户管理',
    shortName: 'E',
  },
  
  background: {
    image: '/api/wallpaper',
    overlay: 'none',
  },
  
  colors: {
    primary: '#1890ff',
    secondary: '#722ed1',
    success: '#52c41a',
    warning: '#faad14',
    error: '#ff4d4f',
    info: '#1890ff',
  },
  
  glass: {
    background: 'rgba(255, 255, 255, 0.92)',
    blur: '24px',
    border: '1px solid rgba(255, 255, 255, 0.5)',
    shadow: '0 8px 32px rgba(0, 0, 0, 0.08)',
  },
  
  sidebar: {
    background: 'linear-gradient(180deg, #1a1a2e 0%, #16213e 100%)',
    textColor: '#ffffff',
    activeBackground: 'linear-gradient(90deg, #4facfe 0%, #00f2fe 100%)',
    activeTextColor: '#ffffff',
  },
  
  statIcons: {
    users: '#e6f7ff',
    active: '#f6ffed',
    visits: '#fff7e6',
    vip: '#fffbe6',
  },
  
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
  },
}

// 暗色主题
export const darkTheme: ThemeConfig = {
  brand: {
    name: 'Emby用户管理',
    shortName: 'E',
  },
  
  background: {
    image: '/api/wallpaper',
    overlay: 'none',
  },
  
  colors: {
    primary: '#177ddc',
    secondary: '#722ed1',
    success: '#49aa19',
    warning: '#d89614',
    error: '#d32029',
    info: '#177ddc',
  },
  
  glass: {
    background: 'rgba(30, 30, 50, 0.85)',
    blur: '20px',
    border: '1px solid rgba(255, 255, 255, 0.1)',
    shadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
  },
  
  sidebar: {
    background: 'linear-gradient(180deg, #0d0d15 0%, #1a1a2e 100%)',
    textColor: '#ffffff',
    activeBackground: 'linear-gradient(90deg, #177ddc 0%, #1890ff 100%)',
    activeTextColor: '#ffffff',
  },
  
  statIcons: {
    users: 'rgba(23, 125, 220, 0.2)',
    active: 'rgba(73, 170, 25, 0.2)',
    visits: 'rgba(216, 150, 20, 0.2)',
    vip: 'rgba(250, 173, 20, 0.2)',
  },
  
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
  },
}

// 粉色主题
export const pinkTheme: ThemeConfig = {
  brand: {
    name: 'Emby用户管理',
    shortName: 'E',
  },
  
  background: {
    image: '/api/wallpaper',
    overlay: 'none',
  },
  
  colors: {
    primary: '#eb2f96',
    secondary: '#722ed1',
    success: '#52c41a',
    warning: '#faad14',
    error: '#ff4d4f',
    info: '#eb2f96',
  },
  
  glass: {
    background: 'rgba(255, 255, 255, 0.9)',
    blur: '20px',
    border: '1px solid rgba(255, 255, 255, 0.4)',
    shadow: '0 8px 32px rgba(235, 47, 150, 0.1)',
  },
  
  sidebar: {
    background: 'linear-gradient(180deg, #2d1f3d 0%, #1a1a2e 100%)',
    textColor: '#ffffff',
    activeBackground: 'linear-gradient(90deg, #eb2f96 0%, #f759ab 100%)',
    activeTextColor: '#ffffff',
  },
  
  statIcons: {
    users: '#fff0f6',
    active: '#f6ffed',
    visits: '#fff7e6',
    vip: '#fff0f6',
  },
  
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
  },
}

// 所有可用主题
export const themes = {
  default: defaultTheme,
  dark: darkTheme,
  pink: pinkTheme,
}

export type ThemeName = keyof typeof themes
