package web

import (
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const baseURI = "/v1/api"

type Router struct {
	mem     storage.Service
	service *Service
}

func (r *Router) route() {
	http.Handle(baseURI+"/", r.api())
	http.Handle("/", r.static())
	http.Handle("/download/", r.download())

}

func (r *Router) api() http.Handler {
	router := toolbox.NewServiceRouter(
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        fmt.Sprintf("%v/meta", baseURI),
			Handler: func() interface{} {
				resp, err := r.service.Get(&GetRequest{})
				if err != nil {
					resp = &GetResponse{
						Status: "error",
						Error:  err.Error(),
					}
				}
				return resp
			},
			Parameters: []string{},
		},
	)
	return http.HandlerFunc(func(writer http.ResponseWriter, reader *http.Request) {
		if err := router.Route(writer, reader); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

	})
}

func (r *Router) setTextValue(form url.Values, field string, target *string, defaultValue string) {
	if value, ok := form[field]; ok {
		textValue := value[0]
		textValue = strings.Replace(textValue, " ", "", len(textValue))
		*target = textValue
	} else {
		*target = defaultValue
	}
}

func (r *Router) setBoolValue(form url.Values, field string, target *bool) {
	if value, ok := form[field]; ok {
		boolValue := value[0] != ""
		*target = boolValue
	}
}

func (r *Router) download() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		var form = request.Form
		var runRequest = &RunRequest{
			Datastore: &Datastore{},
			Build:     &Build{},
			Testing:   &Testing{},
		}
		r.setTextValue(form, "appTemplate", &runRequest.Build.TemplateApp, "default")
		r.setTextValue(form, "appName", &runRequest.Build.App, "myapp")
		r.setTextValue(form, "origin", &runRequest.Origin, "")
		r.setTextValue(form, "sdk", &runRequest.Build.Sdk, "")
		r.setBoolValue(form, "docker", &runRequest.Build.Docker)

		r.setTextValue(form, "dbEngine", &runRequest.Datastore.Driver, "mysql")
		r.setTextValue(form, "dbName", &runRequest.Datastore.Name, "mydb")
		r.setBoolValue(form, "dbConfig", &runRequest.Datastore.Config)

		r.setBoolValue(form, "http", &runRequest.Testing.HTTP)
		r.setBoolValue(form, "rest", &runRequest.Testing.REST)
		r.setBoolValue(form, "selenium", &runRequest.Testing.Selenium)
		r.setBoolValue(form, "caseData", &runRequest.Testing.UseCaseData)
		r.setBoolValue(form, "mapping", &runRequest.Datastore.MultiTableMapping)

		resp, err := r.service.Run(runRequest)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/zip")
		writer.Write(resp.Data)
	})
}

func (r *Router) static() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		var assetPath = string(request.URL.Path[1:])
		if assetPath == "" {
			assetPath = "index.html"
		}
		var URL = toolbox.URLPathJoin(r.service.baseAssetURL, assetPath)
		if has, _ := r.mem.Exists(URL); !has {
			http.NotFound(writer, request)
			return
		}

		var ext = path.Ext(assetPath)
		if ext != "" {
			ext = string(ext[1:])
		}
		contentType, has := toolbox.FileExtensionMimeType[ext]
		if !has {
			contentType = fmt.Sprintf("text/%v", ext)
		}
		if strings.Contains(contentType, "text") {
			contentType += "; charset=utf-8"
		}
		writer.Header().Set("Content-Type", contentType)
		object, err := r.mem.StorageObject(URL)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		reader, err := r.mem.Download(object)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(writer, reader)
	})
}

func NewRouter(service *Service) *Router {
	srv := storage.NewMemoryService()
	var result = &Router{
		service: service,
		mem:     srv,
	}
	result.route()
	return result
}
