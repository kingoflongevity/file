package file

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"remote-file-manager/internal/connection"
	"remote-file-manager/internal/log"

	"github.com/jlaffaye/ftp"
	sftp "github.com/pkg/sftp"
)

// ZipFileContent 用于存储ZIP文件内容
func ZipFileContent(content []byte, filename string, zw *zip.Writer) error {
	// 创建ZIP文件中的一个文件
	w, err := zw.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file in zip: %v", err)
	}

	// 将文件内容写入ZIP文件
	_, err = w.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content to zip: %v", err)
	}

	return nil
}

// ZipDirectory 递归压缩目录
func (m *FileManager) ZipDirectory(connID string, dirPath string) ([]byte, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 创建一个缓冲区用于存储ZIP文件
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	defer zw.Close()

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

		// 递归压缩目录
		if err := m.zipSFTPDirectory(sftpClient, dirPath, dirPath, zw); err != nil {
			return nil, err
		}

	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 递归压缩目录
		if err := m.zipFTPDirectory(ftpClient, dirPath, dirPath, zw); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}

	// 关闭ZIP写入器
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %v", err)
	}

	log.Info("Zipped directory: %s, size: %d bytes", dirPath, buf.Len())
	return buf.Bytes(), nil
}

// ZipDirectoryWithProgress 带进度的递归压缩目录
func (m *FileManager) ZipDirectoryWithProgress(connID string, dirPath string, updateProgress func(progress int)) ([]byte, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 创建一个缓冲区用于存储ZIP文件
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	defer zw.Close()

	// 更新进度为0%
	updateProgress(0)

	// 计算目录大小和文件数量，用于进度计算
	var totalSize int64
	var fileCount int
	var processedSize int64
	var processedFiles int

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

		// 计算目录大小和文件数量
		if err := m.calculateSFTPDirSize(sftpClient, dirPath, &totalSize, &fileCount); err != nil {
			return nil, err
		}

		// 递归压缩目录，带进度更新
		if err := m.zipSFTPDirectoryWithProgress(sftpClient, dirPath, dirPath, zw, &processedSize, &processedFiles, totalSize, fileCount, updateProgress); err != nil {
			return nil, err
		}

	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 计算目录大小和文件数量
		if err := m.calculateFTPDirSize(ftpClient, dirPath, &totalSize, &fileCount); err != nil {
			return nil, err
		}

		// 递归压缩目录，带进度更新
		if err := m.zipFTPDirectoryWithProgress(ftpClient, dirPath, dirPath, zw, &processedSize, &processedFiles, totalSize, fileCount, updateProgress); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}

	// 关闭ZIP写入器
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %v", err)
	}

	// 更新进度为100%
	updateProgress(100)

	log.Info("Zipped directory: %s, size: %d bytes", dirPath, buf.Len())
	return buf.Bytes(), nil
}

// calculateSFTPDirSize 计算SFTP目录大小和文件数量
func (m *FileManager) calculateSFTPDirSize(sftpClient *sftp.Client, dirPath string, totalSize *int64, fileCount *int) error {
	// 列出目录内容
	files, err := sftpClient.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %v", dirPath, err)
	}

	// 遍历所有文件
	for _, file := range files {
		filePath := path.Join(dirPath, file.Name())

		if file.IsDir() {
			// 递归计算子目录大小
			if err := m.calculateSFTPDirSize(sftpClient, filePath, totalSize, fileCount); err != nil {
				return err
			}
		} else {
			// 累加文件大小和数量
			*totalSize += file.Size()
			*fileCount++
		}
	}

	return nil
}

// calculateFTPDirSize 计算FTP目录大小和文件数量
func (m *FileManager) calculateFTPDirSize(ftpClient *ftp.ServerConn, dirPath string, totalSize *int64, fileCount *int) error {
	// 列出目录内容
	entries, err := ftpClient.List(dirPath)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %v", dirPath, err)
	}

	// 遍历所有文件
	for _, entry := range entries {
		// 跳过当前目录和父目录
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		filePath := path.Join(dirPath, entry.Name)

		if entry.Type == ftp.EntryTypeFolder {
			// 递归计算子目录大小
			if err := m.calculateFTPDirSize(ftpClient, filePath, totalSize, fileCount); err != nil {
				return err
			}
		} else {
			// 累加文件大小和数量
			*totalSize += int64(entry.Size)
			*fileCount++
		}
	}

	return nil
}

