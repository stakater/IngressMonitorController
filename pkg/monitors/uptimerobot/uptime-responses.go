package uptimerobot

// UptimeRobot API v3 DTOs (camelCase JSON).
// v3 has no stat/error envelope anymore: HTTP 2xx means success,
// any other status code is an error.

type UptimeMonitorPagination struct {
	NextLink *string                `json:"nextLink"`
	Data     []UptimeMonitorMonitor `json:"data"`
}

type UptimeMonitorMonitor struct {
	ID                       int                         `json:"id"`
	FriendlyName             string                      `json:"friendlyName"`
	URL                      string                      `json:"url"`
	Type                     string                      `json:"type"`
	Interval                 int                         `json:"interval"`
	Timeout                  int                         `json:"timeout"`
	Status                   string                      `json:"status"`
	KeywordType              string                      `json:"keywordType"`
	KeywordValue             string                      `json:"keywordValue"`
	SuccessHttpResponseCodes []string                    `json:"successHttpResponseCodes"`
	Psps                     []UptimePublicStatusPage    `json:"psps"`
	Tags                     []UptimeMonitorTag          `json:"tags"`
	AssignedAlertContacts    []UptimeMonitorAlertContact `json:"assignedAlertContacts"`
	CreateDateTime           int64                       `json:"createDateTime"`
}

type UptimeMonitorAlertContact struct {
	AlertContactId int `json:"alertContactId"`
	Threshold      int `json:"threshold"`
	Recurrence     int `json:"recurrence"`
}

type UptimeMonitorTag struct {
	Name string `json:"name"`
}

// UptimeMonitorMonitorRequest is the request body for POST /monitors and PATCH /monitors/{id}
type UptimeMonitorMonitorRequest struct {
	FriendlyName             string                      `json:"friendlyName"`
	Type                     string                      `json:"type,omitempty"`
	URL                      string                      `json:"url,omitempty"`
	Interval                 int                         `json:"interval,omitempty"`
	Timeout                  int                         `json:"timeout,omitempty"`
	KeywordType              string                      `json:"keywordType,omitempty"`
	KeywordValue             string                      `json:"keywordValue,omitempty"`
	MaintenanceWindowsIds    []int                       `json:"maintenanceWindowsIds,omitempty"`
	AssignedAlertContacts    []UptimeMonitorAlertContact `json:"assignedAlertContacts,omitempty"`
	SuccessHttpResponseCodes []string                    `json:"successHttpResponseCodes,omitempty"`
}

// UptimePublicStatusPage is the v3 PspDto. It is also reused for the psps
// references embedded in MonitorDto (only id/friendlyName matter there).
type UptimePublicStatusPage struct {
	ID           int    `json:"id"`
	FriendlyName string `json:"friendlyName"`
	MonitorIds   []int  `json:"monitorIds"`
	Status       string `json:"status"`
}

type UptimeStatusPagesPagination struct {
	NextLink *string                  `json:"nextLink"`
	Data     []UptimePublicStatusPage `json:"data"`
}

// pspRequest is the request body for POST /psps
type pspRequest struct {
	FriendlyName string `json:"friendlyName"`
	MonitorIds   []int  `json:"monitorIds"`
	Status       string `json:"status"`
}

// pspPatchRequest is the request body for PATCH /psps/{id};
// the full monitorIds array replaces the existing set
type pspPatchRequest struct {
	FriendlyName string `json:"friendlyName"`
	MonitorIds   []int  `json:"monitorIds"`
}
