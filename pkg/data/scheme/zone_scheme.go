package scheme

type DnsZone struct {
	Zone Zone `json:"zone,omitempty"`
}

type Zone struct {
	ZoneName    string `json:"zoneName,omitempty"`
	Description string `json:"description,omitempty"`
}

