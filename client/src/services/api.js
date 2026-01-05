import axios from 'axios';
import { message } from 'antd';
import { isElectron, getIpcRenderer } from '../utils/electron';

// 创建axios实例
const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 30000, // 30秒超时
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加Authorization头
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    if (error.response) {
      // 服务器返回错误状态码
      const { status, data } = error.response;
      
      // 处理401未授权 - 清除token并重定向到登录页
      if (status === 401) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
      }
      
      return Promise.reject({ status, message: data.error || '请求失败' });
    } else if (error.request) {
      // 请求已发出但没有收到响应
      return Promise.reject({ message: '网络错误，请检查连接' });
    } else {
      // 请求配置错误
      return Promise.reject({ message: '请求配置错误' });
    }
  }
);

// 认证相关API
export const authApi = {
  // 用户登录
  login: (credentials) => api.post('/auth/login', credentials),
  
  // 用户注册
  register: (userData) => api.post('/auth/register', userData),
  
  // 获取当前用户信息
  getMe: () => api.get('/auth/me'),
  
  // 更新个人信息
  updateProfile: (userData) => api.put('/profile', userData),
};

// 连接相关API
export const connectionApi = {
	// 获取所有连接
	getConnections: () => api.get('/connections'),
	
	// 获取单个连接
	getConnection: (id) => api.get(`/connections/${id}`),
	
	// 添加连接
	addConnection: (connectionData) => api.post('/connections', connectionData),
	
	// 更新连接
	updateConnection: (id, connectionData) => api.put(`/connections/${id}`, connectionData),
	
	// 删除连接
	deleteConnection: (id) => api.delete(`/connections/${id}`),
	
	// 测试已保存的连接
	testConnection: (id) => api.post(`/connections/${id}/test`),
	
	// 测试未保存的连接
	testConnectionDirect: (connectionData) => api.post('/connections/test', connectionData),
	
	// 获取所有分类
	getCategories: () => api.get('/categories'),
	
	// 添加分类
	addCategory: (name) => api.post('/categories', { name }),
	
	// 更新分类
	updateCategory: (oldName, newName) => api.put(`/categories/${oldName}`, { new_name: newName }),
	
	// 删除分类
	deleteCategory: (name) => api.delete(`/categories/${name}`),
};

// 保留旧的sshApi名称，保持向后兼容
export const sshApi = connectionApi;

// 任务管理相关API
export const taskApi = {
  // 获取所有任务
  getTasks: () => api.get('/tasks'),
  
  // 获取单个任务状态
  getTaskStatus: (taskId) => api.get(`/tasks/${taskId}`),
  
  // 取消任务
  cancelTask: (taskId) => api.post(`/tasks/${taskId}/cancel`),
  
  // 删除任务
  deleteTask: (taskId) => api.delete(`/tasks/${taskId}`),
};

// 配置管理相关API
export const configApi = {
  // 获取默认下载目录
  getDownloadDir: () => api.get('/config/download-dir'),
  
  // 设置默认下载目录
  setDownloadDir: (downloadDir) => api.post('/config/download-dir', { downloadDir }),
  
  // 获取网站名称
  getSiteName: () => api.get('/config/site-name'),
  
  // 设置网站名称
  setSiteName: (siteName) => api.post('/config/site-name', { siteName }),
};

// 用户管理相关API
export const userApi = {
  // 获取所有用户
  getUsers: () => api.get('/users'),
  
  // 创建用户
  createUser: (userData) => api.post('/users', userData),
  
  // 更新用户
  updateUser: (userId, userData) => api.put(`/users/${userId}`, userData),
  
  // 删除用户
  deleteUser: (userId) => api.delete(`/users/${userId}`),
  
  // 获取用户权限
  getUserPermissions: (userId) => api.get(`/permissions/user/${userId}`),
  
  // 授予用户连接权限
  grantPermission: (permissionData) => api.post('/permissions', permissionData),
  
  // 撤销用户连接权限
  revokePermission: (permissionId) => api.delete(`/permissions/${permissionId}`),
};

