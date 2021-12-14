package powerwall

import "time"

///////////////////////////////////////////////////////////////////////////////

type SystemStatus struct {
	CommandSource                  string  `json:"command_source"`
	BatteryTargetPower             float32 `json:"battery_target_power"`
	BatteryTargetReactivePower     float32     `json:"battery_target_reactive_power"`
	NominalFullPackEnergy          float32     `json:"nominal_full_pack_energy"`
	NominalEnergyRemaining         float32     `json:"nominal_energy_remaining"`
	MaxPowerEnergyRemaining        float32     `json:"max_power_energy_remaining"`
	MaxPowerEnergyToBeCharged      float32     `json:"max_power_energy_to_be_charged"`
	MaxChargePower                 float32     `json:"max_charge_power"`
	MaxDischargePower              float32     `json:"max_discharge_power"`
	MaxApparentPower               float32     `json:"max_apparent_power"`
	InstantaneousMaxDischargePower float32     `json:"instantaneous_max_discharge_power"`
	InstantaneousMaxChargePower    float32     `json:"instantaneous_max_charge_power"`
	GridServicesPower              float32     `json:"grid_services_power"`
	SystemIslandState              string  `json:"system_island_state"`
	AvailableBlocks                int     `json:"available_blocks"`
	BatteryBlocks                  []struct {
		Type                   string        `json:"Type"`
		PackagePartNumber      string        `json:"PackagePartNumber"`
		PackageSerialNumber    string        `json:"PackageSerialNumber"`
		DisabledReasons        []interface{} `json:"disabled_reasons"`
		PinvState              string        `json:"pinv_state"`
		PinvGridState          string        `json:"pinv_grid_state"`
		NominalEnergyRemaining float32           `json:"nominal_energy_remaining"`
		NominalFullPackEnergy  float32           `json:"nominal_full_pack_energy"`
		POut                   float32           `json:"p_out"`
		QOut                   float32           `json:"q_out"`
		VOut                   float32       `json:"v_out"`
		FOut                   float32       `json:"f_out"`
		IOut                   float32       `json:"i_out"`
		EnergyCharged          float32           `json:"energy_charged"`
		EnergyDischarged       float32           `json:"energy_discharged"`
		OffGrid                bool          `json:"off_grid"`
		VfMode                 bool          `json:"vf_mode"`
		WobbleDetected         bool          `json:"wobble_detected"`
		ChargePowerClamped     bool          `json:"charge_power_clamped"`
		BackupReady            bool          `json:"backup_ready"`
		OpSeqState             string        `json:"OpSeqState"`
		Version                string        `json:"version"`
	} `json:"battery_blocks"`
	FfrPowerAvailabilityHigh   float32           `json:"ffr_power_availability_high"`
	FfrPowerAvailabilityLow    float32           `json:"ffr_power_availability_low"`
	LoadChargeConstraint       float32           `json:"load_charge_constraint"`
	MaxSustainedRampRate       float32           `json:"max_sustained_ramp_rate"`
	GridFaults                 []GridFault `json:"grid_faults"`
	CanReboot                  string        `json:"can_reboot"`
	SmartInvDeltaP             float32           `json:"smart_inv_delta_p"`
	SmartInvDeltaQ             float32           `json:"smart_inv_delta_q"`
	LastToggleTimestamp        time.Time        `json:"last_toggle_timestamp"`
	SolarRealPowerLimit        float32           `json:"solar_real_power_limit"`
	Score                      float32           `json:"score"`
	BlocksControlled           int           `json:"blocks_controlled"`
	Primary                    bool          `json:"primary"`
	AuxiliaryLoad              float32           `json:"auxiliary_load"`
	AllEnableLinesHigh         bool          `json:"all_enable_lines_high"`
	InverterNominalUsablePower float32           `json:"inverter_nominal_usable_power"`
	ExpectedEnergyRemaining    float32           `json:"expected_energy_remaining"`
}

func (c *Client) GetSystemStatus() (*SystemStatus, error) {
	c.checkLogin()
	result := SystemStatus{}
	err := c.apiGetJson("system_status", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

type GridFault struct {
	Timestamp              int64  `json:"timestamp"`
	AlertName              string `json:"alert_name"`
	AlertIsFault           bool   `json:"alert_is_fault"`
	DecodedAlert           DecodedAlert `json:"decoded_alert"`
	AlertRaw               int64  `json:"alert_raw"`
	GitHash                string `json:"git_hash"`
	SiteUID                string `json:"site_uid"`
	EcuType                string `json:"ecu_type"`
	EcuPackagePartNumber   string `json:"ecu_package_part_number"`
	EcuPackageSerialNumber string `json:"ecu_package_serial_number"`
}

func (c *Client) GetGridFaults() (*[]GridFault, error) {
	c.checkLogin()
	result := []GridFault{}
	err := c.apiGetJson("system_status/grid_faults", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

type GridStatus struct {
	GridStatus         string `json:"grid_status"`
	GridServicesActive bool   `json:"grid_services_active"`
}

const (
	GridStatusConnected  = "SystemGridConnected"
	GridStatusIslanded   = "SystemIslandedActive"
	GridStatusTransition = "SystemTransitionToGrid"
)

func (c *Client) GetGridStatus() (*GridStatus, error) {
	c.checkLogin()
	result := GridStatus{}
	err := c.apiGetJson("system_status/grid_status", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

type SOEData struct {
	Percentage float32 `json:"percentage"`
}

func (c *Client) GetSOE() (*SOEData, error) {
	c.checkLogin()
	result := SOEData{}
	err := c.apiGetJson("system_status/soe", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

type OperationData struct {
	RealMode                string  `json:"real_mode"`
	BackupReservePercent    float32     `json:"backup_reserve_percent"`
	FreqShiftLoadShedSoe    float32     `json:"freq_shift_load_shed_soe"`
	FreqShiftLoadShedDeltaF float64 `json:"freq_shift_load_shed_delta_f"`
}

func (c *Client) GetOperation() (*OperationData, error) {
	c.checkLogin()
	result := OperationData{}
	err := c.apiGetJson("operation", &result)
	return &result, err
}

