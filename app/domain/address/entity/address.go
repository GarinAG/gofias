package entity

// Объект адреса
type AddressObject struct {
	ID             string `xml:"AOID,attr"`
	AoGuid         string `xml:"AOGUID,attr"`
	ParentGuid     string `xml:"PARENTGUID,attr"`
	FormalName     string `xml:"FORMALNAME,attr"`
	ShortName      string `xml:"SHORTNAME,attr"`
	AoLevel        int    `xml:"AOLEVEL,attr"`
	OffName        string `xml:"OFFNAME,attr"`
	Code           string `xml:"CODE,attr"`
	RegionCode     string `xml:"REGIONCODE,attr"`
	PostalCode     string `xml:"POSTALCODE,attr"`
	Okato          string `xml:"OKATO,attr"`
	Oktmo          string `xml:"OKTMO,attr"`
	ActStatus      string `xml:"ACTSTATUS,attr"`
	LiveStatus     string `xml:"LIVESTATUS,attr"`
	CurrStatus     string `xml:"CURRSTATUS,attr"`
	StartDate      string `xml:"STARTDATE,attr"`
	EndDate        string `xml:"ENDDATE,attr"`
	UpdateDate     string `xml:"UPDATEDATE,attr"`
	FullName       string
	FullAddress    string
	AddressSuggest string
	DistrictGuid   string
	District       string
	DistrictType   string
	DistrictFull   string
	SettlementGuid string
	Settlement     string
	SettlementType string
	SettlementFull string
	Street         string
	StreetType     string
	StreetFull     string
	Location       string
}

// Получить название файла импорта
func (a AddressObject) GetXmlFile() string {
	return "AS_ADDROBJ_"
}

// Получить название таблицы в БД
func (a AddressObject) TableName() string {
	return "fias_address"
}
