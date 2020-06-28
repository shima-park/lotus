package proto

var (
	DefaultSchedule                      = ""
	DefaultCircuitBreakerSamples int64   = 10
	DefaultCircuitBreakerRate    float64 = 0.9
	DefaultBootstrap                     = false
	DefaultConfigOptions                 = ConfigOptions{
		Schedule:              DefaultSchedule,
		CircuitBreakerSamples: DefaultCircuitBreakerSamples,
		CircuitBreakerRate:    DefaultCircuitBreakerRate,
		Bootstrap:             DefaultBootstrap,
	}
)

type ConfigOptions struct {
	Schedule              string   `json:"schedule"`
	CircuitBreakerSamples int64    `json:"circuit_breaker_samples"`
	CircuitBreakerRate    float64  `json:"circuit_breaker_rate"`
	Bootstrap             bool     `json:"bootstrap"`
	Components            []string `json:"components"`
	Processors            []string `json:"processors"`
}

func NewConfigOptions(opts ...ConfigOption) ConfigOptions {
	options := DefaultConfigOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.CircuitBreakerSamples <= 0 {
		options.CircuitBreakerSamples = DefaultCircuitBreakerSamples
	}

	if options.CircuitBreakerRate == 0 {
		options.CircuitBreakerRate = DefaultCircuitBreakerRate
	}

	return options
}

type ConfigOption func(*ConfigOptions)

func WithSchedule(schedule string) ConfigOption {
	return func(c *ConfigOptions) {
		c.Schedule = schedule
	}
}

func WithCircuitBreakerSamples(s int64) ConfigOption {
	return func(c *ConfigOptions) {
		c.CircuitBreakerSamples = s
	}
}

func WithCircuitBreakerRate(f float64) ConfigOption {
	return func(c *ConfigOptions) {
		c.CircuitBreakerRate = f
	}
}

func WithBootstrap(b bool) ConfigOption {
	return func(c *ConfigOptions) {
		c.Bootstrap = b
	}
}

func WithComponents(cs []string) ConfigOption {
	return func(c *ConfigOptions) {
		c.Components = cs
	}
}

func WithProcessor(ps []string) ConfigOption {
	return func(c *ConfigOptions) {
		c.Processors = ps
	}
}
