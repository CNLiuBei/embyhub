import { useState, useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import { Button, Space } from 'antd'
import { 
  UserOutlined, CrownOutlined, SafetyOutlined, CloudServerOutlined,
  ThunderboltOutlined, ApiOutlined, GithubOutlined, SendOutlined,
  QqOutlined, UserAddOutlined, CheckCircleOutlined,
  RocketOutlined, SettingOutlined, TeamOutlined, GiftOutlined,
  BellOutlined, PlayCircleOutlined, DatabaseOutlined, GlobalOutlined,
  MobileOutlined, LockOutlined, SyncOutlined, StarOutlined,
  FireOutlined, TrophyOutlined, HeartOutlined
} from '@ant-design/icons'
import { useSiteSettings } from '../hooks/useSiteSettings'

// 部署方式切换组件
const DeployTabs = ({ siteSettings }: { siteSettings: any }) => {
  const [activeTab, setActiveTab] = useState<'docker' | 'binary' | 'source'>('docker')

  const tabs = [
    { key: 'docker', label: '🐳 Docker', tag: '推荐' },
    { key: 'binary', label: '📦 二进制', tag: '轻量' },
    { key: 'source', label: '🔧 源码编译', tag: '开发者' },
  ]

  const deployContent = {
    docker: {
      gradient: 'from-blue-500 to-cyan-500',
      resultColor: 'text-cyan-400',
      steps: [
        { comment: '# 克隆项目', cmd: '$ git clone https://github.com/CNLiuBei/embyhub.git' },
        { comment: '# 进入目录', cmd: '$ cd embyhub' },
        { comment: '# 启动服务', cmd: '$ docker compose up -d' },
        { comment: '# 🎉 完成！访问系统', cmd: '→ http://localhost:54681', isResult: true },
      ]
    },
    binary: {
      gradient: 'from-green-500 to-emerald-500',
      resultColor: 'text-emerald-400',
      steps: [
        { comment: '# 下载最新版本 (Linux amd64 示例)', cmd: '$ wget https://github.com/CNLiuBei/embyhub/releases/latest/download/server-linux-amd64' },
        { comment: '# 添加执行权限', cmd: '$ chmod +x server-linux-amd64' },
        { comment: '# 配置数据库连接 (需要 PostgreSQL + Redis)', cmd: '$ export DB_HOST=localhost DB_USER=embyhub DB_PASSWORD=xxx' },
        { comment: '# 启动服务', cmd: '$ ./server-linux-amd64' },
        { comment: '# 🎉 后端启动完成', cmd: '→ http://localhost:8080', isResult: true },
      ]
    },
    source: {
      gradient: 'from-purple-500 to-pink-500',
      resultColor: 'text-purple-400',
      steps: [
        { comment: '# 克隆项目', cmd: '$ git clone https://github.com/CNLiuBei/embyhub.git && cd embyhub' },
        { comment: '# 编译后端 (需要 Go 1.24+)', cmd: '$ cd backend && go build -o server ./cmd/server' },
        { comment: '# 编译前端 (需要 Node.js 18+)', cmd: '$ cd ../frontend && npm install && npm run build' },
        { comment: '# 🎉 编译完成', cmd: '→ 后端: backend/server | 前端: frontend/dist', isResult: true },
      ]
    }
  }

  const current = deployContent[activeTab]

  return (
    <>
      {/* Tab 切换按钮 */}
      <div className="flex justify-center gap-4 mb-8 flex-wrap">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key as any)}
            className={`px-6 py-3 rounded-xl font-medium transition-all ${
              activeTab === tab.key
                ? 'bg-purple-500/20 border border-purple-500/50 text-purple-400'
                : 'bg-white/5 border border-white/10 text-gray-400 hover:bg-white/10 hover:text-white'
            }`}
          >
            {tab.label}
            <span className={`ml-2 text-xs px-2 py-0.5 rounded ${
              activeTab === tab.key ? 'bg-purple-500/20' : 'bg-white/10'
            }`}>
              {tab.tag}
            </span>
          </button>
        ))}
      </div>

      {/* 终端窗口 */}
      <div className="relative group">
        <div className={`absolute -inset-1 bg-gradient-to-r ${current.gradient} rounded-3xl blur-xl opacity-20 group-hover:opacity-40 transition-opacity`} />
        <div className="relative rounded-3xl overflow-hidden bg-[#1a1a2e] border border-white/10">
          <div className="flex items-center gap-2 px-4 py-3 bg-white/5 border-b border-white/5">
            <div className="flex gap-2">
              <div className="w-3 h-3 rounded-full bg-red-500" />
              <div className="w-3 h-3 rounded-full bg-yellow-500" />
              <div className="w-3 h-3 rounded-full bg-green-500" />
            </div>
            <span className="text-gray-500 text-sm ml-4">Terminal — bash</span>
          </div>
          <div className="p-6 font-mono text-sm overflow-x-auto">
            {current.steps.map((step, i) => (
              <div key={i}>
                <div className="text-gray-500 mb-2">{step.comment}</div>
                <div className={`${step.isResult ? current.resultColor : 'text-green-400'} mb-4 break-all`}>{step.cmd}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* 底部按钮 */}
      <div className="text-center mt-10">
        <Space size="large">
          {siteSettings.github_url && (
            <a href={siteSettings.github_url} target="_blank" rel="noopener noreferrer">
              <Button size="large" icon={<GithubOutlined />} className="!h-12 !px-8 !rounded-xl !bg-white/5 !border-white/10 !text-white hover:!bg-white/10">
                GitHub
              </Button>
            </a>
          )}
          <Link to="/register">
            <Button type="primary" size="large" icon={<RocketOutlined />} className="!h-12 !px-8 !rounded-xl !bg-gradient-to-r !from-blue-500 !to-purple-500 !border-0">
              立即体验
            </Button>
          </Link>
        </Space>
      </div>
    </>
  )
}

