package scheduler

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-yzs/database"
	"go-yzs/models"
)

const videoDetailURL = "https://api.uboxol.com/lotus/trade/abnormal/detail"
const videoBatchSize = 10
const videoConcurrency = 5
const videoMaxBatches = 500 // 单次最多处理 2000 条，防止意外死循环

var videoHTTPClient = &http.Client{Timeout: 30 * time.Second}

// FillMissingVideoDurations 查找 video_duration 为 NULL 的未处理订单，批量补充视频时长。
// 每批 10 条，最多跑 videoMaxBatches 批。由 SkipIfStillRunning 保证单实例。
func FillMissingVideoDurations() {
	log.Println("[Video] Start filling missing video durations")
	total := 0

	for batch := 0; batch < videoMaxBatches; batch++ {
		var trades []models.TradeAbnormal
		if err := database.DB.
			Select("id, trade_id").
			Where("video_duration IS NULL AND is_handled = 0").
			Order("id DESC").
			Limit(videoBatchSize).
			Find(&trades).Error; err != nil {
			log.Printf("[Video] query error: %v", err)
			return
		}
		if len(trades) == 0 {
			break
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, videoConcurrency)
		for _, t := range trades {
			wg.Add(1)
			go func(trade models.TradeAbnormal) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				fillOneTrade(trade)
			}(t)
		}
		wg.Wait()
		total += len(trades)
	}

	log.Printf("[Video] Done, filled %d records", total)
}

func fillOneTrade(trade models.TradeAbnormal) {
	url, err := fetchFirstVideoURL(trade.TradeID)
	if err != nil {
		log.Printf("[Video] tradeID=%d detail API error: %v", trade.TradeID, err)
		database.DB.Model(&trade).Update("video_duration", 1) // detail 请求失败
		return
	}
	if url == "" {
		database.DB.Model(&trade).Update("video_duration", 0) // 无 MP4 视频
		return
	}

	dur := probeMP4Duration(url)
	if dur == 0 {
		dur = 1 // 有 URL 但 MP4 解析失败
	}
	database.DB.Model(&trade).Update("video_duration", dur)
	log.Printf("[Video] tradeID=%d duration=%ds", trade.TradeID, dur)
}

