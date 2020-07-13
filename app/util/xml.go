package util

import (
	"bufio"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/tamerh/xml-stream-parser"
	"os"
)

type ParseElement func(element *xmlparser.XMLElement) (interface{}, error)

func ParseFile(fileName string, done chan<- bool, c chan<- interface{}, logger interfaces.LoggerInterface, ParseElement ParseElement, xmlName string) {
	logger.WithFields(interfaces.LoggerFields{"fileName": fileName}).Info("Start parse xml file")
	f, err := os.Open(fileName)
	if err != nil {
		logger.WithFields(interfaces.LoggerFields{"error": err}).Error("Error opening file")
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 65536)
	parser := xmlparser.NewXMLParser(br, xmlName).ParseAttributesOnly(xmlName)
	cnt := 0
	for xml := range parser.Stream() {
		if cnt > 5000 {
			break
		}
		cnt++
		data, err := ParseElement(xml)
		if err == nil {
			c <- data
		}
	}

	close(c)
	logger.WithFields(interfaces.LoggerFields{"fileName": fileName}).Info("Parse finished")
	done <- true
}
