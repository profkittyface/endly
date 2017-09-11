package endly

import (
	"github.com/viant/toolbox"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
)

type Request struct {
	Data           map[string]interface{}
	ServiceRequest interface{}
}

type Response struct {
	Status   string
	Error    string
	Response interface{}
	Info     *SessionInfo
}

type Server struct {
	port    string
	manager Manager
}

func (s *Server) requestService(serviceName, method string, httpRequest *http.Request, httpResponse http.ResponseWriter) (*Response, error) {
	var service Service
	var serviceRequest interface{}
	var err error
	service, err = s.manager.Service(toolbox.AsString(serviceName))
	if err != nil {
		return nil, err
	}
	serviceRequest, err = service.NewRequest(toolbox.AsString(method))
	if err != nil {
		return nil, err
	}
	request := &Request{
		ServiceRequest: serviceRequest,
	}
	err = json.NewDecoder(httpRequest.Body).Decode(request)
	if err != nil {
		return nil, err
	}
	context := s.manager.NewContext(toolbox.NewContext())
	defer context.Close()
	state := context.State()
	state.Apply(request.Data)
	serviceResponse := service.Run(context, request.ServiceRequest)

	var sessionInfo = context.SessionInfo()
	var response = &Response{
		Status:   serviceResponse.Status,
		Error:    serviceResponse.Error,
		Response: serviceResponse.Response,
		Info:     sessionInfo,
	}
	return response, nil
}

func (s *Server) routeHandler(serviceRouting *toolbox.ServiceRouting, httpRequest *http.Request, httpResponse http.ResponseWriter, uriParameters map[string]interface{}) (err error) {
	defer func() {
		if err != nil {
			serviceResponse := &Response{
				Error: fmt.Sprintf("%v", err),
			}
			err = toolbox.WriteServiceRoutingResponse(httpResponse, httpRequest, serviceRouting, serviceResponse)
		}

	}()

	serviceName, ok := uriParameters["service"]
	if ! ok {
		return fmt.Errorf("Service name was missing %v", uriParameters)
	}
	method, ok := uriParameters["method"]
	if ! ok {
		return fmt.Errorf("method was missing %v", uriParameters)
	}

	var response *Response
	response, err = s.requestService(toolbox.AsString(serviceName), toolbox.AsString(method), httpRequest, httpResponse)
	if err != nil {
		return err
	}

	err = toolbox.WriteServiceRoutingResponse(httpResponse, httpRequest, serviceRouting, response)

	if err != nil {
		return err
	}
	return nil

}

func (s *Server) Start() error {

	router := toolbox.NewServiceRouter(
		toolbox.ServiceRouting{
			HTTPMethod:     "POST",
			URI:            "/v1/endly/api/{service}/{method}/",
			Handler:        s.requestService,
			HandlerInvoker: s.routeHandler,
			Parameters:     []string{"service", "method", "@httpRequest", "@httpResponseWriter"},
		})

	http.HandleFunc("/v1/", func(response http.ResponseWriter, reader *http.Request) {
		err := router.Route(response, reader)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
		}
	})
	fmt.Printf("Started test server on port %v\n", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, nil))
	return nil
}

func NewServer(port string) *Server {
	return &Server{
		port:    port,
		manager: GetManager(),
	}
}