package main

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
	ID             int                 `json:"id"`
	FriendlyName   string              `json:"friendly_name"`
	URL            string              `json:"url"`
	Type           int                 `json:"type"`
	SubType        string              `json:"sub_type"`
	KeywordType    string              `json:"keyword_type"`
	KeywordValue   string              `json:"keyword_value"`
	HTTPUsername   string              `json:"http_username"`
	HTTPPassword   string              `json:"http_password"`
	Port           string              `json:"port"`
	Interval       int                 `json:"interval"`
	Status         int                 `json:"status"`
	CreateDatetime int                 `json:"create_datetime"`
	MonitorGroup   int                 `json:"monitor_group"`
	IsGroupMain    int                 `json:"is_group_main"`
	Logs           []UptimeMonitorLogs `json:"logs"`
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
