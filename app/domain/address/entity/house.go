package entity

type HouseObject struct {
	ID         string `xml:"HOUSEID,attr"`
	AoGuid     string `xml:"AOGUID,attr"`
	HouseNum   string `xml:"HOUSENUM,attr"`
	RegionCode string `xml:"REGIONCODE,attr"`
	PostalCode string `xml:"POSTALCODE,attr"`
	Okato      string `xml:"OKATO,attr"`
	Oktmo      string `xml:"OKTMO,attr"`
	IfNsFl     string `xml:"IFNSFL,attr"`
	IfNsUl     string `xml:"IFNSUL,attr"`
	TerrIfNsFl string `xml:"TERRIFNSFL,attr"`
	TerrIfNsUl string `xml:"TERRIFNSUL,attr"`
	NormDoc    string `xml:"NORMDOC,attr"`
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
