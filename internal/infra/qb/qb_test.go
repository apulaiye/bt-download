package qb

import (
	"os"
	"sort"
	"testing"

	infraconf "github.com/apulaiye/bt-download/internal/infra/conf"
	"github.com/apulaiye/bt-download/internal/utils"
)

func TestTorrentsOfDownloading(t *testing.T) {
	Init()
	defer Close()

	TorrentsOfDownloading()
}

func TestTorrentsOfCompleted(t *testing.T) {
	os.Setenv("DEBUG", "true")
	infraconf.Init()

	Init()
	defer Close()

	torrents, err := TorrentsOfComplete()
	if err != nil {
		t.Fatal(err)
	}

	var size int64
	count := 0
	for _, torrent := range torrents {
		size += torrent.Size
		count++
		t.Log(torrent.Name, torrent.Size/1024/1024/1024)
		t.Log(torrent.FileDatas)
		for fileName, filePath := range torrent.FileDatas {
			t.Log(fileName, filePath)
		}
		t.Log("------------------------\n")
	}

	t.Log("total size: ", size/1024/1024/1024)
	t.Log("total count: ", count)
}

func TestUploadTorrent(t *testing.T) {
	Init()
	defer Close()

	fileDir := "/Users/winnie/Downloads/0902_video_part_3/"
	torrentMap, err := utils.ListAllFilesAsMap(fileDir)
	if err != nil {
		t.Fatal(err)
	}

	torrents := make([]string, 0)
	for name, _ := range torrentMap {
		torrents = append(torrents, name)
	}

	sort.Strings(torrents)

	onlyCount := 0
	for _, name := range torrents {
		item := torrentMap[name]
		onlyCount++
		err := UploadTorrent(item)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(item)
		if onlyCount > 100 {
			break
		}

	}

}
