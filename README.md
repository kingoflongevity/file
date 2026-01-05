<div align="right">
  <button id="lang-toggle" onclick="toggleLanguage()" style="background-color: #4CAF50; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;">English</button>
</div>

<script>
  // 默认显示中文
  let currentLang = 'zh';
  
  // 初始化页面
  document.addEventListener('DOMContentLoaded', function() {
    toggleLanguage();
  });
  
  // 语言切换函数
  function toggleLanguage() {
    const zhElements = document.querySelectorAll('.zh');
    const enElements = document.querySelectorAll('.en');
    const button = document.getElementById('lang-toggle');
    
    if (currentLang === 'zh') {
      // 切换到英文
      zhElements.forEach(el => el.style.display = 'none');
      enElements.forEach(el => el.style.display = 'block');
      button.textContent = '中文';
      currentLang = 'en';
    } else {
      // 切换到中文
      zhElements.forEach(el => el.style.display = 'block');
      enElements.forEach(el => el.style.display = 'none');
      button.textContent = 'English';
      currentLang = 'zh';
    }
  }
</script>

<div class="zh">
# 🌐 远程连接文件管理系统

## 📖 项目简介

远程连接文件管理系统是一款功能强大的跨平台文件管理工具，支持多种远程连接协议，提供直观的Web界面，方便用户管理远程服务器上的文件。

## 🖼️ 功能展示

### 如何添加项目截图
1. 在项目根目录创建 `screenshots` 文件夹
2. 将项目截图按照以下命名规则放入该文件夹：
   - `login.png` - 登录界面
   - `dashboard.png` - 仪表板
   - `connection-management.png` - 连接管理
   - `file-manager.png` - 文件管理
   - `task-management.png` - 任务管理
   - `settings.png` - 设置界面

### 添加截图到README
在下方添加截图markdown语法，例如：
```markdown
### 登录界面
![登录界面](screenshots/login.png)

### 仪表板
![仪表板](screenshots/dashboard.png)

### 连接管理
![连接管理](screenshots/connection-management.png)

### 文件管理
![文件管理](screenshots/file-manager.png)

### 任务管理
![任务管理](screenshots/task-management.png)

### 设置界面
![设置界面](screenshots/settings.png)
```

## 🔌 支持的服务器类型

### 1. SSH (Secure Shell) 🔒
- **协议介绍**：SSH是一种加密的网络传输协议，用于在不安全的网络上安全地进行远程登录和其他网络服务。
- **主要功能**：
  - 🔐 安全的远程命令执行
  - 📁 文件传输（通过SFTP或SCP）
  - 🔀 端口转发和隧道
  - 🔑 密钥认证支持
- **使用场景**：适用于需要安全远程访问Linux/Unix服务器的场景
- **连接要求**：需要知道服务器的IP地址、端口（默认22）、用户名和密码或SSH密钥

### 2. SFTP (SSH File Transfer Protocol) 📤
- **协议介绍**：SFTP是一种基于SSH的文件传输协议，提供加密的文件传输服务。
- **主要功能**：
  - 🔐 安全的文件上传和下载
  - 📁 文件和目录的创建、删除、重命名
  - 🔑 文件权限管理
  - ⏸️ 断点续传支持
- **使用场景**：适用于需要在本地和远程服务器之间安全传输文件的场景
- **连接要求**：需要知道服务器的IP地址、端口（默认22）、用户名和密码或SSH密钥

### 3. FTP (File Transfer Protocol) 📡
- **协议介绍**：FTP是一种传统的文件传输协议，用于在网络上进行文件传输。
- **主要功能**：
  - 📤 文件上传和下载
  - 📁 文件和目录管理
  - 📦 批量文件操作
- **使用场景**：适用于需要简单文件传输的场景
- **连接要求**：需要知道服务器的IP地址、端口（默认21）、用户名和密码

## 📋 系统要求

- **操作系统**：Windows 10/11 或 Linux（Ubuntu 18.04+、CentOS 7+等）
- **浏览器**：支持现代浏览器（Chrome 90+、Firefox 88+、Edge 90+等）
- **网络**：需要网络连接，以便访问Web界面

