import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { fileApi, sshApi } from '../services/api';
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
  const handleDownload = (file) => {
    try {
      fileApi.downloadFile(connId, file.path);
      message.success('文件下载开始');
    } catch (err) {
      console.error('下载文件失败:', err);
      message.error(err.message || '下载文件失败');
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
        <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
          <Avatar 
            icon={record.is_dir ? <FolderOutlined style={{ color: '#faad14' }} /> : <FileOutlined />} 
            style={{ marginRight: 8 }} 
          />
          <span>{text}</span>
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
    <div style={{ padding: 24 }}>
      {/* 页面标题和连接信息 */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ marginBottom: 16 }}>
          <h1 style={{ fontSize: 24, fontWeight: 'bold', marginBottom: 8 }}>文件管理器</h1>
          {sshConn && (
            <p style={{ color: '#666' }}>
              连接: {sshConn.name} ({sshConn.username}@{sshConn.host}:{sshConn.port})
            </p>
          )}
        </div>

        {/* 文件路径导航 */}
        <Breadcrumb>
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
      </Card>

      {/* 文件操作工具栏 */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 16, alignItems: 'center', marginBottom: 16 }}>
          {/* 搜索框 */}
          <div style={{ flex: 1, minWidth: 200 }}>
            <Search
              placeholder="搜索文件..."
              allowClear
              enterButton={<SearchOutlined />}
              size="middle"
              onSearch={value => setSearchQuery(value)}
              onChange={e => setSearchQuery(e.target.value)}
            />
          </div>
          
          {/* 操作按钮 */}
          <Space>
            {/* 新建目录按钮 */}
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setIsCreateDirModalOpen(true)}
            >
              新建目录
            </Button>

            {/* 上传文件按钮 */}
            <label htmlFor="file-upload">
              <Button
                icon={<UploadOutlined />}
                disabled={isUploading}
              >
                上传文件
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
                <Button type="default">
                  已选择
                </Button>
              </Badge>
            )}
          </Space>

          {/* 批量操作按钮 */}
          <Space>
            {/* 重命名按钮 */}
            <Button
              icon={<EditOutlined />}
              onClick={handleRename}
              disabled={selectedFiles.length !== 1}
            >
              重命名
            </Button>

            {/* 复制按钮 */}
            <Button
              icon={<CopyOutlined />}
              disabled={selectedFiles.length === 0}
            >
              复制
            </Button>

            {/* 删除按钮 */}
            <Button
              icon={<DeleteOutlined />}
              onClick={handleDelete}
              disabled={selectedFiles.length === 0}
              danger
            >
              删除
            </Button>
          </Space>
        </div>

        {/* 上传进度条 */}
        {isUploading && (
          <Card size="small" style={{ marginTop: 16 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <span>上传进度</span>
              <span>{uploadProgress}%</span>
            </div>
            <Progress percent={uploadProgress} status="active" />
          </Card>
        )}
      </Card>

      {/* 文件列表 */}
      <Card>
        <Spin spinning={isLoading}>
          <Table
            columns={columns}
            dataSource={filteredFiles}
            rowKey="path"
            pagination={false}
            onRow={(record) => ({
              onClick: () => handleFileClick(record)
            })}
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