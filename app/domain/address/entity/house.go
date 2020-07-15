package entity

type HouseObject struct {
	ID         string `xml:"HOUSEID,attr"`
	HouseGuid  string `xml:"HOUSEGUID,attr"`
	AoGuid     string `xml:"AOGUID,attr"`
	HouseNum   string `xml:"HOUSENUM,attr"`
	PostalCode string `xml:"POSTALCODE,attr"`
	Okato      string `xml:"OKATO,attr"`
	Oktmo      string `xml:"OKTMO,attr"`
	StartDate  string `xml:"STARTDATE,attr"`
	EndDate    string `xml:"ENDDATE,attr"`
	UpdateDate string `xml:"UPDATEDATE,attr"`
	DivType    string `xml:"DIVTYPE,attr"`
	BuildNum   string `xml:"BUILDNUM,attr"`
	StructNum  string `xml:"STRUCNUM,attr"`
	Counter    string `xml:"COUNTER,attr"`
	CadNum     string `xml:"CADNUM,attr"`
}

func (o HouseObject) GetXmlFile() string {
	return "AS_HOUSE_"
}

func (o HouseObject) TableName() string {
	return "fias_houses"
}
