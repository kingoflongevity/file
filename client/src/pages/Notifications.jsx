import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  List,
  Card,
  Progress,
  Tag,
  Empty,
  Button,
  Spin,
  Space,
} from 'antd';
import {
  DownloadOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  LeftOutlined,
  DeleteOutlined,
  FolderOutlined,
} from '@ant-design/icons';
import { taskApi } from '../services/api';
import { isElectron, getIpcRenderer } from '../utils/electron';

const Notifications = () => {
  const navigate = useNavigate();
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);

  // 获取任务列表
  const fetchTasks = async () => {
    try {
      setLoading(true);
      const response = await taskApi.getTasks();
      if (Array.isArray(response)) {
        setTasks(response);
      } else if (response && response.tasks) {
        setTasks(response.tasks);
      }
    } catch (error) {
      console.error('获取任务列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载时获取任务列表
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

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Button
            type="text"
            icon={<LeftOutlined />}
            onClick={() => navigate(-1)}
            style={{ marginRight: 16 }}
          >
            返回
          </Button>
          <h1 style={{ margin: 0, fontSize: 24 }}>任务通知</h1>
        </div>
        <Button 
          type="primary" 
          onClick={fetchTasks} 
          loading={loading}
        >
          刷新
        </Button>
      </div>

      <Spin spinning={loading}>
        {tasks.length === 0 ? (
          <Empty description="暂无任务" />
        ) : (
          <List
            dataSource={tasks}
            renderItem={(task) => (
              <Card 
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
                            const ipcRenderer = getIpcRenderer();
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
            )}
          />
        )}
      </Spin>
    </div>
  );
};

export default Notifications;