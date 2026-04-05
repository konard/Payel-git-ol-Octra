package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"log"
)

type MergerService struct{}

func RegisterMergerServer(grpcServer interface{}) {
	log.Printf("Merger service registered")
}

func (s *MergerService) MergeProjects(ctx context.Context, projects map[string][]byte) ([]byte, error) {
	log.Printf("Merging %d projects", len(projects))

	allFiles := make(map[string]string)

	for projectName, zipData := range projects {
		files, err := readZipFiles(zipData)
		if err != nil {
			log.Printf("Error reading zip for %s: %v", projectName, err)
			continue
		}

		for path, content := range files {
			newPath := fmt.Sprintf("%s/%s", projectName, path)
			allFiles[newPath] = content
		}
	}

	finalZip, err := createZipArchive(allFiles)
	if err != nil {
		return nil, err
	}

	log.Printf("Merged project created: %d bytes", len(finalZip))
	return finalZip, nil
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

func createZipArchive(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	dirs := make(map[string]bool)
	for path := range files {
		dir := ""
		parts := splitPath(path)
		for i := 0; i < len(parts)-1; i++ {
			if dir == "" {
				dir = parts[i]
			} else {
				dir = dir + "/" + parts[i]
			}
			if !dirs[dir] {
				dirs[dir] = true
				_, err := w.Create(dir + "/")
				if err != nil {
					return nil, err
				}
			}
		}
	}

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

func splitPath(path string) []string {
	var parts []string
	for {
		dir, file := path, ""
		if i := len(path) - 1; i >= 0 {
			for j := i; j >= 0; j-- {
				if path[j] == '/' || path[j] == '\\' {
					dir = path[:j]
					file = path[j+1:]
					break
				}
			}
			if dir == path {
				parts = append([]string{path}, parts...)
				break
			}
			parts = append([]string{file}, parts...)
			path = dir
		}
	}
	return parts
}
