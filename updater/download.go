package updater

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Download(url string, progress func(int64, int64)) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmp := filepath.Join(os.TempDir(), "app-update")

	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	defer out.Close()

	total := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			progress(downloaded, total)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return tmp, nil
}
