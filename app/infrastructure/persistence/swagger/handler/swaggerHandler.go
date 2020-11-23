package handler

import (
	"bytes"
	"context"
	"github.com/GarinAG/gofias/infrastructure/persistence/box"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"net/http"
	"strings"
	"time"
)

// Получить swagger-спецификацию
func RegisterSwaggerHandlers(ctx context.Context, mux *runtime.ServeMux) error {
	// Отдача статических файлов
	files := box.GetKeys()
	for _, file := range files {
		file := strings.Trim(file, "/")
		propList := strings.Split(file, "/")
		ops := []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3, 1, 0, 4, 1, 5, 4, 1, 0, 4, 1, 5, 5, 2, 6, 1, 0, 4, 1, 5, 7, 1, 0, 4, 1, 5, 8, 1, 0, 4, 1, 5, 9, 1, 0, 4, 1, 5, 10, 1, 0, 4, 1, 5, 11, 2, 12, 1, 0, 4, 2, 5, 13, 1, 0, 4, 1, 5, 14, 1, 0, 4, 1, 5, 15, 1, 0, 4, 1, 5, 16, 1, 0, 4, 1, 5, 17, 1, 0, 4, 1, 5, 18}[0 : len(propList)*2]

		mux.Handle(
			"GET",
			runtime.MustPattern(runtime.NewPattern(1, ops, propList, "")),
			func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
				printResponse(w, req, req.URL.Path)
			})
	}

	// Отдача SwaggerUI
	mux.Handle(
		"GET",
		runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{""}, "")),
		func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			printResponse(w, req, "/index.html")
		})

	return nil
}

// Возвращает файлы
func printResponse(w http.ResponseWriter, req *http.Request, filePath string) {
	file := box.Get(filePath)
	// Проверка файла на наличие
	if file == nil {
		http.NotFound(w, req)
	}
	fileName := strings.Trim(filePath, "/")
	// Отдача контента
	http.ServeContent(w, req, fileName, time.Now(),
		bytes.NewReader(file))
}