// zipSFTPDirectory 递归压缩SFTP目录
func (m *FileManager) zipSFTPDirectory(sftpClient *sftp.Client, basePath string, currentPath string, zw *zip.Writer) error {
	// 列出目录内容
	files, err := sftpClient.ReadDir(currentPath)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %v", currentPath, err)
	}

	// 遍历所有文件
	for _, file := range files {
		filePath := path.Join(currentPath, file.Name())

		// 相对于基础路径的相对路径
		relPath, err := filepath.Rel(basePath, filePath)
		if err != nil {
			relPath = file.Name()
		}

		if file.IsDir() {
			// 递归压缩子目录
			if err := m.zipSFTPDirectory(sftpClient, basePath, filePath, zw); err != nil {
				return err
			}
		} else {
			// 打开文件
			srcFile, err := sftpClient.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", filePath, err)
			}
			defer srcFile.Close()

			// 读取文件内容
			content, err := io.ReadAll(srcFile)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", filePath, err)
			}

			// 添加到ZIP文件
			if err := ZipFileContent(content, relPath, zw); err != nil {
				return fmt.Errorf("failed to zip file %s: %v", filePath, err)
			}

			log.Info("Added file to zip: %s", filePath)
		}
	}

	return nil
}

// zipSFTPDirectoryWithProgress 带进度的递归压缩SFTP目录
func (m *FileManager) zipSFTPDirectoryWithProgress(sftpClient *sftp.Client, basePath string, currentPath string, zw *zip.Writer, processedSize *int64, processedFiles *int, totalSize int64, fileCount int, updateProgress func(progress int)) error {
	// 列出目录内容
	files, err := sftpClient.ReadDir(currentPath)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %v", currentPath, err)
	}

	// 遍历所有文件
	for _, file := range files {
		filePath := path.Join(currentPath, file.Name())

		// 相对于基础路径的相对路径
		relPath, err := filepath.Rel(basePath, filePath)
		if err != nil {
			relPath = file.Name()
		}

		if file.IsDir() {
			// 递归压缩子目录
			if err := m.zipSFTPDirectoryWithProgress(sftpClient, basePath, filePath, zw, processedSize, processedFiles, totalSize, fileCount, updateProgress); err != nil {
				return err
			}
		} else {
			// 打开文件
			srcFile, err := sftpClient.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", filePath, err)
			}

			// 读取文件内容
			content, err := io.ReadAll(srcFile)
			srcFile.Close()
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", filePath, err)
			}

			// 添加到ZIP文件
			if err := ZipFileContent(content, relPath, zw); err != nil {
				return fmt.Errorf("failed to zip file %s: %v", filePath, err)
			}

			// 更新进度
			*processedSize += file.Size()
			*processedFiles++

			// 计算进度百分比
			var progress int
			if totalSize > 0 {
				progress = int(float64(*processedSize) / float64(totalSize) * 100)
			} else if fileCount > 0 {
				progress = int(float64(*processedFiles) / float64(fileCount) * 100)
			}

			// 确保进度在0-100之间
			if progress < 0 {
				progress = 0
			} else if progress > 100 {
				progress = 100
			}

			// 更新进度
			updateProgress(progress)

			log.Info("Added file to zip: %s, progress: %d%%", filePath, progress)
		}
	}

	return nil
}

// zipFTPDirectory 递归压缩FTP目录
func (m *FileManager) zipFTPDirectory(ftpClient *ftp.ServerConn, basePath string, currentPath string, zw *zip.Writer) error {
	// 列出目录内容
	entries, err := ftpClient.List(currentPath)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %v", currentPath, err)
	}

	// 遍历所有文件
	for _, entry := range entries {
		// 跳过当前目录和父目录
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		filePath := path.Join(currentPath, entry.Name)

		// 相对于基础路径的相对路径
		relPath := strings.TrimPrefix(filePath, basePath+"/")
		if relPath == filePath {
			relPath = entry.Name
		}

		if entry.Type == ftp.EntryTypeFolder {
			// 递归压缩子目录
			if err := m.zipFTPDirectory(ftpClient, basePath, filePath, zw); err != nil {
				return err
			}
		} else {
			// 打开文件
			srcFile, err := ftpClient.Retr(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", filePath, err)
			}
			defer srcFile.Close()

			// 读取文件内容
			content, err := io.ReadAll(srcFile)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", filePath, err)
			}

			// 添加到ZIP文件
			if err := ZipFileContent(content, relPath, zw); err != nil {
				return fmt.Errorf("failed to zip file %s: %v", filePath, err)
			}

			log.Info("Added file to zip: %s", filePath)
		}
	}

	return nil
}

