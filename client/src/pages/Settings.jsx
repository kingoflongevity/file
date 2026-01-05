import React, { useState, useEffect } from 'react';
import { Card, Button, Form, Input, message, Space } from 'antd';
import { FolderOutlined, SaveOutlined, LeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { configApi } from '../services/api';
import { isElectron, getIpcRenderer } from '../utils/electron';

const Settings = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [downloadDir, setDownloadDir] = useState('');
  const [siteName, setSiteName] = useState('');
  const [loading, setLoading] = useState(false);
  
  // 组件挂载时获取当前下载目录和网站名称
  useEffect(() => {
    // 调试：打印当前环境
    console.log('isElectron:', isElectron());
    console.log('navigator.userAgent:', window.navigator.userAgent);
    
    getDownloadDir();
    getSiteName();
    
    // 监听默认目录变化
    const ipcRenderer = getIpcRenderer();
    if (ipcRenderer) {
      ipcRenderer.on('download-dir-updated', (event, newDir) => {
        setDownloadDir(newDir);
        form.setFieldsValue({ downloadDir: newDir });
        localStorage.setItem('defaultDownloadDir', newDir);
      });
    }
    
    return () => {
      // 清理监听器
      if (ipcRenderer) {
        ipcRenderer.removeAllListeners('download-dir-updated');
      }
    };
  }, []);
  
  // 获取当前下载目录
  const getDownloadDir = async () => {
    try {
      console.log('开始获取下载目录...');
      
      // 直接使用fetch API，绕过axios拦截器
      const token = localStorage.getItem('token');
      const response = await fetch('/api/config/download-dir', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      
      console.log('fetch响应状态:', response.status);
      
      if (response.ok) {
        const result = await response.json();
        console.log('fetch返回结果:', result);
        
        if (result && result.downloadDir) {
          console.log('设置下载目录:', result.downloadDir);
          setDownloadDir(result.downloadDir);
          
          // 使用setTimeout确保form已准备好
          setTimeout(() => {
            form.setFieldsValue({ downloadDir: result.downloadDir });
            console.log('表单已更新');
          }, 0);
          
          localStorage.setItem('defaultDownloadDir', result.downloadDir);
          
          // 同步到Electron主进程
          if (isElectron()) {
            const ipcRenderer = getIpcRenderer();
            if (ipcRenderer) {
              await ipcRenderer.invoke('set-download-dir', result.downloadDir);
            }
          }
        } else {
          console.log('fetch返回的downloadDir为空');
          // 尝试从localStorage获取
          const savedDir = localStorage.getItem('defaultDownloadDir');
          if (savedDir) {
            console.log('从localStorage获取下载目录:', savedDir);
            setDownloadDir(savedDir);
            setTimeout(() => {
              form.setFieldsValue({ downloadDir: savedDir });
            }, 0);
          } else {
            // 如果localStorage也没有，使用默认值
            const defaultDir = 'C:\\Users\\14280\\Downloads\\RemoteFileManager';
            console.log('使用默认下载目录:', defaultDir);
            setDownloadDir(defaultDir);
            setTimeout(() => {
              form.setFieldsValue({ downloadDir: defaultDir });
            }, 0);
            localStorage.setItem('defaultDownloadDir', defaultDir);
          }
        }
      } else {
        console.log('fetch请求失败，状态码:', response.status);
        // 尝试从localStorage获取
        const savedDir = localStorage.getItem('defaultDownloadDir');
        if (savedDir) {
          console.log('从localStorage获取下载目录:', savedDir);
          setDownloadDir(savedDir);
          setTimeout(() => {
            form.setFieldsValue({ downloadDir: savedDir });
          }, 0);
        }
      }
    } catch (apiError) {
      console.error('从API获取下载目录失败:', apiError);
      // API失败时，从localStorage获取
      const savedDir = localStorage.getItem('defaultDownloadDir');
      if (savedDir) {
        console.log('从localStorage获取下载目录:', savedDir);
        setDownloadDir(savedDir);
        setTimeout(() => {
          form.setFieldsValue({ downloadDir: savedDir });
        }, 0);
      } else {
        // 如果localStorage也没有，使用默认值
        const defaultDir = 'C:\\Users\\14280\\Downloads\\RemoteFileManager';
        console.log('使用默认下载目录:', defaultDir);
        setDownloadDir(defaultDir);
        setTimeout(() => {
          form.setFieldsValue({ downloadDir: defaultDir });
        }, 0);
        localStorage.setItem('defaultDownloadDir', defaultDir);
      }
    }
  };
  
  // 保存下载目录
  const handleSaveDir = async () => {
    try {
      console.log('开始保存下载目录...');
      setLoading(true);
      
      // 直接获取表单值，不使用validateFields，避免验证失败
      const newDir = form.getFieldValue('downloadDir');
      console.log('获取到的目录:', newDir);
      
      // 简单验证目录是否为空
      if (!newDir || newDir.trim() === '') {
        console.log('目录为空，显示错误信息');
        message.error('请输入有效的下载目录路径');
        return;
      }
      
      // 所有环境下都保存到后端API
      try {
        console.log('调用API保存目录...');
        const result = await configApi.setDownloadDir(newDir);
        console.log('API返回结果:', result);
        
        if (result && result.downloadDir) {
          console.log('API保存成功，更新状态...');
          setDownloadDir(result.downloadDir);
          localStorage.setItem('defaultDownloadDir', result.downloadDir);
          message.success('下载目录已保存');
          
          // 同步到Electron主进程
          if (isElectron()) {
            console.log('同步到Electron主进程...');
            const ipcRenderer = getIpcRenderer();
            if (ipcRenderer) {
              try {
                const ipcResult = await ipcRenderer.invoke('set-download-dir', result.downloadDir);
                console.log('Electron主进程同步结果:', ipcResult);
              } catch (ipcError) {
                console.error('Electron主进程同步失败:', ipcError);
              }
            }
          }
        } else {
          console.log('API返回结果无效');
          message.error('保存下载目录失败: API返回结果无效');
        }
      } catch (apiError) {
        console.error('通过API保存下载目录失败:', apiError);
        message.error(`保存下载目录失败: ${apiError.message}`);
      }
    } catch (error) {
      console.error('保存下载目录失败:', error);
      message.error(`保存下载目录失败: ${error.message}`);
    } finally {
      console.log('保存目录操作完成');
      setLoading(false);
    }
  };
  
  // 选择下载目录
  const handleSelectDir = async () => {
    try {
      setLoading(true);
      if (isElectron()) {
        const ipcRenderer = getIpcRenderer();
        if (ipcRenderer) {
          const result = await ipcRenderer.invoke('select-download-dir');
          if (result.success && result.downloadDir) {
            // 保存到后端API
            const apiResult = await configApi.setDownloadDir(result.downloadDir);
            if (apiResult.downloadDir) {
              setDownloadDir(apiResult.downloadDir);
              form.setFieldsValue({ downloadDir: apiResult.downloadDir });
              localStorage.setItem('defaultDownloadDir', apiResult.downloadDir);
              message.success('下载目录已选择');
            }
          }
        }
      } else {
        // Web环境下，提示用户手动输入目录
        message.info('请在输入框中手动输入默认下载目录');
      }
    } catch (error) {
      console.error('选择下载目录失败:', error);
      message.error('选择下载目录失败');
    } finally {
      setLoading(false);
    }
  };
  
  // 打开下载目录
  const handleOpenDir = async () => {
    try {
      if (isElectron()) {
        const ipcRenderer = getIpcRenderer();
        if (ipcRenderer) {
          const result = await ipcRenderer.invoke('open-download-dir');
          if (!result.success) {
            message.error('打开目录失败');
          }
        }
      } else {
        message.info('Web端无法直接打开目录');
      }
    } catch (error) {
      console.error('打开下载目录失败:', error);
      message.error('打开目录失败');
    }
  };
  
  // 获取网站名称
  const getSiteName = async () => {
    try {
      const response = await configApi.getSiteName();
      if (response && response.siteName) {
        setSiteName(response.siteName);
        form.setFieldsValue({ siteName: response.siteName });
        localStorage.setItem('siteName', response.siteName);
      }
    } catch (error) {
      console.error('获取网站名称失败:', error);
      // 从localStorage获取，如果没有则使用默认值
      const savedSiteName = localStorage.getItem('siteName') || '远程连接文件管理';
      setSiteName(savedSiteName);
      form.setFieldsValue({ siteName: savedSiteName });
    }
  };
  
  // 保存网站名称
  const handleSaveSiteName = async () => {
    try {
      setLoading(true);
      const newSiteName = form.getFieldValue('siteName');
      
      if (!newSiteName || newSiteName.trim() === '') {
        message.error('请输入有效的网站名称');
        return;
      }
      
      const result = await configApi.setSiteName(newSiteName);
      if (result && result.siteName) {
        setSiteName(result.siteName);
        localStorage.setItem('siteName', result.siteName);
        message.success('网站名称已保存');
        
        // 通知所有页面更新网站名称
        window.dispatchEvent(new CustomEvent('siteNameUpdated', { detail: result.siteName }));
      }
    } catch (error) {
      console.error('保存网站名称失败:', error);
      message.error(`保存网站名称失败: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Button
            type="text"
            icon={<LeftOutlined />}
            onClick={() => navigate(-1)}
            style={{ marginRight: 16 }}
          >
            返回
          </Button>
          <h1 style={{ margin: 0, fontSize: 24 }}>设置</h1>
        </div>
      </div>
      
      <Card title="下载设置" style={{ marginBottom: 24 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ downloadDir }}
        >
          <Form.Item
            label="默认下载目录"
            name="downloadDir"
            rules={[{ required: true, message: '请输入下载目录路径' }]}
            extra="所有客户端将使用此统一配置目录，修改后自动同步到所有客户端"
          >
            <Input
              prefix={<FolderOutlined />}
              placeholder="请输入默认下载目录"
              style={{ marginBottom: 16 }}
            />
          </Form.Item>
          
          <Space>
            <Button
              type="primary"
              icon={<FolderOutlined />}
              onClick={handleSelectDir}
              loading={loading}
              disabled={!isElectron()}
              title={isElectron() ? "选择下载目录" : "Web端无法直接选择目录"}
            >
              选择目录
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              onClick={handleSaveDir}
              loading={loading}
            >
              保存目录
            </Button>
            <Button
              icon={<FolderOutlined />}
              onClick={handleOpenDir}
              disabled={!isElectron()}
              title={isElectron() ? "打开下载目录" : "Web端无法直接打开目录"}
            >
              打开目录
            </Button>
          </Space>
        </Form>
      </Card>
      
      <Card title="主题设置" style={{ marginBottom: 24 }}>
        <div style={{ padding: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span>当前主题：</span>
            <span style={{ fontWeight: '500' }}>
              {window.themeContext?.isDarkTheme ? '暗色主题' : '亮色主题'}
            </span>
          </div>
          <div style={{ marginTop: 16 }}>
            <Button
              onClick={() => window.themeContext?.toggleTheme?.()}
              style={{ width: '100%' }}
            >
              {window.themeContext?.isDarkTheme ? '切换至亮色主题' : '切换至暗色主题'}
            </Button>
          </div>
        </div>
      </Card>
      
      <Card title="网站设置" style={{ marginBottom: 24 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ siteName }}
        >
          <Form.Item
            label="网站名称"
            name="siteName"
            rules={[{ required: true, message: '请输入网站名称' }]}
            extra="修改后将应用于整个网站的标题和页面标题"
          >
            <Input
              placeholder="请输入网站名称"
              style={{ marginBottom: 16 }}
            />
          </Form.Item>
          
          <Button
            type="primary"
            icon={<SaveOutlined />}
            onClick={handleSaveSiteName}
            loading={loading}
          >
            保存网站名称
          </Button>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;