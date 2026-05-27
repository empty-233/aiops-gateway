package model

import (
	amwebhook "github.com/prometheus/alertmanager/notify/webhook"
	amtemplate "github.com/prometheus/alertmanager/template"
)

type AlertPayload = amwebhook.Message

type Alert = amtemplate.Alert

type AlertResult struct {
	AlertCount int     `json:"alertCount"`
	Status     string  `json:"status"`
	Analysis   string  `json:"analysis"`
	Alerts     []Alert `json:"alerts"`
}
