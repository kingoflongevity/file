import React from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import SSHConnections from './pages/SSHConnections';
import FileManager from './pages/FileManager';
import NotFound from './pages/NotFound';
import PrivateRoute from './components/PrivateRoute';

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
        element: <Dashboard />,
      },
      {
        path: 'connections',
        element: <SSHConnections />,
      },
      {
        path: 'files/:connId/*',
        element: <FileManager />,
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
