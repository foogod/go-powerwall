// Functions for reading power meter data:
//
//   (*Client) GetMeters(category string)
//   (*Client) GetMetersAggregates()
//
package powerwall

import (
	"net/url"
	"time"
)

///////////////////////////////////////////////////////////////////////////////

// MeterAggregatesData contains fields returned by the "meters/aggregates" API
// call.  This reflects statistics collected across all of the meters in a
// given category (e.g. "site", "solar", "battery", "load", etc).
//
// This structure is returned by the GetMetersAggregates function.
type MeterAggregatesData struct {
	LastCommunicationTime             time.Time `json:"last_communication_time"`
	InstantPower                      float32   `json:"instant_power"`
	InstantReactivePower              float32   `json:"instant_reactive_power"`
	InstantApparentPower              float32   `json:"instant_apparent_power"`
	Frequency                         float32   `json:"frequency"`
	EnergyExported                    float32   `json:"energy_exported"`
	EnergyImported                    float32   `json:"energy_imported"`
	InstantAverageVoltage             float32   `json:"instant_average_voltage"`
	InstantAverageCurrent             float32   `json:"instant_average_current"`
	IACurrent                         float32   `json:"i_a_current"`
	IBCurrent                         float32   `json:"i_b_current"`
	ICCurrent                         float32   `json:"i_c_current"`
	LastPhaseVoltageCommunicationTime time.Time `json:"last_phase_voltage_communication_time"`
	LastPhasePowerCommunicationTime   time.Time `json:"last_phase_power_communication_time"`
	Timeout                           int       `json:"timeout"`
	NumMetersAggregated               int       `json:"num_meters_aggregated"`
	InstantTotalCurrent               float32   `json:"instant_total_current"`
}

// GetMetersAggregates fetches aggregated meter data for power transferred
// to/from each category of connection ("site", "solar", "battery", "load", etc).
//
// See the MetersAggregatesData type for more information on what fields this returns.
func (c *Client) GetMetersAggregates() (map[string]MeterAggregatesData, error) {
	c.checkLogin()
	result := map[string]MeterAggregatesData{}
	err := c.apiGetJson("meters/aggregates", &result)
	return result, err
}

///////////////////////////////////////////////////////////////////////////////

// MeterData contains fields returned by the "meters/<category>" API call, which
// returns information for each individual meter within that category.
//
// A list of this structure is returned by the GetMeters function.
type MeterData struct {
	ID         int    `json:"id"`
	Location   string `json:"location"`
	Type       string `json:"type"`
	Cts        []bool `json:"cts"`
	Inverted   []bool `json:"inverted"`
	Connection struct {
		ShortID      string `json:"short_id"`
		DeviceSerial string `json:"device_serial"`
		HTTPSConf    struct {
			ClientCert          string `json:"client_cert"`
			ClientKey           string `json:"client_key"`
			ServerCaCert        string `json:"server_ca_cert"`
			MaxIdleConnsPerHost int    `json:"max_idle_conns_per_host"`
		} `json:"https_conf"`
	} `json:"connection"`
	RealPowerScaleFactor float32 `json:"real_power_scale_factor"`
	CachedReadings       struct {
		LastCommunicationTime             time.Time `json:"last_communication_time"`
		InstantPower                      float32   `json:"instant_power"`
		InstantReactivePower              float32   `json:"instant_reactive_power"`
		InstantApparentPower              float32   `json:"instant_apparent_power"`
		Frequency                         float32   `json:"frequency"`
		EnergyExported                    float32   `json:"energy_exported"`
		EnergyImported                    float32   `json:"energy_imported"`
		InstantAverageVoltage             float32   `json:"instant_average_voltage"`
		InstantAverageCurrent             float32   `json:"instant_average_current"`
		IACurrent                         float32   `json:"i_a_current"`
		IBCurrent                         float32   `json:"i_b_current"`
		ICCurrent                         float32   `json:"i_c_current"`
		VL1N                              float32   `json:"v_l1n"`
		VL2N                              float32   `json:"v_l2n"`
		LastPhaseVoltageCommunicationTime time.Time `json:"last_phase_voltage_communication_time"`
		RealPowerA                        float32   `json:"real_power_a"`
		RealPowerB                        float32   `json:"real_power_b"`
		ReactivePowerA                    float32   `json:"reactive_power_a"`
		ReactivePowerB                    float32   `json:"reactive_power_b"`
		LastPhasePowerCommunicationTime   time.Time `json:"last_phase_power_communication_time"`
		SerialNumber                      string    `json:"serial_number"`
		Timeout                           int       `json:"timeout"`
		InstantTotalCurrent               float32   `json:"instant_total_current"`
	} `json:"Cached_readings"`
	CtVoltageReferences struct {
		Ct1 string `json:"ct1"`
		Ct2 string `json:"ct2"`
		Ct3 string `json:"ct3"`
	} `json:"ct_voltage_references"`
}

// GetMeters fetches detailed meter data for each meter under the specified
// category.  Note that as of this writing, only the "site" and "solar"
// categories appear to return any data.
//
// If the API returns no data (i.e. an unsupported category name was provided),
// this will return nil.
//
// See the MeterData type for more information on what fields this returns.
func (c *Client) GetMeters(category string) ([]MeterData, error) {
	c.checkLogin()
	result := []MeterData{}
	err := c.apiGetJson("meters/"+url.PathEscape(category), &result)
	return result, err
}
