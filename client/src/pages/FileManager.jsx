import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { fileApi, sshApi } from '../services/api';
import { isElectron } from '../utils/electron';
import {
  Card, 
  Breadcrumb, 
  Button, 
  Table, 
  Checkbox, 
  Input, 
  Modal, 
  Form, 
  Progress, 
  Space, 
  Avatar, 
  Badge, 
  message, 
  Tooltip,
  Spin
} from 'antd';
import {
  ArrowLeftOutlined, 
  DashboardOutlined,
  PlusOutlined, 
  UploadOutlined, 
  DeleteOutlined, 
  EditOutlined, 
  CopyOutlined, 
  FolderOutlined, 
  FileOutlined, 
  DownloadOutlined, 
  MoreOutlined, 
  SearchOutlined
} from '@ant-design/icons';

const { Search } = Input;

const FileManager = () => {
  const { connId } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  
  // 获取当前路径，从路由参数中提取
  const currentPath = location.pathname.replace(`/files/${connId}`, '') || '/';
  
  // 响应式设计状态
  const [isMobile, setIsMobile] = useState(window.innerWidth < 576);
  const [windowWidth, setWindowWidth] = useState(window.innerWidth);
  const [tableHeight, setTableHeight] = useState(window.innerHeight - 350);
  
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
  const [form] = Form.useForm();
  // 下载相关状态
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [downloadingFiles, setDownloadingFiles] = useState([]);
  const [downloadPath, setDownloadPath] = useState('');
  
  // 监听窗口大小变化
  useEffect(() => {
    const handleResize = () => {
      const width = window.innerWidth;
      const height = window.innerHeight;
      setWindowWidth(width);
      setIsMobile(width < 576);
      setTableHeight(height - 350); // 更新表格高度
    };
    
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);
  
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
    const loadSSHConnection = async () => {
      try {
        const conn = await sshApi.getConnection(connId);
        setSshConn(conn);
      } catch (err) {
        console.error('加载SSH连接失败:', err);
        message.error('加载SSH连接失败');
      }
    };
    
    // 加载当前目录文件列表
    const loadFiles = async () => {
      try {
        setIsLoading(true);
        const fileList = await fileApi.listFiles(connId, currentPath);
        setFiles(fileList);
      } catch (err) {
        console.error('加载文件列表失败:', err);
        message.error(err.message || '加载文件列表失败');
      } finally {
        setIsLoading(false);
      }
    };
    
    loadSSHConnection();
    loadFiles();
  }, [connId, currentPath]);

  // 加载当前目录文件列表
  const loadFiles = async () => {
    try {
      setIsLoading(true);
      const fileList = await fileApi.listFiles(connId, currentPath);
      setFiles(fileList);
    } catch (err) {
      console.error('加载文件列表失败:', err);
      message.error(err.message || '加载文件列表失败');
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
      message.error('目录名称不能为空');
      return;
    }

    try {
      const dirPath = `${currentPath === '/' ? '' : currentPath}/${newDirName.trim()}`;
      await fileApi.createDirectory(connId, dirPath);
      setIsCreateDirModalOpen(false);
      setNewDirName('');
      loadFiles(); // 刷新文件列表
      message.success('目录创建成功');
    } catch (err) {
      console.error('创建目录失败:', err);
      message.error(err.message || '创建目录失败');
    }
  };

  // 删除文件/目录
  const handleDelete = async () => {
    if (selectedFiles.length === 0) {
      message.error('请先选择要删除的文件或目录');
      return;
    }

    Modal.confirm({
      title: '确认删除',
      content: `确定要删除选中的 ${selectedFiles.length} 个项目吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          const paths = selectedFiles.map(file => file.path);
          await fileApi.deleteFiles(connId, paths);
          setSelectedFiles([]);
          loadFiles(); // 刷新文件列表
          message.success('文件删除成功');
        } catch (err) {
          console.error('删除文件失败:', err);
          message.error(err.message || '删除文件失败');
        }
      }
    });
  };

  // 重命名文件/目录
  const handleRename = () => {
    if (selectedFiles.length !== 1) {
      message.error('请选择一个文件或目录进行重命名');
      return;
    }

    const file = selectedFiles[0];
    Modal.prompt({
      title: '重命名',
      placeholder: '请输入新名称',
      defaultValue: file.name,
      onOk: (newName) => {
        if (newName && newName.trim() !== file.name) {
          renameFile(file, newName.trim());
        }
      }
    });
  };

  // 执行重命名操作
  const renameFile = async (file, newName) => {
    try {
      const oldPath = file.path;
      const newPath = `${currentPath === '/' ? '' : currentPath}/${newName}`;
      await fileApi.renameFile(connId, oldPath, newPath);
      setSelectedFiles([]);
      loadFiles(); // 刷新文件列表
      message.success('文件重命名成功');
    } catch (err) {
      console.error('重命名文件失败:', err);
      message.error(err.message || '重命名文件失败');
    }
  };

  // 处理文件下载
  const handleDownload = async (file) => {
    try {
      // 添加到下载文件列表
      setDownloadingFiles(prev => [...prev, file.name]);
      
      const response = await fileApi.downloadFile(connId, file.path);
      
      // 如果返回的是任务ID，说明创建了下载任务
      if (response && response.taskId) {
        message.success('下载任务已创建，请查看通知');
        // 设置默认下载路径提示
        if (isElectron()) {
          setDownloadPath('默认下载目录');
        }
      } else {
        // 单个文件下载成功
        message.success('文件下载开始');
      }
      
      // 2秒后从下载列表中移除，模拟下载完成
      setTimeout(() => {
        setDownloadingFiles(prev => prev.filter(name => name !== file.name));
        if (downloadingFiles.length <= 1) {
          setDownloadPath('');
        }
      }, 2000);
    } catch (err) {
      console.error('下载文件失败:', err);
      message.error(err.message || '下载文件失败');
      // 从下载列表中移除
      setDownloadingFiles(prev => prev.filter(name => name !== file.name));
      if (downloadingFiles.length <= 1) {
        setDownloadPath('');
      }
    }
  };

  // 处理批量下载
  const handleBatchDownload = async () => {
    if (selectedFiles.length === 0) {
      message.error('请先选择要下载的文件');
      return;
    }

    try {
      setIsDownloading(true);
      setDownloadProgress(0);
      
      // 添加到下载文件列表
      const fileNames = selectedFiles.map(file => file.name);
      setDownloadingFiles(prev => [...prev, ...fileNames]);
      
      // 设置默认下载路径提示
      if (isElectron()) {
        setDownloadPath('默认下载目录');
      }

      // 逐个下载文件，使用Promise.allSettled确保所有文件都被处理
      const downloadPromises = selectedFiles.map((file, index) => {
        return new Promise((resolve, reject) => {
          try {
            // 调用单个文件下载API
            fileApi.downloadFile(connId, file.path);
            // 更新进度
            setDownloadProgress(Math.round(((index + 1) / selectedFiles.length) * 100));
            resolve(file.name);
          } catch (err) {
            reject({ file: file.name, error: err });
          }
        });
      });

      // 等待所有下载操作完成
      const results = await Promise.allSettled(downloadPromises);

      // 统计成功和失败的下载
      const successful = results.filter(r => r.status === 'fulfilled').length;
      const failed = results.filter(r => r.status === 'rejected').length;

      // 显示结果
      let messageText = `成功下载 ${successful} 个文件`;
      if (failed > 0) {
        messageText += `，${failed} 个文件下载失败`;
      }
      message.success(messageText);
      
      // 2秒后从下载列表中移除，模拟下载完成
      setTimeout(() => {
        setDownloadingFiles(prev => prev.filter(name => !fileNames.includes(name)));
        setDownloadPath('');
      }, 2000);
    } catch (err) {
      console.error('批量下载失败:', err);
      message.error(err.message || '批量下载失败');
    } finally {
      setIsDownloading(false);
      setDownloadProgress(0);
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
      message.success('文件上传成功');
    } catch (err) {
      console.error('上传文件失败:', err);
      message.error(err.message || '上传文件失败');
    } finally {
      setIsUploading(false);
    }
  };

  // 构建面包屑数据
  const buildBreadcrumbItems = () => {
    const items = [{ title: <span onClick={() => navigateToPath('/')}>根目录</span> }];
    
    if (currentPath !== '/') {
      const segments = currentPath.split('/').filter(segment => segment);
      segments.forEach((segment, index) => {
        const path = '/' + segments.slice(0, index + 1).join('/');
        items.push({
          title: <span onClick={() => navigateToPath(path)}>{segment}</span>
        });
      });
    }
    
    return items;
  };

  // 表格列定义
  const columns = [
    {
      title: (
        <Checkbox
          checked={selectedFiles.length === filteredFiles.length && filteredFiles.length > 0}
          onChange={handleSelectAll}
        />
      ),
      dataIndex: 'select',
      key: 'select',
      width: 40,
      render: (_, record) => (
        <Checkbox
          checked={selectedFiles.some(f => f.path === record.path)}
          onChange={() => handleFileSelect(record)}
        />
      )
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar 
            icon={record.is_dir ? <FolderOutlined style={{ color: '#faad14' }} /> : <FileOutlined />} 
            style={{ marginRight: 8 }} 
          />
          <span 
            style={{ cursor: 'pointer' }}
            onClick={() => handleFileClick(record)}
          >
            {text}
          </span>
          {record.symlink && (
            <span style={{ marginLeft: 8, fontSize: 12, color: '#999' }}>
              → {record.symlink}
            </span>
          )}
        </div>
      )
    },
    {
      title: '大小',
      dataIndex: 'size',
      key: 'size',
      width: 100,
      render: (size, record) => (
        <span>{record.is_dir ? '-' : formatFileSize(size)}</span>
      )
    },
    {
      title: '修改时间',
      dataIndex: 'mod_time',
      key: 'mod_time',
      width: 180,
      render: (modTime) => (
        <span>{new Date(modTime).toLocaleString()}</span>
      )
    },
    {
      title: '权限',
      dataIndex: 'permissions',
      key: 'permissions',
      width: 120,
      render: (permissions) => (
        <span style={{ fontFamily: 'monospace' }}>{permissions}</span>
      )
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space size="middle">
          <Tooltip title="下载">
            <Button 
              type="text" 
              icon={<DownloadOutlined />} 
              onClick={() => handleDownload(record)}
              size="small"
            />
          </Tooltip>
          <Tooltip title="更多">
            <Button 
              type="text" 
              icon={<MoreOutlined />} 
              size="small"
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  return (
    <div style={{ padding: 0, height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* 合并的顶部卡片：标题、连接信息、导航和操作工具栏 */}
      <Card style={{ margin: 0, marginBottom: 16 }}>
        {/* 标题、连接信息和返回按钮 */}
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h1 style={{ fontSize: 24, fontWeight: 'bold', marginBottom: 8 }}>文件管理器</h1>
            {sshConn && (
              <p style={{ color: '#666', margin: 0 }}>
                连接: {sshConn.name} ({sshConn.username}@{sshConn.host}:{sshConn.port})
              </p>
            )}
          </div>
          <Button 
            type="default" 
            icon={<DashboardOutlined />} 
            onClick={() => navigate('/')}
          >
            返回首页
          </Button>
        </div>

        {/* 文件路径导航 */}
        <Breadcrumb style={{ marginBottom: 16 }}>
          <Breadcrumb.Item>
            <Button 
              type="text" 
              icon={<ArrowLeftOutlined />} 
              onClick={goUp} 
              disabled={currentPath === '/'} 
              style={{ marginRight: 8 }}
            />
          </Breadcrumb.Item>
          {buildBreadcrumbItems().map((item, index) => (
            <Breadcrumb.Item key={index}>
              {item.title}
            </Breadcrumb.Item>
          ))}
        </Breadcrumb>

        {/* 文件操作工具栏 */}
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'center' }}>
          {/* 搜索框 */}
          <div style={{ flex: 1, minWidth: 150, marginBottom: isMobile ? 8 : 0 }}>
            <Search
              placeholder="搜索文件..."
              allowClear
              enterButton={<SearchOutlined />}
              size="middle"
              onSearch={value => setSearchQuery(value)}
              onChange={e => setSearchQuery(e.target.value)}
            />
          </div>
          
          {/* 主要操作按钮组 */}
          <Space size={8} wrap>
            {/* 新建目录按钮 */}
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setIsCreateDirModalOpen(true)}
              size={isMobile ? "small" : "middle"}
            >
              {isMobile ? "" : "新建目录"}
            </Button>

            {/* 上传文件按钮 */}
            <label htmlFor="file-upload">
              <Button
                icon={<UploadOutlined />}
                disabled={isUploading}
                size={isMobile ? "small" : "middle"}
              >
                {isMobile ? "" : "上传文件"}
              </Button>
              <input
                type="file"
                id="file-upload"
                multiple
                onChange={handleUpload}
                disabled={isUploading}
                style={{ display: 'none' }}
              />
            </label>

            {/* 选中文件数量 */}
            {selectedFiles.length > 0 && (
              <Badge count={selectedFiles.length} style={{ backgroundColor: '#1890ff' }}>
                <Button type="default" size={isMobile ? "small" : "middle"}>
                  {isMobile ? "" : "已选择"}
                </Button>
              </Badge>
            )}

            {/* 重命名按钮 */}
            <Button
              icon={<EditOutlined />}
              onClick={handleRename}
              disabled={selectedFiles.length !== 1}
              size={isMobile ? "small" : "middle"}
              title="重命名"
            >
              {isMobile ? "" : "重命名"}
            </Button>

            {/* 复制按钮 */}
            <Button
              icon={<CopyOutlined />}
              disabled={selectedFiles.length === 0}
              size={isMobile ? "small" : "middle"}
              title="复制"
            >
              {isMobile ? "" : "复制"}
            </Button>

            {/* 批量下载按钮 */}
            <Button
              icon={<DownloadOutlined />}
              onClick={handleBatchDownload}
              disabled={selectedFiles.length === 0}
              size={isMobile ? "small" : "middle"}
              title="批量下载"
            >
              {isMobile ? "" : "批量下载"}
            </Button>

            {/* 删除按钮 */}
            <Button
              icon={<DeleteOutlined />}
              onClick={handleDelete}
              disabled={selectedFiles.length === 0}
              danger
              size={isMobile ? "small" : "middle"}
              title="删除"
            >
              {isMobile ? "" : "删除"}
            </Button>
          </Space>
        </div>

        {/* 上传进度条 */}
        {isUploading && (
          <div style={{ marginTop: 12 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
              <span>上传进度</span>
              <span>{uploadProgress}%</span>
            </div>
            <Progress percent={uploadProgress} status="active" />
          </div>
        )}
      </Card>

      {/* 文件列表 */}
      <Card style={{ flex: 1, display: 'flex', flexDirection: 'column', margin: 0 }}>
        <Spin spinning={isLoading}>
          <Table
            columns={columns}
            dataSource={filteredFiles}
            rowKey="path"
            pagination={false}
            scroll={{ x: 'max-content', y: `calc(100vh - 400px)` }}
            size="middle"
            locale={{
              emptyText: (
                <div style={{ textAlign: 'center', padding: 40 }}>
                  <FileOutlined style={{ fontSize: 48, color: '#ccc', marginBottom: 16 }} />
                  <h3>{searchQuery ? '没有找到匹配的文件' : '当前目录为空'}</h3>
                  <p>{searchQuery ? '尝试调整搜索条件' : '上传文件或创建新目录'}</p>
                </div>
              )
            }}
          />
        </Spin>
      </Card>

      {/* 下载状态显示 */}
      {(isDownloading || downloadingFiles.length > 0) && (
        <Card style={{ margin: 16, marginBottom: 0 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
            <span>下载进度</span>
            <span>{downloadProgress}%</span>
          </div>
          <Progress percent={downloadProgress} status="active" />
          
          {downloadingFiles.length > 0 && (
            <div style={{ marginTop: 8 }}>
              <div style={{ fontSize: 12, color: '#666', marginBottom: 4 }}>
                正在下载:
              </div>
              <div style={{ fontSize: 12, color: '#333', display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {downloadingFiles.map((name, index) => (
                  <span key={index} style={{ backgroundColor: '#f0f0f0', padding: '2px 8px', borderRadius: 12 }}>
                    {name}
                  </span>
                ))}
              </div>
            </div>
          )}
          
          {downloadPath && (
            <div style={{ marginTop: 8, fontSize: 12, color: '#1890ff' }}>
              下载位置: {downloadPath}
            </div>
          )}
        </Card>
      )}

      {/* 新建目录模态框 */}
      <Modal
        title="新建目录"
        open={isCreateDirModalOpen}
        onOk={createDirectory}
        onCancel={() => setIsCreateDirModalOpen(false)}
        okText="创建"
        cancelText="取消"
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="dirName"
            label="目录名称"
            rules={[{ required: true, message: '请输入目录名称!' }]}
          >
            <Input
              value={newDirName}
              onChange={(e) => setNewDirName(e.target.value)}
              placeholder="输入目录名称"
              autoFocus
            />
          </Form.Item>
        </Form>
      </Modal>
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