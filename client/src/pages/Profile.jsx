import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, message, Space, Avatar } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, SaveOutlined, LeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../services/api';

const Profile = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [user, setUser] = useState(null);

  // 获取当前用户信息
  const fetchUserInfo = async () => {
    try {
      const response = await authApi.getMe();
      setUser(response);
      form.setFieldsValue({
        username: response.username,
        email: response.email,
      });
    } catch (error) {
      message.error('获取用户信息失败');
      console.error('获取用户信息失败:', error);
    }
  };

  // 组件挂载时获取用户信息
  useEffect(() => {
    fetchUserInfo();
  }, []);

  // 保存用户信息
  const handleSave = async (values) => {
    try {
      setLoading(true);
      await authApi.updateProfile(values);
      message.success('个人信息更新成功');
      fetchUserInfo();
    } catch (error) {
      message.error('更新失败');
      console.error('更新个人信息失败:', error);
    } finally {
      setLoading(false);
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
          <h1 style={{ margin: 0, fontSize: 24 }}>个人中心</h1>
        </div>
      </div>

      <Card title="个人信息" style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: 24 }}>
          <Avatar
            size={80}
            style={{ marginRight: 16, backgroundColor: '#1890ff' }}
          >
            {user?.username?.charAt(0).toUpperCase()}
          </Avatar>
          <div>
            <h2 style={{ margin: 0 }}>{user?.username}</h2>
            <p style={{ margin: '4px 0 0 0', color: '#666' }}>{user?.role}</p>
          </div>
        </div>

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
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入用户名"
            />
          </Form.Item>

          <Form.Item
            label="邮箱"
            name="email"
            rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '请输入有效的邮箱地址' }]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="请输入邮箱"
            />
          </Form.Item>

          <Form.Item
            label="新密码"
            name="password"
            rules={[{ min: 6, message: '密码长度不能少于6位' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入新密码（不修改请留空）"
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                icon={<SaveOutlined />}
                onClick={form.submit}
                loading={loading}
              >
                保存修改
              </Button>
              <Button onClick={() => form.resetFields()}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Profile;
