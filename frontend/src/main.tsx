import React from 'react'
import ReactDOM from 'react-dom/client'
import { Provider } from 'react-redux'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, theme, App as AntdApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import App from './App'
import { store } from './store'
import { ThemeProvider } from './theme/ThemeContext'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 10 * 1000, // 10秒后数据过期
      refetchOnWindowFocus: true, // 窗口聚焦时刷新
    },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <ConfigProvider 
            locale={zhCN}
            theme={{
              algorithm: theme.defaultAlgorithm,
              cssVar: true,
              token: {
                colorBgContainer: 'rgba(255, 255, 255, 0.9)',
                colorBgElevated: 'rgba(255, 255, 255, 0.95)',
                colorText: 'rgba(0, 0, 0, 0.88)',
                colorTextSecondary: 'rgba(0, 0, 0, 0.65)',
              },
            }}
          >
            <AntdApp>
              <App />
            </AntdApp>
          </ConfigProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </Provider>
  </React.StrictMode>,
)
