const { app, BrowserWindow } = require('electron');
const path = require('path');
const url = require('url');
const { spawn } = require('child_process');
let serverProcess;

function createWindow() {
  // 创建浏览器窗口
  const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
    },
  });

  // 加载应用
  if (app.isPackaged) {
    // 生产环境：加载打包后的静态文件
    mainWindow.loadURL(
      url.format({
        pathname: path.join(__dirname, '../client/dist/index.html'),
        protocol: 'file:',
        slashes: true,
      })
    );
  } else {
    // 开发环境：加载开发服务器
    mainWindow.loadURL('http://localhost:5173');
    // 打开开发者工具
    mainWindow.webContents.openDevTools();
  }
}

// 启动服务器
function startServer() {
  // 根据平台选择服务器可执行文件
  let serverPath;
  if (process.platform === 'win32') {
    serverPath = path.join(__dirname, '../dist/server-windows.exe');
  } else if (process.platform === 'linux') {
    serverPath = path.join(__dirname, '../dist/server-linux');
  } else {
    console.error('Unsupported platform');
    return;
  }

  // 启动服务器进程
  serverProcess = spawn(serverPath, [], {
    stdio: 'inherit'
  });

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
  // 启动服务器
  startServer();
  
  // 创建浏览器窗口
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
