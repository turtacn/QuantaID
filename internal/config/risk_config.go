package config

// RiskThresholds defines the score boundaries for each risk level.
type RiskThresholds struct {
	Low    float64 `yaml:"low"`
	Medium float64 `yaml:"medium"`
	High   float64 `yaml:"high"`
}

// RiskWeights assigns a weight to each risk factor for scoring.
type RiskWeights struct {
	IPReputation    float64 `yaml:"ip_reputation"`
	GeoReputation   float64 `yaml:"geo_reputation"`
	DeviceChange    float64 `yaml:"device_change"`
	GeoVelocity     float64 `yaml:"geo_velocity"`
	TimeAnomaly     float64 `yaml:"time_anomaly"`
}

// RiskConfig holds the complete configuration for the adaptive risk engine.
type RiskConfig struct {
	Thresholds RiskThresholds `yaml:"thresholds"`
	Weights    RiskWeights    `yaml:"weights"`
}
