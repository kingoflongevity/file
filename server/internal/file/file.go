package file

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"remote-file-manager/internal/connection"
	"remote-file-manager/internal/log"

	"github.com/jlaffaye/ftp"
	sftp "github.com/pkg/sftp"
	ssh_client "golang.org/x/crypto/ssh"
)

// FileInfo 文件信息
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	IsDir       bool      `json:"is_dir"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	ModTime     time.Time `json:"mod_time"`
	AccessTime  time.Time `json:"access_time"`
	UID         string    `json:"uid"`
	GID         string    `json:"gid"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner"`
	Group       string    `json:"group"`
	Symlink     string    `json:"symlink,omitempty"`
}

// FileTree 文件树结构
type FileTree struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Children []*FileTree `json:"children,omitempty"`
}

// FileManager 文件管理器
type FileManager struct {
	connManager *connection.ConnectionManager
}

// NewFileManager 创建文件管理器
func NewFileManager(connManager *connection.ConnectionManager) *FileManager {
	return &FileManager{
		connManager: connManager,
	}
}

// ListFiles 列出指定路径下的文件
func (m *FileManager) ListFiles(connID, remotePath string) ([]*FileInfo, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return nil, err
		}

		// 执行ls命令获取文件列表
		session, err := client.NewSession()
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 使用ls -la命令获取详细文件信息
		cmd := fmt.Sprintf("ls -la %q | tail -n +2", remotePath)
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %v", err)
		}

		// 解析ls输出
		files, err := m.parseLSOutput(string(output), remotePath)
		if err != nil {
			return nil, err
		}

		return files, nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 切换到指定目录
		if remotePath != "/" {
			if err := ftpClient.ChangeDir(remotePath); err != nil {
				return nil, fmt.Errorf("failed to change directory: %v", err)
			}
		}

		// 获取文件列表
		entries, err := ftpClient.List("")
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %v", err)
		}

		// 转换为统一的FileInfo格式
		files := make([]*FileInfo, 0, len(entries))
		for _, entry := range entries {
			// 跳过当前目录和父目录
			if entry.Name == "." || entry.Name == ".." {
				continue
			}

			// 构建完整路径
			filePath := path.Join(remotePath, entry.Name)

			// 解析文件权限
			permissions := "---------"
			if entry.Type == ftp.EntryTypeFolder {
				permissions = "d---------"
			} else if entry.Type == ftp.EntryTypeLink {
				permissions = "l---------"
			}

			// 解析修改时间
			modTime := entry.Time

			// 创建FileInfo对象
			file := &FileInfo{
				Name:        entry.Name,
				Path:        filePath,
				IsDir:       entry.Type == ftp.EntryTypeFolder,
				Size:        int64(entry.Size),
				Mode:        permissions,
				ModTime:     modTime,
				Permissions: permissions[1:],
				Owner:       "",
				Group:       "",
				UID:         "",
				GID:         "",
			}

			files = append(files, file)
		}

		// 按名称排序，目录在前
		sort.Slice(files, func(i, j int) bool {
			if files[i].IsDir != files[j].IsDir {
				return files[i].IsDir
			}
			return files[i].Name < files[j].Name
		})

		return files, nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// GetFileTree 获取文件树结构
