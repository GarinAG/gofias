package service

import (
	"archive/zip"
	"errors"
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

// Сервис управления загрузкой файлов
type DownloadService struct {
	logger interfaces.LoggerInterface // Логгер
	config interfaces.ConfigInterface // Конфиги
}

// Инициализация сервиса
func NewDownloadService(logger interfaces.LoggerInterface, config interfaces.ConfigInterface) *DownloadService {
	return &DownloadService{
		logger: logger,
		config: config,
	}
}

// Очистка директории
func (d *DownloadService) ClearDirectory() {
	dir := d.config.GetConfig().DirectoryFilePath

	// Проверяет наличие директории
	if d.CheckPath(dir) {
		d.logger.WithFields(interfaces.LoggerFields{"dir": dir}).Info("Clear Tmp dir")
		err := os.RemoveAll(dir)
		d.checkFatalError(err)
	}
}

// Создание временной директории
func (d *DownloadService) CreateDirectory() {
	dir := d.config.GetConfig().DirectoryFilePath

	// Проверяет отсутствие директории
	if !d.CheckPath(dir) {
		d.logger.WithFields(interfaces.LoggerFields{"dir": dir}).Info("Create tmp dir")
		// Создает директорию с правами 0777
		err := os.MkdirAll(dir, os.ModePerm)
		d.checkFatalError(err)
	}
}

// Проверяет наличие файла/папки
func (d *DownloadService) CheckPath(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// Получить размер файла
func (d *DownloadService) GetDownloadSize(url string) uint64 {
	// Получает заголовки по URL
	resp, err := http.Head(url)
	if err != nil {
		d.logger.WithFields(interfaces.LoggerFields{"error": err}).Error("Get download file size error")
	} else if resp != nil {
		// Проверяет код ответа сервера
		if resp.StatusCode != http.StatusOK {
			d.logger.WithFields(interfaces.LoggerFields{"code": resp.StatusCode}).Error("Wrong http status code of file")
		} else {
			// Получает размер файла из заголовка
			size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
			return uint64(size)
		}
	}

	return 0
}

// Скачать файл
func (d *DownloadService) DownloadFile(url string, fileName string) (*fileEntity.File, error) {
	// Создает директорию
	d.CreateDirectory()

	filePathLocal := d.config.GetConfig().DirectoryFilePath + fileName
	tmpFile := filePathLocal + ".tmp"
	// Проверяет наличие ранее скачанного файла
	if _, err := os.Stat(filePathLocal); os.IsNotExist(err) {
		d.logger.WithFields(interfaces.LoggerFields{"url": url, "path": filePathLocal}).Info("Download Started")

		// Создает временный файл
		out, err := os.Create(tmpFile)
		if err != nil {
			return nil, err
		}
		defer out.Close()
		defer d.RemoveTmp(tmpFile)

		size := int(d.GetDownloadSize(url))
		if size == 0 {
			return nil, errors.New("unable to get the file size")
		}
		// Создает прогресс-бар для отображение статуса загрузки
		bar := util.StartNewProgress(size, "Downloading", true)

		// Получает файл
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Копирует файл во временный
		if _, err = io.Copy(io.MultiWriter(out, bar.GeBar()), resp.Body); err != nil {
			return nil, err
		}
		// Переименовывает временный файл
		if err = os.Rename(tmpFile, filePathLocal); err != nil {
			return nil, err
		}

		bar.Finish()
		d.logger.Info("Download Finished")
	}

	return &fileEntity.File{Path: filePathLocal}, nil
}

// Удаление временной папки
func (d *DownloadService) RemoveTmp(fileName string) {
	if d.CheckPath(fileName) {
		os.Remove(fileName)
	}
}

// Распаковать файл
func (d *DownloadService) Unzip(file *fileEntity.File, parts ...string) ([]fileEntity.File, error) {
	d.logger.WithFields(interfaces.LoggerFields{"file": file.Path, "parts": parts}).Info("Start unzip file")
	// Проверяет наличие шаблонов названий файлов для распаковки
	if len(parts) == 0 {
		d.logger.Panic("Parts is required field")
		os.Exit(1)
	}

	dest := d.config.GetConfig().DirectoryFilePath
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

	// Создаем обработчик для распаковки файла
	extractAndWriteFile := func(f *zip.File) (interface{}, error) {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				d.checkFatalError(err)
			}
		}()

		savePath := filepath.Join(dest, f.Name)

		// Проверяет является ли объект в архиве директорией
		if f.FileInfo().IsDir() {
			// Создает директорию с правами 0777
			err := os.MkdirAll(savePath, os.ModePerm)
			if err != nil {
				return nil, err
			}
		} else {
			// Проверяет наличие ранее распакованного файла
			if _, err := os.Stat(savePath); err != nil {
				// Создает поддиректорию с правами 0777
				err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
				if err != nil {
					return nil, err
				}
				// Создает временный файл
				f, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
				if err != nil {
					return nil, err
				}
				defer func() {
					if err := f.Close(); err != nil {
						d.checkFatalError(err)
					}
				}()
				// Распаковывает файл
				_, err = io.Copy(f, rc)
				if err != nil {
					return nil, err
				}
			}
		}

		return fileEntity.File{Path: savePath}, nil
	}

	// Проходит по всем файлам в архиве
	for _, f := range r.File {
		// Проходит по всем шаблонам названий файлов
		for _, part := range parts {
			matched, err := regexp.MatchString(part, f.Name)
			if err != nil {
				return filenames, err
			}
			// Проверяет совпадение названия с шаблоном и расширение файла
			if matched && strings.HasSuffix(f.Name, ".XML") {
				// Распаковывает файл
				file, err := extractAndWriteFile(f)

				if err != nil {
					return filenames, err
				}
				filenames = append(filenames, file.(fileEntity.File))
			}
		}
	}

	// Возвращает список распакованных файлов
	return filenames, nil
}

// Проверяет наличие ошибки и логирует ее
func (d *DownloadService) checkFatalError(err error) {
	if err != nil {
		d.logger.Fatal(err.Error())
		os.Exit(1)
	}
}
