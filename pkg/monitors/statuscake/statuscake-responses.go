package statuscake

import (
	statuscake "github.com/StatusCakeDev/statuscake-go"
)

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
	Uptime       float64  `json:"uptime"`
}
type StatusCakeMonitorMetadata struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	PageCount  int `json:"page_count"`
	TotalCount int `json:"total_count"`
}

// TODO use statuscake managed structs, rather than managing own structs
type StatusCakeData struct {
	statuscake.UptimeTest
}
