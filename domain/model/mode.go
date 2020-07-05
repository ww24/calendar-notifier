package model

// RunningMode represents running mode.
type RunningMode string

const (
	// ModeNone is uncategorized running mode.
	ModeNone RunningMode = ""
	// ModeResident is resident running mode.
	ModeResident RunningMode = "resident"
	// ModeOnDemand is on-demand running mode.
	ModeOnDemand RunningMode = "ondemand"
)
