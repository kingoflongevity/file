import React from 'react';
import { Link, Outlet, useNavigate } from 'react-router-dom';
import {
  Layout,
  Menu,
  Avatar,
  Button,
  Dropdown,
  Space,
  Tooltip,
  theme,
} from 'antd';
import {
  DashboardOutlined,
  DatabaseOutlined,
  LogoutOutlined,
  BellOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';

const { Header, Content, Sider } = Layout;

const MainLayout = () => {
  const navigate = useNavigate();
  const { token } = theme.useToken();

  // 获取当前用户信息
  const user = JSON.parse(localStorage.getItem('user'));

  // 处理登出
  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  // 用户菜单
  const userMenu = [
    {
      key: '1',
      label: (
        <div className="flex items-center">
          <UserOutlined className="mr-2" />
          <span>个人中心</span>
        </div>
      ),
    },
    {
      key: '2',
      label: (
        <div className="flex items-center">
          <SettingOutlined className="mr-2" />
          <span>设置</span>
        </div>
      ),
    },
    {
      type: 'divider',
    },
    {
      key: '3',
      label: (
        <div className="flex items-center text-red-600">
          <LogoutOutlined className="mr-2" />
          <span onClick={handleLogout}>退出登录</span>
        </div>
      ),
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 侧边栏 */}
      <Sider
        theme="light"
        width={240}
        style={{
          boxShadow: '2px 0 8px rgba(0, 21, 41, 0.1)',
        }}
      >
        {/* 侧边栏标题 */}
        <div 
          style={{
            padding: '16px 24px',
            fontSize: '18px',
            fontWeight: '600',
            color: token.colorPrimary,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <DatabaseOutlined className="mr-2" />
          SSH 文件管理器
        </div>

        {/* 导航菜单 */}
        <Menu
          mode="inline"
          defaultSelectedKeys={['1']}
          style={{ borderRight: 0, height: '100%' }}
          items={[
            {
              key: '1',
              icon: <DashboardOutlined />,
              label: <Link to="/">仪表盘</Link>,
            },
            {
              key: '2',
              icon: <DatabaseOutlined />,
              label: <Link to="/connections">SSH连接</Link>,
            },
          ]}
        />
      </Sider>

      {/* 主内容区 */}
      <Layout>
        {/* 顶部导航栏 */}
        <Header
          style={{
            background: token.colorBgContainer,
            boxShadow: '0 2px 8px rgba(0, 21, 41, 0.08)',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div style={{ fontSize: '18px', fontWeight: '600' }}>
            <Outlet />
          </div>

          <Space>
            {/* 通知按钮 */}
            <Tooltip title="通知">
              <Button
                type="text"
                icon={<BellOutlined />}
                style={{ fontSize: '18px' }}
              />
            </Tooltip>

            {/* 用户头像下拉菜单 */}
            <Dropdown menu={{ items: userMenu }} placement="bottomRight">
              <Space>
                <Avatar
                  size="default"
                  style={{
                    backgroundColor: token.colorPrimary,
                    marginRight: 8,
                  }}
                >
                  {user?.username?.charAt(0).toUpperCase()}
                </Avatar>
                <span style={{ fontWeight: '500' }}>{user?.username}</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>

        {/* 页面内容 */}
        <Content
          style={{
            padding: 24,
            margin: 0,
            minHeight: 280,
            background: token.colorBgContainer,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
