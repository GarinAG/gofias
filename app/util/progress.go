package util

import (
	"github.com/cheggaaa/pb/v3"
	"time"
)

// Флаг запрета вывода прогресс-бара
var CanPrintProcess = true

// Объект прогресс-бара
type Progress struct {
	bar *pb.ProgressBar
}

// Инициализация прогресс-бара
func StartNewProgress(total int) *Progress {
	if total == 0 {
		total = 1
	}
	bar := pb.New(total)
	// Обновлять раз в секунду
	bar.SetRefreshRate(time.Second)
	// Установить максимальную ширину прогресс-бара
	bar.SetMaxWidth(95)
	if CanPrintProcess {
		bar.Start()
	}

	return &Progress{bar: bar}
}

// Увеличить значение прогресс-бара на 1
func (p *Progress) Increment() {
	if CanPrintProcess {
		p.bar.Increment()
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
	p.bar.SetCurrent(value)
}

// Завершить вывод прогресс-бара
func (p *Progress) Finish() {
	if CanPrintProcess {
		p.bar.Finish()
	}
}

// Установить формат вывода прогресса в байты
func (p *Progress) SetBytes() {
	p.bar.Set(pb.Bytes, true)
	p.bar.Set(pb.SIBytesPrefix, true)
}

// Установить свойства прогресс-бара
func (p *Progress) Set(key interface{}, value interface{}) {
	p.bar.Set(key, value)
}
