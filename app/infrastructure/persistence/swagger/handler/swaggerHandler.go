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
		mux.Handle(
			"GET",
			runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{file}, "")),
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
