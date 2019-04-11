package uptime

type UptimeMonitorGetMonitorsResponse struct {
	Stat       string                  `json:"stat"`
	Pagination UptimeMonitorPagination `json:"pagination"`
	Monitors   []UptimeMonitorMonitor  `json:"monitors"`
}

type UptimeMonitorPagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

type UptimeMonitorMonitor struct {
	PK                      int                        `json:"pk"`
	URL                     string                     `json:"url"`
	Name                    string                     `json:"name"`
	CachedRespTime          float64                    `json:"cached_response_time"`
	CachedUptime            int                        `json:"cached_uptime"`
	ContactGroups           []string                   `json:"contact_groups"`
	CreatedAt               string                     `json:"created_at"`
	ModifiedAt              string                     `json:"modified_at"`
	Locations               []string                   `json:"locations"`
	Tags                    []string                   `json:"tags"`
	CheckType               string                     `json:"check_type"`
	Escalations             []string                   `json:"escalations"`
	Maintainance            []UptimeMonitorMaintenance `json:"monitoring_service_type"`
	IsPaused                bool                       `json:"is_paused"`
	StateIsUp               bool                       `json:"state_is_up"`
	MspIPVersion            string                     `json:"msp_use_ip_version"`
	MspSensitivity          string                     `json:"msp_sensitivity"`
	MspInterval             string                     `json:"msp_interval"`
	MspHeaders              string                     `json:"msp_headers"`
	MspNotes                string                     `json:"msp_notes"`
	MspExpectString         string                     `json:"msp_expect_string"`
	MspAddress              string                     `json:"msp_address"`
	MspSendString           string                     `json:"msp_send_string"`
	MspUsername             string                     `json:"msp_username"`
	MspExpectStringType     string                     `json:"msp_expect_string_type"`
	MspPassword             string                     `json:"msp_password"`
	MspThreshold            int                        `json:"msp_threshold"`
	MspIncludeGlobalMetrics bool                       `json:"msp_include_global_metrics"`
	MspPort                 int                        `json:"msp_port"`
	StatsURL                string                     `json:"stats_url"`
	AlertsURL               string                     `json:"alerts_url"`
}

type UptimeMonitorMaintenance struct {
	Timezone string   `json:"timezone"`
	State    string   `json:"state"`
	Schedule []string `json:"schedule"`
}

type UptimeMonitorLogs struct {
	Type     int `json:"type"`
	Datetime int `json:"datetime"`
	Duration int `json:"duration"`
}

type UptimeMonitorNewMonitorResponse struct {
	Stat    string                     `json:"stat"`
	Monitor UptimeMonitorMonitorStatus `json:"monitor"`
}

type UptimeMonitorMonitorStatus struct {
	ID     int `json:"id"`
	Status int `json:"status"`
}

type UptimeMonitorStatusMonitorResponse struct {
	Stat    string `json:"stat"`
	Monitor struct {
		ID int `json:"id"`
	} `json:"monitor"`
}
