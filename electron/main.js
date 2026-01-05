const { app, BrowserWindow, ipcMain, dialog, shell } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');
const os = require('os');
let serverProcess;

// 默认下载目录
let defaultDownloadDir = path.join(os.homedir(), 'Downloads', 'RemoteFileManager');

// 确保默认下载目录存在
const ensureDownloadDir = () => {
  try {
    if (!fs.existsSync(defaultDownloadDir)) {
      console.log(`Creating download directory: ${defaultDownloadDir}`);
      fs.mkdirSync(defaultDownloadDir, { recursive: true });
      console.log(`Download directory created: ${defaultDownloadDir}`);
    }
  } catch (error) {
    console.error(`Failed to create download directory: ${error.message}`);
    // 尝试使用用户主目录下的Downloads目录作为备选
    const fallbackDir = path.join(os.homedir(), 'Downloads');
    console.log(`Trying fallback directory: ${fallbackDir}`);
    if (!fs.existsSync(fallbackDir)) {
      fs.mkdirSync(fallbackDir, { recursive: true });
    }
    defaultDownloadDir = fallbackDir;
    console.log(`Using fallback download directory: ${defaultDownloadDir}`);
  }
};

// 从后端获取默认下载目录
const fetchDefaultDownloadDir = async () => {
  try {
    console.log('Fetching default download directory from backend...');
    const response = await fetch('http://localhost:8080/api/config/download-dir');
    if (response.ok) {
      const data = await response.json();
      if (data.downloadDir) {
        defaultDownloadDir = data.downloadDir;
        console.log(`Fetched default download directory from backend: ${defaultDownloadDir}`);
        ensureDownloadDir();
      }
    } else {
      console.error(`Failed to fetch download directory from backend: ${response.status}`);
    }
  } catch (error) {
    console.error(`Error fetching download directory from backend: ${error.message}`);
    // 如果获取失败，使用默认目录
    ensureDownloadDir();
  }
};

// 初始化下载目录
ensureDownloadDir();
// 从后端获取最新配置
setTimeout(fetchDefaultDownloadDir, 3000); // 延迟3秒，确保服务器已启动

function createWindow() {
  // 创建浏览器窗口
  const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
    },
    // 隐藏顶部菜单栏
    autoHideMenuBar: true,
    // 设置应用图标
    icon: path.join(__dirname, '../build/file-icon.ico'),
  });

  // 加载应用
  if (app.isPackaged) {
    // 生产模式：连接到本地后端服务器，后端同时提供API和静态文件服务
    mainWindow.loadURL('http://localhost:8080');
  } else {
    // 开发环境：直接连接到 Vite 开发服务器
    mainWindow.loadURL('http://localhost:5173');
    // 打开开发者工具
    mainWindow.webContents.openDevTools();
  }
}

// 启动服务器
function startServer() {
  // 根据平台选择服务器可执行文件
  let serverPath;
  if (app.isPackaged) {
    // 生产环境：服务器在resources目录下
    if (process.platform === 'win32') {
      serverPath = path.join(process.resourcesPath, 'server-windows.exe');
    } else if (process.platform === 'linux') {
      serverPath = path.join(process.resourcesPath, 'server-linux');
    } else {
      console.error('Unsupported platform');
      return;
    }
  } else {
    // 开发环境：服务器在项目根目录的dist目录下
    if (process.platform === 'win32') {
      serverPath = path.join(__dirname, '../dist/server-windows.exe');
    } else if (process.platform === 'linux') {
      serverPath = path.join(__dirname, '../dist/server-linux');
    } else {
      console.error('Unsupported platform');
      return;
    }
  }

  // 启动服务器进程
  serverProcess = spawn(serverPath, [], {
    detached: true,
    stdio: 'ignore'
  });

  serverProcess.unref();

  serverProcess.on('error', (error) => {
    console.error('Failed to start server:', error);
  });

  serverProcess.on('close', (code) => {
    console.log(`Server process exited with code ${code}`);
  });
}

// 停止服务器
function stopServer() {
  if (serverProcess) {
    serverProcess.kill();
    serverProcess = null;
  }
}

