package entity

// Объект версии БД ФИАС
type DownloadFileInfo struct {
	VersionId          int
	TextVersion        string
	FiasCompleteXmlUrl string
	FiasDeltaXmlUrl    string
}
