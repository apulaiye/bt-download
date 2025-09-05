package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func ListAllFilesAsMap(rootDir string) (map[string]string, error) {
	fileMap := make(map[string]string)

	// 使用filepath.Walk会自动递归处理所有子目录
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		// 处理遍历过程中可能出现的错误（如权限问题）
		if err != nil {
			return fmt.Errorf("访问路径 %s 时出错: %v", path, err)
		}

		// 如果是目录则继续递归（filepath.Walk会自动处理），只处理文件
		if !info.IsDir() {
			fileName := info.Name()
			// 如果有同名文件，后面的会覆盖前面的
			fileMap[fileName] = path
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileMap, nil
}
