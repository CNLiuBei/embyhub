import { Avatar } from 'antd'
import { CSSProperties, useMemo } from 'react'

// Telegram 风格渐变色
const TG_GRADIENTS = [
  'linear-gradient(135deg, #FF6B6B 0%, #EE5A5A 100%)', // 红色
  'linear-gradient(135deg, #FFA726 0%, #FB8C00 100%)', // 橙色
  'linear-gradient(135deg, #AB47BC 0%, #8E24AA 100%)', // 紫色
  'linear-gradient(135deg, #42A5F5 0%, #1E88E5 100%)', // 蓝色
  'linear-gradient(135deg, #66BB6A 0%, #43A047 100%)', // 绿色
  'linear-gradient(135deg, #26C6DA 0%, #00ACC1 100%)', // 青色
  'linear-gradient(135deg, #EC407A 0%, #D81B60 100%)', // 粉色
]

interface UserAvatarProps {
  src?: string | null
  name?: string | null
  size?: number | 'small' | 'default' | 'large'
  className?: string
  style?: CSSProperties
  onClick?: () => void
}

/**
 * 统一的用户头像组件
 * - 有头像时显示头像图片
 * - 无头像时显示 Telegram 风格的渐变色 + 首字母
 */
const UserAvatar = ({ src, name, size = 'default', className = '', style, onClick }: UserAvatarProps) => {
  // 计算尺寸
  const sizeValue = useMemo(() => {
    if (typeof size === 'number') return size
    switch (size) {
      case 'small': return 24
      case 'large': return 48
      default: return 32
    }
  }, [size])

  // 计算字体大小
  const fontSize = useMemo(() => {
    if (sizeValue <= 24) return 12
    if (sizeValue <= 32) return 14
    if (sizeValue <= 48) return 18
    return Math.floor(sizeValue * 0.4)
  }, [sizeValue])

  // 获取首字母
  const initial = useMemo(() => {
    if (!name) return 'U'
    // 处理中文名字 - 取第一个字
    const firstChar = name.charAt(0)
    // 如果是英文，转大写
    if (/[a-zA-Z]/.test(firstChar)) {
      return firstChar.toUpperCase()
    }
    return firstChar
  }, [name])

  // 根据名字生成渐变色
  const gradient = useMemo(() => {
    if (!name) return TG_GRADIENTS[0]
    // 使用名字的所有字符计算哈希，确保同一名字总是同一颜色
    let hash = 0
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    return TG_GRADIENTS[Math.abs(hash) % TG_GRADIENTS.length]
  }, [name])

  // 如果有有效的头像URL
  if (src) {
    return (
      <Avatar
        src={src}
        size={sizeValue}
        className={className}
        style={style}
        onClick={onClick}
      />
    )
  }

  // 无头像时显示渐变色 + 首字母
  return (
    <div
      className={`flex items-center justify-center rounded-full text-white font-medium select-none ${className}`}
      style={{
        width: sizeValue,
        height: sizeValue,
        minWidth: sizeValue,
        background: gradient,
        fontSize,
        cursor: onClick ? 'pointer' : 'default',
        ...style,
      }}
      onClick={onClick}
    >
      {initial}
    </div>
  )
}

export default UserAvatar

// 导出工具函数，供需要自定义渲染的场景使用
export const getAvatarGradient = (name: string | null | undefined): string => {
  if (!name) return TG_GRADIENTS[0]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return TG_GRADIENTS[Math.abs(hash) % TG_GRADIENTS.length]
}

export const getAvatarInitial = (name: string | null | undefined): string => {
  if (!name) return 'U'
  const firstChar = name.charAt(0)
  if (/[a-zA-Z]/.test(firstChar)) {
    return firstChar.toUpperCase()
  }
  return firstChar
}