// 粒子背景组件
const ParticleBackground = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    canvas.width = window.innerWidth
    canvas.height = window.innerHeight

    const particles: { x: number; y: number; vx: number; vy: number; size: number; opacity: number }[] = []
    const particleCount = 80

    for (let i = 0; i < particleCount; i++) {
      particles.push({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        vx: (Math.random() - 0.5) * 0.5,
        vy: (Math.random() - 0.5) * 0.5,
        size: Math.random() * 2 + 1,
        opacity: Math.random() * 0.5 + 0.2
      })
    }

    const animate = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      
      particles.forEach((p, i) => {
        p.x += p.vx
        p.y += p.vy
        if (p.x < 0 || p.x > canvas.width) p.vx *= -1
        if (p.y < 0 || p.y > canvas.height) p.vy *= -1

        ctx.beginPath()
        ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
        ctx.fillStyle = `rgba(99, 102, 241, ${p.opacity})`
        ctx.fill()

        // 连线
        particles.slice(i + 1).forEach(p2 => {
          const dx = p.x - p2.x
          const dy = p.y - p2.y
          const dist = Math.sqrt(dx * dx + dy * dy)
          if (dist < 120) {
            ctx.beginPath()
            ctx.moveTo(p.x, p.y)
            ctx.lineTo(p2.x, p2.y)
            ctx.strokeStyle = `rgba(99, 102, 241, ${0.1 * (1 - dist / 120)})`
            ctx.stroke()
          }
        })
      })
      requestAnimationFrame(animate)
    }
    animate()

    const handleResize = () => {
      canvas.width = window.innerWidth
      canvas.height = window.innerHeight
    }
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  return <canvas ref={canvasRef} className="fixed inset-0 pointer-events-none z-0" />
}

