package util

import (
	"bufio"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/tamerh/xml-stream-parser"
	"os"
)

type ParseElement func(element *xmlparser.XMLElement) (interface{}, error)

func ParseFile(fileName string, done chan<- bool, c chan<- interface{}, logger interfaces.LoggerInterface, ParseElement ParseElement, xmlName string, total int) {
	logger.WithFields(interfaces.LoggerFields{"fileName": fileName}).Info("Start parse xml file")
	f, err := os.Open(fileName)
	if err != nil {
		logger.WithFields(interfaces.LoggerFields{"error": err}).Error("Error opening file")
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 65536)
	parser := xmlparser.NewXMLParser(br, xmlName).ParseAttributesOnly(xmlName)

	bar := StartNewProgress(total)

	for xml := range parser.Stream() {
		data, err := ParseElement(xml)
		bar.Increment()
		if err == nil && data != nil {
			c <- data
		}
	}
	close(c)
	logger.WithFields(interfaces.LoggerFields{"fileName": fileName}).Info("Parse finished")
	bar.Finish()
	done <- true
}