## 🚀 运行方式

### Windows系统

1. 📦 解压压缩包
2. 🖱️ 双击运行 `start.bat` 文件
3. 🎉 系统会自动启动服务器并打开浏览器
4. 🌐 在浏览器中访问 `http://localhost:8080`

### Linux系统

1. 📦 解压压缩包
2. 💻 打开终端，进入解压目录
3. 🔧 运行 `chmod +x start.sh` 赋予执行权限
4. 🚀 运行 `./start.sh` 启动系统
5. 🌐 在浏览器中访问 `http://localhost:8080`

## ⚙️ 手动运行方式

### 启动服务器

- **Windows**: 🖱️ 双击 `server-windows.exe`
- **Linux**: 💻 运行 `./server-linux`

### 访问系统

🌐 在浏览器中输入 `http://localhost:8080`

## 📁 项目结构

```
dist/
├── client/          # 前端静态文件 (Frontend static files) 🎨
├── data/            # 数据库文件 (Database files) 📊
├── logs/            # 日志文件 (Log files) 📝
├── permissions/     # 权限配置文件 (Permission configuration files) 🔒
├── users/           # 用户数据 (User data) 👥
├── server-windows.exe  # Windows服务器可执行文件 (Windows server executable) 💻
├── server-linux        # Linux服务器可执行文件 (Linux server executable) 🐧
├── start.bat           # Windows启动脚本 (Windows startup script) 🚀
├── start.sh            # Linux启动脚本 (Linux startup script) 🚀
└── README.md           # 说明文档 (Documentation) 📖
```

## 🔑 默认账号

- 👤 用户名: admin
- 🔒 密码: admin

## 📖 使用指南

### 1. 登录系统

1. 🌐 在浏览器中访问 `http://localhost:8080`
2. 🔑 输入默认用户名和密码（admin/admin）
3. 🖱️ 点击"登录"按钮

### 2. 添加远程连接

1. 📋 登录后，点击左侧菜单栏的"连接管理"
2. ➕ 点击"添加连接"按钮
3. 🔌 选择连接类型（SSH、SFTP或FTP）
4. 📝 填写连接信息：
   - 📛 **连接名称**：给连接起一个易于识别的名称
   - 🌐 **IP地址**：远程服务器的IP地址
   - 🔌 **端口**：连接端口（SSH/SFTP默认22，FTP默认21）
   - 👤 **用户名**：远程服务器的用户名
   - 🔒 **密码**：远程服务器的密码（如果使用密钥认证，可留空）
   - 🔑 **SSH密钥**：如果使用密钥认证，选择SSH私钥文件
5. 🧪 点击"测试连接"按钮，验证连接是否成功
6. 💾 点击"保存"按钮，保存连接配置

### 3. 管理远程文件

1. 📋 在左侧菜单栏的"连接管理"中，点击已添加的连接
2. 📁 进入文件管理界面，您可以：
   - 🔍 **浏览文件**：点击目录名称进入子目录
   - 📤 **上传文件**：点击"上传"按钮，选择本地文件上传到远程服务器
   - 📥 **下载文件**：勾选文件或目录，点击"下载"按钮
   - ➕ **创建文件/目录**：点击"新建"按钮，选择创建文件或目录
   - ✏️ **编辑文件**：点击文件名称，在编辑器中修改文件内容
   - 🗑️ **删除文件**：勾选文件或目录，点击"删除"按钮
   - 🔄 **重命名文件**：右键点击文件或目录，选择"重命名"
   - 📋 **复制/移动文件**：勾选文件或目录，点击"复制"或"移动"按钮

### 4. 管理任务

1. 📋 点击左侧菜单栏的"任务管理"
2. 📊 查看所有上传和下载任务的状态和进度
3. ⏯️ 可以暂停、继续或取消正在进行的任务
4. 📝 查看任务的详细日志

## ⚠️ 注意事项

