import './App.css';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useState, useEffect } from 'react';
import Routes from './routes';
import { configApi } from './services/api';

function App() {
  // 从本地存储获取主题偏好，默认暗色主题
  const [isDarkTheme, setIsDarkTheme] = useState(() => {
    const savedTheme = localStorage.getItem('theme');
    return savedTheme !== 'light';
  });
  
  // 网站名称状态
  const [siteName, setSiteName] = useState('远程连接文件管理');

  // 监听主题变化，更新本地存储
  useEffect(() => {
    localStorage.setItem('theme', isDarkTheme ? 'dark' : 'light');
  }, [isDarkTheme]);
  
  // 获取网站名称并监听更新
  useEffect(() => {
    // 获取网站名称
    const fetchSiteName = async () => {
      try {
        const response = await configApi.getSiteName();
        if (response && response.siteName) {
          setSiteName(response.siteName);
          localStorage.setItem('siteName', response.siteName);
          document.title = response.siteName;
        }
      } catch (error) {
        console.error('获取网站名称失败:', error);
        // 从localStorage获取，如果没有则使用默认值
        const savedSiteName = localStorage.getItem('siteName') || '远程连接文件管理';
        setSiteName(savedSiteName);
        document.title = savedSiteName;
      }
    };
    
    fetchSiteName();
    
    // 监听网站名称更新事件
    const handleSiteNameUpdate = (event) => {
      setSiteName(event.detail);
      document.title = event.detail;
    };
    
    // 监听自定义事件
    window.addEventListener('siteNameUpdated', handleSiteNameUpdate);
    
    return () => {
      window.removeEventListener('siteNameUpdated', handleSiteNameUpdate);
    };
  }, []);

  // 切换主题
  const toggleTheme = () => {
    setIsDarkTheme(!isDarkTheme);
  };

  // 主题配置 - 与登录页面设计风格保持一致
  const themeConfig = {
    algorithm: isDarkTheme ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      // 主色调 - 与登录页面一致
      colorPrimary: '#3b82f6',
      colorPrimaryHover: '#2563eb',
      colorPrimaryActive: '#1d4ed8',
      
      // 圆角 - 现代化设计
      borderRadius: 8,
      borderRadiusLG: 12,
      borderRadiusSM: 6,
      
      // 背景色 - 优化白色主题
      colorBgContainer: isDarkTheme ? '#1e293b' : '#ffffff',
      colorBgElevated: isDarkTheme ? '#334155' : '#f8fafc',
      colorBgLayout: isDarkTheme ? '#0f172a' : '#f1f5f9',
      colorBgSecondary: isDarkTheme ? '#334155' : '#f8fafc',
      
      // 文字颜色 - 增强可读性
      colorText: isDarkTheme ? '#e2e8f0' : '#1e293b',
      colorTextSecondary: isDarkTheme ? '#94a3b8' : '#64748b',
      colorTextTertiary: isDarkTheme ? '#64748b' : '#94a3b8',
      
      // 边框颜色 - 更柔和的视觉效果
      colorBorder: isDarkTheme ? '#334155' : '#e2e8f0',
      colorBorderSecondary: isDarkTheme ? '#334155' : '#cbd5e1',
      
      // 阴影 - 更现代化的阴影效果
      boxShadow: isDarkTheme ? '0 4px 12px rgba(0, 0, 0, 0.3)' : '0 4px 12px rgba(0, 0, 0, 0.08)',
      boxShadowSecondary: isDarkTheme ? '0 2px 8px rgba(0, 0, 0, 0.2)' : '0 2px 8px rgba(0, 0, 0, 0.05)',
      
      // 字体大小 - 优化可读性
      fontSize: 14,
      fontSizeHeading1: 24,
      fontSizeHeading2: 20,
      fontSizeHeading3: 16,
    },
  };

  // 导出主题相关状态和方法，供子组件使用
  window.themeContext = {
    isDarkTheme,
    toggleTheme,
  };

  return (
    <ConfigProvider 
      locale={zhCN}
      theme={themeConfig}
    >
      <Routes />
    </ConfigProvider>
  );
}

export default App;
