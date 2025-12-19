import { useMemo } from 'react'

interface SignInFlameProps {
  days: number
  size?: 'small' | 'medium' | 'large'
  showLabel?: boolean
}

// 根据连续签到天数获取火苗配置
const getFlameConfig = (days: number) => {
  if (days <= 0) return { level: 0, name: '未签到', color: '#9ca3af', bgColor: '#f3f4f6', icon: '💤' }
  if (days <= 2) return { level: 1, name: '小火苗', color: '#fb923c', bgColor: '#fff7ed', icon: '🔥' }
  if (days <= 5) return { level: 2, name: '燃烧中', color: '#f97316', bgColor: '#ffedd5', icon: '🔥' }
  if (days <= 10) return { level: 3, name: '火焰', color: '#ea580c', bgColor: '#fed7aa', icon: '🔥' }
  if (days <= 20) return { level: 4, name: '烈焰', color: '#dc2626', bgColor: '#fee2e2', icon: '🔥' }
  if (days <= 30) return { level: 5, name: '炽焰', color: '#b91c1c', bgColor: '#fecaca', icon: '🔥' }
  if (days <= 60) return { level: 6, name: '神焰', color: '#7c2d12', bgColor: '#fef3c7', icon: '🔥' }
  return { level: 7, name: '传说', color: '#6d28d9', bgColor: '#ede9fe', icon: '👑' }
}