// zipFTPDirectoryWithProgress 带进度的递归压缩FTP目录
func (m *FileManager) zipFTPDirectoryWithProgress(ftpClient *ftp.ServerConn, basePath string, currentPath string, zw *zip.Writer, processedSize *int64, processedFiles *int, totalSize int64, fileCount int, updateProgress func(progress int)) error {
	// 列出目录内容
	entries, err := ftpClient.List(currentPath)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %v", currentPath, err)
	}

	// 遍历所有文件
	for _, entry := range entries {
		// 跳过当前目录和父目录
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		filePath := path.Join(currentPath, entry.Name)

		// 相对于基础路径的相对路径
		relPath := strings.TrimPrefix(filePath, basePath+"/")
		if relPath == filePath {
			relPath = entry.Name
		}

		if entry.Type == ftp.EntryTypeFolder {
			// 递归压缩子目录
			if err := m.zipFTPDirectoryWithProgress(ftpClient, basePath, filePath, zw, processedSize, processedFiles, totalSize, fileCount, updateProgress); err != nil {
				return err
			}
		} else {
			// 打开文件
			srcFile, err := ftpClient.Retr(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", filePath, err)
			}

			// 读取文件内容
			content, err := io.ReadAll(srcFile)
			srcFile.Close()
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", filePath, err)
			}

			// 添加到ZIP文件
			if err := ZipFileContent(content, relPath, zw); err != nil {
				return fmt.Errorf("failed to zip file %s: %v", filePath, err)
			}

			// 更新进度
			*processedSize += int64(entry.Size)
			*processedFiles++

			// 计算进度百分比
			var progress int
			if totalSize > 0 {
				progress = int(float64(*processedSize) / float64(totalSize) * 100)
			} else if fileCount > 0 {
				progress = int(float64(*processedFiles) / float64(fileCount) * 100)
			}

			// 确保进度在0-100之间
			if progress < 0 {
				progress = 0
			} else if progress > 100 {
				progress = 100
			}

			// 更新进度
			updateProgress(progress)

			log.Info("Added file to zip: %s, progress: %d%%", filePath, progress)
		}
	}

	return nil
}

// ZipFilesWithProgress 带进度的批量压缩文件和目录
func (m *FileManager) ZipFilesWithProgress(connID string, paths []string, updateProgress func(progress int)) ([]byte, error) {
	// 获取连接配置
	conn, exists := m.connManager.GetConnection(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// 更新进度为0%
	updateProgress(0)

	// 计算总大小和文件数量，用于进度计算
	var totalSize int64
	var fileCount int

	// 创建一个缓冲区用于存储ZIP文件
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	defer zw.Close()

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

		// 计算总大小和文件数量
		for _, path := range paths {
			if err := m.calculateSFTPPathSize(sftpClient, path, &totalSize, &fileCount); err != nil {
				return nil, err
			}
		}

		// 处理文件和目录
		var processedSize int64
		var processedFiles int

		for _, path := range paths {
			if err := m.processSFTPPath(sftpClient, path, path, zw, &processedSize, &processedFiles, totalSize, fileCount, updateProgress); err != nil {
				return nil, err
			}
		}

		// 关闭ZIP写入器
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("failed to close zip writer: %v", err)
		}

		// 更新进度为100%
		updateProgress(100)

		log.Info("Zipped files: %v, size: %d bytes", paths, buf.Len())
		return buf.Bytes(), nil

	case connection.ConnectionTypeFTP:
		// 使用FTP方式
		ftpClient, err := m.connManager.GetFTPClient(connID)
		if err != nil {
			return nil, err
		}

		// 计算总大小和文件数量
		for _, path := range paths {
			if err := m.calculateFTPPathSize(ftpClient, path, &totalSize, &fileCount); err != nil {
				return nil, err
			}
		}

		// 处理文件和目录
		var processedSize int64
		var processedFiles int

		for _, path := range paths {
			if err := m.processFTPPath(ftpClient, path, path, zw, &processedSize, &processedFiles, totalSize, fileCount, updateProgress); err != nil {
				return nil, err
			}
		}

		// 关闭ZIP写入器
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("failed to close zip writer: %v", err)
		}

		// 更新进度为100%
		updateProgress(100)

		log.Info("Zipped files: %v, size: %d bytes", paths, buf.Len())
		return buf.Bytes(), nil

	default:
		return nil, fmt.Errorf("unsupported connection type: %s", conn.Type)
	}
}

// calculateSFTPPathSize 计算SFTP路径大小和文件数量
func (m *FileManager) calculateSFTPPathSize(sftpClient *sftp.Client, path string, totalSize *int64, fileCount *int) error {
	// 获取文件信息
	fileInfo, err := sftpClient.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %v", path, err)
	}

	if fileInfo.IsDir() {
		// 递归计算目录大小
		return m.calculateSFTPDirSize(sftpClient, path, totalSize, fileCount)
	} else {
		// 累加文件大小和数量
		*totalSize += fileInfo.Size()
		*fileCount++
		return nil
	}
}

