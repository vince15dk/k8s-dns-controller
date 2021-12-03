package scheme

type Records struct {
	RecordSet Recordset `json:"recordset,omitempty"`
}

type Recordset struct {
	RecordsetName string     `json:"recordsetName,omitempty"`
	RecordsetType string     `json:"recordsetType,omitempty"`
	RecordsetTTL  int        `json:"recordsetTtl,omitempty"`
	RecordList    RecordList `json:"recordList,omitempty"`
}

type RecordList []struct {
	RecordDisabled bool   `json:"recordDisabled,omitempty"`
	RecordContent  string `json:"recordContent,omitempty"`
}
