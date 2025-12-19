import { useState } from 'react'
import { Popover, Tabs, Spin } from 'antd'
import { SmileOutlined } from '@ant-design/icons'
import Picker from '@emoji-mart/react'
import data from '@emoji-mart/data'
import { GiphyFetch } from '@giphy/js-fetch-api'
import { Grid } from '@giphy/react-components'

// Giphy API Key (免费公开 key，可替换为自己的)
const gf = new GiphyFetch('Gc7131jiJuvI7IdN0HZ1D7nh0uj5cxxx')

interface EmojiGifPickerProps {
  onEmojiSelect?: (emoji: string) => void
  onGifSelect?: (url: string) => void
  disabled?: boolean
  children?: React.ReactNode
}

const EmojiGifPicker = ({
  onEmojiSelect,
  onGifSelect,
  disabled = false,
  children,
}: EmojiGifPickerProps) => {
  const [open, setOpen] = useState(false)
  const [activeTab, setActiveTab] = useState('emoji')

  const handleEmojiSelect = (emoji: { native: string }) => {
    onEmojiSelect?.(emoji.native)
  }

  const handleGifClick = (gif: { images: { fixed_height: { url: string } } }, e: React.SyntheticEvent) => {
    e.preventDefault()
    const url = gif.images.fixed_height.url
    onGifSelect?.(url)
    setOpen(false)
  }

  const fetchGifs = (offset: number) => gf.trending({ offset, limit: 10 })

  const content = (
    <div style={{ width: 352, height: 400 }}>
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        size="small"
        items={[
          {
            key: 'emoji',
            label: '😀 表情',
            children: (
              <div style={{ height: 340 }}>
                <Picker
                  data={data}
                  onEmojiSelect={handleEmojiSelect}
                  locale="zh"
                  theme="light"
                  previewPosition="none"
                  skinTonePosition="search"
                  navPosition="bottom"
                  perLine={9}
                  emojiSize={24}
                  emojiButtonSize={32}
                />
              </div>
            ),
          },
          {
            key: 'gif',
            label: '🎬 GIF',
            children: (
              <div style={{ height: 340, overflow: 'auto' }}>
                <Grid
                  width={336}
                  columns={2}
                  fetchGifs={fetchGifs}
                  onGifClick={handleGifClick}
                  noLink
                  hideAttribution
                  loader={() => (
                    <div className="flex justify-center py-4">
                      <Spin />
                    </div>
                  )}
                />
              </div>
            ),
          },
        ]}
      />
    </div>
  )

  return (
    <Popover
      content={content}
      trigger="click"
      open={open}
      onOpenChange={setOpen}
      placement="topRight"
      arrow={false}
    >
      {children || (
        <button
          type="button"
          disabled={disabled}
          className="p-1 text-gray-500 hover:text-blue-500 transition-colors disabled:opacity-50"
          title="表情 / GIF"
        >
          <SmileOutlined style={{ fontSize: 18 }} />
        </button>
      )}
    </Popover>
  )
}

export default EmojiGifPicker