const SignInFlame = ({ days, size = 'medium', showLabel = false }: SignInFlameProps) => {
  const config = useMemo(() => getFlameConfig(days), [days])
  
  const sizeMap = {
    small: { box: 32, icon: 16, ring: 2 },
    medium: { box: 48, icon: 24, ring: 3 },
    large: { box: 64, icon: 32, ring: 4 },
  }
  const s = sizeMap[size]

  // Level 0: 未签到 - 灰色静态
  if (config.level === 0) {
    return (
      <div className="flex flex-col items-center gap-1">
        <div 
          className="rounded-full flex items-center justify-center"
          style={{ 
            width: s.box, 
            height: s.box, 
            background: config.bgColor,
            border: `${s.ring}px solid ${config.color}20`
          }}
        >
          <span style={{ fontSize: s.icon }}>{config.icon}</span>
        </div>
        {showLabel && <span className="text-xs text-gray-400">{config.name}</span>}
      </div>
    )
  }

  // Level 1-2: 橙色火苗 - 静态/微光
  if (config.level <= 2) {
    return (
      <div className="flex flex-col items-center gap-1">
        <div 
          className="rounded-full flex items-center justify-center relative"
          style={{ 
            width: s.box, 
            height: s.box, 
            background: `linear-gradient(135deg, ${config.bgColor} 0%, white 100%)`,
            border: `${s.ring}px solid ${config.color}`,
            boxShadow: config.level === 2 ? `0 0 ${s.ring * 3}px ${config.color}40` : 'none'
          }}
        >
          <span style={{ fontSize: s.icon }}>{config.icon}</span>
        </div>
        {showLabel && <span className="text-xs" style={{ color: config.color }}>{config.name}</span>}
      </div>
    )
  }

  // Level 3-4: 红橙火焰 - 脉冲光晕
  if (config.level <= 4) {
    return (
      <div className="flex flex-col items-center gap-1">
        <div className="relative" style={{ width: s.box, height: s.box }}>
          {/* 脉冲光晕 */}
          <div 
            className="absolute inset-0 rounded-full"
            style={{ 
              background: config.color,
              opacity: 0.2,
              animation: 'flamePulse 2s ease-in-out infinite'
            }}
          />
          {/* 主体 */}
          <div 
            className="absolute inset-0 rounded-full flex items-center justify-center"
            style={{ 
              background: `linear-gradient(135deg, ${config.bgColor} 0%, white 100%)`,
              border: `${s.ring}px solid ${config.color}`,
              boxShadow: `0 0 ${s.ring * 4}px ${config.color}60`
            }}
          >
            <span style={{ fontSize: s.icon }}>{config.icon}</span>
          </div>
        </div>
        {showLabel && <span className="text-xs font-medium" style={{ color: config.color }}>{config.name}</span>}
        <style>{`
          @keyframes flamePulse {
            0%, 100% { transform: scale(1); opacity: 0.2; }
            50% { transform: scale(1.15); opacity: 0.1; }
          }
        `}</style>
      </div>
    )
  }

  // Level 5-6: 深红/棕红 - 双层光环 + 火星
  if (config.level <= 6) {
    return (
      <div className="flex flex-col items-center gap-1">
        <div className="relative" style={{ width: s.box + 8, height: s.box + 8 }}>
          {/* 外层光环 */}
          <div 
            className="absolute rounded-full"
            style={{ 
              inset: 0,
              border: `2px solid ${config.color}30`,
              animation: 'flameRing 3s linear infinite'
            }}
          />
          {/* 内层脉冲 */}
          <div 
            className="absolute rounded-full"
            style={{ 
              inset: 4,
              background: config.color,
              opacity: 0.15,
              animation: 'flamePulse2 1.5s ease-in-out infinite'
            }}
          />
          {/* 主体 */}
          <div 
            className="absolute rounded-full flex items-center justify-center"
            style={{ 
              inset: 4,
              background: `linear-gradient(135deg, ${config.bgColor} 0%, white 100%)`,
              border: `${s.ring}px solid ${config.color}`,
              boxShadow: `0 0 ${s.ring * 5}px ${config.color}70, inset 0 0 ${s.ring * 2}px ${config.color}20`
            }}
          >
            <span style={{ fontSize: s.icon }}>{config.icon}</span>
          </div>
          {/* 火星点 */}
          {config.level === 6 && (
            <>
              <div className="absolute w-1 h-1 rounded-full bg-yellow-400" style={{ top: 2, left: '50%', animation: 'sparkFloat 2s ease-in-out infinite' }} />
              <div className="absolute w-1 h-1 rounded-full bg-orange-400" style={{ top: '50%', right: 2, animation: 'sparkFloat 2s ease-in-out infinite 0.5s' }} />
            </>
          )}
        </div>
        {showLabel && <span className="text-xs font-bold" style={{ color: config.color }}>{config.name}</span>}
        <style>{`
          @keyframes flamePulse2 {
            0%, 100% { transform: scale(1); opacity: 0.15; }
            50% { transform: scale(1.1); opacity: 0.08; }
          }
          @keyframes flameRing {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
          }
          @keyframes sparkFloat {
            0%, 100% { opacity: 0.4; transform: translateY(0); }
            50% { opacity: 1; transform: translateY(-3px); }
          }
        `}</style>
      </div>
    )
  }

  // Level 7: 传说 - 紫金色 + 皇冠 + 彩虹边框
  return (
    <div className="flex flex-col items-center gap-1">
      <div className="relative" style={{ width: s.box + 12, height: s.box + 12 }}>
        {/* 彩虹旋转边框 */}
        <div 
          className="absolute rounded-full"
          style={{ 
            inset: 0,
            background: 'conic-gradient(from 0deg, #f97316, #eab308, #22c55e, #3b82f6, #8b5cf6, #ec4899, #f97316)',
            animation: 'rainbowSpin 4s linear infinite',
            opacity: 0.6
          }}
        />
        {/* 内层白色遮罩 */}
        <div 
          className="absolute rounded-full bg-white"
          style={{ inset: 3 }}
        />
        {/* 金色光晕 */}
        <div 
          className="absolute rounded-full"
          style={{ 
            inset: 3,
            background: 'radial-gradient(circle, rgba(251,191,36,0.3) 0%, transparent 70%)',
            animation: 'goldPulse 2s ease-in-out infinite'
          }}
        />
        {/* 主体 */}
        <div 
          className="absolute rounded-full flex items-center justify-center"
          style={{ 
            inset: 6,
            background: `linear-gradient(135deg, #fef3c7 0%, #fde68a 50%, #fbbf24 100%)`,
            border: `${s.ring}px solid ${config.color}`,
            boxShadow: `0 0 ${s.ring * 6}px ${config.color}80, 0 0 ${s.ring * 3}px #fbbf2480`
          }}
        >
          <span style={{ fontSize: s.icon }}>{config.icon}</span>
        </div>
        {/* 星星装饰 */}
        <div className="absolute text-yellow-400" style={{ top: -2, left: '50%', transform: 'translateX(-50%)', fontSize: s.icon * 0.5, animation: 'starTwinkle 1.5s ease-in-out infinite' }}>⭐</div>
        <div className="absolute text-purple-400" style={{ bottom: 0, right: 0, fontSize: s.icon * 0.4, animation: 'starTwinkle 1.5s ease-in-out infinite 0.5s' }}>✨</div>
      </div>
      {showLabel && (
        <span 
          className="text-xs font-bold"
          style={{ 
            background: 'linear-gradient(90deg, #7c2d12, #6d28d9, #7c2d12)',
            backgroundSize: '200% 100%',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            animation: 'textShine 3s linear infinite'
          }}
        >
          {config.name}
        </span>
      )}
      <style>{`
        @keyframes rainbowSpin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        @keyframes goldPulse {
          0%, 100% { opacity: 0.3; transform: scale(1); }
          50% { opacity: 0.5; transform: scale(1.05); }
        }
        @keyframes starTwinkle {
          0%, 100% { opacity: 0.6; transform: scale(1); }
          50% { opacity: 1; transform: scale(1.2); }
        }
        @keyframes textShine {
          0% { background-position: 200% 0; }
          100% { background-position: -200% 0; }
        }
      `}</style>
    </div>
  )
}

export default SignInFlame
export { getFlameConfig }
