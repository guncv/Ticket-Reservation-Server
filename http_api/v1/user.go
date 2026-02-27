package v1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"ticket-reservation/app"
	customError "ticket-reservation/custom_error"
	"ticket-reservation/http_api/response"
	"ticket-reservation/http_api/routes"
)

var UserRoutes = routes.Routes{
	routes.Route{
		Name:        "Register",
		Path:        "/register",
		Method:      "POST",
		HandlerFunc: Register,
	},
}

func init() {
	RouteDefinitions = append(RouteDefinitions, routes.RouteDefinition{
		Routes: UserRoutes,
		Prefix: "",
	})
}

// FIXME
func Register(ctx *app.Context, w http.ResponseWriter, r *http.Request) error {
	var input app.RegisterParams

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &input); err != nil {
		return &customError.UserError{
			Code:           customError.InvalidJSONString,
			Message:        "Invalid JSON string",
			HTTPStatusCode: http.StatusBadRequest,
		}
	}

	resData, err := ctx.Register(input)
	if err != nil {
		return err
	}

	data, err := json.Marshal(&response.Response{
		Code:    0,
		Message: "",
		Data:    resData,
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
