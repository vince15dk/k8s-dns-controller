package scheme

type DnsZone struct {
	Zone Zone `json:"zone,omitempty"`
}

type Zone struct {
	ZoneName    string `json:"zoneName,omitempty"`
	Description string `json:"description,omitempty"`
}

type DnsZoneList struct {
	TotalCount int `json:"totalCount,omitempty"`
	ZoneList   ZoneList `json:"zoneList,omitempty"`
}

type ZoneList []struct {
	ZoneName 	   string    `json:"zoneName,omitempty"`
	ZoneID         string    `json:"zoneId,omitempty"`
}