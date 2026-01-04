import './App.css';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import Routes from './routes';

function App() {
  return (
    <ConfigProvider 
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: '#1890ff',
          borderRadius: 6,
        },
      }}
    >
      <Routes />
    </ConfigProvider>
  );
}

export default App;
