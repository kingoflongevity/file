// 检测是否在Electron环境中
export const isElectron = () => {
  try {
    // 简单可靠的检测方法，适用于Electron渲染进程
    if (typeof window !== 'undefined' && 
        typeof window.process !== 'undefined' && 
        typeof window.process.versions !== 'undefined' && 
        typeof window.process.versions.electron !== 'undefined') {
      console.log('Electron detected via window.process.versions.electron');
      return true;
    }
    
    // 备选方法：检查userAgent
    if (typeof window !== 'undefined' && 
        typeof window.navigator !== 'undefined' && 
        typeof window.navigator.userAgent !== 'undefined') {
      const userAgent = window.navigator.userAgent.toLowerCase();
      if (userAgent.indexOf(' electron/') > -1) {
        console.log('Electron detected via userAgent');
        return true;
      }
    }
    
    // 备选方法：尝试直接访问electron对象（如果预加载脚本暴露了）
    if (typeof window !== 'undefined' && typeof window.electron !== 'undefined') {
      console.log('Electron detected via window.electron');
      return true;
    }
    
    // 备选方法：尝试安全地require electron
    try {
      if (typeof window !== 'undefined' && typeof window.require === 'function') {
        const electron = window.require('electron');
        if (electron && electron.ipcRenderer) {
          console.log('Electron detected via window.require');
          return true;
        }
      }
    } catch (e) {
      // 捕获所有错误，确保函数不会崩溃
      console.log('Failed to detect Electron via window.require:', e.message);
    }
  } catch (error) {
    console.error('Error detecting Electron:', error);
  }
  
  // 默认为非Electron环境
  console.log('Not in Electron environment');
  return false;
};

// 获取Electron的ipcRenderer
export const getIpcRenderer = () => {
  try {
    if (isElectron() && typeof window !== 'undefined') {
      // 尝试不同的获取ipcRenderer的方法
      if (window.require) {
        console.log('Using window.require to get ipcRenderer');
        return window.require('electron').ipcRenderer;
      }
      
      // 如果window.require不可用，尝试使用global.require
      if (typeof global !== 'undefined' && global.require) {
        console.log('Using global.require to get ipcRenderer');
        return global.require('electron').ipcRenderer;
      }
      
      // 如果都不可用，尝试从window.electron获取（如果预加载脚本暴露了）
      if (window.electron && window.electron.ipcRenderer) {
        console.log('Using window.electron.ipcRenderer');
        return window.electron.ipcRenderer;
      }
    }
  } catch (error) {
    console.error('Error getting ipcRenderer:', error);
  }
  
  console.log('ipcRenderer not available');
  return null;
};
