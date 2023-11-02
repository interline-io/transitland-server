package meters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/amberflo/metering-go/v2"
	"github.com/interline-io/transitland-lib/log"
	"github.com/rs/zerolog"
	"github.com/xtgo/uuid"
)

type AmberFlo struct {
	apikey      string
	interval    time.Duration
	client      *metering.Metering
	usageClient *metering.UsageClient
	cfgs        map[string]amberFloConfig
}

func NewAmberFlo(apikey string, interval time.Duration, batchSize int) *AmberFlo {
	afLog := &amberfloLogger{logger: log.Logger}
	meteringClient := metering.NewMeteringClient(
		apikey,
		metering.WithBatchSize(batchSize),
		metering.WithIntervalSeconds(interval),
		metering.WithLogger(afLog),
	)
	usageClient := metering.NewUsageClient(
		apikey,
		metering.WithCustomLogger(afLog),
	)
	return &AmberFlo{
		apikey:      apikey,
		interval:    interval,
		client:      meteringClient,
		usageClient: usageClient,
		cfgs:        map[string]amberFloConfig{},
	}
}

type amberFloConfig struct {
	Name          string     `json:"name,omitempty"`
	DefaultUser   string     `json:"default_user,omitempty"`
	ExternalIDKey string     `json:"external_id_key,omitempty"`
	Dimensions    Dimensions `json:"dimensions,omitempty"`
}

func (m *AmberFlo) LoadConfig(path string) error {
	cfgs := map[string]amberFloConfig{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &cfgs); err != nil {
		return err
	}
	m.cfgs = cfgs
	return nil
}

func (m *AmberFlo) NewMeter(user MeterUser) ApiMeter {
	return &amberFloMeter{
		user: user,
		mp:   m,
	}
}

func (m *AmberFlo) Close() error {
	return m.client.Shutdown()
}

func (m *AmberFlo) Flush() error {
	// metering.Flush() // in API docs but not in library
	time.Sleep(m.interval)
	return nil
}

func (m *AmberFlo) getValue(user MeterUser, meterName string) (float64, bool) {
	// TODO: time period and aggregation is hardcoded as 1 day
	cfg, ok := m.getcfg(meterName)
	if !ok {
		return 0, false
	}
	customerId, ok := m.getCustomerID(cfg, user)
	if !ok {
		return 0, false
	}
	if cfg.Name == "" {
		return 0, false
	}

	startTimeInSeconds := (time.Now().In(time.UTC).UnixNano() / int64(time.Second)) - (24 * 60 * 60)
	timeRange := &metering.TimeRange{
		StartTimeInSeconds: startTimeInSeconds,
	}
	filter := make(map[string][]string)
	filter["customerId"] = []string{customerId}
	usageResult, err := m.usageClient.GetUsage(&metering.UsagePayload{
		MeterApiName:         cfg.Name,
		Aggregation:          metering.Sum,
		TimeGroupingInterval: metering.Day,
		GroupBy:              []string{"customerId"},
		TimeRange:            timeRange,
		Filter:               filter,
	})
	if err != nil {
		log.Error().Err(err).Str("user", user.ID()).Msg("could not get value")
		return 0, false
	}
	if usageResult == nil || len(usageResult.ClientMeters) == 0 || len(usageResult.ClientMeters[0].Values) == 0 {
		log.Error().Err(err).Str("user", user.ID()).Msg("could not get value; no client value meter")
		return 0, false
	}
	cm := usageResult.ClientMeters[0].Values
	cmv := cm[len(cm)-1].Value
	return cmv, true
}

func (m *AmberFlo) sendMeter(user MeterUser, meterName string, value float64, extraDimensions Dimensions) error {
	cfg, ok := m.getcfg(meterName)
	if !ok {
		return nil
	}
	customerId, ok := m.getCustomerID(cfg, user)
	if !ok {
		log.Error().Str("user", user.ID()).Msg("could not meter; no amberflo user id")
		return nil
	}
	uniqueId := uuid.NewRandom().String()
	utcMillis := time.Now().In(time.UTC).UnixNano() / int64(time.Millisecond)
	dimensions := Dimensions{}
	for k, v := range cfg.Dimensions {
		dimensions[k] = v
	}
	for k, v := range extraDimensions {
		dimensions[k] = v
	}
	return m.client.Meter(&metering.MeterMessage{
		MeterApiName:      cfg.Name,
		UniqueId:          uniqueId,
		MeterTimeInMillis: utcMillis,
		CustomerId:        customerId,
		MeterValue:        value,
		Dimensions:        dimensions,
	})
}

func (m *AmberFlo) getCustomerID(cfg amberFloConfig, user MeterUser) (string, bool) {
	customerId := cfg.DefaultUser
	if user != nil {
		eidKey := cfg.ExternalIDKey
		if eidKey == "" {
			eidKey = "amberflo"
		}
		if a, ok := user.GetExternalData(eidKey); ok {
			customerId = a
		}
	}
	if customerId == "" {
		log.Error().Str("user", user.ID()).Str("external_id_key", cfg.ExternalIDKey).Msg("could not get value; no amberflo customer id")
	}
	return customerId, customerId != ""
}

func (m *AmberFlo) getcfg(meterName string) (amberFloConfig, bool) {
	cfg, ok := m.cfgs[meterName]
	if !ok {
		cfg = amberFloConfig{
			Name: meterName,
		}
	}
	if cfg.Name == "" {
		log.Error().Str("meter", meterName).Msg("could not meter; no amberflo config for meter")
		return cfg, false
	}
	return cfg, true
}

//////////

type amberFloMeter struct {
	user MeterUser
	dims []string
	mp   *AmberFlo
}

func (m *amberFloMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	var dm2 Dimensions
	if len(extraDimensions) > 0 || len(m.dims) > 0 {
		dm2 = Dimensions{}
	}
	for k, v := range extraDimensions {
		dm2[k] = v
	}
	for i := 0; i < len(m.dims); i += 3 {
		a := m.dims[i]
		k := m.dims[i+1]
		v := m.dims[i+2]
		if a == "" || a == meterName {
			dm2[k] = v
		}
	}
	log.Trace().
		Str("user", m.user.ID()).
		Str("meter", meterName).
		Float64("meter_value", value).
		Msg("meter")
	return m.mp.sendMeter(m.user, meterName, value, dm2)
}

func (m *amberFloMeter) AddDimension(meterName string, key string, value string) {
	m.dims = append(m.dims, meterName, key, value)
}

func (m *amberFloMeter) GetValue(meterName string, d time.Duration, dims Dimensions) (float64, bool) {
	return m.mp.getValue(m.user, meterName)
}

/////////

type amberfloLogger struct {
	logger zerolog.Logger
}

func (l *amberfloLogger) Log(args ...interface{}) {
	l.logger.Trace().Msgf("amberflo: " + fmt.Sprint(args...))
}

func (l *amberfloLogger) Logf(format string, args ...interface{}) {
	l.logger.Trace().Msgf("amberflo: "+format, args...)
}
