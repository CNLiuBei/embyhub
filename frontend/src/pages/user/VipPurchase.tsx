/**
 * VIP购买页面（已废弃 - 功能已集成到会员中心）
 * @deprecated 此页面已迁移到会员中心的弹窗中
 */

import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Result, Button, Spin } from 'antd'
import { CrownOutlined } from '@ant-design/icons'

const VipPurchase = () => {
  const navigate = useNavigate()

  // 3秒后自动跳转到会员中心
  useEffect(() => {
    const timer = setTimeout(() => {
      navigate('/user/member')
    }, 3000)

    return () => clearTimeout(timer)
  }, [navigate])

  return (
    <div className="flex items-center justify-center" style={{ minHeight: '60vh' }}>
      <Result
        icon={<CrownOutlined style={{ fontSize: 72, color: '#faad14' }} />}
        title="功能已迁移到会员中心"
        subTitle={
          <div>
            <div>VIP购买功能现已集成到会员中心页面</div>
            <div style={{ marginTop: 8, color: '#8c8c8c' }}>
              <Spin size="small" /> 3秒后自动跳转...
            </div>
          </div>
        }
        extra={[
          <Button
            type="primary"
            key="member"
            size="large"
            icon={<CrownOutlined />}
            onClick={() => navigate('/user/member')}
          >
            立即前往会员中心
          </Button>,
        ]}
      />
    </div>
  )
}

export default VipPurchase
