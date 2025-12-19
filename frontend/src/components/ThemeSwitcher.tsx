import { Dropdown, Button } from 'antd'
import { BgColorsOutlined } from '@ant-design/icons'
import { useTheme } from '../theme/ThemeContext'
import { ThemeName } from '../theme'

const themeOptions = [
  { key: 'default', label: '🏔️ 山水风景', description: '默认主题' },
  { key: 'dark', label: '🌌 星空夜景', description: '暗色主题' },
  { key: 'pink', label: '🌸 粉色樱花', description: '粉色主题' },
]

const ThemeSwitcher = () => {
  const { themeName, setTheme } = useTheme()

  const menuItems = themeOptions.map((item) => ({
    key: item.key,
    label: (
      <div className="flex items-center gap-2 py-1">
        <span>{item.label}</span>
        {themeName === item.key && <span className="text-green-500">✓</span>}
      </div>
    ),
    onClick: () => setTheme(item.key as ThemeName),
  }))

  return (
    <Dropdown menu={{ items: menuItems }} placement="bottomRight" trigger={['click']}>
      <Button type="text" icon={<BgColorsOutlined />} className="text-gray-600">
        主题
      </Button>
    </Dropdown>
  )
}

export default ThemeSwitcher
