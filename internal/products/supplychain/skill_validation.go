package supplychain

import (
	"fmt"
	"regexp"
	"time"
)

var fingerprintRegex = regexp.MustCompile(`^[a-f0-9]{64}$`)
var timeRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)

func validateTimeRange(start, end string) error {
	var first, last time.Time
	var err error
	if start != "" {
		first, err = time.Parse(time.RFC3339Nano, start)
		if err != nil {
			return fmt.Errorf("start_time must be RFC3339")
		}
	}
	if end != "" {
		last, err = time.Parse(time.RFC3339Nano, end)
		if err != nil {
			return fmt.Errorf("end_time must be RFC3339")
		}
	}
	if !first.IsZero() && !last.IsZero() && first.After(last) {
		return fmt.Errorf("start_time must not follow end_time")
	}
	return nil
}
