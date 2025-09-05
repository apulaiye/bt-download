package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	infraconf "github.com/apulaiye/bt-download/internal/infra/conf"
	infralark "github.com/apulaiye/bt-download/internal/infra/lark"
	"github.com/apulaiye/bt-download/internal/infra/localstore"
	infraqb "github.com/apulaiye/bt-download/internal/infra/qb"
	infratos "github.com/apulaiye/bt-download/internal/infra/tos"
	"github.com/apulaiye/bt-download/internal/utils"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	UpdateBQTaskPeriod = 1 * time.Hour

	UploadTosTaskPeriod = 1 * time.Hour

	AddBQTaskPeriod = 2 * time.Hour
)

func Runserver(ctx context.Context) error {
	log.Printf("start to run server")
	// 待完成作业扫描
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		log.Println("starting HandleTorrent server")
		if !lo.Contains(infraconf.Get().Features, "EnableUpdateBQTask") {
			return
		}

		if err := HandleTorrent(ctx); err != nil {
			log.Println("server run torrent logic error:", err)
		}
		return
	}, UpdateBQTaskPeriod)

	// 待上传作业扫描
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		time.Sleep(5 * time.Minute)
		log.Println("starting HandleTos server")
		if !lo.Contains(infraconf.Get().Features, "UploadTosTask") {
			return
		}

		if err := HandleUploadingQueue(ctx); err != nil {
			log.Println("server run uploading logic error:", err)
		}
		return
	}, UploadTosTaskPeriod)

	// 自动添加bp pending任务
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		log.Println("starting HandleAddTorrent server")
		if !lo.Contains(infraconf.Get().Features, "AddBQTask") {
			return
		}
		if err := HandleAddTorrent(ctx); err != nil {
			log.Println("server run add torrent logic error:", err)
		}
		return
	}, AddBQTaskPeriod)

	// report to lark
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		log.Println("starting HandleAddTorrent server")
		if !lo.Contains(infraconf.Get().Features, "ReportLarkTask") {
			return
		}
		if err := ReportLark(ctx, infraconf.Get().Qbserver.Name); err != nil {
			log.Println("server run report lark error:", err)
		}
		return
	}, AddBQTaskPeriod)

	<-ctx.Done()
	return nil
}

func ReportLark(ctx context.Context, host string) error {
	sts := localstore.GetLocalStore().Report()

	infralark.SendCardStatistics(infralark.StatisticsData{
		HostName:     host,
		TotalTasks:   sts.TotalTask,
		SuccessTasks: sts.CompletedTask + sts.DownloadedTask,
		Details: map[string]int{
			"下载成功":   sts.CompletedTask + sts.DownloadedTask,
			"上传成功":   sts.CompletedTask,
			"视频容量GB": int(sts.DataSize),
		},
	})
	return nil
}

func HandleAddTorrent(ctx context.Context) error {
	torrentTasks := localstore.GetLocalStore().GetWorkingTorrentTasksWithStatsFilter(localstore.TaskStatusPending, 100)

	for _, task := range torrentTasks {

		err := infraqb.UploadTorrent(task.TorrentFilePath)
		if err != nil {
			log.Printf("upload torrent failed, torrent path: %s, err: %v", task.TorrentFilePath, err)
			continue
		} else {
			log.Printf("upload torrent success, torrent path: %s", task.TorrentFilePath)
			task.Status = localstore.TaskStatusDownloading
			localstore.GetLocalStore().AddWorkingTorrentTask(task)
		}
	}

	return nil
}

func HandleTorrent(ctx context.Context) error {
	qbTorrentsMap, err := infraqb.TorrentsOfAllNotContainerLocalMeta()
	if err != nil {
		log.Printf("get qb torrents error: %v", err)
		return err
	}

	log.Println("TorrentsOfAllNotContainerLocalMeta get total ", len(qbTorrentsMap))

	lo.ForEach(lo.Entries(qbTorrentsMap), func(torrent lo.Entry[string, infraqb.TorrentInfoWrap], _ int) {
		torrentInfoWrap := torrent.Value

		if infraqb.IsTorrentCompleted(torrentInfoWrap) {
			log.Printf("qb torrent %s completed\n", torrent.Key)
		}

		task, ok := localstore.GetLocalStore().GetWorkingTorrentTask(torrentInfoWrap.Hash)
		if !ok {
			log.Printf("qb torrent %s not found in local working tasks store\n", torrent.Key)
			return
		} else {
			log.Printf("qb torrent %s status %s found in local working tasks store\n", torrent.Key, torrentInfoWrap.State)
		}

		if task.VideoName == "" {
			task.VideoName = torrentInfoWrap.TorrentFileName
		}

		if task.DownloadPath == "" {
			task.DownloadPath = torrentInfoWrap.TorrentFilePath
		}

		if infraqb.IsTorrentCompleted(torrentInfoWrap) {
			log.Printf("qb torrent %s completed\n", torrent.Key)
			if lo.Contains([]localstore.TaskStatus{localstore.TaskStatusDownloading, localstore.TaskStatusPending}, task.Status) {
				task.Status = localstore.TaskStatusDownloaded
				task.DownloadPath = torrentInfoWrap.ContentPath
				task.DataSize = torrentInfoWrap.Size
				//task.DownloadPath = torrentInfoWrap.TorrentFilePath
				//task.VideoName = torrentInfoWrap.TorrentFileName

				// localstore.GetLocalStore().AddUploadTosTorrentTask(task)

				log.Printf("qb torrent %s completed, detail %v\n", torrent.Key, task)
			}
		}

		localstore.GetLocalStore().AddWorkingTorrentTask(task)

	})

	localstore.GetLocalStore().PrintTorrentTasks()

	return nil
}

