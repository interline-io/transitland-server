package meters

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/amberflo/metering-go/v2"
	"github.com/interline-io/transitland-lib/log"
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
	// debug := false
	// if log.Logger.GetLevel() == zerolog.TraceLevel {
	// 	debug = true
	// }
	meteringClient := metering.NewMeteringClient(
		apikey,
		metering.WithBatchSize(batchSize),
		metering.WithIntervalSeconds(interval),
		// metering.WithDebug(debug),
	)
	usageClient := metering.NewUsageClient(
		apikey,
		// metering.WithCustomLogger(log.Logger),
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
	Name              string            `json:"name,omitempty"`
	DefaultUser       string            `json:"default_user,omitempty"`
	UserExternalIDKey string            `json:"user_external_id_key,omitempty"`
	Dimensions        map[string]string `json:"dimensions,omitempty"`
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

func (m *AmberFlo) NewMeter(handlerName string) ApiMeter {
	cfg := amberFloConfig{
		Name:              handlerName,
		DefaultUser:       "",
		UserExternalIDKey: "",
	}
	if a, ok := m.cfgs[handlerName]; ok {
		cfg = amberFloConfig{
			Name:              a.Name,
			DefaultUser:       a.DefaultUser,
			UserExternalIDKey: a.UserExternalIDKey,
		}
		for k, v := range a.Dimensions {
			cfg.Dimensions[k] = v
		}
	}
	return &amberFloMeter{
		cfg:         cfg,
		client:      m.client,
		usageClient: m.usageClient,
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

type amberFloMeter struct {
	cfg         amberFloConfig
	client      *metering.Metering
	usageClient *metering.UsageClient
}

func (m *amberFloMeter) getCustomerID(user MeterUser) (string, bool) {
	customerId := m.cfg.DefaultUser
	if user != nil {
		eidKey := m.cfg.UserExternalIDKey
		if eidKey == "" {
			eidKey = "amberflo"
		}
		if a, ok := user.GetExternalID(eidKey); ok {
			customerId = a
		}
	}
	return customerId, customerId != ""
}

func (m *amberFloMeter) Meter(user MeterUser, value float64, extraDimensions map[string]string) error {
	uniqueId := uuid.NewRandom().String()
	utcMillis := time.Now().UnixNano() / int64(time.Millisecond)
	customerId, ok := m.getCustomerID(user)
	if !ok {
		log.Error().Str("user", user.Name()).Msg("could not meter; no amberflo user id")
		return nil
	}
	dimensions := map[string]string{}
	for k, v := range m.cfg.Dimensions {
		dimensions[k] = v
	}
	for k, v := range extraDimensions {
		dimensions[k] = v
	}
	return m.client.Meter(&metering.MeterMessage{
		MeterApiName:      m.cfg.Name,
		UniqueId:          uniqueId,
		MeterTimeInMillis: utcMillis,
		CustomerId:        customerId,
		MeterValue:        value,
		Dimensions:        dimensions,
	})
}

func (m *amberFloMeter) GetValue(user MeterUser) (float64, bool) {
	// TODO: batch and cache
	// TODO: time period and aggregation is hardcoded as 1 day
	startTimeInSeconds := (time.Now().UnixNano() / int64(time.Second)) - (24 * 60 * 60)
	timeRange := &metering.TimeRange{
		StartTimeInSeconds: startTimeInSeconds,
	}
	customerId, ok := m.getCustomerID(user)
	if !ok {
		log.Error().Str("user", user.Name()).Msg("could not get value; no amberflo customer id")
		return 0, false
	}
	filter := make(map[string][]string)
	filter["customerId"] = []string{customerId}
	usageResult, err := m.usageClient.GetUsage(&metering.UsagePayload{
		MeterApiName:         m.cfg.Name,
		Aggregation:          metering.Sum,
		TimeGroupingInterval: metering.Day,
		GroupBy:              []string{"customerId"},
		TimeRange:            timeRange,
		Filter:               filter,
	})
	if err != nil {
		log.Error().Err(err).Str("user", user.Name()).Msg("could not get value")
		return 0, false
	}
	if usageResult == nil || len(usageResult.ClientMeters) == 0 || len(usageResult.ClientMeters[0].Values) == 0 {
		log.Error().Err(err).Str("user", user.Name()).Msg("could not get value; no client value meter")
		return 0, false
	}
	cm := usageResult.ClientMeters[0].Values
	cmv := cm[len(cm)-1].Value
	return cmv, true
}
