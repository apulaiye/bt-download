package localstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var localstore *CacheStore
var once sync.Once

// FileLocalStore 基于文件的LocalStore实现
type FileLocalStore struct{}

func (f *FileLocalStore) GetWorkingTorrentTasksWithStatsFilter(status TaskStatus, limit int) []TorrentTask {
	localstore.locker.Lock()
	defer localstore.locker.Unlock()

	result := make([]TorrentTask, 0)
	for _, task := range localstore.WorkingTorrentTasks {
		if task.Status == status {
			result = append(result, task)
		}

		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func (f *FileLocalStore) GetUploadTosTorrentTasks() []TorrentTask {
	localstore.locker.Lock()
	defer localstore.locker.Unlock()

	result := make([]TorrentTask, 0)
	for _, task := range localstore.QueueUploadTosTorrentTasks {
		result = append(result, task)
	}
	return result
}

func (f *FileLocalStore) AddUploadTosTorrentTask(t TorrentTask) {
	localstore.locker.Lock()
	defer localstore.locker.Unlock()

	// 确保Hash存在
	if t.Hash == "" {
		return
	}

	// 添加到工作队列
	localstore.QueueUploadTosTorrentTasks[t.Hash] = t
}

// NewFileLocalStore 创建新的文件存储实例
func InitFileLocalStore(torrentFileDir string, storeSaveFilePath string,
	downloadingRepoPath string, videoRepoPath string) error {
	once.Do(func() {
		localstore = &CacheStore{}
		localstore.DownloadingRepoPath = downloadingRepoPath
		localstore.VideoRepoPath = videoRepoPath
		// 初始化映射
		localstore.TorrentFileDir = torrentFileDir
		localstore.StoreSaveFilePath = storeSaveFilePath
		localstore.WorkingTorrentTasks = make(map[string]TorrentTask)
		localstore.CompleteTorrentTasks = make(map[string]TorrentTask)
		localstore.locker = sync.RWMutex{}
	})

	return nil
}

func GetCacheStore() *CacheStore {
	return localstore
}

// LoadCacheStore 从文件加载缓存数据
func (f *FileLocalStore) LoadCacheStore() error {
	if localstore.StoreSaveFilePath == "" {
		return errors.New("未设置存储文件路径")
	}

	// 检查文件是否存在
	if _, err := os.Stat(localstore.StoreSaveFilePath); os.IsNotExist(err) {
		// 文件不存在，创建目录
		dir := filepath.Dir(localstore.StoreSaveFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建存储目录失败: %v", err)
		}
		return nil
	}

	// 读取文件内容
	data, err := os.ReadFile(localstore.StoreSaveFilePath)
	if err != nil {
		return fmt.Errorf("读取存储文件失败: %v", err)
	}

	var temp CacheStore
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("解析存储数据失败: %v", err)
	}

	// 加锁更新缓存
	localstore.locker.Lock()
	defer localstore.locker.Unlock()

	// 恢复数据
	localstore.TorrentFileDir = temp.TorrentFileDir
	localstore.StoreSaveFilePath = temp.StoreSaveFilePath
	localstore.DownloadingRepoPath = temp.DownloadingRepoPath
	localstore.VideoRepoPath = temp.VideoRepoPath
	localstore.WorkingTorrentTasks = temp.WorkingTorrentTasks
	localstore.CompleteTorrentTasks = temp.CompleteTorrentTasks

	return nil
}

// SaveCacheStore 将缓存数据保存到文件
func (f *FileLocalStore) SaveCacheStore() error {
	if localstore.StoreSaveFilePath == "" {
		return errors.New("未设置存储文件路径")
	}

	// 加锁读取当前状态
	//localstore.locker.RLock()
	//defer localstore.locker.RUnlock()
	err := os.MkdirAll(filepath.Dir(localstore.StoreSaveFilePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to mkdir %v", err)
	}
	// 准备序列化数据
	data, err := json.MarshalIndent(localstore, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化存储数据失败: %v", err)
	}

	// 打开文件（创建或截断）
	file, err := os.OpenFile(localstore.StoreSaveFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("打开存储文件失败: %v", err)
	}
	defer file.Close()

	// 写入数据
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("写入存储文件失败: %v", err)
	}

	// 关键：刷新到磁盘
	// Sync() 会将文件的修改同步到底层存储设备，确保数据不丢失
	if err := file.Sync(); err != nil {
		return fmt.Errorf("刷新文件到磁盘失败: %v", err)
	}

	return nil
}

// AddWorkingTorrentTask 添加任务到工作队列
func (f *FileLocalStore) AddWorkingTorrentTask(t TorrentTask) {
	localstore.locker.Lock()
	defer localstore.locker.Unlock()

	defer func() {
		f.SaveCacheStore()
	}()

	// 确保Hash存在
	if t.Hash == "" {
		return
	}

	// 添加到工作队列
	localstore.WorkingTorrentTasks[t.Hash] = t
}

// SwitchToCompleteTorrentTask 将任务从工作队列移到完成队列
func (f *FileLocalStore) SwitchToCompleteTorrentTask(t TorrentTask) {
	localstore.locker.Lock()
	defer localstore.locker.Unlock()

	defer func() {
		f.SaveCacheStore()
	}()

	if t.Hash == "" {
		return
	}

	// 从工作队列移除
	delete(localstore.WorkingTorrentTasks, t.Hash)

	// 添加到完成队列
	localstore.CompleteTorrentTasks[t.Hash] = t
}

func (f *FileLocalStore) GetWorkingTorrentTask(hash string) (TorrentTask, bool) {
	localstore.locker.RLock()
	defer localstore.locker.RUnlock()

	task, ok := localstore.WorkingTorrentTasks[hash]
	return task, ok
}

func (f *FileLocalStore) PrintTorrentTasks() {
	localstore.locker.RLock()
	defer localstore.locker.RUnlock()

	log.Println("===== 任务状态报告 =====")
	for hash, task := range localstore.WorkingTorrentTasks {
		log.Println("-working %s: %v", hash, task)
	}

	for hash, task := range localstore.CompleteTorrentTasks {
		log.Println("-complete %s: %v", hash, task)
	}
}

type ReportSts struct {
	TotalTask      int
	CompletedTask  int
	DownloadedTask int
	DataSize       int64
}

// Report 生成任务状态报告
func (f *FileLocalStore) Report() *ReportSts {
	localstore.locker.RLock()
	defer localstore.locker.RUnlock()

	fmt.Println("===== 任务状态报告 =====")
	fmt.Printf("工作目录: %s\n", localstore.TorrentFileDir)
	fmt.Printf("下载中任务数: %d\n", len(localstore.WorkingTorrentTasks))
	fmt.Printf("已完成任务数: %d\n", len(localstore.CompleteTorrentTasks))

	fmt.Println("\n下载中任务:")
	totalTaskCount := len(localstore.WorkingTorrentTasks)
	completeCount := 0
	downloadedCount := 0
	var videoSaveSize int64
	for _, task := range localstore.WorkingTorrentTasks {
		if task.Status == TaskStatusCompleted {
			completeCount++
			videoSaveSize += task.DataSize
		}
		if task.Status == TaskStatusDownloaded {
			downloadedCount++
			videoSaveSize += task.DataSize
		}
	}

	return &ReportSts{
		TotalTask:      totalTaskCount,
		CompletedTask:  completeCount,
		DownloadedTask: downloadedCount,
		DataSize:       videoSaveSize / 1024 / 1024 / 1024,
	}
}
