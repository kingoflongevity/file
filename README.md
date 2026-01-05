# 🌐 远程连接文件管理系统 / Remote Connection File Management System

## 📖 项目简介 / Project Introduction

远程连接文件管理系统是一款功能强大的跨平台文件管理工具，支持多种远程连接协议，提供直观的Web界面，方便用户管理远程服务器上的文件。

Remote Connection File Management System is a powerful cross-platform file management tool that supports multiple remote connection protocols and provides an intuitive web interface for users to manage files on remote servers.

## 🖼️ 功能展示 / Feature Showcase

### 如何添加项目截图 / How to Add Project Screenshots
1. 在项目根目录创建 `screenshots` 文件夹 / Create a `screenshots` folder in the project root directory
2. 将项目截图按照以下命名规则放入该文件夹 / Place project screenshots in this folder according to the following naming rules:
   - `login.png` - 登录界面 / Login interface
   - `dashboard.png` - 仪表板 / Dashboard
   - `connection-management.png` - 连接管理 / Connection management
   - `file-manager.png` - 文件管理 / File management
   - `task-management.png` - 任务管理 / Task management
   - `settings.png` - 设置界面 / Settings interface

### 添加截图到README / Add Screenshots to README

### 登录界面 / Login Interface
![登录界面](screenshots/login.png)

### 仪表板 / Dashboard
![仪表板](screenshots/dashboard.png)

### 连接管理 / Connection Management
<!-- ![连接管理](screenshots/connection-management.png) -->

### 文件管理 / File Management
<!-- ![文件管理](screenshots/file-manager.png) -->

### 任务管理 / Task Management
<!-- ![任务管理](screenshots/task-management.png) -->

### 设置界面 / Settings Interface
<!-- ![设置界面](screenshots/settings.png) -->

## 🔌 支持的服务器类型 / Supported Server Types

### 1. SSH (Secure Shell) 🔒
- **协议介绍 / Protocol Introduction**：SSH是一种加密的网络传输协议，用于在不安全的网络上安全地进行远程登录和其他网络服务。SSH is an encrypted network transmission protocol used for secure remote login and other network services on insecure networks.
- **主要功能 / Main Features**：
  - 🔐 安全的远程命令执行 / Secure remote command execution
  - 📁 文件传输（通过SFTP或SCP） / File transfer (via SFTP or SCP)
  - 🔀 端口转发和隧道 / Port forwarding and tunneling
  - 🔑 密钥认证支持 / Key authentication support
- **使用场景 / Usage Scenarios**：适用于需要安全远程访问Linux/Unix服务器的场景 / Suitable for scenarios requiring secure remote access to Linux/Unix servers
- **连接要求 / Connection Requirements**：需要知道服务器的IP地址、端口（默认22）、用户名和密码或SSH密钥 / Need to know the server's IP address, port (default 22), username and password or SSH key

### 2. SFTP (SSH File Transfer Protocol) 📤
- **协议介绍 / Protocol Introduction**：SFTP是一种基于SSH的文件传输协议，提供加密的文件传输服务。SFTP is an SSH-based file transfer protocol that provides encrypted file transfer services.
- **主要功能 / Main Features**：
  - 🔐 安全的文件上传和下载 / Secure file upload and download
  - 📁 文件和目录的创建、删除、重命名 / Create, delete, rename files and directories
  - 🔑 文件权限管理 / File permission management
  - ⏸️ 断点续传支持 / Resume transfer support
- **使用场景 / Usage Scenarios**：适用于需要在本地和远程服务器之间安全传输文件的场景 / Suitable for scenarios requiring secure file transfer between local and remote servers
- **连接要求 / Connection Requirements**：需要知道服务器的IP地址、端口（默认22）、用户名和密码或SSH密钥 / Need to know the server's IP address, port (default 22), username and password or SSH key

### 3. FTP (File Transfer Protocol) 📡
- **协议介绍 / Protocol Introduction**：FTP是一种传统的文件传输协议，用于在网络上进行文件传输。FTP is a traditional file transfer protocol used for file transfer over networks.
- **主要功能 / Main Features**：
  - 📤 文件上传和下载 / File upload and download
  - 📁 文件和目录管理 / File and directory management
  - 📦 批量文件操作 / Batch file operations
- **使用场景 / Usage Scenarios**：适用于需要简单文件传输的场景 / Suitable for scenarios requiring simple file transfer
- **连接要求 / Connection Requirements**：需要知道服务器的IP地址、端口（默认21）、用户名和密码 / Need to know the server's IP address, port (default 21), username and password

## 📋 系统要求 / System Requirements

