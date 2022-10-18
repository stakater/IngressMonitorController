package statuscake

// StatusCakeMonitor response Structure for GetAll and GetByName API's for Statuscake

type StatusCakeMonitor struct {
	StatusCakeData     []StatusCakeMonitorData   `json:"data"`
	StatusCakeMetadata StatusCakeMonitorMetadata `json:"metadata"`
}

type StatusCakeMonitorData struct {
	TestID       string   `json:"id"`
	Paused       bool     `json:"paused"`
	WebsiteName  string   `json:"name"`
	WebsiteURL   string   `json:"website_url"`
	TestType     string   `json:"test_type"`
	CheckRate    int      `json:"check_rate"`
	ContactGroup []string `json:"contact_groups"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
	Uptime       int      `json:"uptime"`
}
type StatusCakeMonitorMetadata struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	PageCount  int `json:"page_count"`
	TotalCount int `json:"total_count"`
}

// StatusCakeUpsertResponse response Structure for Insert API for Statuscake

// we dont need it anymore because of statuscake api changes
type StatusCakeUpsertResponse struct {
	Issues   interface{} `json:"Issues"`
	Success  bool        `json:"Success"`
	Message  string      `json:"Message"`
	InsertID int         `json:"InsertID"`
}