func (m *FileManager) GetFileTree(connID, remotePath string, depth int) (*FileTree, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 检查连接类型
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return nil, err
		}

		// 基本文件信息
		info, err := m.getFileInfo(client, remotePath)
		if err != nil {
			return nil, err
		}

		tree := &FileTree{
			Name:  filepath.Base(remotePath),
			Path:  remotePath,
			IsDir: info.IsDir,
		}

		// 如果不是目录或达到最大深度，直接返回
		if !info.IsDir || depth <= 0 {
			return tree, nil
		}

		// 列出目录内容
		files, err := m.ListFiles(connID, remotePath)
		if err != nil {
			return nil, err
		}

		// 递归获取子目录
		tree.Children = make([]*FileTree, 0)
		for _, file := range files {
			if file.Name != "." && file.Name != ".." {
				childPath := filepath.Join(remotePath, file.Name)
				childTree, err := m.GetFileTree(connID, childPath, depth-1)
				if err != nil {
					log.Warn("Failed to get file tree for %s: %v", childPath, err)
					continue
				}
				tree.Children = append(tree.Children, childTree)
			}
		}

		return tree, nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 列出当前路径内容，检查是否为目录
		entries, err := ftpClient.List(remotePath)
		if err != nil {
			return nil, fmt.Errorf("failed to list remote path: %v", err)
		}

		// 确定是否为目录
		isDir := false
		if len(entries) > 0 && (entries[0].Name != "" || strings.HasSuffix(remotePath, "/")) {
			isDir = true
		}

		tree := &FileTree{
			Name:  filepath.Base(remotePath),
			Path:  remotePath,
			IsDir: isDir,
		}

		// 如果不是目录或达到最大深度，直接返回
		if !isDir || depth <= 0 {
			return tree, nil
		}

		// 列出目录内容
		files, err := m.ListFiles(connID, remotePath)
		if err != nil {
			return nil, err
		}

		// 递归获取子目录
		tree.Children = make([]*FileTree, 0)
		for _, file := range files {
			if file.Name != "." && file.Name != ".." {
				childPath := filepath.Join(remotePath, file.Name)
				childTree, err := m.GetFileTree(connID, childPath, depth-1)
				if err != nil {
					log.Warn("Failed to get file tree for %s: %v", childPath, err)
					continue
				}
				tree.Children = append(tree.Children, childTree)
			}
		}

		return tree, nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// CreateDirectory 创建目录