// 通用下载处理函数
const handleFileDownload = async (taskId, token) => {
  console.log('handleFileDownload called with taskId:', taskId);
  
  // 检查当前环境
  const electronEnv = isElectron();
  console.log('Is Electron environment:', electronEnv);
  
  if (electronEnv) {
    // Electron环境：使用IPC调用主进程下载到默认目录
    console.log('In Electron environment, preparing to download via IPC');
    const ipcRenderer = getIpcRenderer();
    console.log('ipcRenderer:', ipcRenderer);
    
    if (ipcRenderer) {
      console.log('Invoking download-file IPC with taskId:', taskId);
      try {
        const result = await ipcRenderer.invoke('download-file', taskId, token);
        console.log('download-file IPC result:', result);
        
        if (result.success) {
          console.log(`文件已下载到: ${result.filePath}`);
          // 可以在这里显示通知
        } else {
          console.error('Electron下载失败:', result.error);
        }
      } catch (error) {
        console.error('Error invoking download-file IPC:', error);
      }
    } else {
      console.error('ipcRenderer is null, cannot download via IPC');
      // 降级到Web环境下载
      console.log('Falling back to Web download');
      await webDownload(taskId, token);
    }
  } else {
    // Web环境：触发浏览器下载
    console.log('In Web environment, preparing to download via fetch');
    await webDownload(taskId, token);
  }
};

// Web环境下载函数
const webDownload = async (taskId, token) => {
  try {
    // 使用相对URL，让浏览器自动处理域名和端口
    const downloadUrl = `/api/tasks/${taskId}/download`;
    console.log('Web download URL:', downloadUrl);
    
    // 使用window.location.origin来构建完整URL，确保在不同环境下都能正确访问
    const fullUrl = `${window.location.origin}${downloadUrl}`;
    console.log('Full download URL:', fullUrl);
    
    const downloadResponse = await fetch(fullUrl, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
    
    console.log('Web download response status:', downloadResponse.status);
    
    if (downloadResponse.ok) {
      // 获取文件名称
      const contentDisposition = downloadResponse.headers.get('Content-Disposition');
      let fileName = 'downloaded-file';
      console.log('Web download Content-Disposition:', contentDisposition);
      
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="([^"]+)"/);
        if (match && match[1]) {
          fileName = match[1];
          console.log('Web download extracted filename:', fileName);
        }
      }
      
      // 获取文件内容
      const blob = await downloadResponse.blob();
      console.log('Web download blob size:', blob.size);
      
      // 创建临时URL，触发浏览器下载
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      console.log('Web download completed:', fileName);
    } else {
      console.error('Web download failed with status:', downloadResponse.status);
      // 尝试获取错误信息
      try {
        const errorText = await downloadResponse.text();
        console.error('Web download error:', errorText);
      } catch (error) {
        console.error('Failed to get error text:', error);
      }
    }
  } catch (error) {
    console.error('Web download failed:', error);
  }
};

