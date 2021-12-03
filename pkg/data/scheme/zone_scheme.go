package scheme

type DnsZone struct {
	Zone Zone `json:"zone,omitempty"`
}

type Zone struct {
	ZoneName    string `json:"zoneName,omitempty"`
	Description string `json:"description,omitempty"`
}

type DnsZoneList struct {
	TotalCount int `json:"totalCount"`
	ZoneList   ZoneList `json:"zoneList"`
}

type ZoneList []struct {
	ZoneName 	   string    `json:"zoneName"`
	ZoneID         string    `json:"zoneId"`
}