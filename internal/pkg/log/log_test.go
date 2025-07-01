package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name          string
		level         string
		expectedLevel logrus.Level
	}{
		{
			name:          "Success Case - Info Level",
			level:         "info",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "Success Case - Debug Level",
			level:         "debug",
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "Success Case - Warn Level",
			level:         "warn",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "Failure Case - Invalid Level",
			level:         "invalid-level",
			expectedLevel: logrus.InfoLevel, // Should default to Info
		},
		{
			name:          "Failure Case - Empty Level",
			level:         "",
			expectedLevel: logrus.InfoLevel, // Should default to Info
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Execution
			logger := New(tc.level)

			// 2. Assertion
			assert.NotNil(t, logger)
			assert.Equal(t, tc.expectedLevel, logger.GetLevel())

			// Verify formatter by checking for the absence of a timestamp
			var buf bytes.Buffer
			logger.SetOutput(&buf)
			logger.Info("test message")

			assert.NotContains(t, strings.ToLower(buf.String()), "time=", "Timestamp should be disabled")
		})
	}
}
