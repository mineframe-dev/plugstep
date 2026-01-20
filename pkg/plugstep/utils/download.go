package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFile(url, destPath string) error {
	return DownloadFileWithProgress(url, destPath, nil)
}

type ProgressFunc func(downloaded, total int64)

func DownloadFileWithProgress(url, destPath string, onProgress ProgressFunc) error {
	resp, err := DownloadClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if onProgress == nil || resp.ContentLength <= 0 {
		_, err = io.Copy(out, resp.Body)
		return err
	}

	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			onProgress(downloaded, resp.ContentLength)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}
