import { useEffect, useRef, useCallback, useImperativeHandle, forwardRef, useState } from 'react'
import { App, Image as AntImage } from 'antd'
import Vditor from 'vditor'
import 'vditor/dist/index.css'
import { imageHostApi } from '../services/api'
import { createRoot } from 'react-dom/client'
import { createPortal } from 'react-dom'

interface CommentEditorProps {
  value?: string
  onChange?: (value: string) => void
  placeholder?: string
  minHeight?: number
  maxLength?: number
  disabled?: boolean
  renderActions?: () => React.ReactNode // 渲染操作按钮
}

export interface CommentEditorRef {
  insertEmoji: (emoji: string) => void
  insertGif: (url: string) => void
}

const CommentEditor = forwardRef<CommentEditorRef, CommentEditorProps>(({
  value = '',
  onChange,
  placeholder = '写下你的评论... (可粘贴图片)',
  minHeight = 120,
  maxLength = 1000,
  disabled = false,
  renderActions,
}, ref) => {
  const { message } = App.useApp()
  const editorRef = useRef<Vditor | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const valueRef = useRef(value)
  const [toolbarActionsContainer, setToolbarActionsContainer] = useState<HTMLElement | null>(null)

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
        const result = await imageHostApi.upload(file, '', 'forum-comment')
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

  useEffect(() => {
    if (!containerRef.current) return

    const vditor = new Vditor(containerRef.current, {
      minHeight: minHeight,
      mode: 'wysiwyg',
      placeholder: placeholder,
      theme: 'classic',
      toolbar: ['bold', 'italic', 'link', 'upload', '|', 'undo', 'redo'],
      toolbarConfig: {
        pin: false,
      },
      counter: {
        enable: false,
      },
      cache: {
        enable: false,
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
        // 创建工具栏右侧的操作按钮容器
        const toolbar = containerRef.current?.querySelector('.vditor-toolbar') as HTMLElement
        if (toolbar && renderActions) {
          // 确保工具栏是 flex 布局
          toolbar.style.display = 'flex'
          toolbar.style.alignItems = 'center'
          toolbar.style.flexWrap = 'nowrap'
          
          const actionsContainer = document.createElement('div')
          actionsContainer.className = 'comment-editor-actions'
          actionsContainer.style.cssText = 'margin-left: auto; display: flex; align-items: center; gap: 8px; flex-shrink: 0;'
          toolbar.appendChild(actionsContainer)
          setToolbarActionsContainer(actionsContainer)
        }
      },
    })

    return () => {
      editorRef.current?.destroy()
      editorRef.current = null
    }
  }, [minHeight, placeholder, maxLength, disabled, uploadImage, onChange])

  useEffect(() => {
    if (editorRef.current && value !== editorRef.current.getValue()) {
      editorRef.current.setValue(value)
    }
  }, [value])

  useEffect(() => {
    if (editorRef.current) {
      if (disabled) {
        editorRef.current.disabled()
      } else {
        editorRef.current.enable()
      }
    }
  }, [disabled])

  // 暴露方法给父组件
  useImperativeHandle(ref, () => ({
    insertEmoji: (emoji: string) => {
      if (editorRef.current) {
        editorRef.current.insertValue(emoji)
      }
    },
    insertGif: (url: string) => {
      if (editorRef.current) {
        editorRef.current.insertValue(`![GIF](${url})`)
      }
    },
  }))

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

  // 计算纯文本长度（去除 markdown 语法）
  const getTextLength = (text: string) => {
    // 移除图片语法 ![xxx](url)，然后去除首尾空白
    const cleaned = text.replace(/!\[.*?\]\(.*?\)/g, '[图片]').trim()
    return cleaned.length
  }

  const currentLength = getTextLength(value)
  const isOverLimit = currentLength > maxLength

  return (
    <div className="comment-editor-wrapper relative">
      <div ref={containerRef} className="comment-vditor" />
      {/* 字数统计 - 输入框内右下角 */}
      <div
        className={`absolute bottom-2 right-3 text-xs ${isOverLimit ? 'text-red-500' : 'text-gray-400'}`}
        style={{ pointerEvents: 'none' }}
      >
        {currentLength}/{maxLength}
      </div>
      {/* 通过 Portal 将操作按钮渲染到工具栏 */}
      {toolbarActionsContainer && renderActions && createPortal(renderActions(), toolbarActionsContainer)}
    </div>
  )
})

CommentEditor.displayName = 'CommentEditor'

export default CommentEditor