// fetchFirstVideoURL 调外部 detail 接口，按优先级返回第一个 .mp4 URL：
// doorCloseFileUrlList → doorOpenFileUrlList → fileUrlList
// 没有 MP4 视频时返回空字符串。
func fetchFirstVideoURL(tradeID int64) (string, error) {
	body, err := json.Marshal(map[string]any{"id": tradeID})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, videoDetailURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := videoHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			DoorCloseFileUrlList []string `json:"doorCloseFileUrlList"`
			DoorOpenFileUrlList  []string `json:"doorOpenFileUrlList"`
			FileUrlList          []string `json:"fileUrlList"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}
	if !result.Success {
		return "", fmt.Errorf("detail API returned success=false")
	}

	for _, list := range [][]string{
		result.Data.DoorCloseFileUrlList,
		result.Data.DoorOpenFileUrlList,
		result.Data.FileUrlList,
	} {
		for _, u := range list {
			if isVideoURL(u) {
				return u, nil
			}
		}
	}
	return "", nil
}

// isVideoURL 判断 URL 是否为视频链接，兼容 OSS .mp4 和支付宝动态媒体地址。
func isVideoURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, ".mp4") ||
		strings.Contains(lower, "mass.alipay.com/antmedia")
}

// probeMP4Duration 通过 HTTP Range 请求读取 MP4 moov atom，返回视频秒数。
// 先试文件头（faststart），再通过 atom 布局定位 moov 的精确偏移量直接读取。
// 失败返回 0。
func probeMP4Duration(url string) int {
	const chunk = int64(256 * 1024)

	data, totalSize, err := httpRangeGetWithSize(url, 0, chunk-1)
	if err != nil {
		return 0
	}
	// Faststart: moov 在文件头部
	if dur, err := parseMp4Duration(data); err == nil && dur > 0 {
		return dur
	}

	// 遍历顶层 atom 计算 moov 的精确偏移（普通录制：ftyp→[free]→mdat→moov）
	moovOff := computeMoovOffset(data, totalSize)
	if moovOff <= 0 || moovOff >= totalSize {
		return 0
	}
	end := moovOff + chunk - 1
	if end >= totalSize {
		end = totalSize - 1
	}
	data, _, err = httpRangeGetWithSize(url, moovOff, end)
	if err != nil {
		return 0
	}
	dur, _ := parseMp4Duration(data)
	return dur
}

// computeMoovOffset 从文件头 chunk 遍历顶层 atom，返回 moov atom 的文件偏移量。
// 如果遇到超出 chunk 的大 atom（通常是 mdat），则 moov 紧随其后，返回该 atom 的结束偏移。
func computeMoovOffset(data []byte, totalSize int64) int64 {
	i := 0
	for i+8 <= len(data) {
		size32 := binary.BigEndian.Uint32(data[i : i+4])
		atomType := string(data[i+4 : i+8])
		var atomSize int64
		switch size32 {
		case 0:
			atomSize = totalSize - int64(i)
		case 1:
			if i+16 > len(data) {
				return -1
			}
			atomSize = int64(binary.BigEndian.Uint64(data[i+8 : i+16]))
		default:
			if size32 < 8 {
				return -1
			}
			atomSize = int64(size32)
		}
		atomEnd := int64(i) + atomSize
		if atomType == "moov" {
			return int64(i)
		}
		if atomEnd > int64(len(data)) {
			return atomEnd
		}
		i = int(atomEnd)
	}
	return -1
}

// httpRangeGetWithSize 发起 Range GET 请求，同时从 Content-Range 响应头解析文件总大小。
func httpRangeGetWithSize(url string, start, end int64) ([]byte, int64, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := videoHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Content-Range: bytes 0-262143/19533826
	var totalSize int64
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		if idx := strings.LastIndex(cr, "/"); idx >= 0 {
			totalSize, _ = strconv.ParseInt(cr[idx+1:], 10, 64)
		}
	}

	data, err := io.ReadAll(resp.Body)
	return data, totalSize, err
}

// parseMp4Duration 从字节流中解析 moov/mvhd，返回视频秒数。
func parseMp4Duration(data []byte) (int, error) {
	moov := findAtom(data, "moov")
	if moov == nil {
		return 0, fmt.Errorf("moov not found")
	}
	mvhd := findAtom(moov, "mvhd")
	if mvhd == nil || len(mvhd) < 20 {
		return 0, fmt.Errorf("mvhd not found or too short")
	}

	version := mvhd[0]
	var timeScale uint32
	var duration uint64

	switch version {
	case 0:
		timeScale = binary.BigEndian.Uint32(mvhd[12:16])
		duration = uint64(binary.BigEndian.Uint32(mvhd[16:20]))
	case 1:
		if len(mvhd) < 32 {
			return 0, fmt.Errorf("mvhd v1 too short")
		}
		timeScale = binary.BigEndian.Uint32(mvhd[20:24])
		duration = binary.BigEndian.Uint64(mvhd[24:32])
	default:
		return 0, fmt.Errorf("unknown mvhd version %d", version)
	}

	if timeScale == 0 {
		return 0, fmt.Errorf("timescale is zero")
	}
	return int(duration / uint64(timeScale)), nil
}

// findAtom 在字节流顶层查找指定类型的 MP4 atom，返回其 body（不含 header）。
func findAtom(data []byte, typ string) []byte {
	i := 0
	for i+8 <= len(data) {
		size := binary.BigEndian.Uint32(data[i : i+4])
		atomType := string(data[i+4 : i+8])
		headerLen := 8
		var bodyLen uint64

		switch size {
		case 0:
			bodyLen = uint64(len(data)-i) - 8
		case 1:
			if i+16 > len(data) {
				return nil
			}
			bodyLen = binary.BigEndian.Uint64(data[i+8:i+16]) - 16
			headerLen = 16
		default:
			if size < 8 {
				return nil
			}
			bodyLen = uint64(size) - 8
		}

		end := i + headerLen + int(bodyLen)
		if atomType == typ {
			if end > len(data) {
				return data[i+headerLen:]
			}
			return data[i+headerLen : end]
		}
		if end <= i {
			return nil
		}
		i = end
	}
	return nil
}
