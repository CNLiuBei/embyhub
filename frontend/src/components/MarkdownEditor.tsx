import { useEffect, useRef, useCallback } from 'react'
import { App, Image as AntImage } from 'antd'
import Vditor from 'vditor'
import 'vditor/dist/index.css'
import { imageHostApi } from '../services/api'
import { createRoot } from 'react-dom/client'
import EmojiGifPicker from './EmojiGifPicker'

interface MarkdownEditorProps {
  value?: string
  onChange?: (value: string) => void
  placeholder?: string
  height?: number
  minHeight?: number
  disabled?: boolean
  mode?: 'ir' | 'wysiwyg' | 'sv' // ir=即时渲染, wysiwyg=所见即所得, sv=分屏预览
}

const MarkdownEditor = ({
  value = '',
  onChange,
  placeholder = '支持 Markdown 格式，可直接粘贴或拖拽图片...',
  height,
  minHeight = 300,
  disabled = false,
  mode = 'wysiwyg', // 默认所见即所得模式，图片直接显示
}: MarkdownEditorProps) => {
  const { message } = App.useApp()
  const editorRef = useRef<Vditor | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const valueRef = useRef(value)

  // 保持 value 引用最新
  useEffect(() => {
    valueRef.current = value
  }, [value])

  // 预加载图片
  const preloadImage = (url: string): Promise<void> => {
    return new Promise((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve()
      img.onerror = () => reject(new Error('图片加载失败'))
      img.src = url
    })
  }

  // 自定义图片上传
  const uploadImage = useCallback(
    async (files: File[]): Promise<string | null> => {
      const file = files[0]
      if (!file) return null

      if (!file.type.startsWith('image/')) {
        message.error('只能上传图片文件')
        return null
      }
      if (file.size > 10 * 1024 * 1024) {
        message.error('图片大小不能超过 10MB')
        return null
      }

      const hide = message.loading('图片上传中...', 0)
      try {
        const result = await imageHostApi.upload(file, '', 'forum')
        if (result.success && result.data?.url) {
          // 等待图片加载完成
          await preloadImage(result.data.url)
          hide()
          message.success('图片上传成功')
          return result.data.url
        } else {
          hide()
          message.error(result.error || '上传失败')
          return null
        }
      } catch {
        hide()
        message.error('上传失败')
        return null
      }
    },
    [message]
  )

  // 初始化编辑器
  useEffect(() => {
    if (!containerRef.current) return

    const vditor = new Vditor(containerRef.current, {
      height: height,
      minHeight: minHeight,
      mode: mode,
      placeholder: placeholder,
      theme: 'classic',
      icon: 'ant',
      toolbar: [
        'emoji',
        'headings',
        'bold',
        'italic',
        'strike',
        '|',
        'line',
        'quote',
        'list',
        'ordered-list',
        'check',
        '|',
        'code',
        'inline-code',
        'link',
        'upload',
        'table',
        '|',
        'undo',
        'redo',
        '|',
        'edit-mode',
        'fullscreen',
        'preview',
      ],
      toolbarConfig: {
        pin: false,
      },
      counter: {
        enable: true,
        max: 10000,
      },
      cache: {
        enable: false,
      },
      preview: {
        markdown: {
          toc: true,
          autoSpace: true,
        },
        hljs: {
          lineNumber: true,
          style: 'github',
        },
      },
      hint: {
        emoji: {
          // 笑脸表情
          'smile': '😄', 'laugh': '😂', 'grin': '😁', 'joy': '🤣', 'rofl': '🤣',
          'wink': '😉', 'blush': '😊', 'innocent': '😇', 'love': '🥰', 'kiss': '😘',
          'heart_eyes': '😍', 'star_eyes': '🤩', 'yum': '😋', 'stuck_out_tongue': '😛',
          'crazy': '🤪', 'thinking': '🤔', 'shush': '🤫', 'hand_over_mouth': '🤭',
          // 负面表情
          'unamused': '😒', 'roll_eyes': '🙄', 'grimace': '😬', 'exhale': '😮‍💨',
          'lying': '🤥', 'relieved': '😌', 'pensive': '😔', 'sleepy': '😪',
          'drool': '🤤', 'sleeping': '😴', 'mask': '😷', 'sick': '🤒',
          'hurt': '🤕', 'nauseated': '🤢', 'sneeze': '🤧', 'hot': '🥵',
          'cold': '🥶', 'woozy': '🥴', 'dizzy': '😵', 'exploding': '🤯',
          'cowboy': '🤠', 'party': '🥳', 'cool': '😎', 'nerd': '🤓',
          'monocle': '🧐', 'confused': '😕', 'worried': '😟', 'frown': '🙁',
          'open_mouth': '😮', 'hushed': '😯', 'astonished': '😲', 'flushed': '😳',
          'pleading': '🥺', 'cry': '😢', 'sob': '😭', 'scream': '😱',
          'confounded': '😖', 'persevere': '😣', 'disappointed': '😞', 'sweat': '😓',
          'weary': '😩', 'tired': '😫', 'angry': '😠', 'rage': '😡',
          'cursing': '🤬', 'devil': '😈', 'skull': '💀', 'poop': '💩',
          // 手势
          '+1': '👍', '-1': '👎', 'ok': '👌', 'pinch': '🤌', 'victory': '✌️',
          'crossed_fingers': '🤞', 'love_you': '🤟', 'rock': '🤘', 'call': '🤙',
          'point_left': '👈', 'point_right': '👉', 'point_up': '👆', 'point_down': '👇',
          'wave': '👋', 'raised_hand': '✋', 'clap': '👏', 'handshake': '🤝',
          'pray': '🙏', 'writing': '✍️', 'muscle': '💪', 'leg': '🦵',
          // 心形
          'heart': '❤️', 'orange_heart': '🧡', 'yellow_heart': '💛', 'green_heart': '💚',
          'blue_heart': '💙', 'purple_heart': '💜', 'black_heart': '🖤', 'white_heart': '🤍',
          'broken_heart': '💔', 'heart_fire': '❤️‍🔥', 'sparkling_heart': '💖', 'heartbeat': '💓',
          // 其他
          'fire': '🔥', 'star': '⭐', 'sparkles': '✨', 'zap': '⚡', 'rainbow': '🌈',
          'sun': '☀️', 'moon': '🌙', 'cloud': '☁️', 'rain': '🌧️', 'snow': '❄️',
          'gift': '🎁', 'balloon': '🎈', 'tada': '🎉', 'trophy': '🏆', 'medal': '🏅',
          'crown': '👑', 'gem': '💎', 'money': '💰', 'dollar': '💵', 'coin': '🪙',
          'bell': '🔔', 'music': '🎵', 'mic': '🎤', 'headphones': '🎧', 'guitar': '🎸',
          'camera': '📷', 'video': '📹', 'tv': '📺', 'phone': '📱', 'laptop': '💻',
          'bulb': '💡', 'book': '📖', 'bookmark': '🔖', 'link': '🔗', 'paperclip': '📎',
          'scissors': '✂️', 'lock': '🔒', 'key': '🔑', 'hammer': '🔨', 'wrench': '🔧',
          'check': '✅', 'x': '❌', 'question': '❓', 'exclamation': '❗', 'warning': '⚠️',
          '100': '💯', 'zzz': '💤', 'boom': '💥', 'collision': '💢', 'sweat_drops': '💦',
          // 动物
          'dog': '🐕', 'cat': '🐈', 'mouse': '🐁', 'rabbit': '🐇', 'fox': '🦊',
          'bear': '🐻', 'panda': '🐼', 'koala': '🐨', 'tiger': '🐯', 'lion': '🦁',
          'cow': '🐄', 'pig': '🐷', 'frog': '🐸', 'monkey': '🐵', 'chicken': '🐔',
          'penguin': '🐧', 'bird': '🐦', 'eagle': '🦅', 'duck': '🦆', 'owl': '🦉',
          'butterfly': '🦋', 'bee': '🐝', 'bug': '🐛', 'snail': '🐌', 'octopus': '🐙',
          'fish': '🐟', 'dolphin': '🐬', 'whale': '🐳', 'shark': '🦈', 'crab': '🦀',
          // 食物
          'apple': '🍎', 'orange': '🍊', 'lemon': '🍋', 'banana': '🍌', 'watermelon': '🍉',
          'grapes': '🍇', 'strawberry': '🍓', 'peach': '🍑', 'cherry': '🍒', 'mango': '🥭',
          'pizza': '🍕', 'burger': '🍔', 'fries': '🍟', 'hotdog': '🌭', 'taco': '🌮',
          'sushi': '🍣', 'ramen': '🍜', 'rice': '🍚', 'curry': '🍛', 'bento': '🍱',
          'egg': '🥚', 'cooking': '🍳', 'bread': '🍞', 'croissant': '🥐', 'cake': '🎂',
          'cookie': '🍪', 'chocolate': '🍫', 'candy': '🍬', 'lollipop': '🍭', 'icecream': '🍦',
          'coffee': '☕', 'tea': '🍵', 'beer': '🍺', 'wine': '🍷', 'cocktail': '🍸',
        },
      },
      upload: {
        accept: 'image/*',
        multiple: false,
        handler: async (files: File[]) => {
          const url = await uploadImage(files)
          if (url && editorRef.current) {
            editorRef.current.insertValue(`![图片](${url})`)
          }
          return null
        },
      },
      input: (val: string) => {
        onChange?.(val)
      },
      after: () => {
        editorRef.current = vditor
        if (valueRef.current) {
          vditor.setValue(valueRef.current)
        }
        if (disabled) {
          vditor.disabled()
        }
      },
    })

    return () => {
      editorRef.current?.destroy()
      editorRef.current = null
    }
  }, [height, minHeight, mode, placeholder, disabled, uploadImage, onChange])

  // 同步外部 value 变化
  useEffect(() => {
    if (editorRef.current && value !== editorRef.current.getValue()) {
      editorRef.current.setValue(value)
    }
  }, [value])

  // 同步 disabled 状态
  useEffect(() => {
    if (editorRef.current) {
      if (disabled) {
        editorRef.current.disabled()
      } else {
        editorRef.current.enable()
      }
    }
  }, [disabled])

  // 点击图片放大
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const handleImageClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      if (target.tagName === 'IMG') {
        e.preventDefault()
        e.stopPropagation()
        const src = (target as HTMLImageElement).src
        if (src) {
          // 创建临时容器显示大图
          const div = document.createElement('div')
          document.body.appendChild(div)
          const root = createRoot(div)
          root.render(
            <App>
              <AntImage
                src={src}
                style={{ display: 'none' }}
                preview={{
                  visible: true,
                  onVisibleChange: (visible) => {
                    if (!visible) {
                      root.unmount()
                      div.remove()
                    }
                  },
                }}
              />
            </App>
          )
        }
      }
    }

    container.addEventListener('click', handleImageClick)
    return () => container.removeEventListener('click', handleImageClick)
  }, [])

  // 插入表情
  const handleEmojiSelect = (emoji: string) => {
    if (editorRef.current) {
      editorRef.current.insertValue(emoji)
    }
  }

  // 插入 GIF
  const handleGifSelect = (url: string) => {
    if (editorRef.current) {
      editorRef.current.insertValue(`![GIF](${url})`)
    }
  }

  return (
    <div className="markdown-editor-wrapper">
      <div ref={containerRef} className="vditor-container" />
      <div className="markdown-editor-toolbar">
        <EmojiGifPicker
          onEmojiSelect={handleEmojiSelect}
          onGifSelect={handleGifSelect}
          disabled={disabled}
        />
        <span className="text-xs text-gray-400 ml-2">点击选择表情或 GIF</span>
      </div>
    </div>
  )
}

export default MarkdownEditor
