package scheme

import "time"

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

type RecordSetLists struct {
	TotalCount    int `json:"totalCount,omitempty"`
	RecordsetList []RecordSetList `json:"recordsetList,omitempty"`
}

type RecordSetList struct {
	RecordsetID     string    `json:"recordsetId,omitempty"`
	RecordsetName   string    `json:"recordsetName,omitempty"`
	RecordsetType   string    `json:"recordsetType,omitempty"`
	RecordsetTTL    int       `json:"recordsetTtl,omitempty"`
	RecordsetStatus string    `json:"recordsetStatus,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}

