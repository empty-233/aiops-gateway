package logs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	batchSize int64 = 64 * 1024       // 64KB
	maxSize   int64 = 5 * 1024 * 1024 // 5MB
)

type FileReader struct{}

func NewFileReader() *FileReader {
	return &FileReader{}
}

func (r *FileReader) Read(ctx context.Context, path string, maxLines int, options *Options) ([]LogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("日志文件不存在: %s", path)
		}
		if errors.Is(err, os.ErrPermission) {
			return nil, fmt.Errorf("没有权限访问日志文件: %s", path)
		}
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取日志文件信息失败: %w", err)
	}

	var (
		fileSize     = stat.Size()
		offset       = fileSize
		readBytes    int64
		carry        string
		results      []LogEntry
		reachedStart bool
		allHaveTime  = true
	)

outer:
	for offset > 0 && readBytes < maxSize && !reachedStart {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		readSize := batchSize
		if offset < readSize {
			readSize = offset
		}

		if readBytes+readSize > maxSize {
			readSize = maxSize - readBytes
		}

		if readSize <= 0 {
			break
		}

		offset -= readSize
		readBytes += readSize

		buf := make([]byte, int(readSize))

		n, err := file.ReadAt(buf, offset)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		buf = buf[:n]

		chunk := string(buf) + carry
		lines := strings.Split(chunk, "\n")

		if offset > 0 {
			carry = lines[0]
			lines = lines[1:]
		} else {
			carry = ""
		}

		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}

			logTime, ok := parseLogTime(line)
			if !ok {
				allHaveTime = false
				results = append(results, LogEntry{Message: line})
			} else {
				if logTime.After(options.EndTime) {
					continue
				}

				if logTime.Before(options.StartTime) {
					reachedStart = true
					break
				}

				results = append(results, LogEntry{
					Time:    logTime,
					Message: line,
				})
			}
			if maxLines > 0 && len(results) >= maxLines {
				break outer
			}
		}
	}

	if allHaveTime {
		slices.Reverse(results)
	} else {
		fmt.Printf("Warning: 日志文件 %s 中部分行缺少时间戳，无法保证日志顺序\n", path)
	}

	return results, nil
}

func parseLogTime(line string) (time.Time, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return time.Time{}, false
	}

	if t, err := time.Parse(time.RFC3339Nano, fields[0]); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, fields[0]); err == nil {
		return t, true
	}

	return time.Time{}, false
}
