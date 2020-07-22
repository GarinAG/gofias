package service

import (
	"archive/zip"
	fileEntity "github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type WriteCounter struct {
	Total    uint64
	Progress *util.Progress
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	wc.Progress.Add(int64(wc.Total))
}

type DownloadService struct {
	logger interfaces.LoggerInterface
	config interfaces.ConfigInterface
}

func NewDownloadService(logger interfaces.LoggerInterface, config interfaces.ConfigInterface) *DownloadService {
	return &DownloadService{
		logger: logger,
		config: config,
	}
}

func (d *DownloadService) ClearDirectory() error {
	dir := d.config.GetString("directory.filePath")

	if _, err := os.Stat(dir); err == nil {
		d.logger.WithFields(interfaces.LoggerFields{"dir": dir}).Info("Clear Tmp dir")
		err = os.RemoveAll(dir)
		if err != nil {
			d.logger.Fatal(err.Error())
			os.Exit(1)
		}
	}

	return nil
}

func (d *DownloadService) CreateDirectory() error {
	dir := d.config.GetString("directory.filePath")

	if _, err := os.Stat(dir); err != nil {
		d.logger.WithFields(interfaces.LoggerFields{"dir": dir}).Info("Create tmp dir")
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			d.logger.Fatal(err.Error())
			os.Exit(1)
		}
	}

	return nil
}

func (d *DownloadService) GetDownloadSize(url string) uint64 {
	resp, err := http.Head(url)
	if err != nil {
		d.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Get download file size error")
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		d.logger.WithFields(interfaces.LoggerFields{"code": resp.StatusCode}).Fatal("Wrong http status code of file")
	}
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	return uint64(size)
}

func (d *DownloadService) DownloadFile(url string, fileName string) (*fileEntity.File, error) {
	err := d.CreateDirectory()
	if err != nil {
		d.logger.Fatal(err.Error())
		os.Exit(1)
	}

	filePathLocal := d.config.GetString("directory.filePath") + fileName
	if _, err := os.Stat(filePathLocal); os.IsNotExist(err) {
		d.logger.WithFields(interfaces.LoggerFields{"url": url, "path": filePathLocal}).Info("Download Started")

		out, err := os.Create(filePathLocal + ".tmp")
		if err != nil {
			return nil, err
		}
		defer out.Close()

		// Create our progress reporter and pass it to be used alongside our writer
		counter := &WriteCounter{
			Progress: util.StartNewProgress(int(d.GetDownloadSize(url))),
		}
		counter.Progress.SetBytes()

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			return nil, err
		}

		if err = os.Rename(filePathLocal+".tmp", filePathLocal); err != nil {
			return nil, err
		}

		counter.Progress.Finish()
		d.logger.Info("Download Finished")
	}

	return &fileEntity.File{Path: filePathLocal}, nil
}

func (d *DownloadService) Unzip(file *fileEntity.File, parts ...string) ([]fileEntity.File, error) {
	d.logger.WithFields(interfaces.LoggerFields{"file": file.Path, "parts": parts}).Info("Start unzip file")
	if len(parts) == 0 {
		d.logger.Panic("Parts is required field")
		os.Exit(1)
	}

	dest := d.config.GetString("directory.filePath")
	var filenames []fileEntity.File
	r, err := zip.OpenReader(file.Path)
	if err != nil {
		return filenames, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			d.logger.WithFields(interfaces.LoggerFields{"error": err}).Panic("Open zip error")
			os.Exit(1)
		}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) (interface{}, error) {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				d.logger.Panic(err.Error())
				os.Exit(1)
			}
		}()

		savePath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(savePath, os.ModePerm)
			if err != nil {
				return nil, err
			}
		} else {
			if _, err := os.Stat(savePath); err != nil {
				err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
				if err != nil {
					return nil, err
				}
				f, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					return nil, err
				}
				defer func() {
					if err := f.Close(); err != nil {
						d.logger.Panic(err.Error())
						os.Exit(1)
					}
				}()

				_, err = io.Copy(f, rc)
				if err != nil {
					return nil, err
				}
			}
		}

		return fileEntity.File{Path: savePath}, nil
	}

	for _, f := range r.File {
		for _, part := range parts {
			matched, err := regexp.MatchString(part, f.Name)
			if err != nil {
				return filenames, err
			}
			if matched && strings.HasSuffix(f.Name, ".XML") {
				file, err := extractAndWriteFile(f)

				if err != nil {
					return filenames, err
				}
				filenames = append(filenames, file.(fileEntity.File))
			}
		}
	}

	return filenames, nil
}
