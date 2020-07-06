package main

import (
	"archive/zip"
	"github.com/dustin/go-humanize"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"
)

type WriteCounter struct {
	Total uint64
	Size  uint64
	Begin time.Time
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	PrintProcess(wc.Begin, wc.Total, wc.Size, "bytes")
}

func GetDownloadSize(url string) uint64 {
	resp, err := http.Head(url)
	if err != nil {
		logFatalln(err)
	}
	if resp.StatusCode != http.StatusOK {
		logFatalf("Wrong http status code of file: %d", resp.StatusCode)
	}
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	return uint64(size)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadFile(filepath string, url string) error {
	logPrintf("Download Started: %s to %s", url, filepath)

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{
		Size:  GetDownloadSize(url),
		Begin: time.Now(),
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		return err
	}

	if *status {
		fmtPrint("\n")
	}

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}

	return nil
}

func Unzip(src, dest, part string) error {
	logPrintf("Start unzip file: %s with part %s", src, part)
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			if _, err := os.Stat(path); err != nil {
				os.MkdirAll(filepath.Dir(path), os.ModePerm)
				f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil {
						panic(err)
					}
				}()

				_, err = io.Copy(f, rc)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	var hasFiles = false
	for _, f := range r.File {
		if part != "" {
			matched, err := regexp.MatchString(part, f.Name)
			if err != nil {
				return err
			}
			if matched {
				hasFiles = true
				err := extractAndWriteFile(f)

				if err != nil {
					return err
				}
			}
		} else {
			hasFiles = true
			err := extractAndWriteFile(f)
			if err != nil {
				return err
			}
		}
	}
	if !hasFiles {
		logPrintln("Files not found")
	}

	return nil
}

func PrintProcess(begin time.Time, total uint64, size uint64, unit string) {
	if *status {
		// Simple progress
		current := atomic.AddUint64(&total, 1)
		dur := time.Since(begin).Seconds()
		sec := int(dur)
		pps := int64(float64(current) / dur)

		currentPrint := strconv.FormatUint(current, 10)
		sizePrint := strconv.FormatUint(size, 10)
		PpsPrint := strconv.FormatInt(pps, 10)

		if unit == "bytes" {
			currentPrint = humanize.Bytes(current)
			sizePrint = humanize.Bytes(size)
			PpsPrint = humanize.Bytes((uint64)(pps))
			unit = ""
		} else {
			unit = " " + unit
		}

		if size > 0 {
			fmtPrintf("%s/%s | %s%s/s | %02d:%02d     \r", currentPrint, sizePrint, PpsPrint, unit, sec/60, sec%60)
		} else {
			fmtPrintf("%s | %s%s/s | %02d:%02d     \r", currentPrint, PpsPrint, unit, sec/60, sec%60)
		}
	}
}
