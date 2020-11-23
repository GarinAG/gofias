//go:generate go run generator.go

package box

type embedBox struct {
    storage map[string][]byte
}

// Создает новый контейнер для встраивания файлов
func newEmbedBox() *embedBox {
    return &embedBox{storage: make(map[string][]byte)}
}

// Добавляет файл в контейнер
func (e *embedBox) Add(file string, content []byte) {
    e.storage[file] = content
}

// Возвращает файл из контейнера
func (e *embedBox) Get(file string) []byte {
    if f, ok := e.storage[file]; ok {
        return f
    }
    return nil
}

// Инициализация контейнера
var box = newEmbedBox()

// Добавляет файл в контейнер
func Add(file string, content []byte) {
    box.Add(file, content)
}

// Возвращает файл из контейнера
func Get(file string) []byte {
    return box.Get(file)
}

// Возвращает список всех файлов в контейнере
func GetKeys() []string {
    var keys []string
    for key := range box.storage{
        keys = append(keys, key)
    }

    return keys
}
