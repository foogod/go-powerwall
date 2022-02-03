// Functions for getting info about the system state:
//
//   (*Client) GetSystemStatus()
//   (*Client) GetGridFaults()
//   (*Client) GetGridStatus()
//   (*Client) GetSOE()
//   (*Client) GetOperation()
//
package powerwall

import "time"

///////////////////////////////////////////////////////////////////////////////

// SystemStatusData contains fields returned by the "system_status" API call.
// This contains a lot of information about the general state of the system and
// how it is operating, such as battery charge, utility power status, etc.
//
// This structure is returned by the GetSystemStatus function.
type SystemStatusData struct {
	CommandSource                  string  `json:"command_source"`
	BatteryTargetPower             float32 `json:"battery_target_power"`
	BatteryTargetReactivePower     float32 `json:"battery_target_reactive_power"`
	NominalFullPackEnergy          float32 `json:"nominal_full_pack_energy"`
	NominalEnergyRemaining         float32 `json:"nominal_energy_remaining"`
	MaxPowerEnergyRemaining        float32 `json:"max_power_energy_remaining"`
	MaxPowerEnergyToBeCharged      float32 `json:"max_power_energy_to_be_charged"`
	MaxChargePower                 float32 `json:"max_charge_power"`
	MaxDischargePower              float32 `json:"max_discharge_power"`
	MaxApparentPower               float32 `json:"max_apparent_power"`
	InstantaneousMaxDischargePower float32 `json:"instantaneous_max_discharge_power"`
	InstantaneousMaxChargePower    float32 `json:"instantaneous_max_charge_power"`
	GridServicesPower              float32 `json:"grid_services_power"`
	SystemIslandState              string  `json:"system_island_state"`
	AvailableBlocks                int     `json:"available_blocks"`
	BatteryBlocks                  []struct {
		Type                   string        `json:"Type"`
		PackagePartNumber      string        `json:"PackagePartNumber"`
		PackageSerialNumber    string        `json:"PackageSerialNumber"`
		DisabledReasons        []interface{} `json:"disabled_reasons"` // TODO: Unclear what type these entries are when present.
		PinvState              string        `json:"pinv_state"`
		PinvGridState          string        `json:"pinv_grid_state"`
		NominalEnergyRemaining float32       `json:"nominal_energy_remaining"`
		NominalFullPackEnergy  float32       `json:"nominal_full_pack_energy"`
		POut                   float32       `json:"p_out"`
		QOut                   float32       `json:"q_out"`
		VOut                   float32       `json:"v_out"`
		FOut                   float32       `json:"f_out"`
		IOut                   float32       `json:"i_out"`
		EnergyCharged          float32       `json:"energy_charged"`
		EnergyDischarged       float32       `json:"energy_discharged"`
		OffGrid                bool          `json:"off_grid"`
		VfMode                 bool          `json:"vf_mode"`
		WobbleDetected         bool          `json:"wobble_detected"`
		ChargePowerClamped     bool          `json:"charge_power_clamped"`
		BackupReady            bool          `json:"backup_ready"`
		OpSeqState             string        `json:"OpSeqState"`
		Version                string        `json:"version"`
	} `json:"battery_blocks"`
	FfrPowerAvailabilityHigh   float32         `json:"ffr_power_availability_high"`
	FfrPowerAvailabilityLow    float32         `json:"ffr_power_availability_low"`
	LoadChargeConstraint       float32         `json:"load_charge_constraint"`
	MaxSustainedRampRate       float32         `json:"max_sustained_ramp_rate"`
	GridFaults                 []GridFaultData `json:"grid_faults"`
	CanReboot                  string          `json:"can_reboot"`
	SmartInvDeltaP             float32         `json:"smart_inv_delta_p"`
	SmartInvDeltaQ             float32         `json:"smart_inv_delta_q"`
	LastToggleTimestamp        time.Time       `json:"last_toggle_timestamp"`
	SolarRealPowerLimit        float32         `json:"solar_real_power_limit"`
	Score                      float32         `json:"score"`
	BlocksControlled           int             `json:"blocks_controlled"`
	Primary                    bool            `json:"primary"`
	AuxiliaryLoad              float32         `json:"auxiliary_load"`
	AllEnableLinesHigh         bool            `json:"all_enable_lines_high"`
	InverterNominalUsablePower float32         `json:"inverter_nominal_usable_power"`
	ExpectedEnergyRemaining    float32         `json:"expected_energy_remaining"`
}

