import axios from 'axios';

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
};

// SSH连接相关API
export const sshApi = {
  // 获取所有SSH连接
  getConnections: () => api.get('/ssh/connections'),
  
  // 获取单个SSH连接
  getConnection: (id) => api.get(`/ssh/connections/${id}`),
  
  // 添加SSH连接
  addConnection: (connectionData) => api.post('/ssh/connections', connectionData),
  
  // 更新SSH连接
  updateConnection: (id, connectionData) => api.put(`/ssh/connections/${id}`, connectionData),
  
  // 删除SSH连接
  deleteConnection: (id) => api.delete(`/ssh/connections/${id}`),
  
  // 测试SSH连接
  testConnection: (id) => api.post(`/ssh/connections/${id}/test`),
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
  downloadFile: (connId, path) => {
    // 使用window.open直接下载，避免axios处理二进制数据的问题
    const token = localStorage.getItem('token');
    window.open(`http://localhost:8080/api/files/download/${connId}${path}?token=${token}`, '_blank');
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
