import { useState, useEffect } from 'react'
import { Card, Form, Input, InputNumber, Switch, Button, Divider, Space, Alert, Collapse, App, Tag, Tabs, Radio, Upload } from 'antd'
import { MailOutlined, SendOutlined, SaveOutlined, QuestionCircleOutlined, GlobalOutlined, PlusOutlined, SettingOutlined, UserAddOutlined, PlayCircleOutlined, ApiOutlined, HomeOutlined, UploadOutlined, PictureOutlined } from '@ant-design/icons'
import { adminApi } from '../../services/api'

interface EmailSettings {
  enabled: boolean
  host: string
  port: number
  username: string
  password: string
  from: string
  from_name: string
}

interface DomainSettings {
  enabled: boolean
  domains: string[]
}

interface RegisterSettings {
  enabled: boolean
  gift_member_days: number
  auto_disable_on_exp: boolean
}

interface EmbySettings {
  enabled: boolean
  mode: string
  base_url: string
  api_key: string
  admin_user: string
  admin_pass: string
  template_user: string
}

interface SiteSettings {
  title: string
  description: string
  keywords: string
  logo: string
  favicon: string
  footer: string
  icp: string
}

const Settings = () => {
  const [activeTab, setActiveTab] = useState('site')
  
  const [emailLoading, setEmailLoading] = useState(false)
  const [emailSaving, setEmailSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [emailForm] = Form.useForm()
  const [testEmail, setTestEmail] = useState('')
  
  const [domainLoading, setDomainLoading] = useState(false)
  const [domainSaving, setDomainSaving] = useState(false)
  const [domainForm] = Form.useForm()
  const [newDomain, setNewDomain] = useState('')
  
  const [registerLoading, setRegisterLoading] = useState(false)
  const [registerSaving, setRegisterSaving] = useState(false)
  const [registerForm] = Form.useForm()
  
  const [embyLoading, setEmbyLoading] = useState(false)
  const [embySaving, setEmbySaving] = useState(false)
  const [embyTesting, setEmbyTesting] = useState(false)
  const [embyForm] = Form.useForm()
  const [embyMode, setEmbyMode] = useState('emby')
  
  const [siteLoading, setSiteLoading] = useState(false)
  const [siteSaving, setSiteSaving] = useState(false)
  const [siteForm] = Form.useForm()
  
  const { message } = App.useApp()

  useEffect(() => {
    // 根据当前激活的Tab加载对应的数据
    if (activeTab === 'site') {
      loadSiteSettings()
    } else if (activeTab === 'domain') {
      loadDomainSettings()
    } else if (activeTab === 'email') {
      loadEmailSettings()
    } else if (activeTab === 'register') {
      loadRegisterSettings()
    } else if (activeTab === 'emby') {
      loadEmbySettings()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab])

  const loadEmailSettings = async () => {
    try {
      setEmailLoading(true)
      const response = await adminApi.getEmailSettings()
      const data = response.data.data as EmailSettings
      // 转换字段名：enabled -> email_enabled
      emailForm.setFieldsValue({
        ...data,
        email_enabled: data.enabled
      })
    } catch (err) {
      message.error('加载设置失败')
    } finally {
      setEmailLoading(false)
    }
  }

  const loadDomainSettings = async () => {
    try {
      setDomainLoading(true)
      const response = await adminApi.getDomainSettings()
      const data = response.data.data as DomainSettings
      // 转换字段名：enabled -> domain_enabled
      domainForm.setFieldsValue({
        domain_enabled: data.enabled,
        domains: data.domains || []
      })
    } catch (err) {
      message.error('加载域名设置失败')
    } finally {
      setDomainLoading(false)
    }
  }

  const loadRegisterSettings = async () => {
    try {
      setRegisterLoading(true)
      const response = await adminApi.getRegisterSettings()
      const data = response.data.data as RegisterSettings
      registerForm.setFieldsValue({
        register_enabled: data.enabled,
        gift_member_days: data.gift_member_days,
        auto_disable_on_exp: data.auto_disable_on_exp
      })
    } catch (err) {
      message.error('加载注册设置失败')
    } finally {
      setRegisterLoading(false)
    }
  }

  const loadEmbySettings = async () => {
    try {
      setEmbyLoading(true)
      const response = await adminApi.getEmbySettings()
      const data = response.data.data as EmbySettings
      setEmbyMode(data.mode || 'emby')
      embyForm.setFieldsValue({
        emby_enabled: data.enabled,
        mode: data.mode || 'emby',
        base_url: data.base_url,
        api_key: data.api_key,
        admin_user: data.admin_user,
        admin_pass: data.admin_pass,
        template_user: data.template_user
      })
    } catch (err) {
      message.error('加载Emby设置失败')
    } finally {
      setEmbyLoading(false)
    }
  }

  const loadSiteSettings = async () => {
    try {
      setSiteLoading(true)
      const response = await adminApi.getSiteSettings()
      const data = response.data.data as SiteSettings
      siteForm.setFieldsValue(data)
    } catch (err) {
      message.error('加载网站设置失败')
    } finally {
      setSiteLoading(false)
    }
  }

  const handleRegisterSave = async () => {
    try {
      const values = await registerForm.validateFields()
      setRegisterSaving(true)
      const payload = {
        enabled: values.register_enabled,
        gift_member_days: values.gift_member_days || 0,
        auto_disable_on_exp: values.auto_disable_on_exp
      }
      await adminApi.saveRegisterSettings(payload)
      message.success('注册设置保存成功')
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(String(err) || '保存失败')
    } finally {
      setRegisterSaving(false)
    }
  }

  const handleSiteSave = async () => {
    try {
      const values = await siteForm.validateFields()
      setSiteSaving(true)
      await adminApi.saveSiteSettings(values)
      message.success('网站设置保存成功')
      // 更新页面标题
      if (values.title) {
        document.title = values.title
      }
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(String(err) || '保存失败')
    } finally {
      setSiteSaving(false)
    }
  }

  const handleEmbySave = async () => {
    try {
      const values = await embyForm.validateFields()
      setEmbySaving(true)
      const payload = {
        enabled: values.emby_enabled,
        mode: values.mode,
        base_url: values.base_url,
        api_key: values.api_key || '',
        admin_user: values.admin_user || '',
        admin_pass: values.admin_pass || '',
        template_user: values.template_user || ''
      }
      await adminApi.saveEmbySettings(payload)
      message.success('Emby设置保存成功')
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(String(err) || '保存失败')
    } finally {
      setEmbySaving(false)
    }
  }

  const handleEmbyTest = async () => {
    try {
      setEmbyTesting(true)
      const response = await adminApi.testEmbyConnection()
      const data = response.data.data as { message: string; user_count: number }
      message.success(`${data.message}，共 ${data.user_count} 个用户`)
    } catch (err: unknown) {
      message.error(String(err) || '连接测试失败')
    } finally {
      setEmbyTesting(false)
    }
  }

  const handleEmailSave = async () => {
    try {
      const values = await emailForm.validateFields()
      setEmailSaving(true)
      // 转换字段名：email_enabled -> enabled
      const payload = {
        enabled: values.email_enabled,
        host: values.host,
        port: values.port,
        username: values.username,
        password: values.password,
        from: values.from,
        from_name: values.from_name
      }
      await adminApi.saveEmailSettings(payload)
      message.success('保存成功')
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(String(err) || '保存失败')
    } finally {
      setEmailSaving(false)
    }
  }

  const handleDomainSave = async () => {
    try {
      const values = await domainForm.validateFields()
      setDomainSaving(true)
      // 转换字段名：domain_enabled -> enabled
      const payload = {
        enabled: values.domain_enabled,
        domains: values.domains || []
      }
      await adminApi.saveDomainSettings(payload)
      message.success('域名设置保存成功')
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) return
      message.error(String(err) || '保存失败')
    } finally {
      setDomainSaving(false)
    }
  }

  const handleAddDomain = () => {
    if (!newDomain.trim()) {
      message.warning('请输入域名')
      return
    }
    const currentDomains = domainForm.getFieldValue('domains') || []
    if (currentDomains.includes(newDomain.trim())) {
      message.warning('该域名已存在')
      return
    }
    domainForm.setFieldsValue({ domains: [...currentDomains, newDomain.trim()] })
    setNewDomain('')
  }

  const handleRemoveDomain = (domain: string) => {
    const currentDomains = domainForm.getFieldValue('domains') || []
    domainForm.setFieldsValue({ domains: currentDomains.filter((d: string) => d !== domain) })
  }

  const handleAddCurrentDomain = () => {
    const currentDomain = window.location.hostname
    const currentDomains = domainForm.getFieldValue('domains') || []
    if (currentDomains.includes(currentDomain)) {
      message.info('当前域名已在白名单中')
      return
    }
    domainForm.setFieldsValue({ domains: [...currentDomains, currentDomain] })
    message.success(`已添加当前域名：${currentDomain}`)
  }

  const handleTest = async () => {
    if (!testEmail) {
      message.warning('请输入测试邮箱')
      return
    }
    try {
      setTesting(true)
      await adminApi.testEmailSettings(testEmail)
      message.success('测试邮件已发送')
    } catch (err: unknown) {
      message.error(String(err) || '发送失败')
    } finally {
      setTesting(false)
    }
  }

  const tabItems = [
    {
      key: 'site',
      label: (
        <span>
          <HomeOutlined className="mr-2" />
          网站设置
        </span>
      ),
      children: (
        <div className="p-4">
          {siteLoading ? (
            <div className="text-center py-20">加载中...</div>
          ) : (
            <>
              <Alert
                message="功能说明"
                description={
                  <div className="text-sm">
                    <p>配置网站的基本信息，包括标题、描述、页脚等。</p>
                    <p className="text-blue-600 mt-2">💡 修改标题后会立即更新浏览器标签页显示</p>
                  </div>
                }
                type="info"
                showIcon
                className="mb-6"
              />

              <Form 
                form={siteForm} 
                layout="vertical" 
                className="max-w-2xl"
                name="siteSettingsForm"
                initialValues={{ 
                  title: 'EmbyHub - 用户管理系统',
                  description: 'EmbyHub用户管理系统',
                  footer: '© 2025 EmbyHub'
                }}
              >
                <Divider orientation="left" plain>品牌信息</Divider>

                <Form.Item 
                  label="网站 Logo"
                  extra="支持 PNG、JPG、GIF、SVG、WebP 格式，最大 2MB"
                >
                  <Space direction="vertical" className="w-full">
                    <Space.Compact className="w-full">
                      <Form.Item name="logo" noStyle>
                        <Input 
                          placeholder="输入 Logo 图片 URL，或点击上传" 
                          prefix={<PictureOutlined className="text-gray-400" />}
                          className="flex-1"
                        />
                      </Form.Item>
                      <Upload
                        accept=".png,.jpg,.jpeg,.gif,.svg,.webp,.ico"
                        showUploadList={false}
                        beforeUpload={async (file) => {
                          if (file.size > 2 * 1024 * 1024) {
                            message.error('文件大小不能超过 2MB')
                            return false
                          }
                          try {
                            const response = await adminApi.uploadLogo(file)
                            const url = response.data?.data?.url
                            if (url) {
                              siteForm.setFieldValue('logo', url)
                              message.success('Logo 上传成功')
                            }
                          } catch (err) {
                            message.error('上传失败')
                          }
                          return false
                        }}
                      >
                        <Button icon={<UploadOutlined />}>上传</Button>
                      </Upload>
                    </Space.Compact>
                    <Form.Item noStyle shouldUpdate>
                      {() => {
                        const logoUrl = siteForm.getFieldValue('logo')
                        return logoUrl ? (
                          <div className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
                            <img src={logoUrl} alt="Logo预览" className="w-16 h-16 object-contain rounded-lg border" />
                            <div className="flex-1">
                              <div className="text-gray-600 text-sm font-medium">Logo 预览</div>
                              <div className="text-gray-400 text-xs truncate max-w-xs">{logoUrl}</div>
                            </div>
                            <Button 
                              type="text" 
                              danger 
                              size="small"
                              onClick={() => siteForm.setFieldValue('logo', '')}
                            >
                              清除
                            </Button>
                          </div>
                        ) : (
                          <div className="p-4 bg-gray-50 rounded-lg text-center text-gray-400 text-sm">
                            暂未设置 Logo，将显示默认样式
                          </div>
                        )
                      }}
                    </Form.Item>
                  </Space>
                </Form.Item>

                <Divider orientation="left" plain>基本信息</Divider>

                <Form.Item 
                  name="title" 
                  label="网站标题" 
                  rules={[{ required: true, message: '请输入网站标题' }]}
                  extra="显示在浏览器标签页和页面顶部"
                >
                  <Input placeholder="如：EmbyHub - 用户管理系统" />
                </Form.Item>

                <Form.Item 
                  name="description" 
                  label="网站描述"
                  extra="用于SEO优化，显示在搜索引擎结果中"
                >
                  <Input.TextArea placeholder="网站描述" rows={2} />
                </Form.Item>

                <Form.Item 
                  name="keywords" 
                  label="SEO关键词"
                  extra="多个关键词用逗号分隔"
                >
                  <Input placeholder="如：EmbyHub,Emby,媒体服务" />
                </Form.Item>

                <Divider orientation="left" plain>页脚信息</Divider>

                <Form.Item 
                  name="footer" 
                  label="页脚文字"
                >
                  <Input placeholder="如：© 2025 EmbyHub" />
                </Form.Item>

                <Form.Item 
                  name="icp" 
                  label="ICP备案号"
                  extra="如有备案，填写备案号"
                >
                  <Input placeholder="如：京ICP备XXXXXXXX号" />
                </Form.Item>

                <Divider orientation="left" plain>社交链接</Divider>

                <Form.Item 
                  name="github_url" 
                  label="GitHub 链接"
                  extra="显示在关于页面"
                >
                  <Input placeholder="如：https://github.com/yourname/project" />
                </Form.Item>

                <Form.Item 
                  name="telegram_url" 
                  label="Telegram 链接"
                  extra="Telegram 群组或频道链接"
                >
                  <Input placeholder="如：https://t.me/yourgroup" />
                </Form.Item>

                <Form.Item 
                  name="qq_url" 
                  label="QQ 群链接"
                  extra="QQ 群加群链接"
                >
                  <Input placeholder="如：https://qm.qq.com/..." />
                </Form.Item>

                <Form.Item>
                  <Button type="primary" icon={<SaveOutlined />} onClick={handleSiteSave} loading={siteSaving}>
                    保存设置
                  </Button>
                </Form.Item>
              </Form>
            </>
          )}
        </div>
      ),
    },
    {
      key: 'domain',
      label: (
        <span>
          <GlobalOutlined className="mr-2" />
          域名访问控制
        </span>
      ),
      children: (
        <div className="p-4">
          {domainLoading ? (
            <div className="text-center py-20">加载中...</div>
          ) : (
            <>
              <Alert
                message="功能说明"
                description={
                  <div className="text-sm">
                    <p>启用域名白名单后，只有列表中的域名才能访问系统API（无法登录和使用功能）。</p>
                    <p className="text-blue-600 mt-2">💡 当前访问域名：<strong>{window.location.hostname}</strong></p>
                    <p className="text-orange-500 mt-1">⚠️ 修改配置后立即生效，无需重启服务</p>
                    <p className="text-green-600 mt-1">✓ 绿色标签表示当前域名已在白名单中</p>
                    <p className="text-red-600 mt-1 font-semibold">⚠️ 请务必先添加当前域名再启用，否则将被锁定！</p>
                    <p className="text-gray-500 mt-1 text-xs">Go后端动态控制，配置立即生效</p>
                    <p className="text-yellow-600 mt-1 text-xs">💡 生产环境建议配合Nginx使用以完全阻止页面访问</p>
                  </div>
                }
                type="warning"
                showIcon
                className="mb-6"
              />

              <Form form={domainForm} layout="vertical" className="max-w-2xl" name="domainSettingsForm" initialValues={{ enabled: false, domains: [] }}>
                <Form.Item name="domain_enabled" label="启用域名白名单" valuePropName="checked">
                  <Switch checkedChildren="开启" unCheckedChildren="关闭" />
                </Form.Item>

                <Form.Item label="域名列表">
                  <Space direction="vertical" className="w-full">
                    <Space className="w-full">
                      <Space.Compact className="flex-1">
                        <Input 
                          placeholder="输入允许访问的域名，如：example.com 或 *.example.com" 
                          value={newDomain}
                          onChange={(e) => setNewDomain(e.target.value)}
                          onPressEnter={handleAddDomain}
                        />
                        <Button type="primary" icon={<PlusOutlined />} onClick={handleAddDomain}>
                          添加
                        </Button>
                      </Space.Compact>
                      <Button onClick={handleAddCurrentDomain} type="dashed">
                        添加当前域名
                      </Button>
                    </Space>
                    
                    <Form.Item noStyle shouldUpdate={(prev, curr) => prev.domains !== curr.domains}>
                      {() => {
                        const domains = domainForm.getFieldValue('domains') || []
                        const currentDomain = window.location.hostname
                        return (
                          <div className="flex flex-wrap gap-2 mt-2">
                            {domains.length === 0 ? (
                              <div className="text-gray-400 text-sm">暂无域名，请添加至少一个域名</div>
                            ) : (
                              domains.map((domain: string) => {
                                // 检查是否为当前访问域名
                                const isCurrentDomain = domain === currentDomain || 
                                  (domain.startsWith('*.') && currentDomain.endsWith(domain.substring(1)))
                                
                                return (
                                  <Tag
                                    key={domain}
                                    closable
                                    onClose={() => handleRemoveDomain(domain)}
                                    color={isCurrentDomain ? 'green' : 'blue'}
                                    className="px-3 py-1"
                                  >
                                    {domain}
                                    {isCurrentDomain && <span className="ml-1">✓</span>}
                                  </Tag>
                                )
                              })
                            )}
                          </div>
                        )
                      }}
                    </Form.Item>
                    
                    {/* 隐藏的表单字段用于存储域名列表 */}
                    <Form.Item name="domains" hidden>
                      <Input />
                    </Form.Item>

                    <div className="text-xs text-gray-400 mt-2">
                      <div>• <strong className="text-green-600">绿色标签</strong>：当前访问域名（已在白名单中）</div>
                      <div>• <strong className="text-blue-600">蓝色标签</strong>：其他已添加的域名</div>
                      <div>• 支持通配符，如：*.example.com 表示所有子域名</div>
                      <div>• 建议添加：localhost、127.0.0.1（开发环境）</div>
                    </div>
                  </Space>
                </Form.Item>

                <Form.Item>
                  <Button type="primary" icon={<SaveOutlined />} onClick={handleDomainSave} loading={domainSaving}>
                    保存设置
                  </Button>
                </Form.Item>
              </Form>
            </>
          )}
        </div>
      ),
    },
    {
      key: 'email',
      label: (
        <span>
          <MailOutlined className="mr-2" />
          邮件服务配置
        </span>
      ),
      children: (
        <div className="p-4">
          {emailLoading ? (
            <div className="text-center py-20">加载中...</div>
          ) : (
            <>
              <Alert
                message="配置说明"
                description={
                  <div className="text-sm">
                    <p className="mb-2">邮件服务用于发送密码重置验证码等系统邮件。配置前请先获取邮箱的 SMTP 授权码。</p>
                    <Collapse 
                      size="small" 
                      ghost
                      items={[
                        {
                          key: '1',
                          label: <span className="text-blue-500"><QuestionCircleOutlined className="mr-1" />常用邮箱配置参考</span>,
                          children: (
                            <div className="space-y-2 text-gray-600">
                              <div><strong>QQ邮箱：</strong>smtp.qq.com，端口 587 或 465(SSL)</div>
                              <div><strong>163邮箱：</strong>smtp.163.com，端口 25 或 465(SSL)</div>
                              <div><strong>Gmail：</strong>smtp.gmail.com，端口 587</div>
                              <div><strong>阿里企业邮箱：</strong>smtp.qiye.aliyun.com，端口 465(SSL)</div>
                              <div className="text-orange-500 mt-2">⚠️ 密码请填写 SMTP 授权码，而非邮箱登录密码</div>
                            </div>
                          ),
                        },
                      ]}
                    />
                  </div>
                }
                type="info"
                showIcon
                className="mb-6"
              />

              <Form form={emailForm} layout="vertical" className="max-w-2xl" name="emailSettingsForm">
                <Form.Item name="email_enabled" label="启用邮件服务" valuePropName="checked">
                  <Switch checkedChildren="开启" unCheckedChildren="关闭" />
                </Form.Item>

                <Divider orientation="left" plain>SMTP 服务器</Divider>

                <div className="grid grid-cols-2 gap-4">
                  <Form.Item name="host" label="SMTP服务器" rules={[{ required: true, message: '请输入SMTP服务器' }]}>
                    <Input placeholder="如: smtp.qq.com" />
                  </Form.Item>
                  <Form.Item name="port" label="端口" rules={[{ required: true, message: '请输入端口' }]}>
                    <InputNumber placeholder="587" min={1} max={65535} className="w-full" />
                  </Form.Item>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <Form.Item name="username" label="用户名/邮箱" rules={[{ required: true, message: '请输入用户名' }]}>
                    <Input placeholder="发件邮箱账号" />
                  </Form.Item>
                  <Form.Item name="password" label="密码/授权码" rules={[{ required: true, message: '请输入密码' }]}>
                    <Input.Password placeholder="SMTP授权码" autoComplete="new-password" />
                  </Form.Item>
                </div>

                <Divider orientation="left" plain>发件人信息</Divider>

                <div className="grid grid-cols-2 gap-4">
                  <Form.Item name="from" label="发件人邮箱" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
                    <Input placeholder="发件人邮箱地址" />
                  </Form.Item>
                  <Form.Item name="from_name" label="发件人名称">
                    <Input placeholder="如: EmbyHub" />
                  </Form.Item>
                </div>

                <Form.Item>
                  <Button type="primary" icon={<SaveOutlined />} onClick={handleEmailSave} loading={emailSaving}>
                    保存设置
                  </Button>
                </Form.Item>
              </Form>

              <Divider orientation="left" plain>测试邮件</Divider>
              
              <div className="max-w-2xl">
                <Space.Compact className="w-full">
                  <Input 
                    placeholder="输入接收测试邮件的邮箱" 
                    value={testEmail}
                    onChange={(e) => setTestEmail(e.target.value)}
                    className="flex-1"
                  />
                  <Button 
                    type="primary" 
                    icon={<SendOutlined />} 
                    onClick={handleTest} 
                    loading={testing}
                  >
                    发送测试
                  </Button>
                </Space.Compact>
                <p className="text-gray-400 text-xs mt-2">发送一封测试邮件以验证配置是否正确</p>
              </div>
            </>
          )}
        </div>
      ),
    },
    {
      key: 'register',
      label: (
        <span>
          <UserAddOutlined className="mr-2" />
          注册设置
        </span>
      ),
      children: (
        <div className="p-4">
          {registerLoading ? (
            <div className="text-center py-20">加载中...</div>
          ) : (
            <>
              <Alert
                message="功能说明"
                description={
                  <div className="text-sm">
                    <p>控制用户注册功能，包括是否允许注册、注册赠送会员天数等。</p>
                    <p className="text-orange-500 mt-2">⚠️ 关闭注册后，新用户将无法自行注册账号</p>
                    <p className="text-blue-600 mt-1">💡 注册赠送会员天数：新用户注册成功后自动获得指定天数的会员</p>
                    <p className="text-red-600 mt-1">⚠️ 会员到期后用户将被自动禁用，需使用卡密续费后才能登录</p>
                  </div>
                }
                type="info"
                showIcon
                className="mb-6"
              />

              <Form 
                form={registerForm} 
                layout="vertical" 
                className="max-w-2xl"
                name="registerSettingsForm" 
                initialValues={{ register_enabled: true, gift_member_days: 0, auto_disable_on_exp: true }}
              >
                <Form.Item name="register_enabled" label="允许用户注册" valuePropName="checked">
                  <Switch checkedChildren="开启" unCheckedChildren="关闭" />
                </Form.Item>

                <Form.Item 
                  name="gift_member_days" 
                  label="注册赠送会员天数"
                  extra="设置为0表示不赠送会员，新用户需要使用卡密激活会员后才能登录"
                >
                  <Space.Compact className="w-full">
                    <InputNumber min={0} max={365} placeholder="0" className="flex-1" />
                    <Button disabled>天</Button>
                  </Space.Compact>
                </Form.Item>

                <Form.Item 
                  name="auto_disable_on_exp" 
                  label="会员到期后自动禁用账户" 
                  valuePropName="checked"
                  extra="启用后，会员到期的用户将无法登录，需使用卡密续费后才能重新登录"
                >
                  <Switch checkedChildren="开启" unCheckedChildren="关闭" />
                </Form.Item>

                <Form.Item>
                  <Button type="primary" icon={<SaveOutlined />} onClick={handleRegisterSave} loading={registerSaving}>
                    保存设置
                  </Button>
                </Form.Item>
              </Form>
            </>
          )}
        </div>
      ),
    },
    {
      key: 'emby',
      label: (
        <span>
          <PlayCircleOutlined className="mr-2" />
          媒体服务配置
        </span>
      ),
      children: (
        <div className="p-4">
          {embyLoading ? (
            <div className="text-center py-20">加载中...</div>
          ) : (
            <>
              <Alert
                message="配置说明"
                description={
                  <div className="text-sm">
                    <p>配置 Emby 媒体服务器，用于用户管理和媒体库访问。</p>
                    <p className="text-blue-600 mt-2">💡 <strong>Emby模式</strong>：使用 Emby 官方 API，需要提供 API 密钥</p>
                    <p className="text-green-600 mt-1">💡 <strong>fnOS模式</strong>：使用fnOS影视 API，需要提供管理员账号密码</p>
                    <Collapse 
                      size="small" 
                      ghost
                      items={[
                        {
                          key: '1',
                          label: <span className="text-blue-500"><QuestionCircleOutlined className="mr-1" />如何获取 Emby API 密钥？</span>,
                          children: (
                            <div className="space-y-2 text-gray-600">
                              <div>1. 登录 Emby 管理后台</div>
                              <div>2. 进入 <strong>设置 → API 密钥</strong></div>
                              <div>3. 点击 <strong>新建 API 密钥</strong></div>
                              <div>4. 输入应用名称（如：用户管理系统）</div>
                              <div>5. 复制生成的 API 密钥</div>
                            </div>
                          ),
                        },
                      ]}
                    />
                  </div>
                }
                type="info"
                showIcon
                className="mb-6"
              />

              <Form 
                form={embyForm} 
                layout="vertical" 
                className="max-w-2xl"
                name="embySettingsForm"
                initialValues={{ emby_enabled: false, mode: 'emby', base_url: 'http://localhost:8096' }}
              >
                <Form.Item name="emby_enabled" label="启用媒体服务" valuePropName="checked">
                  <Switch checkedChildren="开启" unCheckedChildren="关闭" />
                </Form.Item>

                <Form.Item name="mode" label="服务模式">
                  <Radio.Group onChange={(e) => setEmbyMode(e.target.value)}>
                    <Radio.Button value="emby">
                      <ApiOutlined className="mr-1" />
                      Emby 官方
                    </Radio.Button>
                    <Radio.Button value="feiniu">
                      <PlayCircleOutlined className="mr-1" />
                      fnOS影视
                    </Radio.Button>
                  </Radio.Group>
                </Form.Item>

                <Divider orientation="left" plain>服务器配置</Divider>

                <Form.Item 
                  name="base_url" 
                  label="服务器地址" 
                  rules={[{ required: true, message: '请输入服务器地址' }]}
                  extra="包含协议和端口，如：http://192.168.1.100:8096"
                >
                  <Input placeholder="http://localhost:8096" />
                </Form.Item>

                {embyMode === 'emby' ? (
                  <Form.Item 
                    name="api_key" 
                    label="API 密钥"
                    extra="从 Emby 管理后台获取的 API 密钥"
                  >
                    <Input.Password placeholder="请输入 API 密钥" autoComplete="new-password" />
                  </Form.Item>
                ) : (
                  <>
                    <Form.Item 
                      name="admin_user" 
                      label="管理员用户名"
                    >
                      <Input placeholder="fnOS影视管理员用户名" />
                    </Form.Item>
                    <Form.Item 
                      name="admin_pass" 
                      label="管理员密码"
                    >
                      <Input.Password placeholder="fnOS影视管理员密码" autoComplete="new-password" />
                    </Form.Item>
                  </>
                )}

                <Divider orientation="left" plain>新用户权限模板</Divider>

                <Form.Item 
                  name="template_user" 
                  label="模板用户"
                  extra="新注册用户将复制此用户的权限配置（如媒体库访问权限等）。留空则使用默认权限。"
                >
                  <Input placeholder="如：test（Emby中已存在的用户名）" />
                </Form.Item>

                <Form.Item>
                  <Space>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handleEmbySave} loading={embySaving}>
                      保存设置
                    </Button>
                    <Button icon={<ApiOutlined />} onClick={handleEmbyTest} loading={embyTesting}>
                      测试连接
                    </Button>
                  </Space>
                </Form.Item>
              </Form>

            </>
          )}
        </div>
      ),
    },
  ]

  return (
    <Card 
      title={
        <span>
          <SettingOutlined className="mr-2" />
          系统设置
        </span>
      }
    >
      <Tabs 
        items={tabItems} 
        activeKey={activeTab}
        onChange={setActiveTab}
        size="large"
        destroyInactiveTabPane
      />
    </Card>
  )
}

export default Settings
