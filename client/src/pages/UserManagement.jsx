import React, { useState, useEffect } from 'react';
import { Card, Button, Table, Form, Input, Modal, message, Space, Tag, Switch, Select, Checkbox, TreeSelect } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SaveOutlined, LeftOutlined, KeyOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { userApi, connectionApi } from '../services/api';

const UserManagement = () => {
  const navigate = useNavigate();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isEditMode, setIsEditMode] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);
  // 权限管理相关状态
  const [isPermissionModalVisible, setIsPermissionModalVisible] = useState(false);
  const [permissions, setPermissions] = useState([]);
  const [connections, setConnections] = useState([]);
  const [selectedConnectionIds, setSelectedConnectionIds] = useState([]);

  // 获取用户列表
  const fetchUsers = async () => {
    try {
      setLoading(true);
      const response = await userApi.getUsers();
      setUsers(response.users || []);
    } catch (error) {
      message.error('获取用户列表失败');
      console.error('获取用户列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载时获取用户列表
  useEffect(() => {
    fetchUsers();
  }, []);

  // 打开创建用户模态框
  const showCreateModal = () => {
    setIsEditMode(false);
    setCurrentUser(null);
    form.resetFields();
    setIsModalVisible(true);
  };

  // 打开编辑用户模态框
  const showEditModal = (user) => {
    setIsEditMode(true);
    setCurrentUser(user);
    form.setFieldsValue({
      username: user.username,
      email: user.email,
      role: user.role,
      active: user.active,
    });
    setIsModalVisible(true);
  };

  // 关闭模态框
  const handleCancel = () => {
    setIsModalVisible(false);
  };

  // 保存用户
  const handleSave = async (values) => {
    try {
      setLoading(true);
      if (isEditMode && currentUser) {
        // 更新用户
        await userApi.updateUser(currentUser.id, values);
        message.success('用户更新成功');
      } else {
        // 创建用户
        await userApi.createUser(values);
        message.success('用户创建成功');
      }
      setIsModalVisible(false);
      fetchUsers();
    } catch (error) {
      message.error(isEditMode ? '用户更新失败' : '用户创建失败');
      console.error('保存用户失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 删除用户
  const handleDelete = (userId, username) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除用户 "${username}" 吗？`,
      onOk: async () => {
        try {
          setLoading(true);
          await userApi.deleteUser(userId);
          message.success('用户删除成功');
          fetchUsers();
        } catch (error) {
          message.error('用户删除失败');
          console.error('删除用户失败:', error);
        } finally {
          setLoading(false);
        }
      },
    });
  };

  // 切换用户状态
  const handleToggleActive = async (user) => {
    try {
      setLoading(true);
      await userApi.updateUser(user.id, {
        active: !user.active,
      });
      message.success('用户状态更新成功');
      fetchUsers();
    } catch (error) {
      message.error('用户状态更新失败');
      console.error('更新用户状态失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 获取所有连接
  const fetchConnections = async () => {
    try {
      const response = await connectionApi.getConnections();
      setConnections(response || []);
    } catch (error) {
      message.error('获取连接列表失败');
      console.error('获取连接列表失败:', error);
    }
  };

  // 获取用户权限
  const fetchUserPermissions = async (userId) => {
    try {
      const response = await userApi.getUserPermissions(userId);
      return response.permissions || [];
    } catch (error) {
      message.error('获取用户权限失败');
      console.error('获取用户权限失败:', error);
      return [];
    }
  };

  // 打开权限管理模态框
  const showPermissionModal = async (user) => {
    setCurrentUser(user);
    // 获取所有连接
    await fetchConnections();
    // 获取用户当前权限
    const userPermissions = await fetchUserPermissions(user.id);
    setPermissions(userPermissions);
    // 提取已授权的连接ID
    const authorizedIds = userPermissions.map(perm => perm.connection_id);
    setSelectedConnectionIds(authorizedIds);
    // 显示模态框
    setIsPermissionModalVisible(true);
  };

  // 关闭权限管理模态框
  const handlePermissionCancel = () => {
    setIsPermissionModalVisible(false);
  };

  // 保存用户权限
  const handleSavePermissions = async () => {
    try {
      setLoading(true);
      
      // 先删除用户所有现有权限
      for (const perm of permissions) {
        await userApi.revokePermission(perm.id);
      }
      
      // 再添加新的权限
      for (const connId of selectedConnectionIds) {
        await userApi.grantPermission({
          user_id: currentUser.id,
          connection_id: connId,
        });
      }
      
      message.success('用户权限更新成功');
      setIsPermissionModalVisible(false);
    } catch (error) {
      message.error('用户权限更新失败');
      console.error('更新用户权限失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 表格列配置
  const columns = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      ellipsis: true,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      ellipsis: true,
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role) => (
        <Tag color={role === 'admin' ? 'red' : 'blue'}>
          {role === 'admin' ? '管理员' : '用户'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'active',
      key: 'active',
      render: (active, record) => (
        <Switch
          checked={active}
          onChange={() => handleToggleActive(record)}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (createdAt) => new Date(createdAt).toLocaleString(),
      sorter: (a, b) => new Date(a.created_at) - new Date(b.created_at),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => showEditModal(record)}
          >
            编辑
          </Button>
          <Button
            type="link"
            icon={<KeyOutlined />}
            onClick={() => showPermissionModal(record)}
          >
            授权
          </Button>
          <Button
            type="link"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record.id, record.username)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

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
          <h1 style={{ margin: 0, fontSize: 24 }}>用户管理</h1>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={showCreateModal}
        >
          创建用户
        </Button>
      </div>

      <Card title="用户列表">
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      {/* 用户创建/编辑模态框 */}
      <Modal
        title={isEditMode ? '编辑用户' : '创建用户'}
        open={isModalVisible}
        onCancel={handleCancel}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
        >
          <Form.Item
            label="用户名"
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>

          <Form.Item
            label="邮箱"
            name="email"
            rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '请输入有效的邮箱地址' }]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          {!isEditMode && (
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '密码长度不能少于6位' }]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}

          <Form.Item
            label="角色"
            name="role"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select placeholder="请选择角色">
              <Select.Option value="user">用户</Select.Option>
              <Select.Option value="admin">管理员</Select.Option>
            </Select>
          </Form.Item>

          {isEditMode && (
            <Form.Item
              label="状态"
              name="active"
            >
              <Select placeholder="请选择状态">
                <Select.Option value={true}>启用</Select.Option>
                <Select.Option value={false}>禁用</Select.Option>
              </Select>
            </Form.Item>
          )}

          <Form.Item style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button onClick={handleCancel}>取消</Button>
              <Button
                type="primary"
                icon={<SaveOutlined />}
                loading={loading}
                htmlType="submit"
              >
                保存
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 用户权限管理模态框 */}
      <Modal
        title={`${currentUser?.username} 的服务器访问权限`}
        open={isPermissionModalVisible}
        onCancel={handlePermissionCancel}
        footer={null}
        width={600}
      >
        <div style={{ marginBottom: 16 }}>
          <h4>选择用户可以访问的服务器：</h4>
        </div>
        <Form
          layout="vertical"
          onFinish={handleSavePermissions}
        >
          <Form.Item
            label="可访问服务器"
            name="connections"
            rules={[{ required: true, message: '请至少选择一个服务器' }]}
          >
            <Checkbox.Group
              options={connections.map(conn => ({
                label: `${conn.name} (${conn.host}:${conn.port})`,
                value: conn.id
              }))}
              value={selectedConnectionIds}
              onChange={(values) => setSelectedConnectionIds(values)}
              style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}
            />
          </Form.Item>

          <Form.Item style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button onClick={handlePermissionCancel}>取消</Button>
              <Button
                type="primary"
                icon={<SaveOutlined />}
                loading={loading}
                htmlType="submit"
              >
                保存权限
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserManagement;