// GetSystemStatus performs a "system_status" API call to fetch general
// information about the system operation and state.
//
// See the SystemStatusData type for more information on what fields this returns.
func (c *Client) GetSystemStatus() (*SystemStatusData, error) {
	c.checkLogin()
	result := SystemStatusData{}
	err := c.apiGetJson("system_status", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

// GridFaultData contains fields returned by the "system_status/grid_faults" API call.
//
// This structure is returned by the GetSystemStatus and GetGridFaults functions.
type GridFaultData struct {
	Timestamp              int64        `json:"timestamp"`
	AlertName              string       `json:"alert_name"`
	AlertIsFault           bool         `json:"alert_is_fault"`
	DecodedAlert           DecodedAlert `json:"decoded_alert"`
	AlertRaw               int64        `json:"alert_raw"`
	GitHash                string       `json:"git_hash"`
	SiteUID                string       `json:"site_uid"`
	EcuType                string       `json:"ecu_type"`
	EcuPackagePartNumber   string       `json:"ecu_package_part_number"`
	EcuPackageSerialNumber string       `json:"ecu_package_serial_number"`
}

// GetGridFaults returns a list of any current "grid fault" events detected by
// the system.  These generally indicate some issue with the utility power,
// such as being over or undervoltage, etc.
//
// This same information is also returned by GetSystemStatus in the GridFaults
// field.
//
// See the GridFaultData type for more information on what fields this returns.
func (c *Client) GetGridFaults() ([]GridFaultData, error) {
	c.checkLogin()
	result := []GridFaultData{}
	err := c.apiGetJson("system_status/grid_faults", &result)
	return result, err
}

///////////////////////////////////////////////////////////////////////////////

// GridStatusData contains fields returned by the "system_status/grid_status" API call.
//
// This structure is returned by the GetGridStatus function.
type GridStatusData struct {
	GridStatus         string `json:"grid_status"`
	GridServicesActive bool   `json:"grid_services_active"`
}

// Possible options for the GridStatus field of GridStatusData:
//
// The powerwall can be in one of three states.  "Connected" indicates that it
// is receiving power from the utility and functioning normally.  "Islanded"
// means that it has gone off-grid and is generating power entirely
// independently of the utility power (if utility power has been lost, or if it
// has been put into "off-grid" mode).  "Transitioning" means that it is in the
// process of going from being off-grid to back on-grid, which can take a
// little bit of time to verify the supplied power is clean and synchronize
// with it, etc.
const (
	GridStatusConnected  = "SystemGridConnected"
	GridStatusIslanded   = "SystemIslandedActive"
	GridStatusTransition = "SystemTransitionToGrid"
)

// GetGridStatus returns information about the current state of the Powerwall's
// connection to the utility power grid.
//
// See the GridStatusData type for more information on what fields this returns.
func (c *Client) GetGridStatus() (*GridStatusData, error) {
	c.checkLogin()
	result := GridStatusData{}
	err := c.apiGetJson("system_status/grid_status", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

// SOEData contains fields returned by the "system_status/soe" API call.
// This currently just returns a single "Percentage" field, indicating the
// total amount of charge across all batteries.
//
// This structure is returned by the GetSOE function.
type SOEData struct {
	Percentage float32 `json:"percentage"`
}

// GetSOE returns information about the current "State Of Energy" of the
// system.  That is, how much total charge (as a percentage) is present across
// all batteries.
//
// See the SOEData type for more information on what fields this returns.
func (c *Client) GetSOE() (*SOEData, error) {
	c.checkLogin()
	result := SOEData{}
	err := c.apiGetJson("system_status/soe", &result)
	return &result, err
}

///////////////////////////////////////////////////////////////////////////////

// OperationData contains fields returned by the "operation" API call.
//
// This structure is returned by the GetOperation function.
type OperationData struct {
	RealMode                string  `json:"real_mode"`
	BackupReservePercent    float32 `json:"backup_reserve_percent"`
	FreqShiftLoadShedSoe    float32 `json:"freq_shift_load_shed_soe"`
	FreqShiftLoadShedDeltaF float64 `json:"freq_shift_load_shed_delta_f"`
}

// Possible options for the RealMode field of OperationData:
const (
	OperationModeSelf      = "self_consumption" // Reported as "Self Powered" in the app
	OperationModeTimeBased = "autonomous"       // Reported as "Time-Based Control" in the app
)

// GetOperation returns information about the current operation mode
// configuration.  This includes whether the Powerwall is configured for "Self
// Powered" mode or "Time Based Control", how much of the battery is reserved
// for backup use, etc.
//
// See the OperationData type for more information on what fields this returns.
func (c *Client) GetOperation() (*OperationData, error) {
	c.checkLogin()
	result := OperationData{}
	err := c.apiGetJson("operation", &result)
	return &result, err
}