1. 🎉 首次运行时会自动创建数据库和默认用户
2. 🔌 服务器默认监听8080端口
3. 🚫 请勿修改或删除data、logs、permissions和users目录
4. 💾 建议定期备份data目录下的数据库文件
5. 🔑 使用SSH密钥认证比密码认证更安全
6. 📡 对于FTP连接，建议使用被动模式以避免防火墙问题
7. 🔒 请勿在生产环境中使用默认密码，建议登录后立即修改

## �️ 技术栈

- 🎨 **前端**：React + Ant Design + Vite
- ⚙️ **后端**：Go + Gin框架
- 📊 **数据库**：SQLite
- 🔒 **认证**：JWT
- 🔌 **支持协议**：SSH、SFTP、FTP
- 💻 **跨平台**：支持Windows和Linux

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         浏览器 / Browser 🌐                     │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Web服务器 / Web Server 🎨                │
│                       (React + Ant Design)                      │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API服务器 / API Server ⚙️                 │
│                        (Go + Gin Framework)                     │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         数据库 / Database 📊                    │
│                           (SQLite)                              │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         远程服务器 / Remote Servers 🔌           │
│                     (SSH / SFTP / FTP协议)                       │
└─────────────────────────────────────────────────────────────────┘
```

## 📧 联系方式

如有问题或建议，请联系开发团队。

## 🚀 未来计划

### 支持更多连接协议
- 📁 **SMB/CIFS** - Windows文件共享协议
- ☁️ **Amazon S3** - 对象存储服务
- 📤 **WebDAV** - 基于HTTP的文件管理协议
- 📡 **FTPS** - 安全的FTP协议（FTP over SSL/TLS）
- 📦 **Dropbox** - 云存储服务
- 💾 **Google Drive** - 云存储服务
- 📄 **OneDrive** - 云存储服务

### 完善更多功能
- 🎨 **更现代化的UI设计** - 提升用户体验
- 📱 **响应式设计** - 支持移动端访问
- 🔐 **双因素认证** - 增强安全性
- ⚡ **批量文件操作** - 提升效率
- 📊 **详细的统计报表** - 提供使用数据分析
- 🔄 **文件版本控制** - 支持文件历史记录
- 📤 **断点续传优化** - 提升大文件传输体验
- 🌐 **多语言支持** - 支持国际化
- 📱 **移动端应用** - 开发移动端客户端
- 🤖 **智能搜索** - 支持文件内容搜索

## 📝 更新日志

### v1.0.0 (2026-01-05)
- 🎉 初始版本发布
- 🔌 支持SSH、SFTP和FTP协议
- 🎨 提供Web界面进行文件管理
- 📊 支持任务管理和进度显示
- 💻 支持跨平台运行（Windows和Linux）
</div>

<div class="en" style="display: none;">
# 🌐 Remote Connection File Management System

## 📖 Project Introduction

Remote Connection File Management System is a powerful cross-platform file management tool that supports multiple remote connection protocols and provides an intuitive web interface for users to manage files on remote servers.

## 🖼️ Feature Showcase

### How to Add Project Screenshots
1. Create a `screenshots` folder in the project root directory
2. Place project screenshots in this folder according to the following naming rules:
   - `login.png` - Login interface
   - `dashboard.png` - Dashboard
   - `connection-management.png` - Connection management
   - `file-manager.png` - File management
   - `task-management.png` - Task management
   - `settings.png` - Settings interface

### Add Screenshots to README
Add screenshot markdown syntax below, for example:
```markdown
### Login Interface
![Login Interface](screenshots/login.png)

### Dashboard
![Dashboard](screenshots/dashboard.png)

### Connection Management
![Connection Management](screenshots/connection-management.png)

### File Management
![File Management](screenshots/file-manager.png)

### Task Management
![Task Management](screenshots/task-management.png)

