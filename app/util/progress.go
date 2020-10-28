package util

import (
	"github.com/schollz/progressbar/v3"
	"os"
	"time"
)

// Флаг запрета вывода прогресс-бара
var CanPrintProcess = true

// Объект прогресс-бара
type Progress struct {
	bar *progressbar.ProgressBar
}

// Инициализация прогресс-бара
func StartNewProgress(total int, desc string, isBytes bool) *Progress {
	if total <= 0 {
		total = -1
	}

	options := []progressbar.Option{
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(time.Second),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionClearOnFinish(),
	}

	if isBytes {
		options = append(options, progressbar.OptionShowBytes(true))
	} else {
		options = append(options, progressbar.OptionShowIts())
	}

	if !CanPrintProcess {
		options = append(options, progressbar.OptionSetVisibility(false))
	}

	bar := progressbar.NewOptions(total, options...)

	return &Progress{bar: bar}
}

// Увеличить значение прогресс-бара на 1
func (p *Progress) Increment() {
	if CanPrintProcess {
		p.bar.Add(1)
	}
}

// Добавить произвольное значение прогресс-бара
func (p *Progress) Add(value int64) {
	if CanPrintProcess {
		p.bar.Add64(value)
	}
}

// Установить значение прогресс-бара
func (p *Progress) SetCurrent(value int64) {
	p.bar.Set64(value)
}

// Завершить вывод прогресс-бара
func (p *Progress) Finish() {
	if CanPrintProcess {
		p.bar.Clear()
		p.bar.Finish()
	}
}

// Получить прогрессбар
func (p *Progress) GeBar() *progressbar.ProgressBar {
	return p.bar
}
