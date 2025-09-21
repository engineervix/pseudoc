package config

import "time"

// Document generation limits
const (
	// MaxPages defines the maximum number of pages allowed for document generation
	MaxPages = 100

	// MaxSheets defines the maximum number of sheets allowed for spreadsheet generation
	MaxSheets = 50

	// MaxFilenameLength defines the maximum allowed length for custom filenames
	MaxFilenameLength = 255
)

// Request ID configuration
const (
	// RequestIDSeparator is the character used to separate parts of the request ID
	RequestIDSeparator = "-"

	// RequestIDLength is the length of the random string component in the request ID
	RequestIDLength = 8
)

// File naming constants
const (
	// DefaultFilenamePrefix is the default prefix used for generated filenames
	DefaultFilenamePrefix = "pseudoc_"

	// DefaultTimestampFormat is the format used for timestamps in default filenames
	DefaultTimestampFormat = "2006-01-02_15-04-05"
)

// Windows reserved filenames (case-insensitive)
var WindowsReservedFilenames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

// Time conversion constants
const (
	// NanosecondsToMilliseconds conversion factor
	NanosecondsToMilliseconds = 1000000
)

// Performance tracking constants
const (
	// MaxPerformanceSamples defines the maximum number of response time samples to keep
	MaxPerformanceSamples = 1000

	// RequestsPerMinuteWindow defines the time window for calculating requests per minute
	RequestsPerMinuteWindow = 60 * time.Second

	// MinSamplesForMetrics defines the minimum samples needed before showing metrics
	MinSamplesForMetrics = 5
)
