package utils

import (
	"log"

	"github.com/anacrolix/torrent/metainfo"
)

type MetaInfo struct {
	InfoHash        string
	TorrentFileName string
	TorrentFilePath string
}

func QBHashInfo(torrentPath string) (string, error) {
	metaInfo, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		return "", err
	}

	// 直接获取 InfoHash
	infoHash := metaInfo.HashInfoBytes().HexString()

	return infoHash, nil
}

func LoadTorrents(torrentDir string) (map[string]MetaInfo, error) {
	log.Println("start to load torrents")
	torrents := make(map[string]MetaInfo)
	fileMap, err := ListAllFilesAsMap(torrentDir)
	if err != nil {
		log.Println("list all files error", err)
		return nil, err
	}

	log.Println("ListAllFilesAsMap get ", len(fileMap))

	//count := 10
	for name, path := range fileMap {
		infoHash, err := QBHashInfo(path)
		if err != nil {
			log.Println("load info hash error", err)
			continue
		}
		torrents[infoHash] = MetaInfo{
			InfoHash:        infoHash,
			TorrentFileName: name,
			TorrentFilePath: path,
		}
		//if count > 10 {
		//	break
		//}
		//count++
	}

	return torrents, nil
}
