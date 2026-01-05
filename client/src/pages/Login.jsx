import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, Alert } from 'antd';
import { LockOutlined, UserOutlined, DatabaseOutlined } from '@ant-design/icons';
import { authApi } from '../services/api';

const Login = () => {
  const [form] = Form.useForm();
  const [error, setError] = React.useState('');
  const [isLoading, setIsLoading] = React.useState(false);
  const navigate = useNavigate();
  
  // 获取当前主题状态 - 从localStorage读取，与App.jsx保持一致
  const [isDarkTheme, setIsDarkTheme] = useState(() => {
    const savedTheme = localStorage.getItem('theme');
    return savedTheme === 'dark';
  });
  
  // 监听主题变化
  useEffect(() => {
    const handleThemeChange = () => {
      const savedTheme = localStorage.getItem('theme');
      setIsDarkTheme(savedTheme === 'dark');
    };
    
    // 监听storage变化
    window.addEventListener('storage', handleThemeChange);
    
    return () => {
      window.removeEventListener('storage', handleThemeChange);
    };
  }, []);

  // 处理表单提交
  const handleSubmit = async (values) => {
    setIsLoading(true);
    setError('');

    try {
      // 调用登录API
      const response = await authApi.login(values);
      
      // 保存token和用户信息
      localStorage.setItem('token', response.token);
      localStorage.setItem('user', JSON.stringify(response.user));
      
      // 重定向到仪表板
      navigate('/');
    } catch (err) {
      setError(err.message || '登录失败，请检查用户名和密码');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      minHeight: '100vh',
      alignItems: 'center',
      justifyContent: 'center',
      background: isDarkTheme 
        ? 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)' 
        : 'linear-gradient(135deg, #ffffff 0%, #f1f5f9 100%)',
      position: 'relative',
      overflow: 'hidden',
      fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
      color: isDarkTheme ? '#e2e8f0' : '#1e293b'
    }}>
      {/* 动态背景光效 */}
      <div style={{
        position: 'absolute',
        top: '50%',
        left: '50%',
        width: '800px',
        height: '800px',
        background: isDarkTheme 
          ? 'radial-gradient(circle, rgba(59, 130, 246, 0.15) 0%, transparent 70%)' 
          : 'radial-gradient(circle, rgba(59, 130, 246, 0.08) 0%, transparent 70%)',
        transform: 'translate(-50%, -50%)',
        zIndex: 0
      }} />
      
      {/* 网格背景 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundImage: `
          linear-gradient(${isDarkTheme ? 'rgba(59, 130, 246, 0.08)' : 'rgba(59, 130, 246, 0.05)'} 1px, transparent 1px),
          linear-gradient(90deg, ${isDarkTheme ? 'rgba(59, 130, 246, 0.08)' : 'rgba(59, 130, 246, 0.05)'} 1px, transparent 1px)
        `,
        backgroundSize: '40px 40px',
        zIndex: 0
      }} />
      
      {/* 发光装饰环 */}
      <div style={{
        position: 'absolute',
        top: '-20%',
        right: '-20%',
        width: '600px',
        height: '600px',
        border: `2px solid ${isDarkTheme ? 'rgba(59, 130, 246, 0.15)' : 'rgba(59, 130, 246, 0.1)'}`,
        borderRadius: '50%',
        opacity: isDarkTheme ? 0.5 : 0.3,
        animation: 'rotate 30s linear infinite',
        zIndex: 0
      }} />
      
      <div style={{
        position: 'absolute',
        bottom: '-25%',
        left: '-25%',
        width: '700px',
        height: '700px',
        border: `2px solid ${isDarkTheme ? 'rgba(147, 197, 253, 0.12)' : 'rgba(147, 197, 253, 0.08)'}`,
        borderRadius: '50%',
        opacity: isDarkTheme ? 0.4 : 0.2,
        animation: 'rotate 45s linear infinite reverse',
        zIndex: 0
      }} />
      
      {/* 主登录卡片 */}
      <Card
        style={{
          width: 400,
          borderRadius: '20px',
          background: isDarkTheme 
            ? 'rgba(15, 23, 42, 0.95)' 
            : 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(30px)',
          border: `1px solid ${isDarkTheme ? 'rgba(59, 130, 246, 0.2)' : 'rgba(59, 130, 246, 0.15)'}`,
          boxShadow: isDarkTheme 
            ? '0 20px 60px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(59, 130, 246, 0.1) inset' 
            : '0 20px 60px rgba(0, 0, 0, 0.08), 0 0 0 1px rgba(59, 130, 246, 0.1) inset',
          zIndex: 1,
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          overflow: 'hidden'
        }}
        bordered={false}
      >
        {/* 卡片顶部渐变装饰条 */}
        <div style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '4px',
          background: 'linear-gradient(90deg, #3b82f6 0%, #60a5fa 50%, #93c5fd 100%)',
        }} />
        
        {/* 系统图标和标题 */}
        <div style={{
          textAlign: 'center',
          padding: '32px 0 24px',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center'
        }}>
          <div style={{
            width: '80px',
            height: '80px',
            borderRadius: '16px',
            background: 'linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: isDarkTheme ? '0 8px 24px rgba(59, 130, 246, 0.4)' : '0 8px 24px rgba(59, 130, 246, 0.2)',
            marginBottom: '16px'
          }}>
            <DatabaseOutlined style={{ fontSize: '36px', color: '#ffffff' }} />
          </div>
          <h1 style={{
            margin: 0,
            fontSize: '24px',
            fontWeight: '700',
            color: isDarkTheme ? '#ffffff' : '#1e293b',
            letterSpacing: '0.5px',
            marginBottom: '8px'
          }}>
            远程文件管理系统
          </h1>
          <p style={{
            margin: 0,
            fontSize: '14px',
            color: isDarkTheme ? '#94a3b8' : '#64748b',
            letterSpacing: '0.3px'
          }}>
            安全连接 · 高效管理 · 智能协作
          </p>
        </div>
        
        {/* 错误提示 */}
        {error && (
          <Alert
            message="登录失败"
            description={error}
            type="error"
            showIcon
            style={{
              margin: '0 32px 24px',
              background: isDarkTheme ? 'rgba(239, 68, 68, 0.1)' : 'rgba(239, 68, 68, 0.05)',
              border: isDarkTheme ? '1px solid rgba(239, 68, 68, 0.2)' : '1px solid rgba(239, 68, 68, 0.15)',
              borderRadius: '10px',
              color: isDarkTheme ? '#fca5a5' : '#dc2626',
              fontSize: '14px'
            }}
          />
        )}
        
        {/* 登录表单 */}
        <Form
          form={form}
          name="login"
          onFinish={handleSubmit}
          autoComplete="on"
          style={{ padding: '0 32px 32px' }}
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名!' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#60a5fa', fontSize: '16px', marginRight: '8px' }} />}
              placeholder="用户名"
              autoComplete="username"
              style={{
                background: isDarkTheme ? 'rgba(255, 255, 255, 0.05)' : 'rgba(255, 255, 255, 0.8)',
                border: isDarkTheme ? '1px solid rgba(59, 130, 246, 0.2)' : '1px solid rgba(59, 130, 246, 0.15)',
                borderRadius: '12px',
                color: isDarkTheme ? '#e2e8f0' : '#1e293b',
                height: '50px',
                fontSize: '15px',
                transition: 'all 0.3s ease',
                padding: '12px 16px 12px 48px',
                '&:focus': {
                  borderColor: '#3b82f6',
                  boxShadow: '0 0 0 3px rgba(59, 130, 246, 0.1)',
                  background: isDarkTheme ? 'rgba(255, 255, 255, 0.08)' : 'rgba(255, 255, 255, 0.95)'
                },
                '&:hover': {
                  borderColor: '#60a5fa',
                  background: isDarkTheme ? 'rgba(255, 255, 255, 0.07)' : 'rgba(255, 255, 255, 0.9)'
                }
              }}
              placeholderStyle={{
                color: isDarkTheme ? '#94a3b8' : '#94a3b8'
              }}
            />
          </Form.Item>
          
          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码!' }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#60a5fa', fontSize: '16px', marginRight: '8px' }} />}
              placeholder="密码"
              autoComplete="current-password"
              style={{
                background: isDarkTheme ? 'rgba(255, 255, 255, 0.05)' : 'rgba(255, 255, 255, 0.8)',
                border: isDarkTheme ? '1px solid rgba(59, 130, 246, 0.2)' : '1px solid rgba(59, 130, 246, 0.15)',
                borderRadius: '12px',
                color: isDarkTheme ? '#e2e8f0' : '#1e293b',
                height: '50px',
                fontSize: '15px',
                transition: 'all 0.3s ease',
                padding: '12px 16px 12px 48px',
                '&:focus': {
                  borderColor: '#3b82f6',
                  boxShadow: '0 0 0 3px rgba(59, 130, 246, 0.1)',
                  background: isDarkTheme ? 'rgba(255, 255, 255, 0.08)' : 'rgba(255, 255, 255, 0.95)'
                },
                '&:hover': {
                  borderColor: '#60a5fa',
                  background: isDarkTheme ? 'rgba(255, 255, 255, 0.07)' : 'rgba(255, 255, 255, 0.9)'
                }
              }}
              placeholderStyle={{
                color: isDarkTheme ? '#94a3b8' : '#94a3b8'
              }}
              iconRender={(visible) => (
                <span style={{ color: isDarkTheme ? '#94a3b8' : '#64748b', fontSize: '16px', cursor: 'pointer' }}>
                  {visible ? '👁️' : '👁️‍🗨️'}
                </span>
              )}
            />
          </Form.Item>
          
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={isLoading}
              block
              size="large"
              style={{
                background: 'linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%)',
                border: 'none',
                borderRadius: '12px',
                fontSize: '16px',
                fontWeight: '600',
                color: '#ffffff',
                boxShadow: isDarkTheme ? '0 8px 24px rgba(59, 130, 246, 0.4)' : '0 8px 24px rgba(59, 130, 246, 0.2)',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                height: '52px',
                marginTop: '16px',
                '&:hover': {
                  background: 'linear-gradient(135deg, #2563eb 0%, #3b82f6 100%)',
                  boxShadow: isDarkTheme ? '0 12px 32px rgba(59, 130, 246, 0.5)' : '0 12px 32px rgba(59, 130, 246, 0.3)',
                  transform: 'translateY(-2px)'
                },
                '&:active': {
                  transform: 'translateY(0)',
                  boxShadow: isDarkTheme ? '0 6px 20px rgba(59, 130, 246, 0.4)' : '0 6px 20px rgba(59, 130, 246, 0.25)'
                },
                '&[disabled]': {
                  opacity: 0.6,
                  boxShadow: 'none',
                  transform: 'none'
                }
              }}
            >
              {isLoading ? '登录中...' : '登录'}
            </Button>
          </Form.Item>
        </Form>
        
        {/* 版权信息 */}
        <div style={{
          textAlign: 'center',
          padding: '16px 0',
          borderTop: `1px solid ${isDarkTheme ? 'rgba(59, 130, 246, 0.1)' : 'rgba(59, 130, 246, 0.1)'}`,
          fontSize: '13px',
          color: isDarkTheme ? '#64748b' : '#94a3b8'
        }}>
          © 2026 Remote File Manager. All rights reserved.
        </div>
      </Card>
      
      {/* 动态粒子效果 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        pointerEvents: 'none',
        zIndex: 0
      }}>
        {Array.from({ length: 20 }).map((_, index) => (
          <div
            key={index}
            style={{
              position: 'absolute',
              width: `${Math.random() * 3 + 1}px`,
              height: `${Math.random() * 3 + 1}px`,
              background: isDarkTheme ? '#60a5fa' : '#3b82f6',
              borderRadius: '50%',
              opacity: Math.random() * 0.5 + 0.1,
              animation: `particleMove${index % 3} ${Math.random() * 10 + 10}s linear infinite`,
              boxShadow: `0 0 ${Math.random() * 10 + 5}px rgba(59, 130, 246, ${Math.random() * 0.5 + 0.3})`
            }}
          />
        ))}
      </div>
      
      {/* 动态网格线 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundImage: `
          linear-gradient(${isDarkTheme ? 'rgba(59, 130, 246, 0.1)' : 'rgba(59, 130, 246, 0.05)'} 1px, transparent 1px),
          linear-gradient(90deg, ${isDarkTheme ? 'rgba(59, 130, 246, 0.1)' : 'rgba(59, 130, 246, 0.05)'} 1px, transparent 1px)
        `,
        backgroundSize: '40px 40px',
        animation: 'gridMove 20s linear infinite',
        zIndex: 0
      }} />
      
      {/* 增强光效 */}
      <div style={{
        position: 'absolute',
        top: '50%',
        left: '50%',
        width: '1000px',
        height: '1000px',
        background: isDarkTheme 
          ? 'radial-gradient(circle, rgba(59, 130, 246, 0.15) 0%, transparent 70%)' 
          : 'radial-gradient(circle, rgba(59, 130, 246, 0.1) 0%, transparent 70%)',
        transform: 'translate(-50%, -50%)',
        zIndex: 0,
        animation: 'pulse 4s ease-in-out infinite'
      }} />
      
      {/* 装饰线条 */}
      <div style={{
        position: 'absolute',
        bottom: '10%',
        left: '10%',
        width: '2px',
        height: '100px',
        background: `linear-gradient(to bottom, ${isDarkTheme ? 'rgba(59, 130, 246, 0.3)' : 'rgba(59, 130, 246, 0.15)'} 0%, transparent 100%)`,
        zIndex: 0,
        animation: 'linePulse 3s ease-in-out infinite'
      }} />
      
      <div style={{
        position: 'absolute',
        top: '10%',
        right: '10%',
        width: '2px',
        height: '100px',
        background: `linear-gradient(to top, ${isDarkTheme ? 'rgba(59, 130, 246, 0.3)' : 'rgba(59, 130, 246, 0.15)'} 0%, transparent 100%)`,
        zIndex: 0,
        animation: 'linePulse 3s ease-in-out infinite 1s'
      }} />
      
      {/* 动画样式 */}
      <style>{`
        @keyframes rotate {
          0% {
            transform: rotate(0deg) scale(1);
          }
          50% {
            transform: rotate(180deg) scale(1.1);
          }
          100% {
            transform: rotate(360deg) scale(1);
          }
        }
        
        @keyframes particleMove0 {
          0% {
            transform: translate(0, 100vh);
            opacity: 0;
          }
          10% {
            opacity: 0.5;
          }
          90% {
            opacity: 0.5;
          }
          100% {
            transform: translate(${Math.random() * 100 - 50}vw, -100px);
            opacity: 0;
          }
        }
        
        @keyframes particleMove1 {
          0% {
            transform: translate(100vw, ${Math.random() * 100}vh);
            opacity: 0;
          }
          10% {
            opacity: 0.5;
          }
          90% {
            opacity: 0.5;
          }
          100% {
            transform: translate(-100px, ${Math.random() * 100 - 50}vh);
            opacity: 0;
          }
        }
        
        @keyframes particleMove2 {
          0% {
            transform: translate(${Math.random() * 100}vw, 0);
            opacity: 0;
          }
          10% {
            opacity: 0.5;
          }
          90% {
            opacity: 0.5;
          }
          100% {
            transform: translate(${Math.random() * 100 - 50}vw, 100vh);
            opacity: 0;
          }
        }
        
        @keyframes gridMove {
          0% {
            backgroundPosition: 0 0;
          }
          100% {
            backgroundPosition: 40px 40px;
          }
        }
        
        @keyframes pulse {
          0%, 100% {
            transform: translate(-50%, -50%) scale(1);
            opacity: 0.3;
          }
          50% {
            transform: translate(-50%, -50%) scale(1.2);
            opacity: 0.5;
          }
        }
        
        @keyframes linePulse {
          0%, 100% {
            opacity: 0.3;
          }
          50% {
            opacity: 0.8;
            boxShadow: 0 0 20px rgba(59, 130, 246, 0.5);
          }
        }
      `}</style>
    </div>
  );
};

export default Login;