// 打字机效果
const TypeWriter = ({ texts, className }: { texts: string[]; className?: string }) => {
  const [displayText, setDisplayText] = useState('')
  const [textIndex, setTextIndex] = useState(0)
  const [charIndex, setCharIndex] = useState(0)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    const currentText = texts[textIndex]
    const timeout = setTimeout(() => {
      if (!isDeleting) {
        if (charIndex < currentText.length) {
          setDisplayText(currentText.slice(0, charIndex + 1))
          setCharIndex(charIndex + 1)
        } else {
          setTimeout(() => setIsDeleting(true), 2000)
        }
      } else {
        if (charIndex > 0) {
          setDisplayText(currentText.slice(0, charIndex - 1))
          setCharIndex(charIndex - 1)
        } else {
          setIsDeleting(false)
          setTextIndex((textIndex + 1) % texts.length)
        }
      }
    }, isDeleting ? 50 : 100)
    return () => clearTimeout(timeout)
  }, [charIndex, isDeleting, textIndex, texts])

  return <span className={className}>{displayText}<span className="animate-pulse">|</span></span>
}

// 数字滚动动画
const CountUp = ({ end, duration = 2000, suffix = '' }: { end: number; duration?: number; suffix?: string }) => {
  const [count, setCount] = useState(0)
  const countRef = useRef<HTMLSpanElement>(null)
  const [hasAnimated, setHasAnimated] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !hasAnimated) {
          setHasAnimated(true)
          let start = 0
          const step = end / (duration / 16)
          const timer = setInterval(() => {
            start += step
            if (start >= end) {
              setCount(end)
              clearInterval(timer)
            } else {
              setCount(Math.floor(start))
            }
          }, 16)
        }
      },
      { threshold: 0.5 }
    )
    if (countRef.current) observer.observe(countRef.current)
    return () => observer.disconnect()
  }, [end, duration, hasAnimated])

  return <span ref={countRef}>{count}{suffix}</span>
}