// calculateFTPPathSize 计算FTP路径大小和文件数量
func (m *FileManager) calculateFTPPathSize(ftpClient *ftp.ServerConn, path string, totalSize *int64, fileCount *int) error {
	// 尝试获取文件信息
	entries, err := ftpClient.List(path)
	if err != nil {
		return fmt.Errorf("failed to list path %s: %v", path, err)
	}

	if len(entries) == 1 && entries[0].Name == filepath.Base(path) {
		// 是文件
		*totalSize += int64(entries[0].Size)
		*fileCount++
		return nil
	} else {
		// 是目录
		return m.calculateFTPDirSize(ftpClient, path, totalSize, fileCount)
	}
}

// processSFTPPath 处理SFTP路径（文件或目录）
func (m *FileManager) processSFTPPath(sftpClient *sftp.Client, basePath string, currentPath string, zw *zip.Writer, processedSize *int64, processedFiles *int, totalSize int64, fileCount int, updateProgress func(progress int)) error {
	// 获取文件信息
	fileInfo, err := sftpClient.Stat(currentPath)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %v", currentPath, err)
	}

	if fileInfo.IsDir() {
		// 递归处理目录
		return m.zipSFTPDirectoryWithProgress(sftpClient, basePath, currentPath, zw, processedSize, processedFiles, totalSize, fileCount, updateProgress)
	} else {
		// 处理单个文件
		// 相对于基础路径的相对路径
		relPath, err := filepath.Rel(basePath, currentPath)
		if err != nil {
			relPath = fileInfo.Name()
		}

		// 打开文件
		srcFile, err := sftpClient.Open(currentPath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", currentPath, err)
		}

		// 读取文件内容
		content, err := io.ReadAll(srcFile)
		srcFile.Close()
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", currentPath, err)
		}

		// 添加到ZIP文件
		if err := ZipFileContent(content, relPath, zw); err != nil {
			return fmt.Errorf("failed to zip file %s: %v", currentPath, err)
		}

		// 更新进度
		*processedSize += fileInfo.Size()
		*processedFiles++

		// 计算进度百分比
		var progress int
		if totalSize > 0 {
			progress = int(float64(*processedSize) / float64(totalSize) * 100)
		} else if fileCount > 0 {
			progress = int(float64(*processedFiles) / float64(fileCount) * 100)
		}

		// 确保进度在0-100之间
		if progress < 0 {
			progress = 0
		} else if progress > 100 {
			progress = 100
		}

		// 更新进度
		updateProgress(progress)

		log.Info("Added file to zip: %s, progress: %d%%", currentPath, progress)
		return nil
	}
}

// processFTPPath 处理FTP路径（文件或目录）
func (m *FileManager) processFTPPath(ftpClient *ftp.ServerConn, basePath string, currentPath string, zw *zip.Writer, processedSize *int64, processedFiles *int, totalSize int64, fileCount int, updateProgress func(progress int)) error {
	// 尝试获取文件信息
	entries, err := ftpClient.List(currentPath)
	if err != nil {
		return fmt.Errorf("failed to list path %s: %v", currentPath, err)
	}

	if len(entries) == 1 && entries[0].Name == filepath.Base(currentPath) {
		// 是文件
		// 相对于基础路径的相对路径
		relPath := filepath.Base(currentPath)

		// 打开文件
		srcFile, err := ftpClient.Retr(currentPath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", currentPath, err)
		}

		// 读取文件内容
		content, err := io.ReadAll(srcFile)
		srcFile.Close()
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", currentPath, err)
		}

		// 添加到ZIP文件
		if err := ZipFileContent(content, relPath, zw); err != nil {
			return fmt.Errorf("failed to zip file %s: %v", currentPath, err)
		}

		// 更新进度
		*processedSize += int64(entries[0].Size)
		*processedFiles++

		// 计算进度百分比
		var progress int
		if totalSize > 0 {
			progress = int(float64(*processedSize) / float64(totalSize) * 100)
		} else if fileCount > 0 {
			progress = int(float64(*processedFiles) / float64(fileCount) * 100)
		}

		// 确保进度在0-100之间
		if progress < 0 {
			progress = 0
		} else if progress > 100 {
			progress = 100
		}

		// 更新进度
		updateProgress(progress)

		log.Info("Added file to zip: %s, progress: %d%%", currentPath, progress)
		return nil
	} else {
		// 是目录
		return m.zipFTPDirectoryWithProgress(ftpClient, basePath, currentPath, zw, processedSize, processedFiles, totalSize, fileCount, updateProgress)
	}
}
