import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { publicApi } from '../services/api'

export interface SiteSettings {
  title: string
  description: string
  keywords: string
  logo: string
  favicon: string
  footer: string
  icp: string
  github_url: string
  telegram_url: string
  qq_url: string
}

const defaultSettings: SiteSettings = {
  title: 'EmbyHub - 用户管理系统',
  description: 'EmbyHub用户管理系统',
  keywords: 'EmbyHub,Emby,媒体服务',
  logo: '/uploads/logo/lightsail-logo.svg',
  favicon: '/uploads/logo/lightsail-logo.svg',
  footer: '© 2025 EmbyHub',
  icp: '',
  github_url: '',
  telegram_url: '',
  qq_url: ''
}

interface SiteSettingsContextType {
  settings: SiteSettings
  loading: boolean
}

const SiteSettingsContext = createContext<SiteSettingsContextType>({
  settings: defaultSettings,
  loading: true
})

export const SiteSettingsProvider = ({ children }: { children: ReactNode }) => {
  const [settings, setSettings] = useState<SiteSettings>(defaultSettings)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const response = await publicApi.getSiteSettings()
        const data = response.data?.data as SiteSettings
        if (data) {
          setSettings(data)
          if (data.title) {
            document.title = data.title
          }
          if (data.description) {
            let metaDesc = document.querySelector('meta[name="description"]')
            if (!metaDesc) {
              metaDesc = document.createElement('meta')
              metaDesc.setAttribute('name', 'description')
              document.head.appendChild(metaDesc)
            }
            metaDesc.setAttribute('content', data.description)
          }
          if (data.keywords) {
            let metaKeywords = document.querySelector('meta[name="keywords"]')
            if (!metaKeywords) {
              metaKeywords = document.createElement('meta')
              metaKeywords.setAttribute('name', 'keywords')
              document.head.appendChild(metaKeywords)
            }
            metaKeywords.setAttribute('content', data.keywords)
          }
          // 更新 favicon - 优先使用 favicon，其次使用 logo
          const faviconUrl = data.favicon || data.logo
          if (faviconUrl) {
            let link = document.querySelector('link[rel="icon"]') as HTMLLinkElement
            if (!link) {
              link = document.createElement('link')
              link.rel = 'icon'
              document.head.appendChild(link)
            }
            link.href = faviconUrl
          }
        }
      } catch (err) {
        console.error('加载网站设置失败:', err)
      } finally {
        setLoading(false)
      }
    }

    loadSettings()
  }, [])

  return (
    <SiteSettingsContext.Provider value={{ settings, loading }}>
      {children}
    </SiteSettingsContext.Provider>
  )
}

export const useSiteSettings = () => useContext(SiteSettingsContext)

export default useSiteSettings