func (m *FileManager) CreateDirectory(connID, remotePath string) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 使用mkdir -p命令创建目录
		cmd := fmt.Sprintf("mkdir -p %q", remotePath)
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v, output: %s", err, string(output))
		}

		log.Info("Created directory: %s", remotePath)
		return nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return err
		}

		// 创建目录
		if err := ftpClient.MakeDir(remotePath); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		log.Info("Created directory: %s", remotePath)
		return nil
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// DeleteFiles 删除文件或目录
func (m *FileManager) DeleteFiles(connID string, paths []string) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 构建rm命令
		var cmd string
		for _, p := range paths {
			cmd += fmt.Sprintf("%q ", p)
		}
		cmd = fmt.Sprintf("rm -rf %s", cmd)

		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to delete files: %v, output: %s", err, string(output))
		}

		log.Info("Deleted files: %v", paths)
		return nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return err
		}

		// 遍历所有路径，逐个删除
		for _, filePath := range paths {
			// 检查文件类型
			entries, err := ftpClient.List(filePath)
			if err != nil {
				// 如果获取列表失败，尝试作为文件删除
				if err := ftpClient.Delete(filePath); err != nil {
					return fmt.Errorf("failed to delete file %s: %v", filePath, err)
				}
				continue
			}

			// 如果能获取到列表，说明是目录
			if len(entries) > 0 {
				// 删除目录内容
				for _, entry := range entries {
					if entry.Name == "." || entry.Name == ".." {
						continue
					}
					subPath := path.Join(filePath, entry.Name)
					if err := m.DeleteFiles(connID, []string{subPath}); err != nil {
						return err
					}
				}

				// 删除空目录
				if err := ftpClient.RemoveDir(filePath); err != nil {
					return fmt.Errorf("failed to delete directory %s: %v", filePath, err)
				}
			} else {
				// 尝试作为文件删除
				if err := ftpClient.Delete(filePath); err != nil {
					return fmt.Errorf("failed to delete file %s: %v", filePath, err)
				}
			}
		}

		log.Info("Deleted files: %v", paths)
		return nil
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// RenameFile 重命名文件或目录
func (m *FileManager) RenameFile(connID, oldPath, newPath string) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 执行mv命令
		cmd := fmt.Sprintf("mv %q %q", oldPath, newPath)
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to rename file: %v, output: %s", err, string(output))
		}

		log.Info("Renamed file from %s to %s", oldPath, newPath)
		return nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return err
		}

		// 重命名文件或目录
		if err := ftpClient.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("failed to rename file: %v", err)
		}

		log.Info("Renamed file from %s to %s", oldPath, newPath)
		return nil
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// CopyFiles 复制文件或目录
func (m *FileManager) CopyFiles(connID string, srcPaths []string, destPath string) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 构建cp命令
		var cmd string
		for _, src := range srcPaths {
			cmd += fmt.Sprintf("%q ", src)
		}
		cmd = fmt.Sprintf("cp -r %s %q", cmd, destPath)

		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to copy files: %v, output: %s", err, string(output))
		}

		log.Info("Copied files %v to %s", srcPaths, destPath)
		return nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return err
		}

		// 遍历所有源路径，逐个复制
		for _, srcPath := range srcPaths {
			// 检查源路径是否存在
			entries, err := ftpClient.List(srcPath)
			if err != nil {
				return fmt.Errorf("failed to list source path %s: %v", srcPath, err)
			}

			// 构建目标路径
			fileName := path.Base(srcPath)
			finalDestPath := path.Join(destPath, fileName)

			// 如果是目录
			if len(entries) > 0 && entries[0].Name != "" {
				// 创建目标目录
				if err := ftpClient.MakeDir(finalDestPath); err != nil {
					return fmt.Errorf("failed to create destination directory %s: %v", finalDestPath, err)
				}

				// 递归复制目录内容
				for _, entry := range entries {
					if entry.Name == "." || entry.Name == ".." {
						continue
					}
					srcSubPath := path.Join(srcPath, entry.Name)
					destSubPath := path.Join(finalDestPath, entry.Name)
					if entry.Type == ftp.EntryTypeFolder {
						// 递归复制子目录
						if err := m.CopyFiles(connID, []string{srcSubPath}, destSubPath); err != nil {
							return err
						}
					} else {
						// 复制文件
						// 下载源文件
						srcFile, err := ftpClient.Retr(srcSubPath)
						if err != nil {
							return fmt.Errorf("failed to download source file %s: %v", srcSubPath, err)
						}
						defer srcFile.Close()

						// 读取文件内容
						content, err := ioutil.ReadAll(srcFile)
						if err != nil {
							return fmt.Errorf("failed to read source file %s: %v", srcSubPath, err)
						}

						// 上传到目标路径
					if err := ftpClient.Stor(destSubPath, strings.NewReader(string(content))); err != nil {
						return fmt.Errorf("failed to upload destination file %s: %v", destSubPath, err)
					}
					}
				}
			} else {
				// 复制单个文件
				// 下载源文件
				srcFile, err := ftpClient.Retr(srcPath)
				if err != nil {
					return fmt.Errorf("failed to download source file %s: %v", srcPath, err)
				}
				defer srcFile.Close()

				// 读取文件内容
				content, err := ioutil.ReadAll(srcFile)
				if err != nil {
					return fmt.Errorf("failed to read source file %s: %v", srcPath, err)
				}

				// 上传到目标路径
				if err := ftpClient.Stor(finalDestPath, strings.NewReader(string(content))); err != nil {
					return fmt.Errorf("failed to upload destination file %s: %v", finalDestPath, err)
				}
			}
		}

		log.Info("Copied files %v to %s", srcPaths, destPath)
		return nil
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// MoveFiles 移动文件或目录
func (m *FileManager) MoveFiles(connID string, srcPaths []string, destPath string) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 构建mv命令
		var cmd string
		for _, src := range srcPaths {
			cmd += fmt.Sprintf("%q ", src)
		}
		cmd = fmt.Sprintf("mv %s %q", cmd, destPath)

		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to move files: %v, output: %s", err, string(output))
		}

		log.Info("Moved files %v to %s", srcPaths, destPath)
		return nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return err
		}

		// 遍历所有源路径，逐个移动
		for _, srcPath := range srcPaths {
			// 构建目标路径
			fileName := path.Base(srcPath)
			finalDestPath := path.Join(destPath, fileName)

			// 移动文件或目录
			if err := ftpClient.Rename(srcPath, finalDestPath); err != nil {
				return fmt.Errorf("failed to move file %s to %s: %v", srcPath, finalDestPath, err)
			}
		}

		log.Info("Moved files %v to %s", srcPaths, destPath)
		return nil
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// UploadFile 上传文件到远程服务器
func (m *FileManager) UploadFile(connID, remotePath, filename string, fileContent []byte) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		// 创建SFTP客户端
		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			return fmt.Errorf("failed to create SFTP client: %v", err)
		}
		defer sftpClient.Close()

		// 打开远程文件用于写入
		remoteFilePath := filepath.Join(remotePath, filename)
		dstFile, err := sftpClient.Create(remoteFilePath)
		if err != nil {
			return fmt.Errorf("failed to create remote file: %v", err)
		}
		defer dstFile.Close()

		// 写入文件内容
		_, err = dstFile.Write(fileContent)
		if err != nil {
			return fmt.Errorf("failed to write to remote file: %v", err)
		}

		log.Info("Uploaded file %s to %s", filename, remoteFilePath)
		return nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return err
		}

		// 切换到指定目录
		if remotePath != "/" {
			if err := ftpClient.ChangeDir(remotePath); err != nil {
				return fmt.Errorf("failed to change directory: %v", err)
			}
		}

		// 构建远程文件路径
		remoteFilePath := filepath.Join(remotePath, filename)

		// 上传文件
		if err := ftpClient.Stor(remoteFilePath, strings.NewReader(string(fileContent))); err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}

		log.Info("Uploaded file %s to %s", filename, remoteFilePath)
		return nil
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// DownloadFile 从远程服务器下载文件
func (m *FileManager) DownloadFile(connID, remotePath string) ([]byte, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 使用SSH/SFTP方式
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return nil, err
		}

		// 创建SFTP客户端
		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			return nil, fmt.Errorf("failed to create SFTP client: %v", err)
		}
		defer sftpClient.Close()

		// 清理路径，确保路径格式正确
		cleanPath := strings.TrimSpace(remotePath)
		if cleanPath == "" {
			return nil, fmt.Errorf("invalid file path: empty")
		}

		// 确保路径以/开头（绝对路径）
		if !strings.HasPrefix(cleanPath, "/") {
			cleanPath = "/" + cleanPath
		}

		// 打开远程文件用于读取
		srcFile, err := sftpClient.Open(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open remote file: %v", err)
		}
		defer srcFile.Close()

		// 获取文件信息，检查是否为目录
		fileInfo, err := srcFile.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %v", err)
		}

		srcFile.Close()

		if fileInfo.IsDir() {
			// 如果是目录，返回错误，让API路由处理
			log.Info("Detected directory, will create zip task: %s", cleanPath)
			return nil, fmt.Errorf("cannot download directory: %s", cleanPath)
		}

		// 重新打开文件
		srcFile, err = sftpClient.Open(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open remote file: %v", err)
		}
		defer srcFile.Close()

		// 读取文件内容
		content, err := ioutil.ReadAll(srcFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read remote file: %v", err)
		}

		log.Info("Downloaded file %s, size: %d bytes", cleanPath, len(content))
		return content, nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 清理路径，确保路径格式正确
		cleanPath := strings.TrimSpace(remotePath)
		if cleanPath == "" {
			return nil, fmt.Errorf("invalid file path: empty")
		}

		// 确保路径以/开头（绝对路径）
		if !strings.HasPrefix(cleanPath, "/") {
			cleanPath = "/" + cleanPath
		}

		// 检查文件是否为目录
		_, err = ftpClient.List(cleanPath)
		if err == nil {
			// 如果能获取到列表（无论是否为空），说明是目录，返回错误，让API路由处理
			log.Info("Detected directory, will create zip task: %s", cleanPath)
			return nil, fmt.Errorf("cannot download directory: %s", cleanPath)
		}

		// 下载文件
		srcFile, err := ftpClient.Retr(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open remote file: %v", err)
		}
		defer srcFile.Close()

		// 读取文件内容
		content, err := ioutil.ReadAll(srcFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read remote file: %v", err)
		}

		log.Info("Downloaded file %s, size: %d bytes", cleanPath, len(content))
		return content, nil
	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// DownloadFiles 从远程服务器批量下载文件
