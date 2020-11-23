//+build ignore

package main

import (
    "bytes"
    "fmt"
    "go/format"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
    "text/template"
)

const (
    blobFileName string = "blob.go"
    embedFolder  string = "../../../swagger"
)

// Переменные для генерации шаблона
var conv = map[string]interface{}{"conv": fmtByteSlice}
var tmpl = template.Must(template.New("").Funcs(conv).Parse(`package box

// Код сгенерирован автоматически; НЕ РЕДАКТИРОВАТЬ.

func init() {
    {{- range $name, $file := . }}
        box.Add("{{ $name }}", []byte{ {{ conv $file }} })
    {{- end }}
}`),
)

func fmtByteSlice(s []byte) string {
    builder := strings.Builder{}

    for _, v := range s {
        builder.WriteString(fmt.Sprintf("%d,", int(v)))
    }

    return builder.String()
}

func main() {
    // Проверка директории на наличие файлов
    if _, err := os.Stat(embedFolder); os.IsNotExist(err) {
        log.Fatal("Configs directory does not exists!")
    }

    // Создаем массив для файлов
    configs := make(map[string][]byte)

    // Проходим по файлам в директории
    err := filepath.Walk(embedFolder, func(path string, info os.FileInfo, err error) error {
        relativePath := filepath.ToSlash(strings.TrimPrefix(path, embedFolder))

        if info.IsDir() {
            // Пропускаем директории
            log.Println(path, "is a directory, skipping...")
            return nil
        } else {
            // Проверяем файл
            log.Println(path, "is a file, packing in...")

            b, err := ioutil.ReadFile(path)
            if err != nil {
                // Если файл не доступен для чтения, возвращаем ошибку
                log.Printf("Error reading %s: %s", path, err)
                return err
            }

            // Добавляем файл в массив
            configs[relativePath] = b
        }

        return nil
    })
    if err != nil {
        log.Fatal("Error walking through embed directory:", err)
    }

    // Создаем временный файл
    f, err := os.Create(blobFileName)
    if err != nil {
        log.Fatal("Error creating blob file:", err)
    }
    defer f.Close()

    // Создаем буфер
    builder := &bytes.Buffer{}

    // Получаем шаблон для генерации файла
    if err = tmpl.Execute(builder, configs); err != nil {
        log.Fatal("Error executing template", err)
    }

    // Форматируем код
    data, err := format.Source(builder.Bytes())
    if err != nil {
        log.Fatal("Error formatting generated code", err)
    }

    // Сохраняем файл
    if err = ioutil.WriteFile(blobFileName, data, os.ModePerm); err != nil {
        log.Fatal("Error writing blob file", err)
    }
}
