package boss

import (
	"archive/zip"
	"bytes"
)

func mergeZipArchives(zipDatas [][]byte) ([]byte, error) {
	if len(zipDatas) == 0 {
		return make([]byte, 0), nil
	}
	if len(zipDatas) == 1 {
		return zipDatas[0], nil
	}

	allFiles := make(map[string]string)
	for _, zipData := range zipDatas {
		files, _ := readZipFiles(zipData)
		for path, content := range files {
			allFiles[path] = content
		}
	}

	return CreateZipArchive(allFiles)
}

func readZipFiles(zipData []byte) (map[string]string, error) {
	files := make(map[string]string)
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return files, err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)
		files[f.Name] = buf.String()
		rc.Close()
	}

	return files, nil
}

// CreateZipArchive creates ZIP from files
func CreateZipArchive(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for path, content := range files {
		f, err := w.Create(path)
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
