package dto

type JsonDownloadFileInfo struct {
	VersionId          int    `json:"VersionId"`
	TextVersion        string `json:"TextVersion"`
	FiasCompleteXmlUrl string ` json:"FiasCompleteXmlUrl"`
	FiasDeltaXmlUrl    string `json:"FiasDeltaXmlUrl"`
}
