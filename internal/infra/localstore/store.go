package localstore

import "sync"

type TaskStatus string

const (
	TaskStatusPending     TaskStatus = "pending"
	TaskStatusDownloading TaskStatus = "downloading"
	TaskStatusDownloaded  TaskStatus = "downloaded"
	TaskStausUploading    TaskStatus = "uploading"
	TaskStatusCompleted   TaskStatus = "completed"
)

type TorrentTask struct {
	Hash            string            `json:"hash"`
	Name            string            `json:"name"` // torrent name
	VideoName       string            `json:"video_name"`
	DataSize        int64             `json:"data_size"` // MiB
	Labels          map[string]string `json:"labels"`
	Status          TaskStatus        `json:"status"` // pending downloading downloaded uploading uploaded
	DownloadPath    string            `json:"download_path"`
	UploadPath      string            `json:"upload_path"`
	LocalSavePath   string            `json:"local_save_path"`
	TorrentFilePath string            `json:"torrent_file_path"`
}

type CacheStore struct {
	TorrentFileDir      string `json:"torrent_file_dir"`
	StoreSaveFilePath   string `json:"store_save_file_path"`
	DownloadingRepoPath string `json:"downloading_repo_path"`
	VideoRepoPath       string `json:"video_repo_path"`

	WorkingTorrentTasks        map[string]TorrentTask `json:"working_torrent_tasks"` // all tasks status
	QueueUploadTosTorrentTasks map[string]TorrentTask `json:"queue_uploadTos_torrent_tasks"`
	CompleteTorrentTasks       map[string]TorrentTask `json:"complete_torrent_tasks"`

	locker sync.RWMutex
}

type Store interface {
	LoadCacheStore() error
	SaveCacheStore() error

	SwitchToCompleteTorrentTask(t TorrentTask)

	AddWorkingTorrentTask(t TorrentTask)
	GetWorkingTorrentTask(hash string) (TorrentTask, bool)
	GetWorkingTorrentTasksWithStatsFilter(status TaskStatus, limit int) []TorrentTask

	AddUploadTosTorrentTask(t TorrentTask)
	GetUploadTosTorrentTasks() []TorrentTask
	Report() *ReportSts
	PrintTorrentTasks()
}

func InitCacheStore(torrentFileDir string, storeSaveFilePath string,
	downloadingRepoPath string, videoRepoPath string) error {
	return InitFileLocalStore(torrentFileDir, storeSaveFilePath, downloadingRepoPath, videoRepoPath)
}

func GetLocalStore() Store {
	return &FileLocalStore{}
}
