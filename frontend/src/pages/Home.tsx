import { Link } from 'react-router-dom'
import { Button, Space, Tag } from 'antd'
import { 
  UserOutlined, 
  CrownOutlined, 
  SafetyOutlined,
  CloudServerOutlined,
  ThunderboltOutlined,
  ApiOutlined,
  GithubOutlined,
  SendOutlined,
  QqOutlined,
  LoginOutlined,
  UserAddOutlined
} from '@ant-design/icons'
import { useSiteSettings } from '../hooks/useSiteSettings'

const Home = () => {
  const { settings: siteSettings } = useSiteSettings()
  
  const brandName = siteSettings.title?.split(' - ')[0] || 'EmbyHub'
  const shortName = brandName.charAt(0)

  const features = [
    { 
      icon: <UserOutlined className="text-4xl" />, 
      title: '用户管理', 
      desc: '完整的用户注册、登录、个人中心功能，支持邮箱验证和密码找回',
      color: 'from-blue-500 to-cyan-500'
    },
    { 
      icon: <CrownOutlined className="text-4xl" />, 
      title: 'VIP会员', 
      desc: '灵活的会员等级和套餐管理，支持卡密充值和在线支付',
      color: 'from-yellow-500 to-orange-500'
    },
    { 
      icon: <SafetyOutlined className="text-4xl" />, 
      title: '安全防护', 
      desc: 'IP黑名单、登录日志、操作审计，多重安全保障',
      color: 'from-green-500 to-emerald-500'
    },
    { 
      icon: <CloudServerOutlined className="text-4xl" />, 
      title: 'Emby集成', 
      desc: '与Emby媒体服务器深度集成，自动同步用户和权限',
      color: 'from-purple-500 to-pink-500'
    },
    { 
      icon: <ThunderboltOutlined className="text-4xl" />, 
      title: '积分系统', 
      desc: '签到、邀请、兑换，丰富的积分玩法提升用户活跃度',
      color: 'from-red-500 to-rose-500'
    },
    { 
      icon: <ApiOutlined className="text-4xl" />, 
      title: '开放接口', 
      desc: '提供完整的API接口，支持第三方系统对接',
      color: 'from-indigo-500 to-violet-500'
    },
  ]

  const techStack = [
    { name: 'Go', color: 'cyan' },
    { name: 'Gin', color: 'blue' },
    { name: 'React 18', color: 'geekblue' },
    { name: 'TypeScript', color: 'blue' },
    { name: 'Ant Design', color: 'purple' },
    { name: 'TailwindCSS', color: 'cyan' },
    { name: 'PostgreSQL', color: 'green' },
    { name: 'Docker', color: 'blue' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      {/* 导航栏 */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-slate-900/80 backdrop-blur-md border-b border-slate-700/50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            {siteSettings.logo ? (
              <img src={siteSettings.logo} alt="Logo" className="w-10 h-10 rounded-lg object-contain" />
            ) : (
              <div className="w-10 h-10 rounded-lg bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center">
                <span className="text-white font-bold text-lg">{shortName}</span>
              </div>
            )}
            <span className="text-white text-xl font-bold">{brandName}</span>
          </div>
          <Space size="middle">
            <Link to="/about">
              <Button type="text" className="!text-slate-300 hover:!text-white">关于我们</Button>
            </Link>
            <Link to="/login">
              <Button type="default" icon={<LoginOutlined />} className="!border-slate-600 !text-slate-300 hover:!border-blue-500 hover:!text-blue-400">
                登录
              </Button>
            </Link>
            <Link to="/register">
              <Button type="primary" icon={<UserAddOutlined />}>
                注册
              </Button>
            </Link>
          </Space>
        </div>
      </nav>

      {/* Hero 区域 */}
      <section className="pt-32 pb-20 px-6">
        <div className="max-w-4xl mx-auto text-center">
          {/* Logo */}
          <div className="mb-8 flex justify-center">
            {siteSettings.logo ? (
              <img src={siteSettings.logo} alt="Logo" className="w-28 h-28 rounded-2xl object-contain shadow-2xl shadow-blue-500/20" />
            ) : (
              <div className="w-28 h-28 rounded-2xl bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 flex items-center justify-center shadow-2xl shadow-purple-500/30">
                <span className="text-white text-5xl font-bold">{shortName}</span>
              </div>
            )}
          </div>
          
          {/* 标题 */}
          <h1 className="text-5xl md:text-6xl font-bold text-white mb-6">
            {brandName}
          </h1>
          <p className="text-xl md:text-2xl text-slate-400 mb-4">
            {siteSettings.description || '专业的 Emby 用户管理系统'}
          </p>
          <p className="text-slate-500 mb-10 max-w-2xl mx-auto">
            功能强大、安全可靠、易于部署的媒体服务用户管理解决方案，支持用户注册、VIP会员、支付系统、积分系统等完整功能
          </p>
          
          {/* CTA 按钮 */}
          <Space size="large" className="flex-wrap justify-center">
            <Link to="/register">
              <Button type="primary" size="large" icon={<UserAddOutlined />} className="!h-12 !px-8 !text-base">
                立即体验
              </Button>
            </Link>
            <Link to="/about">
              <Button size="large" className="!h-12 !px-8 !text-base !bg-slate-800 !border-slate-600 !text-slate-300 hover:!border-blue-500 hover:!text-blue-400">
                了解更多
              </Button>
            </Link>
          </Space>
        </div>
      </section>

      {/* 功能特性 */}
      <section className="py-20 px-6 bg-slate-800/30">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-white text-center mb-4">功能特性</h2>
          <p className="text-slate-400 text-center mb-12">为您的媒体服务提供完整的用户管理解决方案</p>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <div 
                key={index} 
                className="p-6 rounded-2xl bg-slate-800/50 border border-slate-700/50 hover:border-slate-600 transition-all hover:transform hover:-translate-y-1"
              >
                <div className={`w-16 h-16 rounded-xl bg-gradient-to-r ${feature.color} flex items-center justify-center text-white mb-4`}>
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold text-white mb-2">{feature.title}</h3>
                <p className="text-slate-400">{feature.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 技术栈 */}
      <section className="py-20 px-6">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl font-bold text-white mb-4">技术栈</h2>
          <p className="text-slate-400 mb-8">采用现代化技术栈，确保系统稳定高效</p>
          <Space size={[8, 16]} wrap className="justify-center">
            {techStack.map((tech, index) => (
              <Tag key={index} color={tech.color} className="!text-base !px-4 !py-1">
                {tech.name}
              </Tag>
            ))}
          </Space>
        </div>
      </section>

      {/* 底部 */}
      <footer className="py-12 px-6 border-t border-slate-700/50">
        <div className="max-w-6xl mx-auto">
          <div className="flex flex-col md:flex-row justify-between items-center gap-6">
            {/* 品牌 */}
            <div className="flex items-center gap-3">
              {siteSettings.logo ? (
                <img src={siteSettings.logo} alt="Logo" className="w-8 h-8 rounded-lg object-contain" />
              ) : (
                <div className="w-8 h-8 rounded-lg bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center">
                  <span className="text-white font-bold">{shortName}</span>
                </div>
              )}
              <span className="text-slate-300 font-semibold">{brandName}</span>
            </div>
            
            {/* 链接 */}
            <Space size="large" className="text-slate-400">
              {siteSettings.github_url && (
                <a 
                  href={siteSettings.github_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="hover:text-white transition-colors flex items-center gap-1"
                >
                  <GithubOutlined /> GitHub
                </a>
              )}
              {siteSettings.telegram_url && (
                <a 
                  href={siteSettings.telegram_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="hover:text-white transition-colors flex items-center gap-1"
                >
                  <SendOutlined /> Telegram
                </a>
              )}
              {siteSettings.qq_url && (
                <a 
                  href={siteSettings.qq_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="hover:text-white transition-colors flex items-center gap-1"
                >
                  <QqOutlined /> QQ群
                </a>
              )}
            </Space>
          </div>
          
          {/* 版权 */}
          <div className="mt-8 pt-6 border-t border-slate-700/50 text-center text-slate-500 text-sm">
            {siteSettings.footer || `© 2025 ${brandName} · 保留所有权利`}
            {siteSettings.icp && (
              <span className="ml-2">| {siteSettings.icp}</span>
            )}
          </div>
        </div>
      </footer>
    </div>
  )
}

export default Home
