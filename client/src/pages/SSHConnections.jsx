import React, { useEffect, useState } from 'react';
import { sshApi } from '../services/api';
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  InputNumber,
  Switch,
  message,
  Space,
  Badge,
  Card,
  Divider,
  Typography,
  Avatar,
  Spin
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  DatabaseOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;

const SSHConnections = () => {
  const [connections, setConnections] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [editingConn, setEditingConn] = useState(null);
  const [form] = Form.useForm();

  useEffect(() => {
    // 加载SSH连接列表
    loadConnections();
  }, []);

  // 加载SSH连接列表
  const loadConnections = async () => {
    try {
      setIsLoading(true);
      const connList = await sshApi.getConnections();
      setConnections(connList);
    } catch (err) {
      console.error('加载SSH连接失败:', err);
      message.error('加载SSH连接失败');
    } finally {
      setIsLoading(false);
    }
  };

  // 打开添加连接模态框
  const openAddModal = () => {
    setEditingConn(null);
    form.setFieldsValue({
      name: '',
      host: '',
      port: 22,
      username: '',
      password: '',
      privateKey: '',
      passphrase: '',
      isActive: true,
    });
    setIsModalOpen(true);
    setTestResult(null);
  };

  // 打开编辑连接模态框
  const openEditModal = (conn) => {
    setEditingConn(conn);
    form.setFieldsValue({
      name: conn.name,
      host: conn.host,
      port: conn.port,
      username: conn.username,
      password: '', // 不显示密码
      privateKey: '', // 不显示私钥
      passphrase: '', // 不显示密码短语
      isActive: conn.isActive,
    });
    setIsModalOpen(true);
    setTestResult(null);
  };

  // 关闭模态框
  const closeModal = () => {
    setIsModalOpen(false);
    setEditingConn(null);
    setTestResult(null);
    form.resetFields();
  };

  // 测试SSH连接
  const testConnection = async () => {
    try {
      const values = await form.validateFields();
      setIsTesting(true);
      setTestResult(null);
      // 创建临时连接进行测试
      const tempConn = {
        ...values,
        id: 'temp-test-connection',
      };

      // 调用添加连接API（会自动测试连接）
      await sshApi.addConnection(tempConn);
      
      // 测试成功
      setTestResult({ success: true, message: '连接测试成功' });
      message.success('连接测试成功');
    } catch (err) {
      // 测试失败
      setTestResult({ success: false, message: err.message || '连接测试失败' });
      message.error(err.message || '连接测试失败');
    } finally {
      setIsTesting(false);
    }
  };

  // 保存SSH连接
  const saveConnection = async () => {
    try {
      const values = await form.validateFields();

      if (editingConn) {
        // 更新现有连接
        await sshApi.updateConnection(editingConn.id, values);
        message.success('SSH连接更新成功');
      } else {
        // 添加新连接
        await sshApi.addConnection(values);
        message.success('SSH连接添加成功');
      }

      // 关闭模态框并刷新列表
      closeModal();
      loadConnections();
    } catch (err) {
      message.error(err.message || '保存SSH连接失败');
    }
  };

  // 删除SSH连接
  const deleteConnection = async (connId, connName) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除SSH连接 "${connName}" 吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await sshApi.deleteConnection(connId);
          message.success('SSH连接删除成功');
          loadConnections();
        } catch (err) {
          console.error('删除SSH连接失败:', err);
          message.error('删除SSH连接失败');
        }
      }
    });
  };

  // 切换连接状态
  const toggleConnectionStatus = async (conn) => {
    try {
      await sshApi.updateConnection(conn.id, {
        ...conn,
        isActive: !conn.isActive,
      });
      message.success(`连接已${!conn.isActive ? '启用' : '停用'}`);
      loadConnections();
    } catch (err) {
      console.error('更新连接状态失败:', err);
      message.error('更新连接状态失败');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar icon={<DatabaseOutlined />} style={{ marginRight: 8 }} />
          <Text strong>{text}</Text>
        </div>
      )
    },
    {
      title: '连接信息',
      dataIndex: 'host',
      key: 'host',
      render: (host, record) => (
        <Text>{record.username}@{host}:{record.port}</Text>
      )
    },
    {
      title: '状态',
      dataIndex: 'isActive',
      key: 'isActive',
      render: (isActive) => (
        <Badge status={isActive ? 'success' : 'default'} text={isActive ? '活跃' : '非活跃'} />
      )
    },
    {
      title: '最后使用',
      dataIndex: 'last_used',
      key: 'last_used',
      render: (lastUsed) => (
        <Text type="secondary">{new Date(lastUsed).toLocaleString()}</Text>
      )
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button
            type="link"
            size="small"
            icon={record.isActive ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            onClick={() => toggleConnectionStatus(record)}
          >
            {record.isActive ? '停用' : '启用'}
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => deleteConnection(record.id, record.name)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* 页面标题和操作按钮 */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Title level={2} style={{ margin: 0 }}>SSH连接管理</Title>
            <Text type="secondary">管理您的SSH连接配置</Text>
          </div>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={openAddModal}
          >
            添加连接
          </Button>
        </div>
      </Card>

      {/* SSH连接列表 */}
      <Card>
        <Table
          columns={columns}
          dataSource={connections}
          rowKey="id"
          loading={isLoading}
          bordered
          pagination={false}
          locale={{
            emptyText: (
              <div style={{ textAlign: 'center', padding: 40 }}>
                <DatabaseOutlined style={{ fontSize: 48, color: '#ccc', marginBottom: 16 }} />
                <p>没有SSH连接</p>
                <p>点击"添加连接"按钮创建您的第一个SSH连接</p>
              </div>
            )
          }}
        />
      </Card>

      {/* 添加/编辑SSH连接模态框 */}
      <Modal
        title={editingConn ? '编辑SSH连接' : '添加SSH连接'}
        open={isModalOpen}
        onCancel={closeModal}
        footer={null}
        width={700}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            port: 22,
            isActive: true,
          }}
        >
          <Form.Item
            name="name"
            label="连接名称"
            rules={[{ required: true, message: '请输入连接名称!' }]}
          >
            <Input placeholder="输入连接名称" />
          </Form.Item>

          <Divider orientation="left" style={{ margin: '16px 0' }}>主机信息</Divider>
          
          <Space.Compact style={{ width: '100%' }}>
            <Form.Item
              name="host"
              label="主机地址"
              rules={[{ required: true, message: '请输入主机地址!' }]}
              style={{ flex: 2 }}
            >
              <Input placeholder="输入主机地址" />
            </Form.Item>
            <Form.Item
              name="port"
              label="端口"
              rules={[{ required: true, message: '请输入端口!' }]}
              style={{ flex: 1 }}
            >
              <InputNumber min={1} max={65535} style={{ width: '100%' }} />
            </Form.Item>
          </Space.Compact>

          <Divider orientation="left" style={{ margin: '16px 0' }}>认证信息</Divider>
          
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名!' }]}
          >
            <Input placeholder="输入用户名" />
          </Form.Item>

          <Form.Item
            name="password"
            label="密码"
          >
            <Input.Password placeholder="输入密码（可选）" />
          </Form.Item>

          <Form.Item
            name="privateKey"
            label="私钥"
          >
            <Input.TextArea
              rows={6}
              placeholder="输入SSH私钥（可选）"
            />
          </Form.Item>

          <Form.Item
            name="passphrase"
            label="私钥密码短语"
          >
            <Input.Password placeholder="输入私钥密码短语（可选）" />
          </Form.Item>

          <Form.Item
            name="isActive"
            label="设为活跃连接"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          {/* 连接测试结果 */}
          {testResult && (
            <div style={{ 
              padding: 12, 
              borderRadius: 6, 
              marginBottom: 16,
              backgroundColor: testResult.success ? '#f6ffed' : '#fff1f0',
              color: testResult.success ? '#52c41a' : '#ff4d4f',
              border: `1px solid ${testResult.success ? '#b7eb8f' : '#ffccc7'}`
            }}>
              {testResult.message}
            </div>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 24, gap: 8 }}>
          <Button
            onClick={testConnection}
            loading={isTesting}
            icon={<CheckCircleOutlined />}
          >
            {isTesting ? '测试中...' : '测试连接'}
          </Button>
            <Button onClick={closeModal}>
              取消
            </Button>
            <Button
              type="primary"
              onClick={saveConnection}
            >
              保存
            </Button>
          </div>
        </Form>
      </Modal>
    </div>
  );
};

export default SSHConnections;
