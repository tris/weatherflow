package weatherflow

import (
	"encoding/json"
	"fmt"
)

type Message interface {
	GetType() string
	GetDeviceID() (int, bool)
}

type MessageObsSt struct {
	Status   ObsStStatus  `json:"status"`
	DeviceID int          `json:"device_id"`
	Type     string       `json:"type"`
	Source   string       `json:"source"`
	Summary  ObsStSummary `json:"summary"`
	Obs      []ObsStData  `json:"obs"`
}

type MessageRapidWind struct {
	DeviceID     int           `json:"device_id"`
	SerialNumber string        `json:"serial_number"`
	Type         string        `json:"type"`
	HubSN        string        `json:"hub_sn"`
	Ob           RapidWindData `json:"ob"`
}

type MessageConnectionOpened struct {
	Type string `json:"type"`
}

type MessageAck struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ObsStStatus struct {
	StatusCode    int    `json:"status_code"`
	StatusMessage string `json:"status_message"`
}

type ObsStSummary struct {
	PressureTrend                  string  `json:"pressure_trend"`
	StrikeCount1h                  int     `json:"strike_count_1h"`
	StrikeCount3h                  int     `json:"strike_count_3h"`
	PrecipTotal1h                  float64 `json:"precip_total_1h"`
	StrikeLastDist                 int     `json:"strike_last_dist"`
	StrikeLastEpoch                int     `json:"strike_last_epoch"`
	PrecipAccumLocalYesterday      float64 `json:"precip_accum_local_yesterday"`
	PrecipAccumLocalYesterdayFinal float64 `json:"precip_accum_local_yesterday_final"`
	PrecipAnalysisTypeYesterday    int     `json:"precip_analysis_type_yesterday"`
	RainingMinutes                 []int   `json:"raining_minutes"`
	PrecipMinutesLocalDay          int     `json:"precip_minutes_local_day"`
	PrecipMinutesLocalYesterday    int     `json:"precip_minutes_local_yesterday"`
}

type ObsStData struct {
	TimeEpoch                       int      `json:"time_epoch"`
	WindLull                        float64  `json:"wind_lull"`
	WindAvg                         float64  `json:"wind_avg"`
	WindGust                        float64  `json:"wind_gust"`
	WindDirection                   int      `json:"wind_direction"`
	WindSampleInterval              int      `json:"wind_sample_interval"`
	StationPressure                 *float64 `json:"station_pressure"`
	AirTemperature                  *float64 `json:"air_temperature"`
	RelativeHumidity                *float64 `json:"relative_humidity"`
	Illuminance                     int      `json:"illuminance"`
	UV                              int      `json:"uv"`
	SolarRadiation                  int      `json:"solar_radiation"`
	RainAccumulated                 float64  `json:"rain_accumulated"`
	PrecipitationType               int      `json:"precipitation_type"`
	LightningStrikeAvgDistance      int      `json:"lightning_strike_avg_distance"`
	LightningStrikeCount            int      `json:"lightning_strike_count"`
	Battery                         float64  `json:"battery"`
	ReportInterval                  int      `json:"report_interval"`
	LocalDailyRainAccumulation      float64  `json:"local_daily_rain_accumulation"`
	RainAccumulatedFinal            float64  `json:"rain_accumulated_final"`
	LocalDailyRainAccumulationFinal float64  `json:"local_daily_rain_accumulation_final"`
	PrecipitationAnalysisType       int      `json:"precipitation_analysis_type"`
}

type RapidWindData struct {
	TimeEpoch     int     `json:"time_epoch"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection int     `json:"wind_direction"`
}

func (obs *ObsStData) UnmarshalJSON(data []byte) error {
	var obsArray []interface{}
	err := json.Unmarshal(data, &obsArray)
	if err != nil {
		return err
	}

	obs.TimeEpoch = int(obsArray[0].(float64))
	obs.WindLull = obsArray[1].(float64)
	obs.WindAvg = obsArray[2].(float64)
	obs.WindGust = obsArray[3].(float64)
	obs.WindDirection = int(obsArray[4].(float64))
	obs.WindSampleInterval = int(obsArray[5].(float64))

	if obsArray[6] != nil {
		staPressure := obsArray[6].(float64)
		obs.StationPressure = &staPressure
	}

	if obsArray[7] != nil {
		airTemp := obsArray[7].(float64)
		obs.AirTemperature = &airTemp
	}

	if obsArray[8] != nil {
		relHumidity := obsArray[8].(float64)
		obs.RelativeHumidity = &relHumidity
	}

	obs.Illuminance = int(obsArray[9].(float64))
	obs.UV = int(obsArray[10].(float64))
	obs.SolarRadiation = int(obsArray[11].(float64))
	obs.RainAccumulated = obsArray[12].(float64)
	obs.PrecipitationType = int(obsArray[13].(float64))
	obs.LightningStrikeAvgDistance = int(obsArray[14].(float64))
	obs.LightningStrikeCount = int(obsArray[15].(float64))
	obs.Battery = obsArray[16].(float64)
	obs.ReportInterval = int(obsArray[17].(float64))
	obs.LocalDailyRainAccumulation = obsArray[18].(float64)
	obs.RainAccumulatedFinal = obsArray[19].(float64)
	obs.LocalDailyRainAccumulationFinal = obsArray[20].(float64)
	obs.PrecipitationAnalysisType = int(obsArray[21].(float64))

	return nil
}

func (rw *RapidWindData) UnmarshalJSON(data []byte) error {
	var rwArray []interface{}
	err := json.Unmarshal(data, &rwArray)
	if err != nil {
		return err
	}

	rw.TimeEpoch = int(rwArray[0].(float64))
	rw.WindSpeed = rwArray[1].(float64)
	rw.WindDirection = int(rwArray[2].(float64))

	return nil
}

func (w *MessageObsSt) GetType() string {
	return w.Type
}

func (w *MessageRapidWind) GetType() string {
	return w.Type
}

func (w *MessageConnectionOpened) GetType() string {
	return w.Type
}

func (w *MessageAck) GetType() string {
	return w.Type
}

func (w *MessageObsSt) GetDeviceID() (int, bool) {
	return w.DeviceID, true
}

func (w *MessageRapidWind) GetDeviceID() (int, bool) {
	return w.DeviceID, true
}

func (w *MessageConnectionOpened) GetDeviceID() (int, bool) {
	return -1, false
}

func (w *MessageAck) GetDeviceID() (int, bool) {
	return -1, false
}

func UnmarshalMessage(data []byte) (Message, error) {
	var rawMessage map[string]interface{}
	err := json.Unmarshal(data, &rawMessage)
	if err != nil {
		return nil, err
	}

	messageType, ok := rawMessage["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing 'type' field in message")
	}

	switch messageType {
	case "obs_st":
		var message MessageObsSt
		err := json.Unmarshal(data, &message)
		return &message, err
	case "rapid_wind":
		var message MessageRapidWind
		err := json.Unmarshal(data, &message)
		return &message, err
	case "connection_opened":
		var message MessageConnectionOpened
		err := json.Unmarshal(data, &message)
		return &message, err
	case "ack":
		var message MessageAck
		err := json.Unmarshal(data, &message)
		return &message, err
	default:
		return nil, fmt.Errorf("unsupported message type: %s", messageType)
	}
}
