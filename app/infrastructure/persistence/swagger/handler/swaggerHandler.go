package handler

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

// Получить swagger-спецификацию
func RegisterSwaggerHandlers(ctx context.Context, mux *runtime.ServeMux) error {
	files, _ := ioutil.ReadDir("swagger")

	// Отдача статических файлов
	for _, f := range files {
		mux.Handle(
			"GET",
			runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{f.Name()}, "")),
			func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
				staticPath, _ := filepath.Abs("swagger/" + req.URL.Path)
				http.ServeFile(w, req, staticPath)
			})
	}
	// Отдача спецификаций
	mux.Handle(
		"GET",
		runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"swagger", "config"}, "")),
		func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			configFile, ok := req.URL.Query()["config"]
			// Список спецификаций проекта
			list := map[string]string{
				"fias.swagger.json": "interfaces/grpc/proto/v1/fias/fias.swagger.json",
			}
			if !ok || len(configFile[0]) < 1 || len(list[configFile[0]]) < 1 {
				http.Error(w, "Config not found", 404)
				return
			}
			staticPath, _ := filepath.Abs(list[configFile[0]])
			http.ServeFile(w, req, staticPath)
		})

	// Отдача SwaggerUI
	mux.Handle(
		"GET",
		runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{""}, "")),
		func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			staticPath, _ := filepath.Abs("swagger")
			http.ServeFile(w, req, staticPath+"/index.html")
		})

	return nil
}
