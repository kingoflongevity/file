import axios from 'axios';

// 创建axios实例
const api = axios.create({
  baseURL: 'http://localhost:8082/api',
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
    const token = localStorage.getItem('token');
    try {
      // 发送请求，获取JSON响应
      const response = await fetch(`http://localhost:8082/api/files/download/${connId}${path}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      // 解析JSON响应
      const data = await response.json();
      return data;
    } catch (error) {
      console.error('下载失败:', error);
      throw error;
    }
  },
  
  // 批量下载文件
  downloadFiles: async (connId, paths, onProgress) => {
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

export default api;
