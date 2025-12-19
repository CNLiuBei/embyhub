// 域名白名单检查脚本
// 在页面加载前执行，阻止未授权域名访问

(function() {
  // 获取当前域名
  const currentHost = window.location.hostname;
  
  // 域名白名单检查
  async function checkDomain() {
    try {
      // 调用后端API检查域名
      const response = await fetch('/api/v1/domain-check', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ domain: currentHost })
      });
      
      // 如果响应失败，允许访问（避免因网络问题导致无法访问）
      if (!response.ok && response.status !== 403) {
        console.warn('Domain check failed, allowing access');
        return;
      }
      
      const data = await response.json();
      
      // 只有明确返回不允许时才阻止访问
      if (response.status === 403 || (data.code === 403)) {
        // 域名未授权，显示错误页面
        document.body.innerHTML = `
          <div style="
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            margin: 0;
            padding: 20px;
            text-align: center;
          ">
            <div style="
              background: rgba(255, 255, 255, 0.1);
              backdrop-filter: blur(10px);
              border-radius: 20px;
              padding: 40px 60px;
              box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
              max-width: 500px;
            ">
              <div style="font-size: 64px; margin-bottom: 20px;">🚫</div>
              <h1 style="margin: 0 0 10px 0; font-size: 32px;">域名未授权</h1>
              <p style="margin: 0 0 20px 0; font-size: 18px; opacity: 0.9;">
                Domain Not Authorized
              </p>
              <div style="
                background: rgba(255, 255, 255, 0.2);
                border-radius: 10px;
                padding: 15px;
                margin-bottom: 20px;
                font-family: 'Courier New', monospace;
              ">
                <strong>${currentHost}</strong>
              </div>
              <p style="margin: 0; font-size: 14px; opacity: 0.8;">
                此域名未在系统白名单中<br/>
                请联系系统管理员添加授权
              </p>
            </div>
          </div>
        `;
        
        // 阻止页面继续加载
        throw new Error('Domain not authorized');
      }
      
      // 如果允许访问，正常继续
      console.log('Domain check passed for:', currentHost);
      
    } catch (error) {
      if (error.message === 'Domain not authorized') {
        // 这是我们主动抛出的错误，不要继续
        throw error;
      }
      // 其他错误（如网络错误），允许继续访问
      console.warn('Domain check error, allowing access:', error);
    }
  }
  
  // 在DOM加载后执行检查
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', checkDomain);
  } else {
    checkDomain();
  }
})();
