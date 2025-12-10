package main

import (
	"github.com/WiiLink24/DemaeJustEat/logger"
	"github.com/getsentry/sentry-go"
	"net/http"
	"strings"
)

type Route struct {
	Actions []Action
}

// Action contains information about how a specified action should be handled.
type Action struct {
	ActionName  string
	Callback    func(*Response)
	XMLType     XMLType
	ServiceType string
}

func NewRoute() Route {
	return Route{}
}

// RoutingGroup defines a group of actions for a given service type.
type RoutingGroup struct {
	Route       *Route
	ServiceType string
}

// HandleGroup returns a routing group type for the given service type.
func (r *Route) HandleGroup(serviceType string) RoutingGroup {
	return RoutingGroup{
		Route:       r,
		ServiceType: serviceType,
	}
}

func (r *RoutingGroup) NormalResponse(action string, function func(*Response)) {
	r.Route.Actions = append(r.Route.Actions, Action{
		ActionName:  action,
		Callback:    function,
		XMLType:     Normal,
		ServiceType: r.ServiceType,
	})
}

func (r *RoutingGroup) MultipleRootNodes(action string, function func(*Response)) {
	r.Route.Actions = append(r.Route.Actions, Action{
		ActionName:  action,
		Callback:    function,
		XMLType:     MultipleRootNodes,
		ServiceType: r.ServiceType,
	})
}

func (r *RoutingGroup) ServeImage(function func(*Response)) {
	r.Route.Actions = append(r.Route.Actions, Action{
		Callback:    function,
		XMLType:     None,
		ServiceType: r.ServiceType,
	})
}

func (r *Route) Handle() http.Handler {
	return sentryHandler.HandleFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Debug("HTTP", req.Method, req.URL.String())

		// If this is a POST request it is either an actual request or an error.
		var actionName string
		var serviceType string
		var userAgent string
		if req.Method == "POST" {
			req.ParseForm()
			actionName = req.PostForm.Get("action")
			userAgent = req.PostForm.Get("platform")
			serviceType = "nwapi.php"
		} else {
			actionName = req.URL.Query().Get("action")
			userAgent = req.URL.Query().Get("platform")
			serviceType = strings.Replace(req.URL.Path, "/", "", -1)
		}

		if userAgent != "wii" && !strings.Contains(req.URL.Path, "img") {
			printError(w, "Invalid request.", http.StatusBadRequest)
			return
		}

		// Ensure we can route to this action before processing.
		// Search all registered actions and find a matching action.
		var action Action
		for _, routeAction := range r.Actions {
			if routeAction.ActionName == actionName && strings.Contains(serviceType, routeAction.ServiceType) {
				action = routeAction
			}
		}

		// Action is only properly populated if we found it previously.
		if action.ActionName == "" && action.ServiceType == "" {
			printError(w, "Unknown action was passed.", http.StatusBadRequest)
			return
		}

		resp := NewResponse(req, &w, action.XMLType)
		action.Callback(resp)

		if action.XMLType == None {
			// Already written by function
			return
		}

		if resp.hasError {
			// Response was already written by callback function.
			return
		}

		contents, err := resp.toXML()
		if err != nil {
			printError(w, err.Error(), http.StatusInternalServerError)
			sentry.CaptureException(err)
			return
		}

		w.Write([]byte(contents))
	})
}
