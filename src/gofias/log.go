package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func fmtPrintf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func fmtPrint(v ...interface{}) {
	fmt.Print(v...)
}

func fmtPrintln(v ...interface{}) {
	fmt.Println(v...)
}

func logPrintf(format string, v ...interface{}) {
	if checkLogPath() {
		fmtPrintf(format, v...)
		fmtPrintln("")
	}
	log.Printf(format, v...)
}

func logPrint(v ...interface{}) {
	if checkLogPath() {
		fmtPrint(v...)
	}
	log.Print(v...)
}

func logPrintln(v ...interface{}) {
	if checkLogPath() {
		fmtPrintln(v...)
	}
	log.Println(v...)
}

func logFatalf(format string, v ...interface{}) {
	if checkLogPath() {
		fmtPrintf(format, v...)
	}
	log.Fatalf(format, v...)
}

func logFatal(v ...interface{}) {
	if checkLogPath() {
		fmtPrint(v...)
	}
	log.Fatal(v...)
}

func logFatalln(v ...interface{}) {
	if checkLogPath() {
		fmtPrintln(v...)
	}
	log.Fatalln(v...)
}

func setLogPath() {
	*logPath = strings.TrimRight(*logPath, "/")
	if checkLogPath() {
		if _, err := os.Stat(*logPath); err != nil {
			err := os.MkdirAll(*logPath, os.ModePerm)
			if err != nil {
				logFatal(err)
			}
		}

		f, err := os.OpenFile(
			fmt.Sprintf("%s/%s-%s.log", *logPath, "fias-log", time.Now().Format("2006-01-02")),
			os.O_RDWR|os.O_CREATE|os.O_APPEND,
			os.ModePerm)

		if err != nil {
			logFatalf("Error opening file: %v", err)
		}
		log.SetOutput(f)
	}
}

func checkLogPath() bool {
	return *logPath != ""
}