// 当 Electron 完成初始化并准备创建浏览器窗口时调用此方法
app.whenReady().then(() => {
  // 启动后端服务器（提供 API 服务）
  startServer();
  
  // 创建浏览器窗口（连接 Vite 前端）
  createWindow();

  // 在 macOS 上，当点击 dock 图标并且没有其他窗口打开时，通常会在应用中重新创建一个窗口
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

// 关闭所有窗口后退出应用（Windows & Linux）
app.on('window-all-closed', () => {
  stopServer();
  if (process.platform !== 'darwin') app.quit();
});

// 应用退出时停止服务器
app.on('quit', () => {
  stopServer();
});

// IPC处理程序
ipcMain.handle('download-file', async (event, taskId, token) => {
  console.log(`Received download request for task: ${taskId}`);
  try {
    // 确保下载目录存在
    ensureDownloadDir();
    console.log(`Using download directory: ${defaultDownloadDir}`);
    
    // 下载文件
    const downloadUrl = `http://localhost:8080/api/tasks/${taskId}/download`;
    console.log(`Downloading from URL: ${downloadUrl}`);
    const response = await fetch(downloadUrl, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    // 获取文件名称
    const contentDisposition = response.headers.get('Content-Disposition');
    let fileName = `downloaded-file-${Date.now()}`;
    console.log(`Content-Disposition: ${contentDisposition}`);
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="([^"]+)"/);
      if (match && match[1]) {
        fileName = match[1];
        console.log(`Extracted filename: ${fileName}`);
      }
    }
    
    // 读取文件内容
    console.log('Reading file content...');
    const buffer = await response.arrayBuffer();
    console.log(`File content size: ${buffer.byteLength} bytes`);
    
    // 保存文件到默认下载目录
    const filePath = path.join(defaultDownloadDir, fileName);
    console.log(`Saving file to: ${filePath}`);
    
    // 写入文件
    fs.writeFileSync(filePath, Buffer.from(buffer));
    console.log(`File saved successfully: ${filePath}`);
    
    // 验证文件是否存在
    if (fs.existsSync(filePath)) {
      console.log(`File exists after write: ${filePath}`);
      const stats = fs.statSync(filePath);
      console.log(`File size after write: ${stats.size} bytes`);
    } else {
      console.error(`File does not exist after write: ${filePath}`);
    }
    
    return {
      success: true,
      filePath: filePath,
      fileName: fileName
    };
  } catch (error) {
    console.error('Download failed:', error);
    return {
      success: false,
      error: error.message
    };
  }
});

// 处理设置下载目录请求
ipcMain.handle('set-download-dir', async (event, newDir) => {
  try {
    // 验证目录是否存在，不存在则创建
    if (!fs.existsSync(newDir)) {
      fs.mkdirSync(newDir, { recursive: true });
    }
    
    // 更新默认下载目录
    defaultDownloadDir = newDir;
    
    // 向所有窗口发送目录更新事件
    for (const window of BrowserWindow.getAllWindows()) {
      window.webContents.send('download-dir-updated', defaultDownloadDir);
    }
    
    return {
      success: true,
      downloadDir: defaultDownloadDir
    };
  } catch (error) {
    console.error('Failed to set download directory:', error);
    return {
      success: false,
      error: error.message
    };
  }
});

// 处理获取下载目录请求
ipcMain.handle('get-download-dir', async (event) => {
  return {
    downloadDir: defaultDownloadDir
  };
});

// 处理打开下载目录请求
ipcMain.handle('open-download-dir', async (event) => {
  try {
    await shell.openPath(defaultDownloadDir);
    return {
      success: true
    };
  } catch (error) {
    console.error('Failed to open download directory:', error);
    return {
      success: false,
      error: error.message
    };
  }
});

// 处理选择下载目录请求
ipcMain.handle('select-download-dir', async (event) => {
  try {
    const result = await dialog.showOpenDialog({
      properties: ['openDirectory'],
      title: '选择默认下载目录',
      defaultPath: defaultDownloadDir
    });
    
    if (!result.canceled && result.filePaths && result.filePaths.length > 0) {
      // 更新默认下载目录
      defaultDownloadDir = result.filePaths[0];
      ensureDownloadDir();
      
      // 向所有窗口发送目录更新事件
      for (const window of BrowserWindow.getAllWindows()) {
        window.webContents.send('download-dir-updated', defaultDownloadDir);
      }
      
      return {
        success: true,
        downloadDir: defaultDownloadDir
      };
    } else {
      return {
        success: false,
        error: '用户取消了选择'
      };
    }
  } catch (error) {
    console.error('Failed to select download directory:', error);
    return {
      success: false,
      error: error.message
    };
  }
});