func (m *FileManager) DownloadFiles(connID string, paths []string) (map[string][]byte, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 创建结果map
	fileContents := make(map[string][]byte)

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
		// 获取SSH客户端
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH client: %v", err)
		}

		// 创建SFTP客户端
		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			return nil, fmt.Errorf("failed to create SFTP client: %v", err)
		}
		defer sftpClient.Close()

		// 遍历所有路径，逐个下载文件或目录
		for _, filePath := range paths {
			// 清理路径，确保路径格式正确
			cleanPath := strings.TrimSpace(filePath)
			if cleanPath == "" {
				log.Warn("Skipping empty file path")
				continue
			}

			// 确保路径以/开头（绝对路径）
			if !strings.HasPrefix(cleanPath, "/") {
				cleanPath = "/" + cleanPath
			}

			log.Info("Attempting to download: %s", cleanPath)

			// 打开远程文件用于读取
			srcFile, err := sftpClient.Open(cleanPath)
			if err != nil {
				log.Error("Failed to open remote file %s: %v", cleanPath, err)
				// 继续下载其他文件，不中断整个批量下载
				continue
			}

			// 获取文件信息，检查是否为目录
			fileInfo, err := srcFile.Stat()
			if err != nil {
				log.Error("Failed to get file info for %s: %v", cleanPath, err)
				srcFile.Close()
				continue
			}

			srcFile.Close()

			if fileInfo.IsDir() {
				// 如果是目录，压缩后下载
				log.Info("Detected directory, zipping: %s", cleanPath)
				zipContent, err := m.ZipDirectory(connID, cleanPath)
				if err != nil {
					log.Error("Failed to zip directory %s: %v", cleanPath, err)
					continue
				}

				// 将压缩后的内容添加到结果map中
				zipFileName := filepath.Base(cleanPath) + ".zip"
				fileContents[zipFileName] = zipContent
				log.Info("Successfully zipped directory %s, size: %d bytes", cleanPath, len(zipContent))
			} else {
				// 如果是文件，直接下载
				// 重新打开文件
				srcFile, err := sftpClient.Open(cleanPath)
				if err != nil {
					log.Error("Failed to re-open remote file %s: %v", cleanPath, err)
					continue
				}

				// 读取文件内容
				content, err := ioutil.ReadAll(srcFile)
				srcFile.Close()
				if err != nil {
					log.Error("Failed to read remote file %s: %v", cleanPath, err)
					// 继续下载其他文件，不中断整个批量下载
					continue
				}

				// 将文件内容添加到结果map中
				fileName := filepath.Base(cleanPath)
				fileContents[fileName] = content
				log.Info("Successfully downloaded file %s, size: %d bytes", cleanPath, len(content))
			}
		}

	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 遍历所有路径，逐个下载文件或目录
		for _, filePath := range paths {
			// 清理路径，确保路径格式正确
			cleanPath := strings.TrimSpace(filePath)
			if cleanPath == "" {
				log.Warn("Skipping empty file path")
				continue
			}

			// 确保路径以/开头（绝对路径）
			if !strings.HasPrefix(cleanPath, "/") {
				cleanPath = "/" + cleanPath
			}

			log.Info("Attempting to download: %s", cleanPath)

			// 检查文件是否为目录
			_, err = ftpClient.List(cleanPath)
			if err == nil {
				// 如果能获取到列表（无论是否为空），说明是目录，压缩后下载
				log.Info("Detected directory, zipping: %s", cleanPath)
				zipContent, err := m.ZipDirectory(connID, cleanPath)
				if err != nil {
					log.Error("Failed to zip directory %s: %v", cleanPath, err)
					continue
				}

				// 将压缩后的内容添加到结果map中
				zipFileName := filepath.Base(cleanPath) + ".zip"
				fileContents[zipFileName] = zipContent
				log.Info("Successfully zipped directory %s, size: %d bytes", cleanPath, len(zipContent))
			} else {
				// 如果是文件，直接下载
				// 下载文件
				srcFile, err := ftpClient.Retr(cleanPath)
				if err != nil {
					log.Error("Failed to open remote file %s: %v", cleanPath, err)
					// 继续下载其他文件，不中断整个批量下载
					continue
				}

				// 读取文件内容
				content, err := ioutil.ReadAll(srcFile)
				srcFile.Close()
				if err != nil {
					log.Error("Failed to read remote file %s: %v", cleanPath, err)
					// 继续下载其他文件，不中断整个批量下载
					continue
				}

				// 将文件内容添加到结果map中
				fileName := filepath.Base(cleanPath)
				fileContents[fileName] = content
				log.Info("Successfully downloaded file %s, size: %d bytes", cleanPath, len(content))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}

	// 如果所有文件都下载失败，返回错误
	if len(fileContents) == 0 && len(paths) > 0 {
		return nil, fmt.Errorf("failed to download any files from provided paths")
	}

	return fileContents, nil
}

