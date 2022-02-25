package directions

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/model"
)

type Handler interface {
	Request(model.DirectionRequest) (*model.Directions, error)
}

type handlerFunc func() Handler

var handlersLock sync.Mutex
var handlers = map[string]handlerFunc{}

func RegisterRouter(name string, f handlerFunc) error {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	if _, ok := handlers[name]; ok {
		return fmt.Errorf("handler '%s' already registered", name)
	}
	log.Debugf("registering routing handler: %s", name)
	handlers[name] = f
	return nil
}

func getHandler(name string) (handlerFunc, bool) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	a, ok := handlers[name]
	return a, ok
}

func HandleRequest(pref string, req model.DirectionRequest) (*model.Directions, error) {
	var handler Handler
	handler = &lineRouter{}
	// Always use line
	if req.Mode == model.StepModeLine {
		pref = "line"
	}
	// Realtime auto requires aws
	if req.Mode == model.StepModeAuto && req.DepartAt == nil {
		pref = "aws"
	}
	// Default
	if pref == "" {
		pref = os.Getenv("TL_DEFAULT_ROUTER")
		// will default to line if invalid or empty
	}
	if hf, ok := getHandler(pref); ok {
		handler = hf()
	}
	return handler.Request(req)
}

func validateDirectionRequest(req model.DirectionRequest) error {
	if req.From == nil || req.To == nil {
		return errors.New("from and to waypoints required")
	}
	return nil
}

func wpiWaypoint(w *model.WaypointInput) *model.Waypoint {
	if w == nil {
		return nil
	}
	return &model.Waypoint{
		Lon:  w.Lon,
		Lat:  w.Lat,
		Name: w.Name,
	}
}