### Settings Interface
![Settings Interface](screenshots/settings.png)
```

## 🔌 Supported Server Types

### 1. SSH (Secure Shell) 🔒
- **Protocol Introduction**: SSH is an encrypted network transmission protocol used for secure remote login and other network services on insecure networks.
- **Main Features**:
  - 🔐 Secure remote command execution
  - 📁 File transfer (via SFTP or SCP)
  - 🔀 Port forwarding and tunneling
  - 🔑 Key authentication support
- **Usage Scenarios**: Suitable for scenarios requiring secure remote access to Linux/Unix servers
- **Connection Requirements**: Need to know the server's IP address, port (default 22), username and password or SSH key

### 2. SFTP (SSH File Transfer Protocol) 📤
- **Protocol Introduction**: SFTP is an SSH-based file transfer protocol that provides encrypted file transfer services.
- **Main Features**:
  - 🔐 Secure file upload and download
  - 📁 Create, delete, rename files and directories
  - 🔑 File permission management
  - ⏸️ Resume transfer support
- **Usage Scenarios**: Suitable for scenarios requiring secure file transfer between local and remote servers
- **Connection Requirements**: Need to know the server's IP address, port (default 22), username and password or SSH key

### 3. FTP (File Transfer Protocol) 📡
- **Protocol Introduction**: FTP is a traditional file transfer protocol used for file transfer over networks.
- **Main Features**:
  - 📤 File upload and download
  - 📁 File and directory management
  - 📦 Batch file operations
- **Usage Scenarios**: Suitable for scenarios requiring simple file transfer
- **Connection Requirements**: Need to know the server's IP address, port (default 21), username and password

## 📋 System Requirements

- **Operating System**: Windows 10/11 or Linux (Ubuntu 18.04+, CentOS 7+, etc.)
- **Browser**: Supports modern browsers (Chrome 90+, Firefox 88+, Edge 90+, etc.)
- **Network**: Requires network connection to access the web interface

## 🚀 Running Methods

### Windows System

1. 📦 Extract the compressed package
2. �️ Double-click to run the `start.bat` file
3. 🎉 The system will automatically start the server and open the browser
4. 🌐 Access `http://localhost:8080` in the browser

### Linux System

1. 📦 Extract the compressed package
2. 💻 Open the terminal and enter the extracted directory
3. 🔧 Run `chmod +x start.sh` to grant execution permissions
4. 🚀 Run `./start.sh` to start the system
5. 🌐 Access `http://localhost:8080` in the browser

## ⚙️ Manual Running Method

### Start the Server

- **Windows**: 🖱️ Double-click `server-windows.exe`
- **Linux**: 💻 Run `./server-linux`

### Access the System

🌐 Enter `http://localhost:8080` in the browser

## 📁 Project Structure

```
dist/
├── client/          # 前端静态文件 (Frontend static files) 🎨
├── data/            # 数据库文件 (Database files) 📊
├── logs/            # 日志文件 (Log files) 📝
├── permissions/     # 权限配置文件 (Permission configuration files) 🔒
├── users/           # 用户数据 (User data) 👥
├── server-windows.exe  # Windows服务器可执行文件 (Windows server executable) 💻
├── server-linux        # Linux服务器可执行文件 (Linux server executable) 🐧
├── start.bat           # Windows启动脚本 (Windows startup script) 🚀
├── start.sh            # Linux启动脚本 (Linux startup script) 🚀
└── README.md           # 说明文档 (Documentation) 📖
```

## 🔑 Default Account

- 👤 Username: admin
- 🔒 Password: admin

## 📖 Usage Guide

### 1. Login to the System

1. 🌐 Access `http://localhost:8080` in the browser
2. 🔑 Enter the default username and password (admin/admin)
3. 🖱️ Click the "Login" button

### 2. Add Remote Connection

1. 📋 After logging in, click "Connection Management" in the left menu bar
2. ➕ Click the "Add Connection" button
3. 🔌 Select the connection type (SSH, SFTP, or FTP)
4. 📝 Fill in the connection information:
   - 📛 **Connection Name**: Give the connection an easy-to-identify name
   - 🌐 **IP Address**: The IP address of the remote server
   - 🔌 **Port**: Connection port (default 22 for SSH/SFTP, 21 for FTP)
   - 👤 **Username**: Username for the remote server
   - 🔒 **Password**: Password for the remote server (can be left blank if using key authentication)
   - 🔑 **SSH Key**: If using key authentication, select the SSH private key file
