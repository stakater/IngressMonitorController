package statuscake

//StatusCakeMonitorMonitor response Structure for GetAll and GetByName API's for Statuscake
type StatusCakeMonitorMonitor struct {
	TestID       int      `json:"TestID"`
	Paused       bool     `json:"Paused"`
	TestType     string   `json:"TestType"`
	WebsiteName  string   `json:"WebsiteName"`
	WebsiteURL   string   `json:"WebsiteURL"`
	ContactGroup []string `json:"ContactGroup"`
	ContactID    int      `json:"ContactID"`
	Status       string   `json:"Status"`
	Uptime       float64  `json:"Uptime"`
	Tags         []string `json:"Tags"`
	WebsiteHost  string   `json:"WebsiteHost"`
}

// StatusCakeUpsertResponse response Structure for Insert API for Statuscake
type StatusCakeUpsertResponse struct {
	Issues   interface{} `json:"Issues"`
	Success  bool        `json:"Success"`
	Message  string      `json:"Message"`
	InsertID int         `json:"InsertID"`
}
