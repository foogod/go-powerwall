// Functions for reading network interface info:
//
//   (*Client) GetNetworks()
//
package powerwall

///////////////////////////////////////////////////////////////////////////////

// NetworkData contains information returned by the "networks" API call for a
// particular network interface.
//
// A list of this structure is returned by the GetNetworks function.
type NetworkData struct {
	NetworkName string `json:"network_name"`
	Interface   string `json:"interface"`
	Dhcp        bool   `json:"dhcp"`
	Enabled     bool   `json:"enabled"`
	ExtraIps    []struct {
		IP      string `json:"ip"`
		Netmask int    `json:"netmask"`
	} `json:"extra_ips,omitempty"`
	Active                bool `json:"active"`
	Primary               bool `json:"primary"`
	LastTeslaConnected    bool `json:"lastTeslaConnected"`
	LastInternetConnected bool `json:"lastInternetConnected"`
	IfaceNetworkInfo      struct {
		NetworkName string `json:"network_name"`
		IPNetworks  []struct {
			IP   string `json:"ip"`
			Mask string `json:"mask"`
		} `json:"ip_networks"`
		Gateway        string `json:"gateway"`
		Interface      string `json:"interface"`
		State          string `json:"state"`
		StateReason    string `json:"state_reason"`
		SignalStrength int    `json:"signal_strength"`
		HwAddress      string `json:"hw_address"`
	} `json:"iface_network_info"`
	SecurityType string `json:"security_type"`
	Username     string `json:"username"`
}

// GetNetworks returns information about all of the network interfaces in the
// system, and their statuses.
//
// See the NetworkData type for more information on what fields this returns.
func (c *Client) GetNetworks() ([]NetworkData, error) {
	c.checkLogin()
	result := []NetworkData{}
	err := c.apiGetJson("networks", &result)
	return result, err
}
