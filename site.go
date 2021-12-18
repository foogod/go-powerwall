// Functions for getting general info about the gateway and site:
//
//   (*Client) GetStatus()
//   (*Client) GetSiteInfo()
//   (*Client) GetSitemaster()
//
package powerwall

///////////////////////////////////////////////////////////////////////////////

// StatusData contains general system information returned by the "status" API
// call, such as the device identification number, type, software version, etc.
//
// This structure is returned by the GetStatus function.
type StatusData struct {
	Din              string      `json:"din"`
	StartTime        NonIsoTime  `json:"start_time"`
	UpTime           Duration    `json:"up_time_seconds"`
	IsNew            bool        `json:"is_new"`
	Version          string      `json:"version"`
	GitHash          string      `json:"git_hash"`
	CommissionCount  int         `json:"commission_count"`
	DeviceType       string      `json:"device_type"`
	SyncType         string      `json:"sync_type"`
	Leader           string      `json:"leader"`
	Followers        interface{} `json:"followers"` // TODO: Unsure what type this returns when present
	CellularDisabled bool        `json:"cellular_disabled"`
}

// GetStatus performs a "status" API call to fetch basic information about the
// system, such as the device identification number (part number + serial
// number), software version, uptime, device type, etc.
//
// Note that this is one of the only API calls which can be made without logging in first.
//
// See the StatusData type for more information on what fields this returns.
func (c *Client) GetStatus() (*StatusData, error) {
	result := StatusData{}
	err := c.apiGetJson("status", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

// SiteInfoData contains information returned by the "site_info" API call.
//
// This structure is returned by the GetSiteInfo function.
type SiteInfoData struct {
	SiteName               string  `json:"site_name"`
	TimeZone               string  `json:"timezone"`
	MaxSiteMeterPowerKW    int     `json:"max_site_meter_power_kW"`
	MinSiteMeterPowerKW    int     `json:"min_site_meter_power_kW"`
	MeasuredFrequency      float32 `json:"measured_frequency"`
	MaxSystemEnergyKWH     float32 `json:"max_system_energy_kWh"`
	MaxSystemPowerKW       float32 `json:"max_system_power_kW"`
	NominalSystemEnergyKWH float32 `json:"nominal_system_energy_kWh"`
	NominalSystemPowerKW   float32 `json:"nominal_system_power_kW"`
	GridData               struct {
		GridCode           string `json:"grid_code"`
		GridVoltageSetting int    `json:"grid_voltage_setting"`
		GridFreqSetting    int    `json:"grid_freq_setting"`
		GridPhaseSetting   string `json:"grid_phase_setting"`
		Country            string `json:"country"`
		State              string `json:"state"`
		Distributor        string `json:"distributor"`
		Utility            string `json:"utility"`
		Retailer           string `json:"retailer"`
		Region             string `json:"region"`
	} `json:"grid_code"`
}

// GetSiteInfo returns information about the "site".
// This includes general information about the installation, configuration of
// the hardware, utility grid it is connected to, etc.
//
// See the SiteInfoData type for more information on what fields this returns.
func (c *Client) GetSiteInfo() (*SiteInfoData, error) {
	c.checkLogin()
	result := SiteInfoData{}
	err := c.apiGetJson("site_info", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

// SitemasterData contains information returned by the "sitemaster" API call.
//
// The CanReboot field indicates whether the sitemaster can be
// stopped/restarted without disrupting anything or not.  It will be either
// "Yes", or it will be a string indicating the reason why it can't be stopped
// right now (such as "Power flow is too high").  (see the SitemasterReboot*
// constants for known possible values)
//
// Note that it is still possible to stop the sitemaster even under these
// conditions, but it is necessary to set the "force" option to true when
// calling the "sitemaster/stop" API in that case.
//
// This structure is returned by the GetSitemaster function.
type SitemasterData struct {
	Status           string `json:"status"`
	Running          bool   `json:"running"`
	ConnectedToTesla bool   `json:"connected_to_tesla"`
	PowerSupplyMode  bool   `json:"power_supply_mode"`
	CanReboot        string `json:"can_reboot"`
}

// Possible values for the Status field of the SitemasterData struct:
const (
	SitemasterStatusUp   = "StatusUp"
	SitemasterStatusDown = "StatusDown"
)

// Known possible values returned for the CanReboot field of the SitemasterData
// struct:
//
// (Note that this list is almost certainly incomplete at this point)
const (
	SitemasterRebootOK           = "Yes"
	SitemasterRebootPowerTooHigh = "Power flow is too high"
)

// GetSitemaster returns information about the Powerwall gateway's "sitemaster"
// process, which handles most of the normal day-to-day tasks of the system
// (monitoring power, directing it to the right places, etc).
//
// See the SitemasterData type for more information on what fields this returns.
func (c *Client) GetSitemaster() (*SitemasterData, error) {
	c.checkLogin()
	result := SitemasterData{}
	err := c.apiGetJson("sitemaster", &result)
	return &result, err
}