5. 🧪 Click the "Test Connection" button to verify the connection
6. 💾 Click the "Save" button to save the connection configuration

### 3. Manage Remote Files

1. 📋 In "Connection Management" in the left menu bar, click on the added connection
2. 📁 Enter the file management interface, where you can:
   - 🔍 **Browse Files**: Click directory names to enter subdirectories
   - 📤 **Upload Files**: Click the "Upload" button and select local files to upload to the remote server
   - 📥 **Download Files**: Select files or directories and click the "Download" button
   - ➕ **Create Files/Directories**: Click the "New" button and select to create a file or directory
   - ✏️ **Edit Files**: Click on file names to modify file content in the editor
   - 🗑️ **Delete Files**: Select files or directories and click the "Delete" button
   - 🔄 **Rename Files**: Right-click files or directories and select "Rename"
   - 📋 **Copy/Move Files**: Select files or directories and click the "Copy" or "Move" button

### 4. Manage Tasks

1. 📋 Click "Task Management" in the left menu bar
2. 📊 View the status and progress of all upload and download tasks
3. ⏯️ Pause, resume, or cancel ongoing tasks
4. 📝 View detailed logs of tasks

## ⚠️ Notes

1. 🎉 The database and default user will be automatically created when running for the first time
2. 🔌 The server listens on port 8080 by default
3. 🚫 Do not modify or delete the data, logs, permissions, and users directories
4. 💾 It is recommended to regularly back up the database files in the data directory
5. 🔑 SSH key authentication is more secure than password authentication
6. 📡 For FTP connections, it is recommended to use passive mode to avoid firewall issues
7. 🔒 Do not use the default password in production environments; it is recommended to change it immediately after login

## 🛠️ Technology Stack

- 🎨 **Frontend**: React + Ant Design + Vite
- ⚙️ **Backend**: Go + Gin Framework
- 📊 **Database**: SQLite
- 🔒 **Authentication**: JWT
- 🔌 **Supported Protocols**: SSH, SFTP, FTP
- 💻 **Cross-platform**: Supports Windows and Linux

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         浏览器 / Browser 🌐                     │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Web服务器 / Web Server 🎨                │
│                       (React + Ant Design)                      │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API服务器 / API Server ⚙️                 │
│                        (Go + Gin Framework)                     │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         数据库 / Database 📊                    │
│                           (SQLite)                              │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         远程服务器 / Remote Servers 🔌           │
│                     (SSH / SFTP / FTP协议)                       │
└─────────────────────────────────────────────────────────────────┘
```

## 📧 Contact Information

For questions or suggestions, please contact the development team.

## 🚀 Future Plans

### Support More Connection Protocols
- 📁 **SMB/CIFS** - Windows file sharing protocol
- ☁️ **Amazon S3** - Object storage service
- 📤 **WebDAV** - HTTP-based file management protocol
- 📡 **FTPS** - Secure FTP protocol (FTP over SSL/TLS)
- 📦 **Dropbox** - Cloud storage service
- 💾 **Google Drive** - Cloud storage service
- 📄 **OneDrive** - Cloud storage service

### Improve More Features
- 🎨 **More modern UI design** - Improve user experience
- 📱 **Responsive design** - Support mobile access
- 🔐 **Two-factor authentication** - Enhance security
- ⚡ **Batch file operations** - Improve efficiency
- 📊 **Detailed statistical reports** - Provide usage data analysis
- 🔄 **File version control** - Support file history
- 📤 **Resume transfer optimization** - Improve large file transfer experience
- 🌐 **Multi-language support** - Support internationalization
- 📱 **Mobile application** - Develop mobile client
- 🤖 **Intelligent search** - Support file content search

## 📝 Changelog

### v1.0.0 (2026-01-05)
- 🎉 Initial release
- 🔌 Support for SSH, SFTP, and FTP protocols
- 🎨 Web interface for file management
- 📊 Task management and progress display
- 💻 Cross-platform support (Windows and Linux)
</div>
