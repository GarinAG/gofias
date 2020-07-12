package service

import (
	"archive/zip"
	"fmt"
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
	"time"
)

type WriteCounter struct {
	Total           uint64
	Size            uint64
	Begin           time.Time
	CanPrintProcess bool `default:"false"`
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	if wc.CanPrintProcess {
		wc.PrintProgress()
	}
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	util.PrintProcess(wc.Begin, wc.Total, wc.Size, "bytes")
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
		d.logger.Info("Clear Tmp dir", dir)
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
		d.logger.Info("Create tmp dir", dir)
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
		d.logger.Fatal("Get download file size error: ", err)
	}
	if resp.StatusCode != http.StatusOK {
		d.logger.Fatal("Wrong http status code of file: %d", resp.StatusCode)
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
		d.logger.Info("Download Started: %s to %s", url, filePathLocal)

		out, err := os.Create(filePathLocal + ".tmp")
		if err != nil {
			return nil, err
		}
		defer out.Close()

		// Create our progress reporter and pass it to be used alongside our writer
		counter := &WriteCounter{
			Size:            d.GetDownloadSize(url),
			Begin:           time.Now(),
			CanPrintProcess: d.config.GetBool("process.print"),
		}

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			return nil, err
		}

		if counter.CanPrintProcess {
			fmt.Print("\n")
		}

		if err = os.Rename(filePathLocal+".tmp", filePathLocal); err != nil {
			return nil, err
		}

		d.logger.Info("Download Finished")
	}

	return &fileEntity.File{Path: filePathLocal}, nil
}

func (d *DownloadService) Unzip(file *fileEntity.File, parts ...string) ([]fileEntity.File, error) {
	d.logger.Info(fmt.Sprintf("Start unzip file: %s with parts %s", file.Path, parts))
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
			d.logger.Panic("Open zip error: ", err.Error())
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
