import React from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Connections from './pages/SSHConnections';
import FileManager from './pages/FileManager';
import Notifications from './pages/Notifications';
import NotFound from './pages/NotFound';
import PrivateRoute from './components/PrivateRoute';
import MainLayout from './components/MainLayout';

// 创建路由
const router = createBrowserRouter([
  // 公共路由
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/register',
    element: <Register />,
  },
  
  // 私有路由 - 需要认证
  {
    path: '/',
    element: <PrivateRoute />,
    children: [
      {
        path: '',
        element: <MainLayout />,
        children: [
          {
            path: '',
            element: <Dashboard />,
          },
          {
            path: 'connections',
            element: <Connections />,
          },
          {
            path: 'files/:connId/*',
            element: <FileManager />,
          },
          {
            path: 'notifications',
            element: <Notifications />,
          },
        ],
      },
    ],
  },
  
  // 404页面
  {
    path: '*',
    element: <NotFound />,
  },
]);

// 路由提供者组件
const Routes = () => {
  return <RouterProvider router={router} />;
};

export default Routes;
