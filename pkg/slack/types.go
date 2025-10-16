package slack

import "strconv"

type Timestamp string

func (t Timestamp) Float64() float64 {
	v, _ := strconv.ParseFloat(string(t), 64)
	return v
}

type ChannelType string

const (
	Public  ChannelType = "public_channel"
	Private ChannelType = "private_channel"
	DM      ChannelType = "im"
	GroupDM ChannelType = "mpim"
)

type User struct {
	ID     string `json:"id,omitempty"`
	TeamID string `json:"team_id,omitempty"`
	Name   string `json:"username,omitempty"`
}

type Team struct {
	ID     string `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type Channel struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
