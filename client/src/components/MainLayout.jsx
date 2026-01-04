import React, { useState, useEffect } from 'react';
import { Link, Outlet, useNavigate } from 'react-router-dom';
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
} from '@ant-design/icons';
import { taskApi } from '../services/api';

const { Header, Content } = Layout;

const MainLayout = () => {
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);

  // 获取当前用户信息
  const user = JSON.parse(localStorage.getItem('user'));

  // 处理登出
  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
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

  // 用户菜单 - 包含导航和设置选项
  const userMenu = [
    {
      key: 'dashboard',
      label: (
        <div className="flex items-center" onClick={() => navigate('/')}>
          <DashboardOutlined className="mr-2" />
          <span>仪表盘</span>
        </div>
      ),
    },
    {
      key: 'connections',
      label: (
        <div className="flex items-center" onClick={() => navigate('/connections')}>
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
    {
      key: 'theme-toggle',
      label: (
        <div className="flex items-center" onClick={() => window.themeContext?.toggleTheme?.()}>
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
          <span onClick={handleLogout}>退出登录</span>
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
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        {/* 页面标题 */}
        <div style={{ fontSize: '18px', fontWeight: '600', color: token.colorPrimary }}>
          远程连接文件管理
        </div>

        <Space>
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
                                        <Button 
                                          type="link" 
                                          size="small" 
                                          icon={<DownloadOutlined />}
                                          onClick={() => {
                                            const token = localStorage.getItem('token');
                                            window.open(`http://localhost:8082/api/tasks/${task.id}/download?token=${token}`, '_blank');
                                          }}
                                        >
                                          下载
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
                  style={{ fontSize: '18px' }}
                />
              </Badge>
            </Dropdown>
          </Tooltip>

          {/* 用户头像下拉菜单 */}
          <Dropdown menu={{ items: userMenu }} placement="bottomRight">
            <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
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
  );
};

export default MainLayout;