func HandleUploadingQueue(ctx context.Context) error {

	uploadingTasks := localstore.GetLocalStore().GetWorkingTorrentTasksWithStatsFilter(localstore.TaskStausUploading, 0)
	for _, task := range uploadingTasks {
		// reset uploading task
		err := infratos.DeleteTosDirAllContext(ctx, task.UploadPath)
		if err != nil {
			log.Printf("reset tos uploading %s of %s file %v", task.UploadPath, task.Name, err)
		} else {
			task.Status = localstore.TaskStatusDownloaded
			localstore.GetLocalStore().AddWorkingTorrentTask(task)
		}
	}

	downloadTasks := localstore.GetLocalStore().GetWorkingTorrentTasksWithStatsFilter(localstore.TaskStatusDownloaded, 0)
	for _, task := range downloadTasks {
		videoFileDir := task.DownloadPath
		TosFileDir := task.UploadPath
		localFileDir := task.LocalSavePath

		log.Printf("start to upload tos videoFileDir %v, TosFileDir %v  localFileDir %v",
			videoFileDir, TosFileDir, localFileDir)

		localFilesMap, err := utils.ListAllFilesAsMap(videoFileDir)
		if err != nil {
			log.Printf("get qb torrent %s error: %v", task.VideoName, err)
			continue
		}

		task.Status = localstore.TaskStausUploading
		localstore.GetLocalStore().AddWorkingTorrentTask(task)

		wg := sync.WaitGroup{}
		wg.Add(len(localFilesMap))
		errs := make([]error, 0)
		errMutex := sync.Mutex{}

		for videoFileName, videoFilePath := range localFilesMap {
			go func() {
				defer wg.Done()

				srcInfo, err := os.Stat(videoFilePath)
				if err != nil {
					log.Printf("stat copy %s error: %v", videoFilePath, err)
					errMutex.Lock()
					errs = append(errs, err)
					errMutex.Unlock()
					return
				}
				if !srcInfo.Mode().IsRegular() {
					err := fmt.Errorf("源路径 %s 不是常规文件", videoFilePath)
					errMutex.Lock()
					errs = append(errs, err)
					errMutex.Unlock()
					return
				}

				tosKey := fmt.Sprintf("%s/%s", TosFileDir, videoFileName)
				err = infratos.UploadTos(ctx, tosKey, videoFilePath)
				if err != nil {
					log.Printf("upload tos %s error: %v\n", tosKey, err)
					errMutex.Lock()
					errs = append(errs, err)
					errMutex.Unlock()
					return
				}
				log.Println("success to upload to tos, ", tosKey)

				err = copyFile(videoFilePath, localFileDir)
				if err != nil {
					log.Printf("upload tos %s error: %v\n", tosKey, err)
					errMutex.Lock()
					errs = append(errs, err)
					errMutex.Unlock()
					return
				}
				log.Println("success to upload to tos, ", tosKey)

			}()
		}

		wg.Wait()

		if len(errs) > 0 {
			log.Printf("upload tos errors: %v\n", errs)
			continue
		}

		task.Status = localstore.TaskStatusCompleted
		localstore.GetLocalStore().AddWorkingTorrentTask(task)
	}

	return nil
}

func copyFile(src, destDir string) (err error) {
	log.Printf("copy file %s to %s\n", src, destDir)

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Printf("stat copy %s error: %v", src, err)
		return err
	}

	destFileName := srcInfo.Name()
	destPath := filepath.Join(destDir, destFileName)

	// 关键修改：如果目标文件已存在，先删除
	if _, err := os.Stat(destPath); err == nil {
		// 文件存在，执行删除
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("删除已存在的目标文件失败: %w", err)
		}
	} else if !os.IsNotExist(err) {
		// 处理其他错误（非"文件不存在"的错误）
		return fmt.Errorf("检查目标文件状态失败: %w", err)
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 复制文件内容
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 确保内容刷新到磁盘
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("刷新目标文件到磁盘失败: %w", err)
	}

	// 复制源文件权限
	if err := os.Chmod(destPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("设置目标文件权限失败: %w", err)
	}

	return nil
}
