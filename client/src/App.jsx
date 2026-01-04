import './App.css';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useState, useEffect } from 'react';
import Routes from './routes';

function App() {
  // 从本地存储获取主题偏好，默认亮色主题
  const [isDarkTheme, setIsDarkTheme] = useState(() => {
    const savedTheme = localStorage.getItem('theme');
    return savedTheme === 'dark';
  });

  // 监听主题变化，更新本地存储
  useEffect(() => {
    localStorage.setItem('theme', isDarkTheme ? 'dark' : 'light');
  }, [isDarkTheme]);

  // 切换主题
  const toggleTheme = () => {
    setIsDarkTheme(!isDarkTheme);
  };

  // 主题配置
  const themeConfig = {
    algorithm: isDarkTheme ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 6,
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
