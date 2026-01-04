import React, { useEffect, useState } from 'react';
import { sshApi } from '../services/api';

const SSHConnections = () => {
  const [connections, setConnections] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [editingConn, setEditingConn] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    host: '',
    port: 22,
    username: '',
    password: '',
    privateKey: '',
    passphrase: '',
    isActive: true,
  });
  const [error, setError] = useState('');

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
      setError('加载SSH连接失败');
    } finally {
      setIsLoading(false);
    }
  };

  // 处理表单输入变化
  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : type === 'number' ? parseInt(value) || 22 : value
    }));
    setError('');
  };

  // 打开添加连接模态框
  const openAddModal = () => {
    setEditingConn(null);
    setFormData({
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
    setFormData({
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
  };

  // 测试SSH连接
  const testConnection = async () => {
    setIsTesting(true);
    setTestResult(null);
    setError('');

    try {
      // 创建临时连接进行测试
      const tempConn = {
        ...formData,
        id: 'temp-test-connection',
      };

      // 调用添加连接API（会自动测试连接）
      await sshApi.addConnection(tempConn);
      
      // 测试成功
      setTestResult({ success: true, message: '连接测试成功' });
    } catch (err) {
      // 测试失败
      setTestResult({ success: false, message: err.message || '连接测试失败' });
    } finally {
      setIsTesting(false);
    }
  };

  // 保存SSH连接
  const saveConnection = async () => {
    setError('');

    // 表单验证
    if (!formData.name || !formData.host || !formData.username || formData.port === 0) {
      setError('请填写所有必填字段');
      return;
    }

    try {
      if (editingConn) {
        // 更新现有连接
        await sshApi.updateConnection(editingConn.id, formData);
      } else {
        // 添加新连接
        await sshApi.addConnection(formData);
      }

      // 关闭模态框并刷新列表
      closeModal();
      loadConnections();
    } catch (err) {
      setError(err.message || '保存SSH连接失败');
    }
  };

  // 删除SSH连接
  const deleteConnection = async (connId, connName) => {
    if (window.confirm(`确定要删除SSH连接 "${connName}" 吗？`)) {
      try {
        await sshApi.deleteConnection(connId);
        loadConnections();
      } catch (err) {
        console.error('删除SSH连接失败:', err);
        setError('删除SSH连接失败');
      }
    }
  };

  // 切换连接状态
  const toggleConnectionStatus = async (conn) => {
    try {
      await sshApi.updateConnection(conn.id, {
        ...conn,
        isActive: !conn.isActive,
      });
      loadConnections();
    } catch (err) {
      console.error('更新连接状态失败:', err);
      setError('更新连接状态失败');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-xl text-gray-600">加载中...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-800">SSH连接管理</h1>
          <p className="mt-1 text-gray-600">管理您的SSH连接配置</p>
        </div>
        <button
          onClick={openAddModal}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors flex items-center"
        >
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          添加连接
        </button>
      </div>

      {error && (
        <div className="p-3 text-red-700 bg-red-100 rounded-lg">
          {error}
        </div>
      )}

      {/* SSH连接列表 */}
      <div className="bg-white rounded-lg shadow-md overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">连接信息</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">最后使用</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {connections.map((conn) => (
              <tr key={conn.id} className="hover:bg-gray-50 transition-colors">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm font-medium text-gray-900">{conn.name}</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm text-gray-900">{conn.username}@{conn.host}:{conn.port}</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                    conn.isActive ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                  }`}>
                    {conn.isActive ? '活跃' : '非活跃'}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm text-gray-500">
                    {new Date(conn.last_used).toLocaleString()}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div className="flex items-center justify-end space-x-2">
                    <button
                      onClick={() => toggleConnectionStatus(conn)}
                      className="px-3 py-1 text-xs text-blue-600 bg-blue-100 rounded-md hover:bg-blue-200 transition-colors"
                    >
                      {conn.isActive ? '停用' : '启用'}
                    </button>
                    <button
                      onClick={() => openEditModal(conn)}
                      className="px-3 py-1 text-xs text-blue-600 bg-blue-100 rounded-md hover:bg-blue-200 transition-colors"
                    >
                      编辑
                    </button>
                    <button
                      onClick={() => deleteConnection(conn.id, conn.name)}
                      className="px-3 py-1 text-xs text-red-600 bg-red-100 rounded-md hover:bg-red-200 transition-colors"
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        
        {/* 空状态 */}
        {connections.length === 0 && (
          <div className="px-6 py-12 text-center">
            <svg className="w-12 h-12 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">没有SSH连接</h3>
            <p className="mt-1 text-sm text-gray-500">点击"添加连接"按钮创建您的第一个SSH连接</p>
          </div>
        )}
      </div>

      {/* 添加/编辑SSH连接模态框 */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            {/* 模态框头部 */}
            <div className="px-6 py-4 border-b">
              <h2 className="text-xl font-bold text-gray-800">
                {editingConn ? '编辑SSH连接' : '添加SSH连接'}
              </h2>
            </div>

            {/* 模态框内容 */}
            <div className="px-6 py-4">
              <form className="space-y-4">
                {/* 名称 */}
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                    连接名称 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    id="name"
                    name="name"
                    value={formData.name}
                    onChange={handleChange}
                    className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    placeholder="输入连接名称"
                  />
                </div>

                {/* 主机信息 */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="host" className="block text-sm font-medium text-gray-700">
                      主机地址 <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      id="host"
                      name="host"
                      value={formData.host}
                      onChange={handleChange}
                      className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="输入主机地址"
                    />
                  </div>
                  <div>
                    <label htmlFor="port" className="block text-sm font-medium text-gray-700">
                      端口 <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="number"
                      id="port"
                      name="port"
                      value={formData.port}
                      onChange={handleChange}
                      className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="22"
                    />
                  </div>
                </div>

                {/* 认证信息 */}
                <div>
                  <h3 className="text-sm font-medium text-gray-700 mb-3">认证信息</h3>
                  
                  {/* 用户名 */}
                  <div className="mb-4">
                    <label htmlFor="username" className="block text-sm font-medium text-gray-700">
                      用户名 <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      id="username"
                      name="username"
                      value={formData.username}
                      onChange={handleChange}
                      className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="输入用户名"
                    />
                  </div>

                  {/* 密码 */}
                  <div className="mb-4">
                    <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                      密码
                    </label>
                    <input
                      type="password"
                      id="password"
                      name="password"
                      value={formData.password}
                      onChange={handleChange}
                      className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="输入密码（可选）"
                    />
                  </div>

                  {/* 私钥 */}
                  <div className="mb-4">
                    <label htmlFor="privateKey" className="block text-sm font-medium text-gray-700">
                      私钥
                    </label>
                    <textarea
                      id="privateKey"
                      name="privateKey"
                      value={formData.privateKey}
                      onChange={handleChange}
                      rows={5}
                      className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="输入SSH私钥（可选）"
                    ></textarea>
                  </div>

                  {/* 私钥密码短语 */}
                  <div>
                    <label htmlFor="passphrase" className="block text-sm font-medium text-gray-700">
                      私钥密码短语
                    </label>
                    <input
                      type="password"
                      id="passphrase"
                      name="passphrase"
                      value={formData.passphrase}
                      onChange={handleChange}
                      className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="输入私钥密码短语（可选）"
                    />
                  </div>
                </div>

                {/* 连接状态 */}
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="isActive"
                    name="isActive"
                    checked={formData.isActive}
                    onChange={handleChange}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <label htmlFor="isActive" className="ml-2 block text-sm text-gray-700">
                    设为活跃连接
                  </label>
                </div>

                {/* 连接测试结果 */}
                {testResult && (
                  <div className={`p-3 rounded-lg ${
                    testResult.success ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                  }`}>
                    {testResult.message}
                  </div>
                )}
              </form>
            </div>

            {/* 模态框底部 */}
            <div className="px-6 py-4 bg-gray-50 border-t flex justify-between">
              <div className="flex space-x-2">
                <button
                  onClick={testConnection}
                  disabled={isTesting || !formData.name || !formData.host || !formData.username}
                  className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                    isTesting || !formData.name || !formData.host || !formData.username
                      ? 'text-gray-400 bg-gray-100 cursor-not-allowed'
                      : 'text-blue-600 bg-blue-100 hover:bg-blue-200'
                  }`}
                >
                  {isTesting ? '测试中...' : '测试连接'}
                </button>
              </div>
              <div className="flex space-x-3">
                <button
                  onClick={closeModal}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
                >
                  取消
                </button>
                <button
                  onClick={saveConnection}
                  disabled={!formData.name || !formData.host || !formData.username}
                  className={`px-4 py-2 text-sm font-medium text-white rounded-md transition-colors ${
                    !formData.name || !formData.host || !formData.username
                      ? 'bg-blue-400 cursor-not-allowed'
                      : 'bg-blue-600 hover:bg-blue-700'
                  }`}
                >
                  保存
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SSHConnections;
