// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package svgtext

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

func GetCData(data []byte) ([]string, error) {
	var cdata []string

	reader := bytes.NewBuffer(data)
	decoder := xml.NewDecoder(reader)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		//if element, ok := token.(xml.StartElement); ok {
		//}
		if charData, ok := token.(xml.CharData); ok {
			str := strings.TrimSpace(string([]byte(charData)))
			if len(str) > 0 {
				cdata = append(cdata, string([]byte(charData)))
			}
		}
	}

	return cdata, nil
}
