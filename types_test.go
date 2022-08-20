package powerwall

import (
	"testing"
)

func TestDecodedAlertUnmarshalJSON(t *testing.T) {
	cases := []struct {
		in         string
		shouldFail bool
		want       map[string]string
	}{
		{
			"\"[{\\\"name\\\":\\\"PINV_alertID\\\",\\\"va",
			true,
			nil,
		},
		{
			"[{\"name\":\"PINV_alertID\",\"value\":\"Non-nested JSON\"}]",
			true,
			nil,
		},
		{"\"[{\\\"name\\\":\\\"PINV_alertID\\\",\\\"value\\\":\\\"PINV_a021_vfCheckSplitPhaseOverVoltage\\\"},{\\\"name\\\":\\\"PINV_alertType\\\",\\\"value\\\":\\\"Warning\\\"},{\\\"name\\\":\\\"PINV_a021_ov_amplitude1\\\",\\\"value\\\":184,\\\"units\\\":\\\"Vrms\\\"},{\\\"name\\\":\\\"PINV_a021_ov_amplitude2\\\",\\\"value\\\":164,\\\"units\\\":\\\"Vrms\\\"}]\"",
			false,
			map[string]string{
				"PINV_alertID":            "PINV_a021_vfCheckSplitPhaseOverVoltage",
				"PINV_alertType":          "Warning",
				"PINV_a021_ov_amplitude1": "184 Vrms",
				"PINV_a021_ov_amplitude2": "164 Vrms",
			},
		},
		{
			"\"[{\\\"name\\\":\\\"PINV_alertID\\\",\\\"value\\\":\\\"PINV_a006_vfCheckUnderFrequency\\\"},{\\\"name\\\":\\\"PINV_alertType\\\",\\\"value\\\":\\\"Warning\\\"},{\\\"name\\\":\\\"PINV_a006_frequency\\\",\\\"value\\\":50.426,\\\"units\\\":\\\"Hz\\\"}]\"",
			false,
			map[string]string{
				"PINV_alertID":        "PINV_a006_vfCheckUnderFrequency",
				"PINV_alertType":      "Warning",
				"PINV_a006_frequency": "50.426 Hz",
			},
		},
		{
			"\"[{\\\"name\\\":\\\"PINV_alertID\\\",\\\"value\\\":\\\"PINV_a008_vfCheckRocof\\\"},{\\\"name\\\":\\\"PINV_alertType\\\",\\\"value\\\":\\\"Warning\\\"}]\"",
			false,
			map[string]string{
				"PINV_alertID":   "PINV_a008_vfCheckRocof",
				"PINV_alertType": "Warning",
			},
		},
		{
			"\"[{\\\"name\\\":\\\"PINV_alertID\\\",\\\"value\\\":\\\"PINV_a020_vfCheckSplitPhaseUnderVoltage\\\"},{\\\"name\\\":\\\"PINV_alertType\\\",\\\"value\\\":\\\"Warning\\\"},{\\\"name\\\":\\\"PINV_a020_uv_amplitude1\\\",\\\"value\\\":31,\\\"units\\\":\\\"Vrms\\\"},{\\\"name\\\":\\\"PINV_a020_uv_amplitude2\\\",\\\"value\\\":30,\\\"units\\\":\\\"Vrms\\\"}]\"",
			false,
			map[string]string{
				"PINV_alertID":            "PINV_a020_vfCheckSplitPhaseUnderVoltage",
				"PINV_alertType":          "Warning",
				"PINV_a020_uv_amplitude1": "31 Vrms",
				"PINV_a020_uv_amplitude2": "30 Vrms",
			},
		},
	}
	for i, c := range cases {
		a := DecodedAlert{}
		if err := a.UnmarshalJSON([]byte(c.in)); err != nil {
			if c.shouldFail {
				continue
			}
			t.Errorf("case %d failed unexpectedly: %v", i, err)
		} else if c.shouldFail {
			t.Errorf("case %d should have failed and did not", i)
		}
		for k, v := range c.want {
			if a[k] != v {
				t.Errorf("case %d: want key %q value %q, got %q (in %v)", i, k, v, a[k], a)
			}
		}
	}
}
