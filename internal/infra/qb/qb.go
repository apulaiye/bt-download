package qb

import (
	"fmt"
	"log"

	infraconf "github.com/apulaiye/bt-download/internal/infra/conf"
	"github.com/apulaiye/bt-download/internal/utils"
	"github.com/samber/lo"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

// const (
// 	Host     = "http://124.88.174.125:18989"
// 	UserName = "admin"
// 	Password = "adminadmin"
// )

var (
	qb *qbt.Client
)

type TorrentInfoWrap struct {
	qbt.TorrentInfo

	Hash            string
	TorrentFileName string
	TorrentFilePath string
	FileDatas       map[string]string
}

func Init() error {

	conf := infraconf.Get()

	qb = qbt.NewClient(conf.Qbserver.Host)

	err := qb.Login(conf.Qbserver.Username, conf.Qbserver.Password)
	if err != nil {
		return err
	}
	log.Println("qb init 1 success")
	return nil
}

func Close() {
	qb.Logout()
	log.Println("qb close success")
}

func torrents(filters qbt.TorrentsOptions) (q []qbt.TorrentInfo, err error) {
	q = make([]qbt.TorrentInfo, 0)

	pageNo := 0
	pageSize := 100

	for {
		filters.Offset = lo.ToPtr(pageNo * pageSize)
		filters.Limit = lo.ToPtr(pageSize)

		log.Println("filters ", *filters.Offset, *filters.Limit)
		torrents, err := qb.Torrents(filters)
		if err != nil {
			return nil, err
		}
		log.Printf("get torrents: %v", len(torrents))
		q = append(q, torrents...)
		if len(torrents) < pageSize {
			break
		}
		pageNo++
	}
	return q, nil
}

func IsTorrentCompleted(torrent TorrentInfoWrap) bool {
	//return torrent.Progress >= 0.999 &&
	//	(torrent.State == "pausedUP" ||
	//		torrent.State == "stalledUP") &&
	//	torrent.Eta == 0
	return (torrent.State == "pausedUP" ||
		torrent.State == "stalledUP")
}

func TorrentsOfAllNotContainerLocalMeta() (map[string]TorrentInfoWrap, error) {
	filters := qbt.TorrentsOptions{
		Sort: lo.ToPtr("dlspeed"), // 按下载速度排序
	}
	torrents, err := torrents(filters)
	if err != nil {
		return nil, err
	}

	torrentsWrapResp := make(map[string]TorrentInfoWrap, 0)
	for _, torrent := range torrents {
		torrentsWrapResp[torrent.Hash] = TorrentInfoWrap{
			TorrentInfo:     torrent,
			Hash:            torrent.Hash,
			TorrentFileName: torrent.Name,
			TorrentFilePath: torrent.ContentPath,
		}
	}
	return torrentsWrapResp, nil
}

func TorrentsOfDownloading() {
	filters := qbt.TorrentsOptions{
		Filter: lo.ToPtr("downloading"), // 过滤正在下载的任务
		Sort:   lo.ToPtr("dlspeed"),     // 按下载速度排序
	}
	torrents, err := torrents(filters)
	if err != nil {
		panic(err)
	}

	// 遍历并打印下载任务信息
	fmt.Printf("当前正在下载的任务共有 %d 个:\n", len(torrents))
	for _, torrent := range torrents {
		fmt.Printf("名称: %s\n", torrent.Name)
		fmt.Printf("哈希: %s\n", torrent.Hash)
		fmt.Printf("下载进度: %.2f%%\n", torrent.Progress*100)
		fmt.Printf("当前下载速度: %d KB/s\n", torrent.Dlspeed/1024)
		fmt.Printf("已下载: %d MB / 总大小: %d MB\n",
			torrent.Downloaded/(1024*1024),
			torrent.TotalSize/(1024*1024))
		fmt.Printf("剩余时间: %d 秒\n", torrent.Eta)
		fmt.Printf("%+v", torrent)
		fmt.Println("------------------------")
		break
	}
}

func TorrentsOfComplete() ([]TorrentInfoWrap, error) {
	filters := qbt.TorrentsOptions{
		Filter: lo.ToPtr("completed"),     // 关键过滤条件：completed 表示已完成
		Sort:   lo.ToPtr("completion_on"), // 按完成时间排序（最新完成的在前）
	}

	completedTorrents, err := torrents(filters)
	if err != nil {
		return nil, err
	}

	// 遍历并打印已完成任务信息
	fmt.Printf("已完成的下载任务共有 %d 个:\n", len(completedTorrents))
	// for _, torrent := range completedTorrents {
	// 	fmt.Printf("名称: %s\n", torrent.Name)
	// 	fmt.Printf("哈希: %s\n", torrent.Hash)
	// 	fmt.Printf("保存路径: %s\n", torrent.SavePath)
	// 	fmt.Printf("文件路径: %s\n", torrent.ContentPath)
	// 	fmt.Printf("总大小: %.2f MB\n", float64(torrent.TotalSize)/(1024*1024))
	// 	fmt.Printf("完成时间: %s\n", time.Unix(torrent.CompletionOn, 0).Format("2006-01-02 15:04:05")) // 格式化时间
	// 	fmt.Println("------------------------")
	// }

	var errs []error
	torrentsWrapResp := make([]TorrentInfoWrap, 0)
	lo.ForEach(completedTorrents, func(torrent qbt.TorrentInfo, index int) {
		fileMap, err := utils.ListAllFilesAsMap(torrent.ContentPath)
		if err != nil {
			errs = append(errs, err)
			return
		}

		torrentsWrapResp = append(torrentsWrapResp, TorrentInfoWrap{
			TorrentInfo:     torrent,
			Hash:            torrent.Hash,
			TorrentFileName: torrent.Name,
			TorrentFilePath: torrent.ContentPath,
			FileDatas:       fileMap,
		})
	})

	if len(errs) > 0 {
		return nil, fmt.Errorf("list all files as map error: %v", errs)
	}

	log.Printf("get torrents wrap resp: %v", len(torrentsWrapResp))

	return torrentsWrapResp, nil
}

func UploadTorrent(torrentPath string) error {
	options := qbt.DownloadOptions{}
	err := qb.DownloadFiles(torrentPath, options)
	if err != nil {
		log.Printf("download torrent file error: %v", err)
		return err
	}
	log.Printf("download torrent file success, torrent path: %s", torrentPath)
	return nil
}
