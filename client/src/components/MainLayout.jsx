import React, { useState, useEffect } from 'react';
import { Link, Outlet, useNavigate } from 'react-router-dom';
import { isElectron } from '../utils/electron';
import {
  Layout,
  Avatar,
  Button,
  Dropdown,
  Space,
  Tooltip,
  theme,
  List,
  Card,
  Progress,
  Tag,
  Empty,
  Badge,
} from 'antd';
import {
  DashboardOutlined,
  DatabaseOutlined,
  LogoutOutlined,
  BellOutlined,
  SettingOutlined,
  UserOutlined,
  DownloadOutlined,
  CloseOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  DeleteOutlined,
  FolderOutlined,
} from '@ant-design/icons';
import { taskApi } from '../services/api';

const { Header, Content } = Layout;

const MainLayout = () => {
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [user, setUser] = useState(null);
  const [siteName, setSiteName] = useState('远程连接文件管理');

  // 获取当前用户信息
  useEffect(() => {
    const fetchUser = () => {
      const userData = localStorage.getItem('user');
      if (userData) {
        const parsedUser = JSON.parse(userData);
        console.log('获取到用户信息:', parsedUser);
        console.log('用户角色:', parsedUser.role);
        setUser(parsedUser);
      }
    };

    // 初始获取
    fetchUser();

    // 监听storage变化，确保用户信息实时更新
    window.addEventListener('storage', fetchUser);

    return () => {
      window.removeEventListener('storage', fetchUser);
    };
  }, []);

  // 获取网站名称并监听更新
  useEffect(() => {
    // 从localStorage获取保存的网站名称
    const savedSiteName = localStorage.getItem('siteName');
    if (savedSiteName) {
      setSiteName(savedSiteName);
    }

    // 监听网站名称更新事件
    const handleSiteNameUpdate = (event) => {
      setSiteName(event.detail);
    };

    // 监听自定义事件
    window.addEventListener('siteNameUpdated', handleSiteNameUpdate);

    return () => {
      window.removeEventListener('siteNameUpdated', handleSiteNameUpdate);
    };
  }, []);

  // 处理登出
  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
    navigate('/login');
  };

  // 获取任务列表
  const fetchTasks = async () => {
    try {
      setLoading(true);
      const response = await taskApi.getTasks();
      // 检查response是否为数组（后端直接返回数组）
      if (Array.isArray(response)) {
        setTasks(response);
      } else if (response && response.tasks) {
        // 兼容可能的对象格式
        setTasks(response.tasks);
      }
    } catch (error) {
      console.error('获取任务列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载时获取任务列表并设置定时刷新
  useEffect(() => {
    fetchTasks();
    // 每2秒刷新一次任务状态
    const interval = setInterval(fetchTasks, 2000);
    return () => clearInterval(interval);
  }, []);

  // 格式化任务状态
  const getTaskStatusText = (status) => {
    switch (status) {
      case 'pending':
        return '等待中';
      case 'running':
        return '运行中';
      case 'completed':
        return '已完成';
      case 'failed':
        return '失败';
      default:
        return status;
    }
  };

  // 获取任务状态图标
  const getTaskStatusIcon = (status) => {
    switch (status) {
      case 'pending':
        return <ClockCircleOutlined />;
      case 'running':
        return <ClockCircleOutlined spin />;
      case 'completed':
        return <CheckCircleOutlined />;
      case 'failed':
        return <ExclamationCircleOutlined />;
      default:
        return null;
    }
  };

  // 获取任务状态标签颜色
  const getTaskStatusColor = (status) => {
    switch (status) {
      case 'pending':
        return 'default';
      case 'running':
        return 'processing';
      case 'completed':
        return 'success';
      case 'failed':
        return 'error';
      default:
        return 'default';
    }
  };

  // 处理菜单点击事件
  const handleMenuClick = ({ key }) => {
    switch (key) {
      case 'dashboard':
        navigate('/');
        break;
      case 'connections':
        navigate('/connections');
        break;
      case 'profile':
        navigate('/profile');
        break;
      case 'settings':
        navigate('/settings');
        break;
      case 'users':
        navigate('/users');
        break;
      case 'theme-toggle':
        window.themeContext?.toggleTheme?.();
        break;
      case 'logout':
        handleLogout();
        break;
      default:
        break;
    }
  };

  // 用户菜单 - 包含导航和设置选项
  const userMenu = [
    {
      key: 'dashboard',
      label: (
        <div className="flex items-center">
          <DashboardOutlined className="mr-2" />
          <span>仪表盘</span>
        </div>
      ),
    },
    {
      key: 'connections',
      label: (
        <div className="flex items-center">
          <DatabaseOutlined className="mr-2" />
          <span>连接管理</span>
        </div>
      ),
    },
    {
      type: 'divider',
    },
    {
      key: 'profile',
      label: (
        <div className="flex items-center">
          <UserOutlined className="mr-2" />
          <span>个人中心</span>
        </div>
      ),
    },
    {
      key: 'settings',
      label: (
        <div className="flex items-center">
          <SettingOutlined className="mr-2" />
          <span>设置</span>
        </div>
      ),
    },
    // 管理员专用菜单
    {
      key: 'users',
      label: (
        <div className="flex items-center">
          <UserOutlined className="mr-2" />
          <span>用户管理</span>
        </div>
      ),
    },
    {
      key: 'theme-toggle',
      label: (
        <div className="flex items-center">
          <div className="mr-2">
            {window.themeContext?.isDarkTheme ? (
              <span>🌞</span>
            ) : (
              <span>🌙</span>
            )}
          </div>
          <span>{window.themeContext?.isDarkTheme ? '切换至亮色主题' : '切换至暗色主题'}</span>
        </div>
      ),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      label: (
        <div className="flex items-center text-red-600">
          <LogoutOutlined className="mr-2" />
          <span>退出登录</span>
        </div>
      ),
    },
  ];

  //{/* 获取未完成任务数量 */}
  const getUnfinishedTasksCount = () => {
    return tasks.filter(task => task.status === 'pending' || task.status === 'running').length;
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 顶部导航栏 */}
      <Header
        style={{
          background: token.colorBgContainer,
          boxShadow: '0 2px 8px rgba(0, 21, 41, 0.08)',
          padding: '0 12px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          position: 'relative',
          zIndex: 1000,
          flexWrap: 'wrap',
          minHeight: 64,
        }}
      >
        {/* 页面标题 */}
        <div style={{ fontSize: '16px', fontWeight: '600', color: token.colorPrimary, whiteSpace: 'nowrap' }}>
          {siteName}
        </div>

        <Space size="middle">
          {/* 通知按钮 */}
          <Tooltip title="通知">
            <Dropdown
              menu={{
                items: [
                  {
                    label: (
                      <div style={{ padding: 16 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                          <h3 style={{ margin: 0, fontSize: 16 }}>任务通知</h3>
                          <Space>
                            <Button 
                              type="link" 
                              size="small" 
                              onClick={fetchTasks} 
                              loading={loading}
                            >
                              刷新
                            </Button>
                            <Button 
                              type="link" 
                              size="small" 
                              onClick={() => navigate('/notifications')}
                            >
                              查看全部
                            </Button>
                          </Space>
                        </div>
                        
                        {tasks.length === 0 ? (
                          <Empty description="暂无任务" />
                        ) : (
                          <div style={{ maxHeight: 300, overflowY: 'auto' }}>
                            {tasks.slice(0, 5).map((task) => (
                              <Card 
                                key={task.id}
                                size="small" 
                                style={{ marginBottom: 12, borderRadius: 8 }}
                              >
                                <div style={{ marginBottom: 8 }}>
                                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                                      <div style={{ display: 'flex', alignItems: 'center' }}>
                                        {getTaskStatusIcon(task.status)}
                                        <span style={{ marginLeft: 8, fontWeight: '500' }}>{task.fileName || (task.type === 'zip' ? '压缩任务' : '下载任务')}</span>
                                      </div>
                                      <Tag color={getTaskStatusColor(task.status)}>
                                        {getTaskStatusText(task.status)}
                                      </Tag>
                                    </div>
                                    <div style={{ fontSize: 12, color: '#666', marginBottom: 8 }}>
                                      {task.path}
                                    </div>
                                  <Progress 
                                    percent={task.progress || 0} 
                                    size="small" 
                                    style={{ marginBottom: 8 }}
                                  />
                                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, color: '#999' }}>
                                    <span>{task.createdAt ? new Date(task.createdAt).toLocaleString() : ''}</span>
                                    <Space size="small">
                                      {task.status === 'completed' && (
                                        <span style={{ fontSize: 12, color: '#666' }}>
                                          文件已下载到默认目录: {task.downloadPath || '默认下载目录'}
                                        </span>
                                      )}
                                      {task.status === 'completed' && isElectron() && (
                                        <Button 
                                          type="link" 
                                          size="small" 
                                          icon={<FolderOutlined />}
                                          onClick={async () => {
                                            const ipcRenderer = window.require('electron').ipcRenderer;
                                            if (ipcRenderer) {
                                              await ipcRenderer.invoke('open-download-dir');
                                            }
                                          }}
                                        >
                                          打开文件所在目录
                                        </Button>
                                      )}
                                      <Button 
                                        type="link" 
                                        size="small" 
                                        icon={<DeleteOutlined />}
                                        onClick={async () => {
                                          try {
                                            await taskApi.deleteTask(task.id);
                                            // 刷新任务列表
                                            fetchTasks();
                                          } catch (error) {
                                            console.error('删除任务失败:', error);
                                          }
                                        }}
                                      >
                                        删除
                                      </Button>
                                    </Space>
                                  </div>
                                </div>
                              </Card>
                            ))}
                            {tasks.length > 5 && (
                              <div style={{ textAlign: 'center', marginTop: 8 }}>
                                <Button 
                                  type="link" 
                                  size="small" 
                                  onClick={() => navigate('/notifications')}
                                >
                                  查看全部 {tasks.length} 条
                                </Button>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    ),
                  },
                ],
              }}
              placement="bottomRight" 
              trigger={['click']}
            >
              <Badge count={getUnfinishedTasksCount() > 0 ? getUnfinishedTasksCount() : 0} offset={[0, -5]}>
                <Button
                  type="text"
                  icon={<BellOutlined />}
                  style={{ fontSize: '18px', cursor: 'pointer' }}
                />
              </Badge>
            </Dropdown>
          </Tooltip>

          {/* 用户头像下拉菜单 */}
          <Dropdown 
            menu={{ 
              items: userMenu, 
              onClick: handleMenuClick 
            }} 
            placement="bottomRight"
            trigger={['click']}
          >
            <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer', padding: '4px 8px', borderRadius: 4, transition: 'background-color 0.2s' }} onMouseEnter={(e) => e.currentTarget.style.backgroundColor = token.colorBgElevated} onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}>
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
            </div>
          </Dropdown>
        </Space>
      </Header>

      {/* 页面内容 */}
      <Content
        className="main-content"
        style={{
          padding: '24px',
          margin: 0,
          minHeight: 280,
          background: token.colorBgContainer,
          transition: 'padding 0.3s ease-in-out',
        }}
      >
        <div className="page-transition">
          <Outlet />
        </div>
      </Content>
    </Layout>
  );
};

export default MainLayout;