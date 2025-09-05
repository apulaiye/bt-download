package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	infraconf "github.com/apulaiye/bt-download/internal/infra/conf"
	infralark "github.com/apulaiye/bt-download/internal/infra/lark"
	"github.com/apulaiye/bt-download/internal/infra/localstore"
	infralocalstore "github.com/apulaiye/bt-download/internal/infra/localstore"
	infraqb "github.com/apulaiye/bt-download/internal/infra/qb"
	infratos "github.com/apulaiye/bt-download/internal/infra/tos"
	"github.com/apulaiye/bt-download/internal/services"
	"github.com/apulaiye/bt-download/internal/utils"
)

func main() {
	fmt.Println("World")

	err := infraconf.Init()
	if err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	log.Println(infraconf.Get())

	err = infraqb.Init()
	defer infraqb.Close()

	tosConf := infraconf.Get().Tos
	err = infratos.Init(tosConf.Endpoint, tosConf.AK, tosConf.SK, tosConf.Bucket)
	if err != nil {
		log.Fatalf("初始化tos失败: %v", err)
	}
	defer infratos.Close()

	infralark.InitLarkClient(infraconf.Get().Lake.Webhook)

	config := infraconf.Get()
	err = infralocalstore.InitFileLocalStore(config.Torrent.TorrentDir,
		config.Torrent.StoreSaveFilePath,
		config.Torrent.DownloadingRepoPath,
		config.Torrent.VideoRepoPath)
	if err != nil {
		log.Fatalf("初始化本地存储失败: %v", err)
	}

	localTorrentsMap, err := utils.LoadTorrents(infraconf.Get().Torrent.TorrentDir)
	if err != nil {
		log.Fatalf("列出种子文件失败: %v", err)
	}

	for hash, localTorrent := range localTorrentsMap {
		if _, isExist := infralocalstore.GetLocalStore().GetWorkingTorrentTask(hash); isExist {
			continue
		}

		// /20250903/data/movies/639412/  889246_1756640022690.torrent
		items := strings.Split(localTorrent.TorrentFileName, "_")
		if len(items) != 2 {
			log.Printf("invalid torrnt file name %s\n", localTorrent.TorrentFileName)
			continue
		}
		repoId := items[0]
		infralocalstore.GetLocalStore().AddWorkingTorrentTask(localstore.TorrentTask{
			Hash:            hash,
			Name:            strings.TrimSuffix(localTorrent.TorrentFileName, ".torrent"),
			TorrentFilePath: localTorrent.TorrentFilePath,
			UploadPath:      fmt.Sprintf("/%s/data/movies/%s/", infraconf.Get().Qbserver.Name, repoId),
			LocalSavePath:   fmt.Sprintf("/data/bt-download/data-repo/%s/data/movies/%s/", infraconf.Get().Qbserver.Name, repoId),
			Status:          localstore.TaskStatusPending,
		})
	}

	infralocalstore.GetLocalStore().PrintTorrentTasks()

	//err = upload(infraconf.Get().Torrent.TorrentDir)
	//if err != nil {
	//	log.Fatalf("上传种子失败: %v", err)
	//}

	ctx, _ := context.WithCancel(context.Background())

	services.Runserver(ctx)

	return

}

func upload(fileDir string) error {
	torrentMap, err := utils.ListAllFilesAsMap(fileDir)
	if err != nil {
		return err
	}

	torrents := make([]string, 0)
	for name, _ := range torrentMap {
		torrents = append(torrents, name)
	}

	sort.Strings(torrents)

	limitCount := infraconf.Get().Torrent.LimitCount

	onlyCount := 0
	for _, name := range torrents {
		item := torrentMap[name]
		onlyCount++
		err = infraqb.UploadTorrent(item)
		if err != nil {
			log.Printf("upload torrent failed, torrent path: %s, err: %v", item, err)
			continue
		}

		log.Printf("upload torrent success, torrent path: %s", item)
		if limitCount != 0 && onlyCount > limitCount {
			break
		}

	}
	return nil
}
