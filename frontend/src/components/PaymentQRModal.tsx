import { useState, useEffect, useCallback } from 'react'
import { Modal, Spin, Result, Button, Typography, Space, Tag } from 'antd'
import { QRCodeSVG } from 'qrcode.react'
import { CheckCircleOutlined, CloseCircleOutlined, AlipayCircleOutlined, ReloadOutlined } from '@ant-design/icons'
import { paymentApi } from '../services/paymentApi'

const { Text, Title } = Typography

interface PaymentQRModalProps {
  open: boolean
  onClose: () => void
  orderNo: string
  qrCode: string
  amount: number
  planName: string
  onSuccess?: () => void
}

type PaymentStatus = 'pending' | 'success' | 'failed' | 'expired'

const PaymentQRModal: React.FC<PaymentQRModalProps> = ({
  open,
  onClose,
  orderNo,
  qrCode,
  amount,
  planName,
  onSuccess,
}) => {
  const [status, setStatus] = useState<PaymentStatus>('pending')
  const [countdown, setCountdown] = useState(30 * 60) // 30分钟倒计时

  // 轮询订单状态
  const pollOrderStatus = useCallback(async () => {
    if (!orderNo || status !== 'pending') return

    try {
      const response = await paymentApi.getOrderStatus(orderNo)
      const data = response.data.data

      if (data.status === 'success') {
        setStatus('success')
        onSuccess?.()
      } else if (data.status === 'failed' || data.status === 'closed') {
        setStatus('failed')
      }
    } catch (error) {
      console.error('查询订单状态失败:', error)
    }
  }, [orderNo, status, onSuccess])

  // 轮询定时器
  useEffect(() => {
    if (!open || status !== 'pending') return

    // 立即查询一次
    pollOrderStatus()

    // 每3秒轮询一次
    const pollInterval = setInterval(pollOrderStatus, 3000)

    return () => clearInterval(pollInterval)
  }, [open, status, pollOrderStatus])

  // 倒计时
  useEffect(() => {
    if (!open || status !== 'pending') return

    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          setStatus('expired')
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(timer)
  }, [open, status])

  // 重置状态
  useEffect(() => {
    if (open) {
      setStatus('pending')
      setCountdown(30 * 60)
    }
  }, [open, orderNo])

  // 格式化倒计时
  const formatCountdown = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  // 格式化金额
  const formatAmount = (cents: number) => {
    return (cents / 100).toFixed(2)
  }

  const renderContent = () => {
    switch (status) {
      case 'success':
        return (
          <Result
            status="success"
            icon={<CheckCircleOutlined style={{ color: '#52c41a', fontSize: 72 }} />}
            title="支付成功"
            subTitle={`您已成功购买 ${planName}`}
            extra={
              <Button type="primary" onClick={onClose}>
                完成
              </Button>
            }
          />
        )

      case 'failed':
        return (
          <Result
            status="error"
            icon={<CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 72 }} />}
            title="支付失败"
            subTitle="订单已关闭或支付失败，请重新下单"
            extra={
              <Button type="primary" onClick={onClose}>
                关闭
              </Button>
            }
          />
        )

      case 'expired':
        return (
          <Result
            status="warning"
            title="订单已过期"
            subTitle="支付超时，请重新下单"
            extra={
              <Button type="primary" onClick={onClose}>
                重新下单
              </Button>
            }
          />
        )

      default:
        return (
          <div className="text-center py-4">
            {/* 支付宝图标 */}
            <div className="mb-4">
              <AlipayCircleOutlined style={{ fontSize: 48, color: '#1677ff' }} />
              <Title level={4} style={{ marginTop: 8, marginBottom: 0 }}>
                支付宝扫码支付
              </Title>
            </div>

            {/* 二维码 */}
            <div className="inline-block p-4 bg-white rounded-lg shadow-sm border">
              {qrCode ? (
                <QRCodeSVG
                  value={qrCode}
                  size={200}
                  level="H"
                  includeMargin
                />
              ) : (
                <div className="w-[200px] h-[200px] flex items-center justify-center">
                  <Spin size="large" />
                </div>
              )}
            </div>

            {/* 订单信息 */}
            <div className="mt-4">
              <Space direction="vertical" size="small">
                <div>
                  <Text type="secondary">套餐：</Text>
                  <Tag color="blue">{planName}</Tag>
                </div>
                <div>
                  <Text type="secondary">金额：</Text>
                  <Text strong style={{ fontSize: 24, color: '#ff4d4f' }}>
                    ¥{formatAmount(amount)}
                  </Text>
                </div>
                <div>
                  <Text type="secondary">订单号：</Text>
                  <Text copyable={{ text: orderNo }}>{orderNo}</Text>
                </div>
              </Space>
            </div>

            {/* 倒计时 */}
            <div className="mt-4 text-gray-500">
              <Text type="secondary">
                请在 <Text strong style={{ color: countdown < 60 ? '#ff4d4f' : '#1677ff' }}>
                  {formatCountdown(countdown)}
                </Text> 内完成支付
              </Text>
            </div>

            {/* 刷新按钮 */}
            <div className="mt-4">
              <Button
                icon={<ReloadOutlined />}
                onClick={pollOrderStatus}
                loading={false}
              >
                刷新状态
              </Button>
            </div>

            {/* 提示 */}
            <div className="mt-4 text-left bg-gray-50 p-3 rounded text-sm text-gray-500">
              <div>温馨提示：</div>
              <ul className="list-disc list-inside mt-1">
                <li>请使用支付宝扫描二维码完成支付</li>
                <li>支付成功后会员将自动开通</li>
                <li>如遇问题请联系客服</li>
              </ul>
            </div>
          </div>
        )
    }
  }

  return (
    <Modal
      title={null}
      open={open}
      onCancel={onClose}
      footer={null}
      width={400}
      centered
      destroyOnClose
      maskClosable={status === 'success' || status === 'failed' || status === 'expired'}
    >
      {renderContent()}
    </Modal>
  )
}

export default PaymentQRModal
