package util

import (
	"github.com/cheggaaa/pb/v3"
	"time"
)

var CanPrintProcess = true

type Progress struct {
	bar *pb.ProgressBar
}

func StartNewProgress(total int) *Progress {
	if total == 0 {
		total = 1
	}
	bar := pb.New(total)
	bar.SetRefreshRate(time.Second)
	bar.SetMaxWidth(95)
	if CanPrintProcess {
		bar.Start()
	}

	return &Progress{bar: bar}
}

func (p *Progress) Increment() {
	if CanPrintProcess {
		p.bar.Increment()
	}
}

func (p *Progress) Add(value int64) {
	if CanPrintProcess {
		p.bar.Add64(value)
	}
}

func (p *Progress) SetCurrent(value int64) {
	p.bar.SetCurrent(value)
}

func (p *Progress) Finish() {
	if CanPrintProcess {
		p.bar.Finish()
	}
}

func (p *Progress) SetBytes() {
	p.bar.Set(pb.Bytes, true)
	p.bar.Set(pb.SIBytesPrefix, true)
}

func (p *Progress) Set(key interface{}, value interface{}) {
	p.bar.Set(key, value)
}
