import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Card, Button, List, Avatar, Badge, Skeleton, Empty } from 'antd';
import { 
  DatabaseOutlined, 
  PlusOutlined,
  FolderOpenOutlined
} from '@ant-design/icons';
import { sshApi } from '../services/api';

const Dashboard = () => {
  const [connections, setConnections] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    // 加载所有SSH连接
    const loadConnections = async () => {
      try {
        const allConnections = await sshApi.getConnections();
        setConnections(allConnections);
      } catch (error) {
        console.error('加载连接失败:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadConnections();
  }, []);

  const handleAddConnection = () => {
    navigate('/connections');
  };

  return (
    <div style={{ padding: 24 }}>
      {/* 页面标题和添加连接按钮 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 style={{ fontSize: 28, fontWeight: 'bold', margin: 0 }}>服务器连接</h1>
        <Button 
          type="primary" 
          icon={<PlusOutlined />} 
          onClick={handleAddConnection}
        >
          添加SSH连接
        </Button>
      </div>

      {/* 服务器连接列表 */}
      <Card>
        {isLoading ? (
          <Skeleton active paragraph={{ rows: 5 }} />
        ) : connections.length > 0 ? (
          <List
            dataSource={connections}
            renderItem={(item) => (
              <List.Item
                actions={[
                  <Button 
                    key="open" 
                    size="small" 
                    type="primary"
                    icon={<FolderOpenOutlined />}
                    onClick={() => navigate(`/files/${item.id}/`)}
                  >
                    管理文件
                  </Button>
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <Badge status={item.is_active ? 'success' : 'default'}>
                      <Avatar icon={<DatabaseOutlined />} />
                    </Badge>
                  }
                  title={<span style={{ fontWeight: 'bold', fontSize: 16 }}>{item.name}</span>}
                  description={`${item.username}@${item.host}:${item.port}`}
                />
              </List.Item>
            )}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              <div>
                <h3 style={{ marginBottom: 16 }}>暂无服务器连接</h3>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />} 
                  onClick={handleAddConnection}
                >
                  添加第一个SSH连接
                </Button>
              </div>
            }
          />
        )}
      </Card>
    </div>
  );
};

export default Dashboard;
