package weatherflow_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tris/weatherflow"
)

func TestUnmarshalWeatherMessage(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      weatherflow.Message
		wantError bool
	}{
		{
			name:  "obs_st message",
			input: `{"status":{"status_code":0,"status_message":"SUCCESS"},"device_id":121037,"type":"obs_st","source":"cache","summary":{"pressure_trend":"steady","strike_count_1h":0,"strike_count_3h":0,"precip_total_1h":0.0,"strike_last_dist":38,"strike_last_epoch":1679435903,"precip_accum_local_yesterday":0.0,"precip_accum_local_yesterday_final":0.0,"precip_analysis_type_yesterday":0,"raining_minutes":[0,0,0,0,0,0,0,0,0,0,0,0],"precip_minutes_local_day":0,"precip_minutes_local_yesterday":0},"obs":[[1681701838,3.71,4.31,5.2,298,3,722.8,null,null,5,0,0,0,0,0,0,2.45,1,0,0,0,0]]}`,
			want: &weatherflow.MessageObsSt{
				Status: weatherflow.ObsStStatus{
					StatusCode:    0,
					StatusMessage: "SUCCESS",
				},
				DeviceID: 121037,
				Type:     "obs_st",
				Source:   "cache",
				Summary: weatherflow.ObsStSummary{
					PressureTrend:                  "steady",
					StrikeCount1h:                  0,
					StrikeCount3h:                  0,
					PrecipTotal1h:                  0,
					StrikeLastDist:                 38,
					StrikeLastEpoch:                1679435903,
					PrecipAccumLocalYesterday:      0,
					PrecipAccumLocalYesterdayFinal: 0,
					PrecipAnalysisTypeYesterday:    0,
					RainingMinutes:                 []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					PrecipMinutesLocalDay:          0,
					PrecipMinutesLocalYesterday:    0,
				},
				Obs: []weatherflow.ObsStData{
					{
						TimeEpoch:                       1681701838,
						WindLull:                        3.71,
						WindAvg:                         4.31,
						WindGust:                        5.2,
						WindDirection:                   298,
						WindSampleInterval:              3,
						StationPressure:                 722.8,
						AirTemperature:                  nil,
						RelativeHumidity:                nil,
						Illuminance:                     5,
						UV:                              0,
						SolarRadiation:                  0,
						RainAccumulated:                 0,
						PrecipitationType:               0,
						LightningStrikeAvgDistance:      0,
						LightningStrikeCount:            0,
						Battery:                         2.45,
						ReportInterval:                  1,
						LocalDailyRainAccumulation:      0,
						RainAccumulatedFinal:            0,
						LocalDailyRainAccumulationFinal: 0,
						PrecipitationAnalysisType:       0,
					},
				},
			},
			wantError: false,
		},
		{
			name:  "rapid_wind message",
			input: `{"device_id":121037,"serial_number":"ST-00026524","type":"rapid_wind","hub_sn":"HB-00039816","ob":[1681701864,4.29,298]}`,
			want: &weatherflow.MessageRapidWind{
				Type:         "rapid_wind",
				DeviceID:     121037,
				SerialNumber: "ST-00026524",
				HubSN:        "HB-00039816",
				Ob: weatherflow.RapidWindData{
					TimeEpoch:     1681701864,
					WindSpeed:     4.29,
					WindDirection: 298,
				},
			},
			wantError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := weatherflow.UnmarshalMessage([]byte(test.input))
			if test.wantError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if diff := cmp.Diff(got, test.want); diff != "" {
					t.Errorf("UnmarshalMessage() mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}
