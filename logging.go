package weatherflow

// Logf is a function type for logging messages in the WeatherFlowClient.
// This is compatible with e.g. log.Printf.
type Logf func(format string, args ...interface{})
