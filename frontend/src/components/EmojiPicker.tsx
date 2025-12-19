import { useState } from 'react'
import { Button, Popover, Tabs } from 'antd'
import { SmileOutlined } from '@ant-design/icons'

// 常用表情分类
const emojiCategories = {
  '常用': ['😀', '😁', '😂', '🤣', '😃', '😄', '😅', '😆', '😉', '😊', '😋', '😎', '😍', '🥰', '😘', '😗', '😙', '😚', '🙂', '🤗', '🤩', '🤔', '🤨', '😐', '😑', '😶', '🙄', '😏', '😣', '😥', '😮', '🤐', '😯', '😪', '😫', '🥱', '😴', '😌', '😛', '😜', '😝', '🤤', '😒', '😓', '😔', '😕', '🙃', '🤑', '😲'],
  '手势': ['👍', '👎', '👌', '🤌', '🤏', '✌️', '🤞', '🤟', '🤘', '🤙', '👈', '👉', '👆', '👇', '☝️', '👋', '🤚', '🖐️', '✋', '🖖', '👏', '🙌', '👐', '🤲', '🤝', '🙏', '✍️', '💪', '🦾', '🦿'],
  '表情': ['😭', '😢', '😤', '😠', '😡', '🤬', '😈', '👿', '💀', '☠️', '💩', '🤡', '👹', '👺', '👻', '👽', '👾', '🤖', '😺', '😸', '😹', '😻', '😼', '😽', '🙀', '😿', '😾', '🙈', '🙉', '🙊'],
  '爱心': ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❣️', '💕', '💞', '💓', '💗', '💖', '💘', '💝', '💟', '♥️', '💌', '💋', '👄', '👅'],
  '物品': ['🎉', '🎊', '🎁', '🎈', '🎀', '🎄', '🎃', '🎗️', '🎟️', '🎫', '🎖️', '🏆', '🏅', '🥇', '🥈', '🥉', '⚽', '🏀', '🏈', '⚾', '🥎', '🎾', '🏐', '🏉', '🎱', '🎮', '🎲', '🎯', '🎳', '🎸'],
  '自然': ['🌸', '💮', '🏵️', '🌹', '🥀', '🌺', '🌻', '🌼', '🌷', '🌱', '🌲', '🌳', '🌴', '🌵', '🌾', '🌿', '☘️', '🍀', '🍁', '🍂', '🍃', '🌍', '🌎', '🌏', '🌐', '🌑', '🌒', '🌓', '🌔', '🌕'],
  '食物': ['🍎', '🍐', '🍊', '🍋', '🍌', '🍉', '🍇', '🍓', '🫐', '🍈', '🍒', '🍑', '🥭', '🍍', '🥥', '🥝', '🍅', '🍆', '🥑', '🥦', '🥬', '🥒', '🌶️', '🫑', '🌽', '🥕', '🫒', '🧄', '🧅', '🥔'],
}

interface EmojiPickerProps {
  onSelect: (emoji: string) => void
  disabled?: boolean
}

const EmojiPicker = ({ onSelect, disabled }: EmojiPickerProps) => {
  const [open, setOpen] = useState(false)

  const handleSelect = (emoji: string) => {
    onSelect(emoji)
    // 不关闭，允许连续选择
  }

  const content = (
    <div className="w-80">
      <Tabs
        size="small"
        items={Object.entries(emojiCategories).map(([name, emojis]) => ({
          key: name,
          label: name,
          children: (
            <div className="grid grid-cols-8 gap-1 max-h-48 overflow-y-auto p-1">
              {emojis.map((emoji, index) => (
                <button
                  key={index}
                  className="w-8 h-8 flex items-center justify-center text-xl hover:bg-gray-100 rounded cursor-pointer transition-colors border-0 bg-transparent"
                  onClick={() => handleSelect(emoji)}
                >
                  {emoji}
                </button>
              ))}
            </div>
          ),
        }))}
      />
    </div>
  )

  return (
    <Popover
      content={content}
      trigger="click"
      open={open}
      onOpenChange={setOpen}
      placement="topLeft"
    >
      <Button
        type="text"
        icon={<SmileOutlined />}
        disabled={disabled}
        className="text-gray-500 hover:text-blue-500"
      />
    </Popover>
  )
}

export default EmojiPicker
