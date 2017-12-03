package isbn

import (
	"encoding/xml"
	"log"
	"os"
	"strconv"
	"strings"
)

// rangeMessageXML is for containing the contents of the RangeMessage.xml
// file. Structure generated using https://github.com/miku/zek/
// `zek -p < RangeMessage.xml > temp_parser.go
type rangeMessageXML struct {
	XMLName       xml.Name `xml:"ISBNRangeMessage"`
	Text          string   `xml:",chardata"`
	MessageSource struct {
		Text string `xml:",chardata"`
	} `xml:"MessageSource"`
	MessageSerialNumber struct {
		Text string `xml:",chardata"`
	} `xml:"MessageSerialNumber"`
	MessageDate struct {
		Text string `xml:",chardata"`
	} `xml:"MessageDate"`
	EANUCCPrefixes struct {
		Text   string `xml:",chardata"`
		EANUCC []struct {
			Text   string `xml:",chardata"`
			Prefix struct {
				Text string `xml:",chardata"`
			} `xml:"Prefix"`
			Agency struct {
				Text string `xml:",chardata"`
			} `xml:"Agency"`
			Rules struct {
				Text string `xml:",chardata"`
				Rule []struct {
					Text  string `xml:",chardata"`
					Range struct {
						Text string `xml:",chardata"`
					} `xml:"Range"`
					Length struct {
						Text string `xml:",chardata"`
					} `xml:"Length"`
				} `xml:"Rule"`
			} `xml:"Rules"`
		} `xml:"EAN.UCC"`
	} `xml:"EAN.UCCPrefixes"`
	RegistrationGroups struct {
		Text  string `xml:",chardata"`
		Group []struct {
			Text   string `xml:",chardata"`
			Prefix struct {
				Text string `xml:",chardata"`
			} `xml:"Prefix"`
			Agency struct {
				Text string `xml:",chardata"`
			} `xml:"Agency"`
			Rules struct {
				Text string `xml:",chardata"`
				Rule []struct {
					Text  string `xml:",chardata"`
					Range struct {
						Text string `xml:",chardata"`
					} `xml:"Range"`
					Length struct {
						Text string `xml:",chardata"`
					} `xml:"Length"`
				} `xml:"Rule"`
			} `xml:"Rules"`
		} `xml:"Group"`
	} `xml:"RegistrationGroups"`
}

type registrant struct {
	Agency string
	Ranges [][]int
}

type rangeData map[string]map[string]registrant

var rmd = make(rangeData)

// HasRangeData is used for indicating whether or not the range data
// has been loaded.
func HasRangeData() bool {
	if len(rmd) > 0 {
		return true
	}
	return false
}

// LoadRangeData loads a RangeMessage.xml file for use in parsing and
// validating ISBNs. While this file does not appear to change often
// it does still change and twould be a shame to have to re-compile
// whenever the contents did change.
//
// The RangeMessage.xml file to load should be available at:
// https://www.isbn-international.org/range_file_generation
func LoadRangeData(filename string) (bool, error) {

	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	var doc rangeMessageXML
	if err := dec.Decode(&doc); err != nil {
		return false, err
	}

	for _, rg := range doc.RegistrationGroups.Group {
		tokens := strings.Split(rg.Prefix.Text, "-")
		prefix := tokens[0]
		group := tokens[1]

		var reg registrant
		reg.Agency = rg.Agency.Text

		for _, rule := range rg.Rules.Rule {
			rLen, err := strconv.Atoi(rule.Length.Text)
			if err != nil {
				log.Println(err)
				continue
			}

			if rLen > 0 {

				tokens := strings.Split(rule.Range.Text, "-")
				rStart, err := strconv.Atoi(tokens[0][:rLen])
				if err != nil {
					log.Println(err)
					continue
				}
				rEnd, err := strconv.Atoi(tokens[1][:rLen])
				if err != nil {
					log.Println(err)
					continue
				}

				if rEnd == 0 {
					continue
				}

				var rng = make([]int, 3)
				rng[0] = rStart
				rng[1] = rEnd
				rng[2] = rLen
				reg.Ranges = append(reg.Ranges, rng)
			}
		}

		if rmd[prefix] == nil {
			rmd[prefix] = make(map[string]registrant)
		}
		rmd[prefix][group] = reg
	}

	return true, nil
}
