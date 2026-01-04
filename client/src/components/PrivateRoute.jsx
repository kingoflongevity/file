import React, { useEffect, useState } from 'react';
import { Outlet, Navigate } from 'react-router-dom';
import { authApi } from '../services/api';

const PrivateRoute = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // 检查用户是否已登录
    const checkAuth = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          setIsAuthenticated(false);
          setIsLoading(false);
          return;
        }

        // 验证token有效性
        const user = await authApi.getMe();
        if (user) {
          setIsAuthenticated(true);
          localStorage.setItem('user', JSON.stringify(user));
        } else {
          setIsAuthenticated(false);
          localStorage.removeItem('token');
          localStorage.removeItem('user');
        }
      } catch (error) {
        console.error('认证检查失败:', error);
        setIsAuthenticated(false);
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  if (isLoading) {
    // 加载中状态
    return (
      <div className="flex items-center justify-center h-screen bg-gray-100">
        <div className="text-xl text-gray-600">加载中...</div>
      </div>
    );
  }

  return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />;
};

export default PrivateRoute;
