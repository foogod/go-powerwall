package powerwall

///////////////////////////////////////////////////////////////////////////////

type TroubleshootingProblemsData struct {
	Problems []interface{} `json:"problems"` // TODO: Unsure what type these values are when present
}

func (c *Client) GetProblems() (*TroubleshootingProblemsData, error) {
	c.checkLogin()
	result := TroubleshootingProblemsData{}
	err := c.apiGetJson("troubleshooting/problems", &result)
	return &result, err
}
