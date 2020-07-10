package entity

type AddressObject struct {
	ID         string `xml:"AOID,attr"`
	AoGuid     string `xml:"AOGUID,attr"`
	ParentGuid string `xml:"PARENTGUID,attr"`
	FormalName string `xml:"FORMALNAME,attr"`
	ShortName  string `xml:"SHORTNAME,attr"`
	AoLevel    string `xml:"AOLEVEL,attr"`
	OffName    string `xml:"OFFNAME,attr"`
	AreaCode   string `xml:"AREACODE,attr"`
	CityCode   string `xml:"CITYCODE,attr"`
	PlaceCode  string `xml:"PLACECODE,attr"`
	AutoCode   string `xml:"AUTOCODE,attr"`
	PlanCode   string `xml:"PLANCODE,attr"`
	StreetCode string `xml:"STREETCODE,attr"`
	CTarCode   string `xml:"CTARCODE,attr"`
	ExtrCode   string `xml:"EXTRCODE,attr"`
	SextCode   string `xml:"SEXTCODE,attr"`
	Code       string `xml:"CODE,attr"`
	RegionCode string `xml:"REGIONCODE,attr"`
	PlainCode  string `xml:"PLAINCODE,attr"`
	PostalCode string `xml:"POSTALCODE,attr"`
	Okato      string `xml:"OKATO,attr"`
	Oktmo      string `xml:"OKTMO,attr"`
	IfNsFl     string `xml:"IFNSFL,attr"`
	IfNsUl     string `xml:"IFNSUL,attr"`
	TerrIfNsFl string `xml:"TERRIFNSFL,attr"`
	TerrIfNsUl string `xml:"TERRIFNSUL,attr"`
	NormDoc    string `xml:"NORMDOC,attr"`
	ActStatus  string `xml:"ACTSTATUS,attr"`
	LiveStatus string `xml:"LIVESTATUS,attr"`
	CurrStatus string `xml:"CURRSTATUS,attr"`
	OperStatus string `xml:"OPERSTATUS,attr"`
	StartDate  string `xml:"STARTDATE,attr"`
	EndDate    string `xml:"ENDDATE,attr"`
	UpdateDate string `xml:"UPDATEDATE,attr"`
}

func (a AddressObject) GetXmlFile() string {
	return "AS_ADDROBJ_"
}

func (a AddressObject) TableName() string {
	return "fias_address"
}
