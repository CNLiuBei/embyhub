import { useState } from 'react'
import { Image } from 'antd'
import MDEditor from '@uiw/react-md-editor'

interface MarkdownContentProps {
  content: string
  className?: string
  compact?: boolean // 紧凑模式，图片更小
}

const MarkdownContent = ({ content, className = '', compact = false }: MarkdownContentProps) => {
  const [previewVisible, setPreviewVisible] = useState(false)
  const [previewSrc, setPreviewSrc] = useState('')

  return (
    <div className={className} data-color-mode="light">
      <MDEditor.Markdown
        source={content}
        style={{ backgroundColor: 'transparent', padding: 0 }}
        components={{
          img: ({ src, alt }) => (
            <img
              src={src}
              alt={alt || '图片'}
              className="rounded-lg cursor-zoom-in hover:opacity-90 transition-opacity"
              style={{
                maxWidth: compact ? '200px' : '100%',
                maxHeight: compact ? '150px' : '500px',
                objectFit: 'contain',
              }}
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
                if (src) {
                  setPreviewSrc(src)
                  setPreviewVisible(true)
                }
              }}
            />
          ),
          a: ({ href, children }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-500 hover:underline"
            >
              {children}
            </a>
          ),
        }}
      />
      <Image
        src={previewSrc}
        style={{ display: 'none' }}
        preview={{
          visible: previewVisible,
          onVisibleChange: setPreviewVisible,
        }}
      />
    </div>
  )
}

export default MarkdownContent
