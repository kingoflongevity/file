import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Card, Statistic, Row, Col, Button, List, Avatar, Badge, Skeleton, Divider } from 'antd';
import { 
  DatabaseOutlined, 
  CheckCircleOutlined, 
  FileOutlined, 
  PlusOutlined, 
  FolderOpenOutlined, 
  UploadOutlined, 
  EditOutlined 
} from '@ant-design/icons';
import { sshApi } from '../services/api';

const Dashboard = () => {
  const [stats, setStats] = useState({
    totalConnections: 0,
    activeConnections: 0,
    recentConnections: [],
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // 加载统计信息
    const loadStats = async () => {
      try {
        const connections = await sshApi.getConnections();
        
        // 计算统计信息
        const totalConnections = connections.length;
        const activeConnections = connections.filter(conn => conn.is_active).length;
        
        // 获取最近使用的连接（按最后使用时间排序）
        const recentConnections = [...connections]
          .sort((a, b) => new Date(b.last_used) - new Date(a.last_used))
          .slice(0, 3);
        
        setStats({
          totalConnections,
          activeConnections,
          recentConnections,
        });
      } catch (error) {
        console.error('加载统计信息失败:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadStats();
  }, []);

  const quickStartItems = [
    {
      icon: <PlusOutlined />,
      title: '添加SSH连接',
      description: '创建新的SSH连接配置，连接到远程服务器',
      color: '#1890ff'
    },
    {
      icon: <FolderOpenOutlined />,
      title: '浏览文件系统',
      description: '通过SSH连接浏览和管理远程服务器上的文件',
      color: '#52c41a'
    },
    {
      icon: <UploadOutlined />,
      title: '上传和下载文件',
      description: '轻松上传本地文件到服务器或从服务器下载文件',
      color: '#722ed1'
    },
    {
      icon: <EditOutlined />,
      title: '管理文件',
      description: '创建、删除、重命名、移动和复制文件和文件夹',
      color: '#fa8c16'
    }
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: 24 }}>
        <h1 style={{ fontSize: 28, fontWeight: 'bold', marginBottom: 8 }}>仪表盘</h1>
        <p style={{ color: '#666' }}>欢迎使用文件管理系统</p>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {/* 总连接数 */}
        <Col xs={24} sm={12} md={8}>
          <Card hoverable>
            <Statistic
              title="总SSH连接"
              value={stats.totalConnections}
              prefix={<DatabaseOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff', fontSize: 32 }}
              footer={
                <Link to="/connections">
                  <Button type="link" size="small">查看所有连接</Button>
                </Link>
              }
            />
          </Card>
        </Col>

        {/* 活跃连接数 */}
        <Col xs={24} sm={12} md={8}>
          <Card hoverable>
            <Statistic
              title="活跃连接"
              value={stats.activeConnections}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a', fontSize: 32 }}
              footer={
                <Link to="/connections">
                  <Button type="link" size="small">管理连接</Button>
                </Link>
              }
            />
          </Card>
        </Col>

        {/* 功能卡片 */}
        <Col xs={24} sm={12} md={8}>
          <Card hoverable>
            <Statistic
              title="快速访问"
              value="文件管理"
              prefix={<FileOutlined style={{ color: '#722ed1' }} />}
              valueStyle={{ fontSize: 20, fontWeight: 'bold' }}
              footer={
                <Link to="/connections">
                  <Button type="link" size="small">开始使用</Button>
                </Link>
              }
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 最近连接 */}
        <Col xs={24} lg={12}>
          <Card 
            title="最近使用的连接"
            extra={
              <Link to="/connections">
                <Button type="link" size="small">查看全部</Button>
              </Link>
            }
          >
            {isLoading ? (
              <Skeleton active paragraph={{ rows: 3 }} />
            ) : stats.recentConnections.length > 0 ? (
              <List
                dataSource={stats.recentConnections}
                renderItem={(item) => (
                  <List.Item
                    actions={[
                      <span key="time" style={{ fontSize: 12, color: '#999' }}>
                        {new Date(item.last_used).toLocaleString()}
                      </span>,
                      <Link key="open" to={`/files/${item.id}/`}>
                        <Button size="small" type="primary">打开</Button>
                      </Link>
                    ]}
                  >
                    <List.Item.Meta
                      avatar={
                        <Badge status={item.is_active ? 'success' : 'default'}>
                          <Avatar icon={<DatabaseOutlined />} />
                        </Badge>
                      }
                      title={<a href={`/files/${item.id}/`}>{item.name}</a>}
                      description={`${item.username}@${item.host}:${item.port}`}
                    />
                  </List.Item>
                )}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: 24, color: '#999' }}>
                暂无连接记录
              </div>
            )}
          </Card>
        </Col>

        {/* 快速开始 */}
        <Col xs={24} lg={12}>
          <Card title="快速开始">
            <List
              grid={{ gutter: 16, column: 1 }}
              dataSource={quickStartItems}
              renderItem={(item) => (
                <List.Item>
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    <Avatar 
                      icon={item.icon} 
                      style={{ 
                        backgroundColor: item.color, 
                        marginRight: 16, 
                        fontSize: 20 
                      }} 
                    />
                    <div>
                      <h3 style={{ margin: 0, fontSize: 16, fontWeight: 'bold' }}>{item.title}</h3>
                      <p style={{ margin: 4, fontSize: 14, color: '#666' }}>{item.description}</p>
                    </div>
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
