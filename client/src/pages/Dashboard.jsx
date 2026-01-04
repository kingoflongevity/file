import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
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

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-xl text-gray-600">加载中...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div>
        <h1 className="text-3xl font-bold text-gray-800">仪表盘</h1>
        <p className="mt-1 text-gray-600">欢迎使用文件管理系统</p>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* 总连接数 */}
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">总SSH连接</p>
              <h3 className="mt-1 text-3xl font-bold text-gray-800">{stats.totalConnections}</h3>
            </div>
            <div className="p-3 bg-blue-100 rounded-full">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
          </div>
          <Link
            to="/connections"
            className="mt-4 inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-500"
          >
            查看所有连接
            <svg className="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        </div>

        {/* 活跃连接数 */}
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">活跃连接</p>
              <h3 className="mt-1 text-3xl font-bold text-gray-800">{stats.activeConnections}</h3>
            </div>
            <div className="p-3 bg-green-100 rounded-full">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
          </div>
          <Link
            to="/connections"
            className="mt-4 inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-500"
          >
            管理连接
            <svg className="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        </div>

        {/* 功能卡片 */}
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">快速访问</p>
              <h3 className="mt-1 text-xl font-bold text-gray-800">文件管理</h3>
            </div>
            <div className="p-3 bg-purple-100 rounded-full">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
              </svg>
            </div>
          </div>
          <p className="mt-2 text-sm text-gray-600">通过SSH连接管理远程文件</p>
          <Link
            to="/connections"
            className="mt-4 inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-500"
          >
            开始使用
            <svg className="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
            </svg>
          </Link>
        </div>
      </div>

      {/* 最近连接 */}
      {stats.recentConnections.length > 0 && (
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-bold text-gray-800">最近使用的连接</h2>
            <Link
              to="/connections"
              className="text-sm font-medium text-blue-600 hover:text-blue-500"
            >
              查看全部
            </Link>
          </div>
          
          <div className="space-y-3">
            {stats.recentConnections.map((conn) => (
              <div
                key={conn.id}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-md hover:bg-gray-100 transition-colors"
              >
                <div className="flex items-center">
                  <div className={`w-3 h-3 rounded-full mr-3 ${
                    conn.is_active ? 'bg-green-500' : 'bg-gray-300'
                  }`}></div>
                  <div>
                    <p className="font-medium text-gray-800">{conn.name}</p>
                    <p className="text-xs text-gray-500">
                      {conn.username}@{conn.host}:{conn.port}
                    </p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-xs text-gray-500">
                    {new Date(conn.last_used).toLocaleString()}
                  </span>
                  <Link
                    to={`/files/${conn.id}/`}
                    className="px-3 py-1 text-xs font-medium text-blue-600 bg-blue-100 rounded-md hover:bg-blue-200 transition-colors"
                  >
                    打开
                  </Link>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 快速指南 */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <h2 className="text-xl font-bold text-gray-800 mb-4">快速开始</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-start">
            <div className="p-2 bg-blue-100 rounded-full mr-3">
              <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <div>
              <h3 className="font-medium text-gray-800">添加SSH连接</h3>
              <p className="mt-1 text-sm text-gray-600">创建新的SSH连接配置，连接到远程服务器</p>
            </div>
          </div>
          <div className="flex items-start">
            <div className="p-2 bg-green-100 rounded-full mr-3">
              <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
              </svg>
            </div>
            <div>
              <h3 className="font-medium text-gray-800">浏览文件系统</h3>
              <p className="mt-1 text-sm text-gray-600">通过SSH连接浏览和管理远程服务器上的文件</p>
            </div>
          </div>
          <div className="flex items-start">
            <div className="p-2 bg-purple-100 rounded-full mr-3">
              <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </div>
            <div>
              <h3 className="font-medium text-gray-800">上传和下载文件</h3>
              <p className="mt-1 text-sm text-gray-600">轻松上传本地文件到服务器或从服务器下载文件</p>
            </div>
          </div>
          <div className="flex items-start">
            <div className="p-2 bg-orange-100 rounded-full mr-3">
              <svg className="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
            </div>
            <div>
              <h3 className="font-medium text-gray-800">管理文件</h3>
              <p className="mt-1 text-sm text-gray-600">创建、删除、重命名、移动和复制文件和文件夹</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
