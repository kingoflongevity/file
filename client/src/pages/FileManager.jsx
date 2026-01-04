import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { fileApi, sshApi } from '../services/api';

const FileManager = () => {
  const { connId } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  
  // 获取当前路径，从路由参数中提取
  const currentPath = location.pathname.replace(`/files/${connId}`, '') || '/';
  
  const [files, setFiles] = useState([]);
  const [filteredFiles, setFilteredFiles] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [sshConn, setSshConn] = useState(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [selectedFiles, setSelectedFiles] = useState([]);
  const [isCreateDirModalOpen, setIsCreateDirModalOpen] = useState(false);
  const [newDirName, setNewDirName] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState('');
  
  // 过滤文件
  useEffect(() => {
    if (!searchQuery.trim()) {
      setFilteredFiles(files);
    } else {
      const query = searchQuery.toLowerCase();
      setFilteredFiles(files.filter(file => 
        file.name.toLowerCase().includes(query)
      ));
    }
  }, [files, searchQuery]);

  useEffect(() => {
    // 加载SSH连接信息
    loadSSHConnection();
    
    // 加载当前目录文件列表
    loadFiles();
  }, [connId, currentPath]);

  // 加载SSH连接信息
  const loadSSHConnection = async () => {
    try {
      const conn = await sshApi.getConnection(connId);
      setSshConn(conn);
    } catch (err) {
      console.error('加载SSH连接失败:', err);
      setError('加载SSH连接失败');
    }
  };

  // 加载当前目录文件列表
  const loadFiles = async () => {
    try {
      setIsLoading(true);
      setError('');
      const fileList = await fileApi.listFiles(connId, currentPath);
      setFiles(fileList);
    } catch (err) {
      console.error('加载文件列表失败:', err);
      setError(err.message || '加载文件列表失败');
    } finally {
      setIsLoading(false);
    }
  };

  // 处理目录导航
  const navigateToPath = useCallback((path) => {
    // 确保路径以/开头
    const fullPath = path.startsWith('/') ? path : `/${path}`;
    navigate(`/files/${connId}${fullPath}`);
  }, [connId, navigate]);

  // 返回上一级目录
  const goUp = () => {
    if (currentPath === '/') return;
    const parentPath = currentPath.substring(0, currentPath.lastIndexOf('/')) || '/';
    navigateToPath(parentPath);
  };

  // 处理文件点击
  const handleFileClick = (file) => {
    if (file.is_dir) {
      // 进入目录
      navigateToPath(file.path);
    } else {
      // 处理文件点击（可以实现预览或下载）
      handleDownload(file);
    }
  };

  // 处理文件选择
  const handleFileSelect = (file) => {
    setSelectedFiles(prev => {
      // 检查是否已选中
      const isSelected = prev.some(f => f.path === file.path);
      if (isSelected) {
        // 取消选择
        return prev.filter(f => f.path !== file.path);
      } else {
        // 选择文件
        return [...prev, file];
      }
    });
  };

  // 处理全选
  const handleSelectAll = () => {
    if (selectedFiles.length === filteredFiles.length) {
      // 取消全选
      setSelectedFiles([]);
    } else {
      // 全选
      setSelectedFiles([...filteredFiles]);
    }
  };

  // 创建目录
  const createDirectory = async () => {
    if (!newDirName.trim()) {
      setError('目录名称不能为空');
      return;
    }

    try {
      const dirPath = `${currentPath === '/' ? '' : currentPath}/${newDirName.trim()}`;
      await fileApi.createDirectory(connId, dirPath);
      setIsCreateDirModalOpen(false);
      setNewDirName('');
      loadFiles(); // 刷新文件列表
    } catch (err) {
      console.error('创建目录失败:', err);
      setError(err.message || '创建目录失败');
    }
  };

  // 删除文件/目录
  const handleDelete = async () => {
    if (selectedFiles.length === 0) {
      setError('请先选择要删除的文件或目录');
      return;
    }

    if (window.confirm(`确定要删除选中的 ${selectedFiles.length} 个项目吗？`)) {
      try {
        const paths = selectedFiles.map(file => file.path);
        await fileApi.deleteFiles(connId, paths);
        setSelectedFiles([]);
        loadFiles(); // 刷新文件列表
      } catch (err) {
        console.error('删除文件失败:', err);
        setError(err.message || '删除文件失败');
      }
    }
  };

  // 重命名文件/目录
  const handleRename = () => {
    if (selectedFiles.length !== 1) {
      setError('请选择一个文件或目录进行重命名');
      return;
    }

    const newName = prompt('请输入新名称:', selectedFiles[0].name);
    if (newName && newName.trim() !== selectedFiles[0].name) {
      renameFile(selectedFiles[0], newName.trim());
    }
  };

  // 执行重命名操作
  const renameFile = async (file, newName) => {
    try {
      const oldPath = file.path;
      const newPath = `${currentPath === '/' ? '' : currentPath}/${newName}`;
      await fileApi.renameFile(connId, oldPath, newPath);
      setSelectedFiles([]);
      loadFiles(); // 刷新文件列表
    } catch (err) {
      console.error('重命名文件失败:', err);
      setError(err.message || '重命名文件失败');
    }
  };

  // 处理文件下载
  const handleDownload = (file) => {
    try {
      fileApi.downloadFile(connId, file.path);
    } catch (err) {
      console.error('下载文件失败:', err);
      setError(err.message || '下载文件失败');
    }
  };

  // 处理文件上传
  const handleUpload = (e) => {
    const filesToUpload = Array.from(e.target.files);
    if (filesToUpload.length === 0) return;

    uploadFiles(filesToUpload);
  };

  // 执行文件上传
  const uploadFiles = async (filesToUpload) => {
    try {
      setIsUploading(true);
      setUploadProgress(0);
      
      // 逐个上传文件
      for (let i = 0; i < filesToUpload.length; i++) {
        const file = filesToUpload[i];
        await fileApi.uploadFile(connId, currentPath, file.name, file);
        
        // 更新上传进度
        const progress = Math.round(((i + 1) / filesToUpload.length) * 100);
        setUploadProgress(progress);
      }
      
      // 上传完成，刷新文件列表
      loadFiles();
      setUploadProgress(0);
    } catch (err) {
      console.error('上传文件失败:', err);
      setError(err.message || '上传文件失败');
    } finally {
      setIsUploading(false);
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
      {/* 页面标题和连接信息 */}
      <div>
        <h1 className="text-3xl font-bold text-gray-800">文件管理器</h1>
        {sshConn && (
          <p className="mt-1 text-gray-600">
            连接: {sshConn.name} ({sshConn.username}@{sshConn.host}:{sshConn.port})
          </p>
        )}
      </div>

      {/* 文件路径导航 */}
      <div className="bg-white p-4 rounded-lg shadow-md">
        <div className="flex items-center space-x-2">
          <button
            onClick={goUp}
            disabled={currentPath === '/'}
            className={`p-2 rounded-md transition-colors ${
              currentPath === '/' 
                ? 'text-gray-400 bg-gray-100 cursor-not-allowed' 
                : 'text-gray-600 bg-gray-100 hover:bg-gray-200'
            }`}
            title="返回上一级"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          
          {/* 路径面包屑 */}
          <div className="flex items-center space-x-2 overflow-x-auto">
            <span 
              className="px-3 py-1 text-sm font-medium text-blue-600 bg-blue-100 rounded-md cursor-pointer hover:bg-blue-200 transition-colors"
              onClick={() => navigateToPath('/')}
            >
              /
            </span>
            
            {currentPath !== '/' && (
              currentPath
                .split('/')
                .filter(segment => segment)
                .map((segment, index, segments) => {
                  const path = '/' + segments.slice(0, index + 1).join('/');
                  return (
                    <React.Fragment key={segment}>
                      <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                      <span 
                        className="px-3 py-1 text-sm font-medium text-blue-600 bg-blue-100 rounded-md cursor-pointer hover:bg-blue-200 transition-colors"
                        onClick={() => navigateToPath(path)}
                      >
                        {segment}
                      </span>
                    </React.Fragment>
                  );
                })
            )}
          </div>
        </div>
      </div>

      {error && (
        <div className="p-3 text-red-700 bg-red-100 rounded-lg">
          {error}
        </div>
      )}

      {/* 文件操作工具栏 */}
      <div className="bg-white p-4 rounded-lg shadow-md">
        <div className="flex flex-wrap items-center justify-between gap-4">
          {/* 搜索框 */}
          <div className="w-full md:w-auto">
            <div className="relative">
              <input
                type="text"
                placeholder="搜索文件..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full md:w-64 px-4 py-2 pr-10 text-sm border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              />
              <div className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
            </div>
          </div>
          
          {/* 左侧操作按钮 */}
          <div className="flex flex-wrap items-center gap-2">
            {/* 新建目录按钮 */}
            <button
              onClick={() => setIsCreateDirModalOpen(true)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors flex items-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              新建目录
            </button>

            {/* 上传文件按钮 */}
            <div className="relative">
              <input
                type="file"
                id="file-upload"
                multiple
                onChange={handleUpload}
                disabled={isUploading}
                className="absolute inset-0 opacity-0 cursor-pointer"
              />
              <button
                type="button"
                disabled={isUploading}
                className={`px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors flex items-center ${
                  isUploading ? 'cursor-not-allowed opacity-75' : ''
                }`}
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                上传文件
              </button>
            </div>

            {/* 选中文件数量 */}
            {selectedFiles.length > 0 && (
              <div className="px-4 py-2 text-sm font-medium text-gray-700 bg-blue-100 rounded-md">
                已选择 {selectedFiles.length} 个项目
              </div>
            )}
          </div>

          {/* 右侧操作按钮 */}
          <div className="flex flex-wrap items-center gap-2">
            {/* 重命名按钮 */}
            <button
              onClick={handleRename}
              disabled={selectedFiles.length !== 1}
              className={`px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors ${
                selectedFiles.length !== 1 ? 'cursor-not-allowed opacity-75' : ''
              }`}
            >
              重命名
            </button>

            {/* 删除按钮 */}
            <button
              onClick={handleDelete}
              disabled={selectedFiles.length === 0}
              className={`px-4 py-2 text-sm font-medium text-red-700 bg-red-100 rounded-md hover:bg-red-200 transition-colors ${
                selectedFiles.length === 0 ? 'cursor-not-allowed opacity-75' : ''
              }`}
            >
              删除
            </button>
          </div>
        </div>

        {/* 上传进度条 */}
        {isUploading && (
          <div className="mt-4">
            <div className="flex items-center justify-between mb-1">
              <span className="text-sm text-gray-600">上传进度</span>
              <span className="text-sm text-gray-600">{uploadProgress}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2.5">
              <div 
                className="bg-blue-600 h-2.5 rounded-full transition-all duration-300"
                style={{ width: `${uploadProgress}%` }}
              ></div>
            </div>
          </div>
        )}
      </div>

      {/* 文件列表 */}
      <div className="bg-white rounded-lg shadow-md overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={selectedFiles.length === filteredFiles.length && filteredFiles.length > 0}
                    onChange={handleSelectAll}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">大小</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">修改时间</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">权限</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {filteredFiles.map((file) => (
              <tr 
                key={file.path} 
                className={`hover:bg-gray-50 transition-colors ${
                  selectedFiles.some(f => f.path === file.path) ? 'bg-blue-50' : ''
                }`}
              >
                <td className="px-6 py-4 whitespace-nowrap">
                  <input
                    type="checkbox"
                    checked={selectedFiles.some(f => f.path === file.path)}
                    onChange={() => handleFileSelect(file)}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center cursor-pointer" onClick={() => handleFileClick(file)}>
                    {file.is_dir ? (
                      <svg className="w-6 h-6 text-yellow-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                      </svg>
                    ) : (
                      <svg className="w-6 h-6 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                    )}
                    <div className="font-medium text-gray-900">{file.name}</div>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm text-gray-500">
                    {file.is_dir ? '-' : formatFileSize(file.size)}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm text-gray-500">
                    {new Date(file.mod_time).toLocaleString()}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm text-gray-500">{file.permissions}</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button
                    onClick={() => handleDownload(file)}
                    className="text-blue-600 hover:text-blue-500"
                    title="下载"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                    </svg>
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* 空状态 */}
        {filteredFiles.length === 0 && (
          <div className="px-6 py-12 text-center">
            <svg className="w-12 h-12 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">
              {searchQuery ? '没有找到匹配的文件' : '当前目录为空'}
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              {searchQuery ? '尝试调整搜索条件' : '上传文件或创建新目录'}
            </p>
          </div>
        )}
      </div>

      {/* 新建目录模态框 */}
      {isCreateDirModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md">
            <div className="px-6 py-4 border-b">
              <h2 className="text-xl font-bold text-gray-800">新建目录</h2>
            </div>
            <div className="px-6 py-4">
              <div>
                <label htmlFor="dirName" className="block text-sm font-medium text-gray-700">
                  目录名称
                </label>
                <input
                  type="text"
                  id="dirName"
                  value={newDirName}
                  onChange={(e) => setNewDirName(e.target.value)}
                  className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  placeholder="输入目录名称"
                  autoFocus
                />
              </div>
              {error && (
                <div className="mt-2 text-sm text-red-600">{error}</div>
              )}
            </div>
            <div className="px-6 py-4 bg-gray-50 border-t flex justify-end space-x-3">
              <button
                onClick={() => {
                  setIsCreateDirModalOpen(false);
                  setNewDirName('');
                  setError('');
                }}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
              >
                取消
              </button>
              <button
                onClick={createDirectory}
                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors"
              >
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

export default FileManager;
