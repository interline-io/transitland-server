package meters

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/amberflo/metering-go/v2"
	"github.com/interline-io/transitland-server/auth"
	"github.com/xtgo/uuid"
)

type AmberFlo struct {
	client   *metering.Metering
	meterMap map[string]amberFloConfig
}

func NewAmberFlo(apikey string) *AmberFlo {
	intervalSeconds := 30 * time.Second
	batchSize := 5
	// debug := false
	// if log.Logger.GetLevel() == zerolog.TraceLevel {
	// 	debug = true
	// }
	meteringClient := metering.NewMeteringClient(
		apikey,
		metering.WithBatchSize(batchSize),
		metering.WithIntervalSeconds(intervalSeconds),
		// metering.WithDebug(debug),
	)
	return &AmberFlo{
		client:   meteringClient,
		meterMap: map[string]amberFloConfig{},
	}
}

type amberFloConfig struct {
	Name           string            `json:"name,omitempty"`
	DefaultUser    string            `json:"default_user,omitempty"`
	UserExternalID string            `json:"user_external_id,omitempty"`
	Dimensions     map[string]string `json:"dimensions,omitempty"`
}

func (m *AmberFlo) LoadConfig(path string) error {
	meterMap := map[string]amberFloConfig{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &meterMap); err != nil {
		return err
	}
	m.meterMap = meterMap
	return nil
}

func (m *AmberFlo) NewMeter(handlerName string) ApiMeter {
	cfg := amberFloConfig{
		Name: handlerName,
	}
	if a, ok := m.meterMap[handlerName]; ok {
		cfg = amberFloConfig{
			Name:           a.Name,
			DefaultUser:    a.DefaultUser,
			UserExternalID: a.UserExternalID,
		}
		for k, v := range a.Dimensions {
			cfg.Dimensions[k] = v
		}
	}
	return &amberFloMeter{
		cfg:    cfg,
		client: m.client,
	}
}

func (m *AmberFlo) Close() error {
	return m.client.Shutdown()
}

type amberFloMeter struct {
	cfg    amberFloConfig
	client *metering.Metering
}

func (m *amberFloMeter) Meter(user auth.User, value float64, extraDimensions map[string]string) error {
	uniqueId := uuid.NewRandom().String()
	utcMillis := time.Now().UnixNano() / int64(time.Millisecond)
	userId := m.cfg.DefaultUser
	if user != nil {
		if a, ok := user.GetExternalID(m.cfg.UserExternalID); ok {
			userId = a
		}
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
		CustomerId:        userId,
		MeterValue:        value,
		Dimensions:        dimensions,
	})
}
