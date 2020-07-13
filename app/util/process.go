package util

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"strconv"
	"time"
)

func PrintProcess(begin time.Time, current uint64, size uint64, unit string) {
	// Simple progress
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
		fmt.Printf("%s/%s | %s%s/s | %02d:%02d     \r", currentPrint, sizePrint, PpsPrint, unit, sec/60, sec%60)
	} else {
		fmt.Printf("%s | %s%s/s | %02d:%02d     \r", currentPrint, PpsPrint, unit, sec/60, sec%60)
	}
}
