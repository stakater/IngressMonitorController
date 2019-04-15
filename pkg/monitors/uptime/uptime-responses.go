package uptime

type UptimeMonitorGetMonitorsResponse struct {
	Count    int                    `json:"count"`
	Next     int                    `json:"next"`
	Previous int                    `json:"previous"`
	Monitors []UptimeMonitorMonitor `json:"results"`
}

type UptimeMonitorPagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

type UptimeMonitorMonitor struct {
	PK                      int                      `json:"pk"`
	URL                     string                   `json:"url"`
	Name                    string                   `json:"name"`
	CachedRespTime          float64                  `json:"cached_response_time"`
	CachedUptime            float64                  `json:"cached_uptime"`
	ContactGroups           []string                 `json:"contact_groups"`
	CreatedAt               string                   `json:"created_at"`
	ModifiedAt              string                   `json:"modified_at"`
	Locations               []string                 `json:"locations"`
	Tags                    []string                 `json:"tags"`
	CheckType               string                   `json:"check_type"`
	Escalations             []string                 `json:"escalations"`
	Maintenance             UptimeMonitorMaintenance `json:"maintenance"`
	MonitoringServiceType   string                   `json:"monitoring_service_type"`
	IsPaused                bool                     `json:"is_paused"`
	StateIsUp               bool                     `json:"state_is_up"`
	MspScript               string                   `json:"msp_script"`
	MspDNSRecordType        string                   `json:"msp_dns_record_type"`
	MspIPVersion            string                   `json:"msp_use_ip_version"`
	MspSensitivity          int                      `json:"msp_sensitivity"`
	MspInterval             int                      `json:"msp_interval"`
	MspHeaders              string                   `json:"msp_headers"`
	MspNotes                string                   `json:"msp_notes"`
	MspEncryption           string                   `json:"msp_encryption"`
	MspExpectString         string                   `json:"msp_expect_string"`
	MspAddress              string                   `json:"msp_address"`
	MspProtocol             string                   `json:"msp_protocol"`
	MspDNSServer            string                   `json:"msp_dns_server"`
	MspSendString           string                   `json:"msp_send_string"`
	MspUsername             string                   `json:"msp_username"`
	MspExpectStringType     string                   `json:"msp_expect_string_type"`
	MspPassword             string                   `json:"msp_password"`
	MspThreshold            int                      `json:"msp_threshold"`
	MspIncludeGlobalMetrics bool                     `json:"msp_include_global_metrics"`
	MspPort                 int                      `json:"msp_port"`
	StatsURL                string                   `json:"stats_url"`
	AlertsURL               string                   `json:"alerts_url"`
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

type UptimeMonitorMonitorResponse struct {
	Errors  bool                 `json:"errors"`
	Details string               `json:"details"`
	Results UptimeMonitorMonitor `json:"results"`
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
