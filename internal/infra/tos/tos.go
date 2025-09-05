package tos

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

var (
	URL        = ""
	ak         = ""
	sk         = ""
	backetName = ""
	region     = "cn-beijing"
)

var client *tos.ClientV2

func Init(_url, _ak, _sk, _bucketName string) error {
	fmt.Println("初始化tos")
	URL = _url
	ak = _ak
	sk = _sk
	backetName = _bucketName

	tosClient, err := tos.NewClientV2(URL, tos.WithRegion(region),
		tos.WithCredentials(tos.NewStaticCredentials(ak, sk)))
	if err != nil {
		log.Printf("failedd to new client v2 url %v ak %v sk %v  error %v", URL, ak, sk, err)
		return err
	}
	client = tosClient
	return nil
}

func Close() {
	client.Close()
	client = nil
}

type event struct{}

func (e event) EventChange(event *tos.UploadEvent) {
	if event != nil && event.UploadPartInfo != nil {
		log.Println("upload part info", event.UploadPartInfo)
	} else {
		log.Println("event change upload %v", event)
	}
}

func (e event) Acquire(want int64) (ok bool, timeToWait time.Duration) {
	return true, 200 * time.Millisecond
}

func UploadTos(ctx context.Context, tosKey string, uploadFilePath string) (_err error) {
	defer func(start time.Time) {
		if _err != nil {
			log.Println("upload tos failed", _err)
		} else {
			log.Println("upload tos success", tosKey, uploadFilePath, time.Since(start))
		}
		log.Println("upload tos done", tosKey, uploadFilePath, time.Since(start))
	}(time.Now())

	log.Printf("upload Tos key=%s uploadFile=%s\n", tosKey, uploadFilePath)

	var event event
	output, err := client.UploadFile(ctx, &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: backetName,
			Key:    tosKey,
		},
		FilePath:            uploadFilePath,
		PartSize:            5 * 1024 * 1024,
		UploadEventListener: event,
		RateLimiter:         event,
	})
	if err != nil {
		return err
	}
	log.Println("upload tos success, upload id: %v", output.UploadID)
	return nil
}

func DeleteTosDirAllContext(ctx context.Context, tosDir string) error {
	// 2. 分页列出目录下的所有对象（TOS 列表接口默认分页，需循环获取所有页）
	var allObjectKeys []string // 存储所有要删除的对象 Key
	marker := ""               // 分页标记（初始为空，后续用前一页的 LastMarker 继续）
	pageSize := 1000           // 每页最大数量（TOS 单页最大支持 1000 个）

	for {
		// 构造列表对象请求（按目录前缀筛选）
		listReq := &tos.ListObjectsV2Input{
			Bucket: backetName,
			ListObjectsInput: tos.ListObjectsInput{
				Prefix:  tosDir,   // 目录前缀（关键：筛选该目录下的所有对象）
				Marker:  marker,   // 分页标记
				MaxKeys: pageSize, // 每页最大数量
			},
		}

		// 调用 TOS 列表接口
		listResp, err := client.ListObjectsV2(context.Background(), listReq)
		if err != nil {
			return fmt.Errorf("列出目录对象失败（marker: %s）: %w", marker, err)
		}

		// 收集当前页的对象 Key
		for _, obj := range listResp.Contents {
			allObjectKeys = append(allObjectKeys, obj.Key)
		}

		// 分页终止条件：没有更多数据（IsTruncated 为 false）
		if !listResp.IsTruncated {
			break
		}
		// 更新分页标记，下一页从 LastMarker 开始
		marker = listResp.NextMarker
	}

	// 3. 若目录下无对象，直接返回
	if len(allObjectKeys) == 0 {
		log.Printf("目录 %s 下无对象，无需删除", tosDir)
		return nil
	}

	log.Printf("目录 %s 下共找到 %d 个对象，开始批量删除", tosDir, len(allObjectKeys))

	// 4. 批量删除对象（TOS 批量删除接口单次最多支持 1000 个对象，需分批次）
	batchSize := 1000
	for i := 0; i < len(allObjectKeys); i += batchSize {
		// 计算当前批次的对象范围（避免越界）
		end := i + batchSize
		if end > len(allObjectKeys) {
			end = len(allObjectKeys)
		}
		batchKeys := allObjectKeys[i:end]

		// 构造批量删除请求
		deleteReq := &tos.DeleteMultiObjectsInput{
			Bucket: backetName,
			Objects: func() []tos.ObjectTobeDeleted {
				var objs []tos.ObjectTobeDeleted
				for _, key := range batchKeys {
					objs = append(objs, tos.ObjectTobeDeleted{Key: key})
				}
				return objs
			}(),
		}

		// 调用 TOS 批量删除接口
		deleteResp, err := client.DeleteMultiObjects(context.Background(), deleteReq)
		if err != nil {
			return fmt.Errorf("批量删除对象失败（批次 %d-%d）: %w", i, end-1, err)
		}

		// 检查是否有删除失败的对象
		if len(deleteResp.Error) > 0 {
			for _, errObj := range deleteResp.Error {
				log.Printf("对象 %s 删除失败: %s（错误码: %s）", errObj.Key, errObj.Message, errObj.Code)
			}
			return fmt.Errorf("存在 %d 个对象删除失败", len(deleteResp.Error))
		}

		log.Printf("批次 %d-%d：成功删除 %d 个对象", i, end-1, len(batchKeys))
	}

	log.Printf("目录 %s 下所有 %d 个对象已全部删除", tosDir, len(allObjectKeys))
	return nil
}