const Home = () => {
  const { settings: siteSettings } = useSiteSettings()
  const [scrollY, setScrollY] = useState(0)
  
  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY)
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])
  
  const brandName = siteSettings.title?.split(' - ')[0] || 'EmbyHub'
  const shortName = brandName.charAt(0)

  const coreFeatures = [
    { icon: <UserOutlined />, title: '用户管理', desc: '完整的用户注册、登录、个人中心，支持邮箱验证', gradient: 'from-blue-500 to-cyan-400' },
    { icon: <CrownOutlined />, title: 'VIP会员', desc: '灵活的会员等级和套餐，支持卡密和在线支付', gradient: 'from-yellow-500 to-orange-400' },
    { icon: <SafetyOutlined />, title: '安全防护', desc: 'IP黑名单、登录日志、操作审计，多重保障', gradient: 'from-green-500 to-emerald-400' },
    { icon: <CloudServerOutlined />, title: 'Emby集成', desc: '与Emby深度集成，自动同步用户和权限', gradient: 'from-purple-500 to-pink-400' },
    { icon: <ThunderboltOutlined />, title: '积分系统', desc: '签到、邀请、兑换，丰富玩法提升活跃度', gradient: 'from-red-500 to-rose-400' },
    { icon: <ApiOutlined />, title: '开放接口', desc: '完整API接口，支持第三方系统对接', gradient: 'from-indigo-500 to-violet-400' },
  ]

  const moreFeatures = [
    { icon: <GiftOutlined />, title: '卡密系统' },
    { icon: <BellOutlined />, title: '公告系统' },
    { icon: <TeamOutlined />, title: '邀请奖励' },
    { icon: <PlayCircleOutlined />, title: '媒体浏览' },
    { icon: <DatabaseOutlined />, title: '数据备份' },
    { icon: <GlobalOutlined />, title: 'CF隧道' },
    { icon: <MobileOutlined />, title: '设备管理' },
    { icon: <LockOutlined />, title: '权限控制' },
    { icon: <SyncOutlined />, title: '自动同步' },
    { icon: <SettingOutlined />, title: '灵活配置' },
    { icon: <StarOutlined />, title: '收藏功能' },
    { icon: <FireOutlined />, title: '热门推荐' },
  ]

  const techStack = [
    { name: 'Go', icon: '🔷' },
    { name: 'React', icon: '⚛️' },
    { name: 'TypeScript', icon: '📘' },
    { name: 'PostgreSQL', icon: '🐘' },
    { name: 'Docker', icon: '🐳' },
    { name: 'TailwindCSS', icon: '🎨' },
  ]

  const stats = [
    { value: 30, suffix: '+', label: '核心功能', icon: <RocketOutlined /> },
    { value: 100, suffix: '%', label: '开源免费', icon: <HeartOutlined /> },
    { value: 99, suffix: '%', label: '稳定运行', icon: <TrophyOutlined /> },
    { value: 5, suffix: 'min', label: '快速部署', icon: <ThunderboltOutlined /> },
  ]


  return (
    <div className="min-h-screen bg-[#0a0a1a] text-white overflow-x-hidden">
      <ParticleBackground />
      
      {/* 导航栏 - 毛玻璃效果 */}
      <nav className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrollY > 50 ? 'bg-[#0a0a1a]/90 backdrop-blur-xl shadow-lg shadow-purple-500/5' : ''
      }`}>
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3 group cursor-pointer">
            <div className="relative">
              {siteSettings.logo ? (
                <img src={siteSettings.logo} alt="Logo" className="w-10 h-10 rounded-xl object-contain transition-transform group-hover:scale-110" />
              ) : (
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 flex items-center justify-center transition-transform group-hover:scale-110 group-hover:rotate-6">
                  <span className="text-white font-bold text-lg">{shortName}</span>
                </div>
              )}
              <div className="absolute -inset-1 bg-gradient-to-r from-blue-500 to-purple-500 rounded-xl blur opacity-0 group-hover:opacity-50 transition-opacity" />
            </div>
            <span className="text-xl font-bold bg-gradient-to-r from-white to-gray-300 bg-clip-text text-transparent">{brandName}</span>
          </div>
          
          <div className="hidden md:flex items-center gap-8">
            {['功能', '技术', '部署'].map((item, i) => (
              <a key={i} href={`#${['features', 'tech', 'deploy'][i]}`} 
                className="text-gray-400 hover:text-white transition-colors relative group">
                {item}
                <span className="absolute -bottom-1 left-0 w-0 h-0.5 bg-gradient-to-r from-blue-500 to-purple-500 group-hover:w-full transition-all" />
              </a>
            ))}
          </div>

          <Space size="middle">
            <Link to="/login">
              <Button className="!bg-transparent !border-gray-600 !text-gray-300 hover:!border-purple-500 hover:!text-purple-400 !rounded-xl">
                登录
              </Button>
            </Link>
            <Link to="/register">
              <Button type="primary" className="!bg-gradient-to-r !from-blue-500 !to-purple-500 !border-0 !rounded-xl hover:!shadow-lg hover:!shadow-purple-500/25 transition-shadow">
                免费注册
              </Button>
            </Link>
          </Space>
        </div>
      </nav>


      {/* Hero 区域 */}
      <section className="relative min-h-screen flex items-center justify-center pt-20 px-6">
        {/* 动态光晕背景 */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/20 rounded-full blur-[100px] animate-pulse" />
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-[100px] animate-pulse" style={{ animationDelay: '1s' }} />
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-full blur-[120px]" />
        </div>

        <div className="relative z-10 text-center max-w-5xl mx-auto">
          {/* 徽章 */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-purple-500/20 mb-8 animate-bounce">
            <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
            <span className="text-sm text-gray-300">🚀 v1.0.0 正式发布</span>
          </div>

          {/* Logo 3D效果 */}
          <div className="mb-8 flex justify-center perspective-1000">
            <div className="relative group" style={{ transform: `rotateY(${scrollY * 0.02}deg)` }}>
              {siteSettings.logo ? (
                <img src={siteSettings.logo} alt="Logo" className="w-36 h-36 rounded-3xl object-contain shadow-2xl shadow-purple-500/30" />
              ) : (
                <div className="w-36 h-36 rounded-3xl bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 flex items-center justify-center shadow-2xl shadow-purple-500/30 transition-transform group-hover:scale-105">
                  <span className="text-white text-7xl font-bold drop-shadow-lg">{shortName}</span>
                </div>
              )}
              <div className="absolute -inset-4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-3xl blur-2xl opacity-30 group-hover:opacity-50 transition-opacity -z-10" />
            </div>
          </div>

          {/* 主标题 - 渐变动画 */}
          <h1 className="text-6xl md:text-8xl font-black mb-6">
            <span className="bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent animate-gradient bg-[length:200%_auto]">
              {brandName}
            </span>
          </h1>

          {/* 打字机效果副标题 */}
          <div className="text-xl md:text-2xl text-gray-400 mb-4 h-8">
            <TypeWriter 
              texts={['专业的 Emby 用户管理系统', '一键 Docker 部署', '完全开源免费', '功能强大易扩展']} 
              className="text-gray-300"
            />
          </div>

          <p className="text-gray-500 mb-10 max-w-2xl mx-auto leading-relaxed">
            集用户管理、VIP会员、支付系统、积分系统、邀请奖励于一体的完整解决方案
          </p>

          {/* CTA 按钮 */}
          <div className="flex flex-wrap justify-center gap-4 mb-16">
            <Link to="/register">
              <Button size="large" className="!h-14 !px-10 !text-lg !rounded-2xl !bg-gradient-to-r !from-blue-500 !to-purple-500 !border-0 !text-white hover:!shadow-xl hover:!shadow-purple-500/30 hover:!scale-105 transition-all">
                <UserAddOutlined className="mr-2" /> 立即体验
              </Button>
            </Link>
            {siteSettings.github_url && (
              <a href={siteSettings.github_url} target="_blank" rel="noopener noreferrer">
                <Button size="large" className="!h-14 !px-10 !text-lg !rounded-2xl !bg-white/5 !border-white/10 !text-white hover:!bg-white/10 hover:!border-white/20 hover:!scale-105 transition-all">
                  <GithubOutlined className="mr-2" /> 查看源码
                </Button>
              </a>
            )}
          </div>

          {/* 统计数据 - 数字滚动 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {stats.map((stat, i) => (
              <div key={i} className="relative group">
                <div className="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl blur-xl opacity-0 group-hover:opacity-100 transition-opacity" />
                <div className="relative p-6 rounded-2xl bg-white/5 border border-white/10 hover:border-purple-500/30 transition-colors">
                  <div className="text-3xl text-purple-400 mb-2">{stat.icon}</div>
                  <div className="text-4xl font-bold bg-gradient-to-r from-white to-gray-300 bg-clip-text text-transparent">
                    <CountUp end={stat.value} suffix={stat.suffix} />
                  </div>
                  <div className="text-gray-500 text-sm mt-1">{stat.label}</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 向下滚动提示 */}
        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 animate-bounce">
          <div className="w-6 h-10 rounded-full border-2 border-gray-600 flex justify-center pt-2">
            <div className="w-1 h-2 bg-gray-400 rounded-full animate-pulse" />
          </div>
        </div>
      </section>


      {/* 核心功能 - 3D卡片 */}
      <section id="features" className="py-32 px-6 relative">
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-purple-500/5 to-transparent" />
        <div className="max-w-6xl mx-auto relative z-10">
          <div className="text-center mb-16">
            <span className="text-purple-400 text-sm font-medium tracking-wider uppercase">Features</span>
            <h2 className="text-4xl md:text-5xl font-bold mt-4 mb-6">
              <span className="bg-gradient-to-r from-white to-gray-400 bg-clip-text text-transparent">核心功能</span>
            </h2>
            <p className="text-gray-500 max-w-2xl mx-auto">为您的媒体服务提供完整的用户管理解决方案</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {coreFeatures.map((feature, i) => (
              <div key={i} className="group perspective-1000">
                <div className="relative p-8 rounded-3xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 hover:border-purple-500/30 transition-all duration-500 hover:transform hover:-translate-y-2 hover:shadow-2xl hover:shadow-purple-500/10">
                  <div className={`w-16 h-16 rounded-2xl bg-gradient-to-r ${feature.gradient} flex items-center justify-center text-white text-3xl mb-6 group-hover:scale-110 group-hover:rotate-3 transition-transform shadow-lg`}>
                    {feature.icon}
                  </div>
                  <h3 className="text-xl font-bold text-white mb-3">{feature.title}</h3>
                  <p className="text-gray-400 leading-relaxed">{feature.desc}</p>
                  <div className={`absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r ${feature.gradient} rounded-b-3xl opacity-0 group-hover:opacity-100 transition-opacity`} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 更多功能 - 悬浮网格 */}
      <section className="py-24 px-6">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-white mb-4">更多功能</h2>
            <p className="text-gray-500">丰富的功能模块，满足各种场景</p>
          </div>
          
          <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {moreFeatures.map((feature, i) => (
              <div key={i} className="group p-4 rounded-2xl bg-white/5 border border-white/5 hover:border-purple-500/30 hover:bg-purple-500/10 transition-all text-center cursor-pointer hover:scale-105">
                <div className="text-2xl text-gray-400 group-hover:text-purple-400 transition-colors mb-2 group-hover:scale-125 transform">
                  {feature.icon}
                </div>
                <div className="text-sm text-gray-400 group-hover:text-white transition-colors">{feature.title}</div>
              </div>
            ))}
          </div>
        </div>
      </section>


      {/* 技术栈 - 浮动图标 */}
      <section id="tech" className="py-24 px-6 relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-r from-blue-500/5 via-purple-500/5 to-pink-500/5" />
        <div className="max-w-4xl mx-auto relative z-10">
          <div className="text-center mb-12">
            <span className="text-purple-400 text-sm font-medium tracking-wider uppercase">Tech Stack</span>
            <h2 className="text-4xl font-bold mt-4 mb-4 text-white">技术栈</h2>
            <p className="text-gray-500">采用现代化技术栈，确保系统稳定高效</p>
          </div>

          <div className="flex flex-wrap justify-center gap-6">
            {techStack.map((tech, i) => (
              <div key={i} className="group relative" style={{ animationDelay: `${i * 0.1}s` }}>
                <div className="absolute inset-0 bg-gradient-to-r from-blue-500 to-purple-500 rounded-2xl blur-xl opacity-0 group-hover:opacity-30 transition-opacity" />
                <div className="relative px-8 py-4 rounded-2xl bg-white/5 border border-white/10 hover:border-purple-500/50 transition-all hover:scale-110 cursor-pointer">
                  <span className="text-2xl mr-2">{tech.icon}</span>
                  <span className="text-white font-medium">{tech.name}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 快速部署 - 多种方式 */}
      <section id="deploy" className="py-32 px-6">
        <div className="max-w-5xl mx-auto">
          <div className="text-center mb-12">
            <span className="text-purple-400 text-sm font-medium tracking-wider uppercase">Quick Start</span>
            <h2 className="text-4xl font-bold mt-4 mb-4 text-white">快速部署</h2>
            <p className="text-gray-500">多种部署方式，灵活选择</p>
          </div>

          {/* 部署方式切换 */}
          <DeployTabs siteSettings={siteSettings} />
        </div>
      </section>


      {/* 功能清单 */}
      <section className="py-24 px-6 relative">
        <div className="absolute inset-0 bg-gradient-to-b from-purple-500/5 to-transparent" />
        <div className="max-w-5xl mx-auto relative z-10">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-white mb-4">完整功能清单</h2>
            <p className="text-gray-500">开箱即用，持续更新</p>
          </div>
          
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {[
              '用户注册/登录', '邮箱验证', '密码找回', 'VIP会员系统',
              '卡密充值', '支付宝支付', '积分签到', '积分兑换',
              '邀请奖励', '邀请排行', '系统公告', '消息通知',
              'Emby同步', '设备管理', '会话限制', '客户端白名单',
              'IP黑名单', '操作日志', '数据备份', 'CF隧道',
              '闲鱼对接', 'API接口', '论坛社区', '私信功能',
            ].map((item, i) => (
              <div key={i} className="flex items-center gap-2 p-3 rounded-xl bg-white/5 border border-white/5 hover:border-green-500/30 hover:bg-green-500/5 transition-all group">
                <CheckCircleOutlined className="text-green-400 group-hover:scale-125 transition-transform" />
                <span className="text-gray-300 text-sm">{item}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-32 px-6 relative overflow-hidden">
        <div className="absolute inset-0">
          <div className="absolute top-0 left-1/4 w-96 h-96 bg-blue-500/20 rounded-full blur-[120px]" />
          <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-[120px]" />
        </div>
        <div className="max-w-3xl mx-auto text-center relative z-10">
          <h2 className="text-4xl md:text-5xl font-bold mb-6">
            <span className="bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
              准备好开始了吗？
            </span>
          </h2>
          <p className="text-gray-400 text-lg mb-10">
            立即注册，体验强大的用户管理系统
          </p>
          <Link to="/register">
            <Button size="large" className="!h-16 !px-12 !text-xl !rounded-2xl !bg-gradient-to-r !from-blue-500 !via-purple-500 !to-pink-500 !border-0 !text-white hover:!shadow-2xl hover:!shadow-purple-500/30 hover:!scale-105 transition-all">
              <UserAddOutlined className="mr-2" /> 免费注册
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 px-6 border-t border-white/5">
        <div className="max-w-6xl mx-auto">
          <div className="flex flex-col md:flex-row justify-between items-center gap-6">
            <div className="flex items-center gap-3">
              {siteSettings.logo ? (
                <img src={siteSettings.logo} alt="Logo" className="w-8 h-8 rounded-lg object-contain" />
              ) : (
                <div className="w-8 h-8 rounded-lg bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center">
                  <span className="text-white font-bold">{shortName}</span>
                </div>
              )}
              <span className="text-gray-300 font-semibold">{brandName}</span>
            </div>
            
            <Space size="large" className="text-gray-500">
              {siteSettings.github_url && (
                <a href={siteSettings.github_url} target="_blank" rel="noopener noreferrer" className="hover:text-purple-400 transition-colors flex items-center gap-1">
                  <GithubOutlined /> GitHub
                </a>
              )}
              {siteSettings.telegram_url && (
                <a href={siteSettings.telegram_url} target="_blank" rel="noopener noreferrer" className="hover:text-purple-400 transition-colors flex items-center gap-1">
                  <SendOutlined /> Telegram
                </a>
              )}
              {siteSettings.qq_url && (
                <a href={siteSettings.qq_url} target="_blank" rel="noopener noreferrer" className="hover:text-purple-400 transition-colors flex items-center gap-1">
                  <QqOutlined /> QQ群
                </a>
              )}
            </Space>
          </div>
          
          <div className="mt-8 pt-6 border-t border-white/5 text-center text-gray-600 text-sm">
            {siteSettings.footer || `© 2025 ${brandName}`}
            {siteSettings.icp && <span className="ml-2">| {siteSettings.icp}</span>}
          </div>
        </div>
      </footer>

      {/* 全局样式 */}
      <style>{`
        @keyframes gradient {
          0%, 100% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
        }
        .animate-gradient {
          animation: gradient 3s ease infinite;
        }
        .perspective-1000 {
          perspective: 1000px;
        }
      `}</style>
    </div>
  )
}

export default Home