- **操作系统 / Operating System**：Windows 10/11 或 Linux（Ubuntu 18.04+、CentOS 7+等） / Windows 10/11 or Linux (Ubuntu 18.04+, CentOS 7+, etc.)
- **浏览器 / Browser**：支持现代浏览器（Chrome 90+、Firefox 88+、Edge 90+等） / Supports modern browsers (Chrome 90+, Firefox 88+, Edge 90+, etc.)
- **网络 / Network**：需要网络连接，以便访问Web界面 / Requires network connection to access the web interface

## 🚀 运行方式 / Running Methods

### Windows系统 / Windows System

1. 📦 解压压缩包 / Extract the compressed package
2. 🖱️ 双击运行 `start.bat` 文件 / Double-click to run the `start.bat` file
3. 🎉 系统会自动启动服务器并打开浏览器 / The system will automatically start the server and open the browser
4. 🌐 在浏览器中访问 `http://localhost:8080` / Access `http://localhost:8080` in the browser

### Linux系统 / Linux System

1. 📦 解压压缩包 / Extract the compressed package
2. 💻 打开终端，进入解压目录 / Open the terminal and enter the extracted directory
3. 🔧 运行 `chmod +x start.sh` 赋予执行权限 / Run `chmod +x start.sh` to grant execution permissions
4. 🚀 运行 `./start.sh` 启动系统 / Run `./start.sh` to start the system
5. 🌐 在浏览器中访问 `http://localhost:8080` / Access `http://localhost:8080` in the browser

## ⚙️ 手动运行方式 / Manual Running Method

### 启动服务器 / Start the Server

- **Windows**: 🖱️ 双击 `server-windows.exe` / Double-click `server-windows.exe`
- **Linux**: 💻 运行 `./server-linux` / Run `./server-linux`

### 访问系统 / Access the System

🌐 在浏览器中输入 `http://localhost:8080` / Enter `http://localhost:8080` in the browser

## 📁 项目结构 / Project Structure

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

## 🔑 默认账号 / Default Account

- 👤 用户名: admin / Username: admin
- 🔒 密码: admin / Password: admin

## 📖 使用指南 / Usage Guide

### 1. 登录系统 / Login to the System

1. 🌐 在浏览器中访问 `http://localhost:8080` / Access `http://localhost:8080` in the browser
2. 🔑 输入默认用户名和密码（admin/admin） / Enter the default username and password (admin/admin)
3. 🖱️ 点击"登录"按钮 / Click the "Login" button

### 2. 添加远程连接 / Add Remote Connection

1. 📋 登录后，点击左侧菜单栏的"连接管理" / After logging in, click "Connection Management" in the left menu bar
2. ➕ 点击"添加连接"按钮 / Click the "Add Connection" button
3. 🔌 选择连接类型（SSH、SFTP或FTP） / Select the connection type (SSH, SFTP, or FTP)
4. 📝 填写连接信息 / Fill in the connection information：
   - 📛 **连接名称 / Connection Name**：给连接起一个易于识别的名称 / Give the connection an easy-to-identify name
   - 🌐 **IP地址 / IP Address**：远程服务器的IP地址 / The IP address of the remote server
   - 🔌 **端口 / Port**：连接端口（SSH/SFTP默认22，FTP默认21） / Connection port (default 22 for SSH/SFTP, 21 for FTP)
   - 👤 **用户名 / Username**：远程服务器的用户名 / Username for the remote server
   - 🔒 **密码 / Password**：远程服务器的密码（如果使用密钥认证，可留空） / Password for the remote server (can be left blank if using key authentication)
   - 🔑 **SSH密钥 / SSH Key**：如果使用密钥认证，选择SSH私钥文件 / If using key authentication, select the SSH private key file
5. 🧪 点击"测试连接"按钮，验证连接是否成功 / Click the "Test Connection" button to verify the connection
6. 💾 点击"保存"按钮，保存连接配置 / Click the "Save" button to save the connection configuration

### 3. 管理远程文件 / Manage Remote Files

1. 📋 在左侧菜单栏的"连接管理"中，点击已添加的连接 / In "Connection Management" in the left menu bar, click on the added connection
2. 📁 进入文件管理界面，您可以 / Enter the file management interface, where you can：
   - 🔍 **浏览文件**：点击目录名称进入子目录 / **Browse Files**: Click directory names to enter subdirectories
   - 📤 **上传文件**：点击"上传"按钮，选择本地文件上传到远程服务器 / **Upload Files**: Click the "Upload" button and select local files to upload to the remote server
   - 📥 **下载文件**：勾选文件或目录，点击"下载"按钮 / **Download Files**: Select files or directories and click the "Download" button
   - ➕ **创建文件/目录**：点击"新建"按钮，选择创建文件或目录 / **Create Files/Directories**: Click the "New" button and select to create a file or directory
   - ✏️ **编辑文件**：点击文件名称，在编辑器中修改文件内容 / **Edit Files**: Click on file names to modify file content in the editor
   - 🗑️ **删除文件**：勾选文件或目录，点击"删除"按钮 / **Delete Files**: Select files or directories and click the "Delete" button
   - 🔄 **重命名文件**：右键点击文件或目录，选择"重命名" / **Rename Files**: Right-click files or directories and select "Rename"
   - 📋 **复制/移动文件**：勾选文件或目录，点击"复制"或"移动"按钮 / **Copy/Move Files**: Select files or directories and click the "Copy" or "Move" button

