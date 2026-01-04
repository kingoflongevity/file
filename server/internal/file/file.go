package file

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"file-manager/internal/log"
	"file-manager/internal/ssh"

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
	sshManager *ssh.SSHManager
}

// NewFileManager 创建文件管理器
func NewFileManager(sshManager *ssh.SSHManager) *FileManager {
	return &FileManager{
		sshManager: sshManager,
	}
}

// ListFiles 列出指定路径下的文件
func (m *FileManager) ListFiles(connID, remotePath string) ([]*FileInfo, error) {
	client, err := m.sshManager.GetClient(connID)
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
}

// GetFileTree 获取文件树结构
func (m *FileManager) GetFileTree(connID, remotePath string, depth int) (*FileTree, error) {
	client, err := m.sshManager.GetClient(connID)
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
}

// CreateDirectory 创建目录
func (m *FileManager) CreateDirectory(connID, remotePath string) error {
	client, err := m.sshManager.GetClient(connID)
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
}

// DeleteFiles 删除文件或目录
func (m *FileManager) DeleteFiles(connID string, paths []string) error {
	client, err := m.sshManager.GetClient(connID)
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
}

// RenameFile 重命名文件或目录
func (m *FileManager) RenameFile(connID, oldPath, newPath string) error {
	client, err := m.sshManager.GetClient(connID)
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
}

// CopyFiles 复制文件或目录
func (m *FileManager) CopyFiles(connID string, srcPaths []string, destPath string) error {
	client, err := m.sshManager.GetClient(connID)
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
}

// MoveFiles 移动文件或目录
func (m *FileManager) MoveFiles(connID string, srcPaths []string, destPath string) error {
	client, err := m.sshManager.GetClient(connID)
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
}

// UploadFile 上传文件到远程服务器
func (m *FileManager) UploadFile(connID, remotePath, filename string, fileContent []byte) error {
	client, err := m.sshManager.GetClient(connID)
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
}

// DownloadFile 从远程服务器下载文件
func (m *FileManager) DownloadFile(connID, remotePath string) ([]byte, error) {
	client, err := m.sshManager.GetClient(connID)
	if err != nil {
		return nil, err
	}

	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// 打开远程文件用于读取
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open remote file: %v", err)
	}
	defer srcFile.Close()

	// 读取文件内容
	content, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote file: %v", err)
	}

	log.Info("Downloaded file %s", remotePath)
	return content, nil
}

// GetFileContent 获取文件内容
func (m *FileManager) GetFileContent(connID, remotePath string) (string, error) {
	client, err := m.sshManager.GetClient(connID)
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
}

// SaveFileContent 保存文件内容
func (m *FileManager) SaveFileContent(connID, remotePath, content string) error {
	client, err := m.sshManager.GetClient(connID)
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
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}

		// 解析文件属性
		mode := parts[0]
		isDir := mode[0] == 'd'
		symlink := ""

		// 处理符号链接
		if mode[0] == 'l' {
			// 符号链接格式：name -> target
			nameParts := strings.SplitN(line, " -> ", 2)
			if len(nameParts) == 2 {
				symlink = nameParts[1]
			}
		}

		// 解析文件大小
		size, _ := strconv.ParseInt(parts[4], 10, 64)

		// 解析修改时间
		modTimeStr := strings.Join(parts[5:8], " ")
		modTime, _ := time.Parse("Jan _2 15:04", modTimeStr)
		// 如果年份不是当前年份，ls会显示年份而不是时间
		if len(parts[7]) == 4 {
			modTime, _ = time.Parse("Jan _2 2006", modTimeStr)
		}

		// 文件名
		name := parts[8]

		// 构建完整路径
		filePath := filepath.Join(basePath, name)

		// 创建FileInfo对象
		file := &FileInfo{
			Name:        name,
			Path:        filePath,
			IsDir:       isDir,
			Size:        size,
			Mode:        mode,
			ModTime:     modTime,
			Permissions: mode[1:10],
			Owner:       parts[2],
			Group:       parts[3],
			UID:         parts[2],
			GID:         parts[3],
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
	files, err := m.parseLSOutput(string(output), filepath.Dir(remotePath))
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
