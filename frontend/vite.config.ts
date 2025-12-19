import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 54681,
    host: '0.0.0.0', // 监听所有网络接口
    strictPort: false,
    // 开发环境：允许所有域名（方便开发，Go后端仍然会验证API）
    // 生产环境：建议使用Nginx在网络层控制域名访问
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://localhost:54680',
        changeOrigin: false, // 保持原始Host头
        // 转发原始Host到后端进行验证
        configure: (proxy, _options) => {
          proxy.on('proxyReq', (proxyReq, req) => {
            if (req.headers.host) {
              proxyReq.setHeader('X-Forwarded-Host', req.headers.host);
            }
          });
        },
      },
      '/uploads': {
        target: 'http://localhost:54680',
        changeOrigin: true,
      },
      '/static': {
        target: 'http://localhost:54680',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': '/src',
    },
  },
  build: {
    // antd 库本身约 1.2MB，调整警告阈值
    chunkSizeWarningLimit: 1200,
    rollupOptions: {
      output: {
        manualChunks: {
          // React 核心
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          // Ant Design
          'vendor-antd': ['antd', '@ant-design/icons'],
          // 图表库
          'vendor-recharts': ['recharts'],
          // 工具库
          'vendor-utils': ['axios', 'dayjs', '@tanstack/react-query'],
        },
      },
    },
  },
})