### 4. 管理任务 / Manage Tasks

1. 📋 点击左侧菜单栏的"任务管理" / Click "Task Management" in the left menu bar
2. 📊 查看所有上传和下载任务的状态和进度 / View the status and progress of all upload and download tasks
3. ⏯️ 可以暂停、继续或取消正在进行的任务 / Pause, resume, or cancel ongoing tasks
4. 📝 查看任务的详细日志 / View detailed logs of tasks

## ⚠️ 注意事项 / Notes

1. 🎉 首次运行时会自动创建数据库和默认用户 / The database and default user will be automatically created when running for the first time
2. 🔌 服务器默认监听8080端口 / The server listens on port 8080 by default
3. 🚫 请勿修改或删除data、logs、permissions和users目录 / Do not modify or delete the data, logs, permissions, and users directories
4. 💾 建议定期备份data目录下的数据库文件 / It is recommended to regularly back up the database files in the data directory
5. 🔑 使用SSH密钥认证比密码认证更安全 / SSH key authentication is more secure than password authentication
6. 📡 对于FTP连接，建议使用被动模式以避免防火墙问题 / For FTP connections, it is recommended to use passive mode to avoid firewall issues
7. 🔒 请勿在生产环境中使用默认密码，建议登录后立即修改 / Do not use the default password in production environments; it is recommended to change it immediately after login

## 🛠️ 技术栈 / Technology Stack

- 🎨 **前端 / Frontend**：React + Ant Design + Vite / React + Ant Design + Vite
- ⚙️ **后端 / Backend**：Go + Gin框架 / Go + Gin Framework
- 📊 **数据库 / Database**：SQLite / SQLite
- 🔒 **认证 / Authentication**：JWT / JWT
- 🔌 **支持协议 / Supported Protocols**：SSH、SFTP、FTP / SSH, SFTP, FTP
- 💻 **跨平台 / Cross-platform**：支持Windows和Linux / Supports Windows and Linux

## 🏗️ 系统架构 / System Architecture

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

## 📧 联系方式 / Contact Information

如有问题或建议，请联系开发团队。For questions or suggestions, please contact the development team.

## 🚀 未来计划 / Future Plans

### 支持更多连接协议 / Support More Connection Protocols
- 📁 **SMB/CIFS** - Windows文件共享协议 / Windows file sharing protocol
- ☁️ **Amazon S3** - 对象存储服务 / Object storage service
- 📤 **WebDAV** - 基于HTTP的文件管理协议 / HTTP-based file management protocol
- 📡 **FTPS** - 安全的FTP协议（FTP over SSL/TLS） / Secure FTP protocol (FTP over SSL/TLS)
- 📦 **Dropbox** - 云存储服务 / Cloud storage service
- 💾 **Google Drive** - 云存储服务 / Cloud storage service
- 📄 **OneDrive** - 云存储服务 / Cloud storage service

### 完善更多功能 / Improve More Features
- 🎨 **更现代化的UI设计** - 提升用户体验 / **More modern UI design** - Improve user experience
- 📱 **响应式设计** - 支持移动端访问 / **Responsive design** - Support mobile access
- 🔐 **双因素认证** - 增强安全性 / **Two-factor authentication** - Enhance security
- ⚡ **批量文件操作** - 提升效率 / **Batch file operations** - Improve efficiency
- 📊 **详细的统计报表** - 提供使用数据分析 / **Detailed statistical reports** - Provide usage data analysis
- 🔄 **文件版本控制** - 支持文件历史记录 / **File version control** - Support file history
- 📤 **断点续传优化** - 提升大文件传输体验 / **Resume transfer optimization** - Improve large file transfer experience
- 🌐 **多语言支持** - 支持国际化 / **Multi-language support** - Support internationalization
- 📱 **移动端应用** - 开发移动端客户端 / **Mobile application** - Develop mobile client
- 🤖 **智能搜索** - 支持文件内容搜索 / **Intelligent search** - Support file content search

## 📝 更新日志 / Changelog

### v1.0.0 (2026-01-05)
- 🎉 初始版本发布 / Initial release
- 🔌 支持SSH、SFTP和FTP协议 / Support for SSH, SFTP, and FTP protocols
- 🎨 提供Web界面进行文件管理 / Web interface for file management
- 📊 支持任务管理和进度显示 / Task management and progress display
- 💻 支持跨平台运行（Windows和Linux） / Cross-platform support (Windows and Linux)