// GetFileContent 获取文件内容
func (m *FileManager) GetFileContent(connID, remotePath string) (string, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return "", fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH:
		// 使用SSH方式执行cat命令
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return "", err
		}

		// 执行cat命令获取文件内容
		session, err := client.NewSession()
		if err != nil {
			return "", fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		output, err := session.CombinedOutput(fmt.Sprintf("cat %q", remotePath))
		if err != nil {
			return "", fmt.Errorf("failed to read file content: %v", err)
		}

		return string(output), nil
	case connection.ConnectionTypeSFTP:
		// 使用SFTP方式读取文件内容
		content, err := m.DownloadFile(connID, remotePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	case connection.ConnectionTypeFTP:
		// 使用FTP方式读取文件内容
		content, err := m.DownloadFile(connID, remotePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	default:
		return "", fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// SaveFileContent 保存文件内容
func (m *FileManager) SaveFileContent(connID, remotePath, content string) error {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// 根据连接类型处理
	switch conn.Type {
	case connection.ConnectionTypeSSH:
		// 使用SSH方式执行echo命令
		client, err := m.connManager.GetClient(connID)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create session: %v", err)
		}
		defer session.Close()

		// 使用echo命令写入文件内容
		cmd := fmt.Sprintf("echo -n %q > %q", content, remotePath)
		output, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("failed to save file content: %v, output: %s", err, string(output))
		}

		log.Info("Saved content to file %s", remotePath)
		return nil
	case connection.ConnectionTypeSFTP:
		// 使用SFTP方式保存文件内容
		return m.UploadFile(connID, filepath.Dir(remotePath), filepath.Base(remotePath), []byte(content))
	case connection.ConnectionTypeFTP:
		// 使用FTP方式保存文件内容
		return m.UploadFile(connID, filepath.Dir(remotePath), filepath.Base(remotePath), []byte(content))
	default:
		return fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// parseLSOutput 解析ls命令输出
func (m *FileManager) parseLSOutput(output, basePath string) ([]*FileInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]*FileInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		// 解析ls -la输出格式
		// 示例：-rw-r--r--  1 user group  1024 Jan  1 00:00 filename
		// 或者：drwxr-xr-x  2 user group  4096 Jan  1 00:00 dir name with spaces

		// 前9个字段是固定的，后面的都是文件名
		// 字段1: mode, 2: links, 3: owner, 4: group, 5: size, 6-8: date/time, 9+: filename

		// 找到第8个字段的结束位置（日期/时间的结束）
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// 解析文件属性
		mode := fields[0]
		isDir := mode[0] == 'd'
		symlink := ""

		// 解析文件大小
		size, _ := strconv.ParseInt(fields[4], 10, 64)

		// 解析修改时间
		modTimeStr := strings.Join(fields[5:8], " ")
		modTime, _ := time.Parse("Jan _2 15:04", modTimeStr)
		// 如果年份不是当前年份，ls会显示年份而不是时间
		if len(fields[7]) == 4 {
			modTime, _ = time.Parse("Jan _2 2006", modTimeStr)
		}

		// 找到第8个字段（时间字段）在原始行中的位置
		// 首先找到第8个字段的内容
		timeField := fields[7]

		// 在原始行中找到这个时间字段的位置
		timePos := strings.Index(line, timeField)
		if timePos == -1 {
			continue
		}

		// 文件名从时间字段结束后的第一个非空格字符开始
		// 先找到时间字段结束的位置
		fieldEndPos := timePos + len(timeField)

		// 跳过时间字段后面的所有空格，找到文件名的开始位置
		nameStartPos := fieldEndPos
		for nameStartPos < len(line) && line[nameStartPos] == ' ' {
			nameStartPos++
		}

		// 提取文件名（从nameStartPos到行尾）
		name := line[nameStartPos:]

		// 处理符号链接
		if mode[0] == 'l' {
			// 符号链接格式：name -> target
			nameParts := strings.SplitN(name, " -> ", 2)
			if len(nameParts) == 2 {
				name = nameParts[0]
				symlink = nameParts[1]
			}
		}

		// 构建完整路径
		filePath := path.Join(basePath, name)

		// 创建FileInfo对象
		file := &FileInfo{
			Name:        name,
			Path:        filePath,
			IsDir:       isDir,
			Size:        size,
			Mode:        mode,
			ModTime:     modTime,
			Permissions: mode[1:10],
			Owner:       fields[2],
			Group:       fields[3],
			UID:         fields[2],
			GID:         fields[3],
			Symlink:     symlink,
		}

		files = append(files, file)
	}

	// 按名称排序，目录在前
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// getFileInfo 获取单个文件的详细信息
func (m *FileManager) getFileInfo(client *ssh_client.Client, remotePath string) (*FileInfo, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// 执行ls -la命令获取单个文件信息
	cmd := fmt.Sprintf("ls -la %q", remotePath)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// 解析输出
	files, err := m.parseLSOutput(string(output), path.Dir(remotePath))
	if err != nil {
		return nil, err
	}

	// 查找匹配的文件
	for _, file := range files {
		if file.Path == remotePath {
			return file, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", remotePath)
}