package tenhou

import (
	"encoding/xml"
)

// 설명
// 설명
type RecordAction struct {
	XMLName xml.Name
	message
}

func (a *RecordAction) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	a.Tag = start.Name.Local
	type action RecordAction // 설명
	return d.DecodeElement((*action)(a), &start)
}

type Record struct {
	XMLName xml.Name        `xml:"mjloggm"`
	Actions []*RecordAction `xml:",any"`
}
