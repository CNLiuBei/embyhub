import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { ThemeConfig, themes, ThemeName, defaultTheme } from './index'

interface ThemeContextType {
  theme: ThemeConfig
  themeName: ThemeName
  setTheme: (name: ThemeName) => void
}

const ThemeContext = createContext<ThemeContextType>({
  theme: defaultTheme,
  themeName: 'default',
  setTheme: () => {},
})

// 将主题配置应用到CSS变量
const applyTheme = (theme: ThemeConfig) => {
  const root = document.documentElement
  
  // 背景
  root.style.setProperty('--bg-image', `url(${theme.background.image})`)
  root.style.setProperty('--bg-overlay', theme.background.overlay)
  
  // 颜色
  root.style.setProperty('--color-primary', theme.colors.primary)
  root.style.setProperty('--color-secondary', theme.colors.secondary)
  root.style.setProperty('--color-success', theme.colors.success)
  root.style.setProperty('--color-warning', theme.colors.warning)
  root.style.setProperty('--color-error', theme.colors.error)
  root.style.setProperty('--color-info', theme.colors.info)
  
  // 毛玻璃
  root.style.setProperty('--glass-bg', theme.glass.background)
  root.style.setProperty('--glass-blur', theme.glass.blur)
  root.style.setProperty('--glass-border', theme.glass.border)
  root.style.setProperty('--glass-shadow', theme.glass.shadow)
  
  // 侧边栏
  root.style.setProperty('--sidebar-bg', theme.sidebar.background)
  root.style.setProperty('--sidebar-text', theme.sidebar.textColor)
  root.style.setProperty('--sidebar-active-bg', theme.sidebar.activeBackground)
  root.style.setProperty('--sidebar-active-text', theme.sidebar.activeTextColor)
  
  // 统计图标
  root.style.setProperty('--stat-users-bg', theme.statIcons.users)
  root.style.setProperty('--stat-active-bg', theme.statIcons.active)
  root.style.setProperty('--stat-visits-bg', theme.statIcons.visits)
  root.style.setProperty('--stat-vip-bg', theme.statIcons.vip)
  
  // 圆角
  root.style.setProperty('--radius-sm', theme.borderRadius.sm)
  root.style.setProperty('--radius-md', theme.borderRadius.md)
  root.style.setProperty('--radius-lg', theme.borderRadius.lg)
  root.style.setProperty('--radius-xl', theme.borderRadius.xl)
}

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [themeName, setThemeName] = useState<ThemeName>(() => {
    const saved = localStorage.getItem('theme') as ThemeName
    return saved && themes[saved] ? saved : 'default'
  })
  
  const theme = themes[themeName]
  
  useEffect(() => {
    applyTheme(theme)
    localStorage.setItem('theme', themeName)
  }, [theme, themeName])
  
  const setTheme = (name: ThemeName) => {
    if (themes[name]) {
      setThemeName(name)
    }
  }
  
  return (
    <ThemeContext.Provider value={{ theme, themeName, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export const useTheme = () => useContext(ThemeContext)

export default ThemeContext
