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

type AmberfloMeterProvider struct {
	apikey      string
	interval    time.Duration
	client      *metering.Metering
	usageClient *metering.UsageClient
	cfgs        map[string]amberFloConfig
}

func NewAmberfloMeterProvider(apikey string, interval time.Duration, batchSize int) *AmberfloMeterProvider {
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
	return &AmberfloMeterProvider{
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

func (m *AmberfloMeterProvider) LoadConfig(path string) error {
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

func (m *AmberfloMeterProvider) NewMeter(user MeterUser) ApiMeter {
	return &amberFloMeter{
		user: user,
		mp:   m,
	}
}

func (m *AmberfloMeterProvider) Close() error {
	return m.client.Shutdown()
}

func (m *AmberfloMeterProvider) Flush() error {
	// metering.Flush() // in API docs but not in library
	time.Sleep(m.interval)
	return nil
}

func (m *AmberfloMeterProvider) getValue(user MeterUser, meterName string, startTime time.Time, endTime time.Time, checkDims Dimensions) (float64, bool) {
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
	timeRange := &metering.TimeRange{
		StartTimeInSeconds: startTime.In(time.UTC).Unix(),
		EndTimeInSeconds:   endTime.In(time.UTC).Unix(),
	}
	if timeRange.EndTimeInSeconds > time.Now().In(time.UTC).Unix() {
		timeRange.EndTimeInSeconds = 0
	}

	filter := make(map[string][]string)
	filter["customerId"] = []string{customerId}
	for _, dim := range checkDims {
		filter[dim.Key] = []string{dim.Value}
	}

	timeGroupingInterval := metering.Hour
	switch timeSpan := endTime.Unix() - startTime.Unix(); {
	case timeSpan > 24*60*60:
		timeGroupingInterval = metering.Month
	case timeSpan > 60*60:
		timeGroupingInterval = metering.Day
	default:
		timeGroupingInterval = metering.Hour
	}

	usageResult, err := m.usageClient.GetUsage(&metering.UsagePayload{
		MeterApiName:         cfg.Name,
		Aggregation:          metering.Sum,
		TimeGroupingInterval: timeGroupingInterval,
		GroupBy:              []string{"customerId"},
		TimeRange:            timeRange,
		Filter:               filter,
	})
	if err != nil {
		log.Error().Err(err).Str("user", user.ID()).Msg("could not get value")
		return 0, false
	}
	// jj, _ := json.Marshal(&usageResult)
	// fmt.Println("usageResult:", string(jj))

	if usageResult == nil || len(usageResult.ClientMeters) == 0 || len(usageResult.ClientMeters[0].Values) == 0 {
		log.Error().Err(err).Str("user", user.ID()).Msg("could not get value; no client value meter")
		return 0, false
	}

	total := usageResult.ClientMeters[0].GroupValue
	return total, true
}

func (m *AmberfloMeterProvider) sendMeter(user MeterUser, meterName string, value float64, extraDimensions Dimensions) error {
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
	amberFloDims := map[string]string{}
	for _, v := range cfg.Dimensions {
		amberFloDims[v.Key] = v.Value
	}
	for _, v := range extraDimensions {
		amberFloDims[v.Key] = v.Value
	}
	return m.client.Meter(&metering.MeterMessage{
		MeterApiName:      cfg.Name,
		UniqueId:          uniqueId,
		MeterTimeInMillis: utcMillis,
		CustomerId:        customerId,
		MeterValue:        value,
		Dimensions:        amberFloDims,
	})
}

func (m *AmberfloMeterProvider) getCustomerID(cfg amberFloConfig, user MeterUser) (string, bool) {
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

func (m *AmberfloMeterProvider) getcfg(meterName string) (amberFloConfig, bool) {
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
	user    MeterUser
	addDims []eventAddDim
	mp      *AmberfloMeterProvider
}

func (m *amberFloMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	var eventDims []Dimension
	// Copy in matching dimensions set through AddDimension
	for _, addDim := range m.addDims {
		if addDim.MeterName == meterName {
			eventDims = append(eventDims, Dimension{Key: addDim.Key, Value: addDim.Value})
		}
	}
	eventDims = append(eventDims, extraDimensions...)
	log.Trace().
		Str("user", m.user.ID()).
		Str("meter", meterName).
		Float64("meter_value", value).
		Msg("meter")
	return m.mp.sendMeter(m.user, meterName, value, eventDims)
}

func (m *amberFloMeter) AddDimension(meterName string, key string, value string) {
	m.addDims = append(m.addDims, eventAddDim{MeterName: meterName, Key: key, Value: value})
}

func (m *amberFloMeter) GetValue(meterName string, startTime time.Time, endTime time.Time, dims Dimensions) (float64, bool) {
	return m.mp.getValue(m.user, meterName, startTime, endTime, dims)
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
