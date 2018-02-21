package main

type UptimeMonitorGetMonitorsResponse struct {
	Stat       string `json:"stat"`
	Pagination struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
		Total  int `json:"total"`
	} `json:"pagination"`
	Monitors []struct {
		ID             int    `json:"id"`
		FriendlyName   string `json:"friendly_name"`
		URL            string `json:"url"`
		Type           int    `json:"type"`
		SubType        string `json:"sub_type"`
		KeywordType    string `json:"keyword_type"`
		KeywordValue   string `json:"keyword_value"`
		HTTPUsername   string `json:"http_username"`
		HTTPPassword   string `json:"http_password"`
		Port           string `json:"port"`
		Interval       int    `json:"interval"`
		Status         int    `json:"status"`
		CreateDatetime int    `json:"create_datetime"`
		MonitorGroup   int    `json:"monitor_group"`
		IsGroupMain    int    `json:"is_group_main"`
		Logs           []struct {
			Type     int `json:"type"`
			Datetime int `json:"datetime"`
			Duration int `json:"duration"`
		} `json:"logs"`
	} `json:"monitors"`
}
