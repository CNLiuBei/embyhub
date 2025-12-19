import { Typography, Divider, Space, Tag } from 'antd'
import { 
  GithubOutlined, 
  CodeOutlined,
  CloudServerOutlined,
  SafetyOutlined,
  ThunderboltOutlined,
  SendOutlined,
  QqOutlined
} from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { useTheme } from '../theme/ThemeContext'
import { useSiteSettings } from '../hooks/useSiteSettings'

const { Title, Paragraph, Text } = Typography

const About = () => {
  const { theme } = useTheme()
  const { settings: siteSettings } = useSiteSettings()
  
  const brandName = siteSettings.title?.split(' - ')[0] || theme.brand.name
  const shortName = brandName.charAt(0)

  const features = [
    { icon: <SafetyOutlined />, title: '安全可靠', desc: '多重安全防护，保障账户安全' },
    { icon: <ThunderboltOutlined />, title: '高效便捷', desc: '简洁的操作界面，流畅的使用体验' },
    { icon: <CloudServerOutlined />, title: '稳定服务', desc: '7x24小时稳定运行，随时随地访问' },
    { icon: <CodeOutlined />, title: '持续更新', desc: '不断优化功能，提升用户体验' },
  ]

  return (
    <div className="login-bg flex items-center justify-center min-h-screen p-4">
      {/* 左上角品牌 */}
      <div className="absolute top-6 left-6 flex items-center gap-3">
        {siteSettings.logo ? (
          <img src={siteSettings.logo} alt="Logo" className="w-10 h-10 rounded-full object-contain shadow-lg" />
        ) : (
          <div className="w-10 h-10 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center shadow-lg">
            <span className="text-white font-bold">{shortName}</span>
          </div>
        )}
        <span className="text-white text-lg font-bold drop-shadow-lg whitespace-nowrap">{brandName}</span>
      </div>

      {/* 主内容 */}
      <div className="glass-card p-8 w-full max-w-2xl shadow-2xl">
        {/* Logo */}
        <div className="flex justify-center mb-6">
          {siteSettings.logo ? (
            <img src={siteSettings.logo} alt="Logo" className="w-24 h-24 rounded-full object-contain shadow-xl" />
          ) : (
            <div className="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 flex items-center justify-center shadow-xl">
              <span className="text-white text-4xl font-bold">{shortName}</span>
            </div>
          )}
        </div>

        <Title level={2} className="text-center !mb-2">{brandName}</Title>
        <Paragraph className="text-center text-gray-500 !mb-6">
          {siteSettings.description || '专业的媒体服务用户管理系统'}
        </Paragraph>

        <Divider />

        {/* 功能特点 */}
        <div className="grid grid-cols-2 gap-4 mb-6">
          {features.map((item, index) => (
            <div key={index} className="flex items-start gap-3 p-3 rounded-xl bg-gray-50 hover:bg-blue-50 transition-colors">
              <div className="text-2xl text-blue-500">{item.icon}</div>
              <div>
                <div className="font-medium text-gray-800">{item.title}</div>
                <div className="text-sm text-gray-500">{item.desc}</div>
              </div>
            </div>
          ))}
        </div>

        <Divider />

        {/* 技术栈 */}
        <div className="mb-6">
          <Title level={5} className="!mb-3">技术栈</Title>
          <Space wrap>
            <Tag color="blue">React 18</Tag>
            <Tag color="cyan">TypeScript</Tag>
            <Tag color="green">Go</Tag>
            <Tag color="orange">Gin</Tag>
            <Tag color="purple">PostgreSQL</Tag>
            <Tag color="magenta">Ant Design</Tag>
            <Tag color="geekblue">TailwindCSS</Tag>
            <Tag color="volcano">Vite</Tag>
          </Space>
        </div>

        {/* 版本信息 */}
        <div className="mb-6">
          <Title level={5} className="!mb-3">版本信息</Title>
          <div className="space-y-1 text-gray-600">
            <div><Text strong>当前版本：</Text>v1.0.0</div>
            <div><Text strong>更新日期：</Text>2025年12月</div>
          </div>
        </div>

        <Divider />

        {/* 底部链接 */}
        <div className="flex justify-center items-center gap-6 text-gray-500 flex-wrap">
          <Link to="/login" className="hover:text-blue-500 transition-colors">
            返回登录
          </Link>
          {siteSettings.github_url && (
            <>
              <span>|</span>
              <a 
                href={siteSettings.github_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-blue-500 transition-colors flex items-center gap-1"
              >
                <GithubOutlined /> GitHub
              </a>
            </>
          )}
          {siteSettings.telegram_url && (
            <>
              <span>|</span>
              <a 
                href={siteSettings.telegram_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-blue-500 transition-colors flex items-center gap-1"
              >
                <SendOutlined /> Telegram
              </a>
            </>
          )}
          {siteSettings.qq_url && (
            <>
              <span>|</span>
              <a 
                href={siteSettings.qq_url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="hover:text-blue-500 transition-colors flex items-center gap-1"
              >
                <QqOutlined /> QQ群
              </a>
            </>
          )}
        </div>
      </div>

      {/* 底部版权 */}
      <div className="absolute bottom-6 text-white/60 text-sm">
        {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
        {siteSettings.icp && (
          <span className="ml-2">| {siteSettings.icp}</span>
        )}
      </div>
    </div>
  )
}

export default About
