package powerwall

///////////////////////////////////////////////////////////////////////////////

// TroubleshootingProblemsData contains info returned by the "troubleshooting/problems" API call.
//
// The format of the entries in the Problems list is currently unknown.  If you
// have any examples, please let us know!
//
// This structure is returned by the GetProblems function.
type TroubleshootingProblemsData struct {
	Problems []interface{} `json:"problems"` // TODO: Unsure what type these values are when present
}

// GetProblems returns info about "troubleshooting problems" currently reported
// by the Powerwall gateway.  The format of these reports is currently not
// documented, but it can potentially still be useful to determine whether the
// list is non-empty (and thus some investigation is required) or not.
//
// See the TroubleshootingProblemsData type for more information on what fields this returns.
func (c *Client) GetProblems() (*TroubleshootingProblemsData, error) {
	c.checkLogin()
	result := TroubleshootingProblemsData{}
	err := c.apiGetJson("troubleshooting/problems", &result)
	return &result, err
}
