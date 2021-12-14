package powerwall

///////////////////////////////////////////////////////////////////////////////

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

func (c *Client) GetStatus() (*StatusData, error) {
	result := StatusData{}
	err := c.apiGetJson("status", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

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

func (c *Client) GetSiteInfo() (*SiteInfoData, error) {
	c.checkLogin()
	result := SiteInfoData{}
	err := c.apiGetJson("site_info", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

type SitemasterData struct {
	Status           string `json:"status"`
	Running          bool   `json:"running"`
	ConnectedToTesla bool   `json:"connected_to_tesla"`
	PowerSupplyMode  bool   `json:"power_supply_mode"`
	CanReboot        string `json:"can_reboot"`
}

func (c *Client) GetSitemaster() (*SitemasterData, error) {
	c.checkLogin()
	result := SitemasterData{}
	err := c.apiGetJson("sitemaster", &result)
	return &result, err
}
