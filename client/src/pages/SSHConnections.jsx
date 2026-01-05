import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { connectionApi, sshApi } from '../services/api';
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
  Spin,
  Select
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  DatabaseOutlined,
  FolderOpenOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;

const Connections = () => {
  const navigate = useNavigate();
  const [connections, setConnections] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isCategoryModalOpen, setIsCategoryModalOpen] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [editingConn, setEditingConn] = useState(null);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [categories, setCategories] = useState([]);
  const [editingCategory, setEditingCategory] = useState(null);
  const [form] = Form.useForm();
  const [categoryForm] = Form.useForm();
  
  // 加载所有分类
  const loadCategories = async () => {
    try {
      const categoryList = await connectionApi.getCategories();
      setCategories(categoryList.map(cat => cat.name));
    } catch (err) {
      console.error('加载分类失败:', err);
      message.error('加载分类失败');
    }
  };
  
  // 筛选后的连接列表
  const filteredConnections = selectedCategory
    ? connections.filter(conn => (conn.category || '未分类') === selectedCategory)
    : connections;

  useEffect(() => {
    // 加载SSH连接列表
    loadConnections();
  }, []);

  // 加载连接列表
  const loadConnections = async () => {
    try {
      setIsLoading(true);
      const connList = await connectionApi.getConnections();
      setConnections(connList);
      // 加载分类
      await loadCategories();
    } catch (err) {
      console.error('加载连接失败:', err);
      message.error('加载连接失败');
    } finally {
      setIsLoading(false);
    }
  };

  // 打开添加连接模态框
  const openAddModal = () => {
    setEditingConn(null);
    form.setFieldsValue({
      name: '',
      type: 'ssh', // 默认连接类型为SSH
      host: '',
      port: 22,
      username: '',
      password: '',
      privateKey: '',
      passphrase: '',
      isActive: true,
      category: '', // 默认分类为空
    });
    setIsModalOpen(true);
    setTestResult(null);
  };

  // 打开编辑连接模态框
  const openEditModal = (conn) => {
    setEditingConn(conn);
    form.setFieldsValue({
      name: conn.name,
      type: conn.type || 'ssh', // 连接类型
      host: conn.host,
      port: conn.port,
      username: conn.username,
      password: '', // 不显示密码
      privateKey: '', // 不显示私钥
      passphrase: '', // 不显示密码短语
      isActive: conn.isActive,
      category: conn.category || '', // 分类字段
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

  // 关闭分类模态框
  const closeCategoryModal = () => {
    setIsCategoryModalOpen(false);
    setEditingCategory(null);
    categoryForm.resetFields();
  };

  // 打开添加分类模态框
  const openAddCategoryModal = () => {
    setEditingCategory(null);
    categoryForm.setFieldsValue({ name: '' });
    setIsCategoryModalOpen(true);
  };

  // 打开编辑分类模态框
  const openEditCategoryModal = (category) => {
    setEditingCategory(category);
    categoryForm.setFieldsValue({ name: category });
    setIsCategoryModalOpen(true);
  };

  // 保存分类
  const saveCategory = async () => {
    try {
      const values = categoryForm.getFieldsValue();
      const categoryName = values.name.trim();
      
      if (!categoryName) {
        message.error('分类名称不能为空');
        return;
      }

      if (editingCategory) {
        // 编辑现有分类
        await sshApi.updateCategory(editingCategory, categoryName);
        message.success('分类已更新');
      } else {
        // 添加新分类
        await sshApi.addCategory(categoryName);
        message.success('分类已添加');
      }
      
      // 重新加载分类和连接
      await loadCategories();
      await loadConnections();
      closeCategoryModal();
    } catch (error) {
      console.error('保存分类失败:', error);
      message.error(error.message || '保存分类失败');
    }
  };

  // 删除分类
  const deleteCategory = (category) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除分类 "${category}" 吗？使用该分类的连接将变为未分类。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await sshApi.deleteCategory(category);
          
          // 重新加载分类和连接
          await loadCategories();
          await loadConnections();
          
          // 如果当前筛选的是被删除的分类，清除筛选
          if (selectedCategory === category) {
            setSelectedCategory('');
          }
          
          message.success('分类已删除');
        } catch (error) {
          console.error('删除分类失败:', error);
          message.error(error.message || '删除分类失败');
        }
      }
    });
  };

  // 测试连接
  const testConnection = async () => {
    try {
      const values = await form.validateFields();
      setIsTesting(true);
      setTestResult(null);

      // 直接调用测试连接API，不保存连接
      await connectionApi.testConnectionDirect(values);
      
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

  // 保存连接
  const saveConnection = async () => {
    try {
      const values = await form.validateFields();

      if (editingConn) {
        // 更新现有连接
        await connectionApi.updateConnection(editingConn.id, values);
        message.success('连接更新成功');
      } else {
        // 添加新连接
        await connectionApi.addConnection(values);
        message.success('连接添加成功');
      }

      // 关闭模态框并刷新列表
      closeModal();
      loadConnections();
    } catch (err) {
      message.error(err.message || '保存连接失败');
    }
  };

  // 删除连接
  const deleteConnection = async (connId, connName) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除连接 "${connName}" 吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await connectionApi.deleteConnection(connId);
          message.success('连接删除成功');
          loadConnections();
        } catch (err) {
          console.error('删除连接失败:', err);
          message.error('删除连接失败');
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
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      render: (category) => (
        <Badge 
          status={category ? 'processing' : 'default'} 
          text={category || '未分类'}
          style={{ backgroundColor: category ? '#1890ff' : '#f0f0f0', color: category ? '#fff' : '#333' }}
        />
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
            icon={<FolderOpenOutlined />}
            onClick={() => navigate(`/files/${record.id}/`)} // 导航到文件管理页面
          >
            文件管理
          </Button>
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
    <div>
      {/* 页面标题和操作按钮 */}
      <Card className="hover-scale" style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 12 }}>
          <div>
          <Title level={2} style={{ margin: 0, fontSize: 'clamp(18px, 4vw, 24px)' }}>连接管理</Title>
          <Text type="secondary">管理您的远程连接配置</Text>
        </div>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={openAddModal}
            className="hover-scale"
          >
            添加连接
          </Button>
        </div>
      </Card>

      {/* 分类筛选和SSH连接列表 */}
      <Card>
        {/* 分类管理和筛选器 */}
        <div style={{ marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
            <Text strong>分类管理:</Text>
            <Button 
              type="primary" 
              size="small" 
              icon={<PlusOutlined />}
              onClick={openAddCategoryModal}
            >
              添加分类
            </Button>
          </div>
          
          {/* 自定义分类列表 */}
          {categories.length > 0 && (
            <div style={{ marginBottom: 16, display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
              <Text strong>自定义分类:</Text>
              <Space size="small">
                {categories.map(category => (
                  <Badge 
                    key={category}
                    status="processing"
                    text={
                      <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                        {category}
                        <span style={{ display: 'flex', gap: 4 }}>
                          <Button 
                            type="text" 
                            size="small"
                            icon={<EditOutlined />}
                            onClick={() => openEditCategoryModal(category)}
                          />
                          <Button 
                            type="text" 
                            size="small"
                            danger
                            icon={<DeleteOutlined />}
                            onClick={() => deleteCategory(category)}
                          />
                        </span>
                      </span>
                    } 
                    style={{ backgroundColor: '#1890ff', color: '#fff', padding: '0 8px', borderRadius: 4 }}
                  />
                ))}
              </Space>
            </div>
          )}
          
          {/* 分类筛选器 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Text strong>按分类筛选:</Text>
            <Select
              placeholder="选择分类"
              value={selectedCategory}
              onChange={setSelectedCategory}
              allowClear
              style={{ width: 200 }}
              options={[
                { value: '未分类', label: '未分类' },
                ...categories.map(category => ({ value: category, label: category }))
              ]}
            />
            {selectedCategory && (
              <Button 
                type="default" 
                size="small"
                onClick={() => setSelectedCategory('')}
              >
                清除筛选
              </Button>
            )}
          </div>
        </div>
        
        {/* SSH连接列表 */}
        <div style={{ overflowX: 'auto', borderRadius: 8, border: '1px solid #f0f0f0' }}>
          <Table
            columns={columns}
            dataSource={filteredConnections}
            rowKey="id"
            loading={isLoading}
            bordered={false}
            pagination={false}
            scroll={{ x: 'max-content' }}
            className="connections-table"
            locale={{
                    emptyText: (
                      <div style={{ textAlign: 'center', padding: 40 }}>
                        <DatabaseOutlined style={{ fontSize: 48, color: '#ccc', marginBottom: 16 }} />
                        <p>没有连接</p>
                        <p>点击"添加连接"按钮创建您的第一个连接</p>
                      </div>
                    )
                  }}
          />
        </div>
      </Card>

      {/* 添加/编辑连接模态框 */}
      <Modal
        title={editingConn ? '编辑连接' : '添加连接'}
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
            category: '',
          }}
        >
          <Form.Item
            name="name"
            label="连接名称"
            rules={[{ required: true, message: '请输入连接名称!' }]}
          >
            <Input placeholder="输入连接名称" />
          </Form.Item>

          <Form.Item
            name="type"
            label="连接类型"
            rules={[{ required: true, message: '请选择连接类型!' }]}
          >
            <Select placeholder="选择连接类型">
              <Select.Option value="ssh">SSH</Select.Option>
              <Select.Option value="sftp">SFTP</Select.Option>
              <Select.Option value="ftp">FTP</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="category"
            label="分类"
          >
            <Select
              placeholder="选择或输入分类"
              showSearch
              allowClear
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
            >
              {categories.map(category => (
                <Select.Option key={category} value={category}>
                  {category}
                </Select.Option>
              ))}
            </Select>
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

      {/* 添加/编辑分类模态框 */}
      <Modal
        title={editingCategory ? '编辑分类' : '添加分类'}
        open={isCategoryModalOpen}
        onCancel={closeCategoryModal}
        footer={null}
        width={400}
        destroyOnClose
      >
        <Form
          form={categoryForm}
          layout="vertical"
          initialValues={{ name: '' }}
        >
          <Form.Item
            name="name"
            label="分类名称"
            rules={[{ required: true, message: '请输入分类名称!' }]}
          >
            <Input placeholder="输入分类名称" />
          </Form.Item>

          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 24, gap: 8 }}>
            <Button onClick={closeCategoryModal}>
              取消
            </Button>
            <Button
              type="primary"
              onClick={saveCategory}
            >
              保存
            </Button>
          </div>
        </Form>
      </Modal>
    </div>
  );
};

export default Connections;
