package tenhou

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	data, err := ioutil.ReadFile("xx.xml")
	if err != nil {
		t.Fatal(err)
	}

	d := Record{}
	if err := xml.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}

	for _, action := range d.Actions {
		fmt.Println(*action)
	}
}
