package powerwall

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Most time values in the API are produced in standard ISO-8601 format, which
// works just fine for unmarshalling to time.Time as well.  However, the
// "start_time" field of the status API call is not returned in this format for
// some reason and thus will not unmarshal directly to a time.Time value.  We
// provide a custom type to handle this case.
type NonIsoTime struct {
	time.Time
}

const nonIsoTimeFormat = "2006-01-02 15:04:05 -0700"

func (v *NonIsoTime) UnmarshalJSON(p []byte) error {
	t, err := time.Parse(nonIsoTimeFormat, strings.Replace(string(p), `"`, ``, -1))
	if err == nil {
		*v = NonIsoTime{t}
	}
	return err
}

// Durations in the API are typically represented as strings in duration-string
// format ("1h23m45.67s", etc).  Go's time.Duration type actually produces this
// format natively, yet will not parse it as an input when unmarshalling JSON
// (grr), so we need a custom type (with a custom UnmarshalJSON function) to
// handle this.
type Duration struct {
	time.Duration
}

func (v *Duration) UnmarshalJSON(p []byte) error {
	d, err := time.ParseDuration(strings.Replace(string(p), `"`, ``, -1))
	if err == nil {
		*v = Duration{d}
	}
	return err
}

func (v *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// The DecodedAlert type is used for unpacking values in the "decoded_alert"
// field of GridFault structures.  These are actually encoded as a string,
// which itself contains a JSON representation of a list of maps, each one
// containing a "name" and "value".  For example:
//
// "[{\"name\":\"PINV_alertID\",\"value\":\"PINV_a008_vfCheckRocof\"},{\"name\":\"PINV_alertType\",\"value\":\"Warning\"}]"
//
// Needless to say, this encoding is rather cumbersome and redundant, so we
// instead provide a custom JSON decoder to decode these into a string/string
// map in the form 'name: value'.
type DecodedAlert map[string]string

func (v *DecodedAlert) UnmarshalJSON(data []byte) error {
	type entry struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
		Units interface{} `json:"units"`
	}

	strvalue := ""
	err := json.Unmarshal(data, &strvalue)
	if err != nil {
		return fmt.Errorf("error unmarshalling grid alert %s into string: %w", strvalue, err)
	}
	if strvalue == "" {
		// For an empty string, just return a nil map
		return nil
	}
	entries := []entry{}
	err = json.Unmarshal([]byte(strvalue), &entries)
	if err != nil {
		return fmt.Errorf("error unmarshalling grid alert %s: %w", strvalue, err)
	}
	*v = make(map[string]string, len(entries))
	for _, e := range entries {
		if iv, ok := e.Value.(int); ok {
			(*v)[e.Name] = strconv.Itoa(iv)
		} else if fv, ok := e.Value.(float64); ok {
			(*v)[e.Name] = strconv.FormatFloat(fv, 'f', -1, 64)
		} else if sv, ok := e.Value.(string); ok {
			(*v)[e.Name] = sv
		} else if bv, ok := e.Value.([]byte); ok {
			(*v)[e.Name] = fmt.Sprintf("%q", bv)
		} else {
			(*v)[e.Name] = fmt.Sprintf("unrecognized type %v", e)
		}
		if uv, ok := e.Units.(string); ok && (*v)[e.Name] != "" && uv != "" {
			(*v)[e.Name] += " " + uv
		}
	}
	return nil
}