// 文件操作相关API
export const fileApi = {
  // 列出文件
  listFiles: (connId, path) => api.get('/files/list', { params: { connId, path } }),
  
  // 获取文件树
  getFileTree: (connId, path, depth = 1) => api.get('/files/tree', { params: { connId, path, depth } }),
  
  // 创建目录
  createDirectory: (connId, path) => api.post('/files/mkdir', { connId, path }),
  
  // 上传文件
  uploadFile: (connId, path, file) => {
    const formData = new FormData();
    formData.append('connId', connId);
    formData.append('path', path);
    formData.append('file', file);
    return api.post('/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        console.log(`Upload progress: ${percentCompleted}%`);
      },
    });
  },
  
  // 下载文件
  downloadFile: async (connId, path) => {
    // 检查是否已设置默认下载目录
    let hasDefaultDir = localStorage.getItem('defaultDownloadDir');
    if (!hasDefaultDir) {
      try {
        // 尝试从后端获取默认目录
        const dirResponse = await configApi.getDownloadDir();
        if (dirResponse.downloadDir) {
          localStorage.setItem('defaultDownloadDir', dirResponse.downloadDir);
          hasDefaultDir = dirResponse.downloadDir;
        }
      } catch (error) {
        console.error('获取默认下载目录失败:', error);
      }
    }
    
    // 如果仍然没有默认目录，提示用户设置
    if (!hasDefaultDir) {
      message.error('请先在设置中设置默认下载目录');
      // 跳转到设置页面
      window.location.href = '/settings';
      return Promise.reject(new Error('请先设置默认下载目录'));
    }
    
    const token = localStorage.getItem('token');
    try {
      // 发送请求，获取JSON响应
      const response = await fetch(`http://localhost:8080/api/files/download/${connId}${path}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      // 解析JSON响应
      const data = await response.json();
      
      // 如果返回的是任务ID，说明创建了下载任务
      if (data && data.taskId) {
        // 轮询任务状态，直到完成
        const checkTaskStatus = async () => {
          const taskResponse = await taskApi.getTaskStatus(data.taskId);
          if (taskResponse.status === 'completed') {
            // 任务完成，下载文件
            await handleFileDownload(data.taskId, token);
          } else if (taskResponse.status === 'failed') {
            throw new Error('下载任务失败');
          } else {
            // 继续轮询
            setTimeout(checkTaskStatus, 1000);
          }
        };
        
        // 开始轮询
        checkTaskStatus();
      }
      
      return data;
    } catch (error) {
      console.error('下载失败:', error);
      throw error;
    }
  },
  
  // 批量下载文件
  downloadFiles: async (connId, paths, onProgress) => {
    // 检查是否已设置默认下载目录
    let hasDefaultDir = localStorage.getItem('defaultDownloadDir');
    if (!hasDefaultDir) {
      try {
        // 尝试从后端获取默认目录
        const dirResponse = await configApi.getDownloadDir();
        if (dirResponse.downloadDir) {
          localStorage.setItem('defaultDownloadDir', dirResponse.downloadDir);
          hasDefaultDir = dirResponse.downloadDir;
        }
      } catch (error) {
        console.error('获取默认下载目录失败:', error);
      }
    }
    
    // 如果仍然没有默认目录，提示用户设置
    if (!hasDefaultDir) {
      message.error('请先在设置中设置默认下载目录');
      // 跳转到设置页面
      window.location.href = '/settings';
      return Promise.reject(new Error('请先设置默认下载目录'));
    }
    
    const token = localStorage.getItem('token');
    try {
      const response = await api.post('/files/download/batch', 
        { connId, paths }, 
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          }
        }
      );
      
      // 如果返回的是任务ID，说明创建了压缩任务
      if (response && response.taskId) {
        // 轮询任务状态，直到完成
        const checkTaskStatus = async () => {
          const taskResponse = await taskApi.getTaskStatus(response.taskId);
          if (taskResponse.status === 'completed') {
            // 任务完成，下载文件
            await handleFileDownload(response.taskId, token);
          } else if (taskResponse.status === 'failed') {
            throw new Error('批量下载任务失败');
          } else {
            // 更新进度
            if (onProgress && taskResponse.progress !== undefined) {
              onProgress(taskResponse.progress);
            }
            // 继续轮询
            setTimeout(checkTaskStatus, 1000);
          }
        };
        
        // 开始轮询
        checkTaskStatus();
      }
      
      return response;
    } catch (error) {
      console.error('批量下载失败:', error);
      throw error;
    }
  },
  
  // 删除文件
  deleteFiles: (connId, paths) => api.delete('/files/delete', { data: { connId, paths } }),
  
  // 重命名文件
  renameFile: (connId, oldPath, newPath) => api.put('/files/rename', { connId, oldPath, newPath }),
  
  // 复制文件
  copyFiles: (connId, srcPaths, destPath) => api.post('/files/copy', { connId, srcPaths, destPath }),
  
  // 移动文件
  moveFiles: (connId, srcPaths, destPath) => api.post('/files/move', { connId, srcPaths, destPath }),
  
  // 获取文件内容
  getFileContent: (connId, path) => api.get('/files/content', { params: { connId, path } }),
  
  // 保存文件内容
  saveFileContent: (connId, path, content) => api.put('/files/content', { connId, path, content }),
};

export default api;