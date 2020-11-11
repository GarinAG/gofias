package util

import (
	"regexp"
	"strings"
)

// Объект названия сокращенных местоположений
type addrShortName struct {
	short  string
	full   string
	prefix bool
}

// Список для замены названий
var replaceList = map[string]string{
	"городское поселение":     "город",
	"муниципальный округ":     "район",
	"поселок городского типа": "поселок",
	"рабочий поселок":         "поселок",
	"ё":                       "е",
}

// Текст для замены названий
var replaceSting = ""

// Список сокращенных названий местоположений
var shortNameList = map[string]addrShortName{
	"г/п":        addrShortName{short: "г.", full: "Городское поселение", prefix: true},
	"ул":         addrShortName{short: "ул.", full: "Улица", prefix: true},
	"пер":        addrShortName{short: "пер.", full: "Переулок", prefix: true},
	"д":          addrShortName{short: "д.", full: "Деревня", prefix: true},
	"тер":        addrShortName{short: "тер.", full: "Территория", prefix: true},
	"с":          addrShortName{short: "село", full: "Село", prefix: true},
	"тер. СНТ":   addrShortName{short: "тер. СНТ", full: "Территория садоводческих некоммерческих товариществ", prefix: true},
	"снт":        addrShortName{short: "СНТ", full: "Садоводческое некоммерческое товарищество", prefix: true},
	"п":          addrShortName{short: "п.", full: "Поселок", prefix: true},
	"проезд":     addrShortName{short: "пр-д", full: "Проезд", prefix: true},
	"кв-л":       addrShortName{short: "кв-л", full: "Квартал", prefix: true},
	"гск":        addrShortName{short: "гск", full: "Гаражно-строительный кооператив", prefix: true},
	"пр-д":       addrShortName{short: "пр-д", full: "Проезд", prefix: true},
	"линия":      addrShortName{short: "лн.", full: "Линия", prefix: true},
	"ряд":        addrShortName{short: "ряд", full: "Ряд", prefix: false},
	"х":          addrShortName{short: "хутор", full: "Хутор", prefix: true},
	"тер. ГСК":   addrShortName{short: "тер. ГСК", full: "Территория гаражно-строительного кооператива", prefix: true},
	"мкр":        addrShortName{short: "мкр.", full: "Микрорайон", prefix: true},
	"туп":        addrShortName{short: "туп.", full: "Тупик", prefix: true},
	"пл":         addrShortName{short: "пл.", full: "Площадь", prefix: true},
	"сад":        addrShortName{short: "сад", full: "Сад", prefix: false},
	"р-н":        addrShortName{short: "р-н", full: "Район", prefix: false},
	"км":         addrShortName{short: "км", full: "Километр", prefix: false},
	"м":          addrShortName{short: "м.", full: "Местечко", prefix: true},
	"лн":         addrShortName{short: "лн.", full: "Линия", prefix: true},
	"ш":          addrShortName{short: "ш.", full: "Шоссе", prefix: true},
	"пр-кт":      addrShortName{short: "пр-кт", full: "Проспект", prefix: true},
	"зона":       addrShortName{short: "зона", full: "Зона (массив)", prefix: false},
	"ал":         addrShortName{short: "ал.", full: "Аллея", prefix: true},
	"дор":        addrShortName{short: "дор.", full: "Дорога", prefix: true},
	"с/с":        addrShortName{short: "с/с", full: "Сельсовет", prefix: true},
	"г":          addrShortName{short: "г.", full: "Город", prefix: true},
	"б-р":        addrShortName{short: "б-р", full: "Бульвар", prefix: true},
	"тер. ТСН":   addrShortName{short: "тер. ТСН", full: "Территория товарищества собственников недвижимости", prefix: true},
	"аллея":      addrShortName{short: "ал.", full: "Аллея", prefix: true},
	"местность":  addrShortName{short: "местность", full: "Местность", prefix: false},
	"ряды":       addrShortName{short: "ряды", full: "Ряды", prefix: false},
	"тер. ДНТ":   addrShortName{short: "тер. ДНТ", full: "Территория дачных некоммерческих товариществ", prefix: true},
	"ст":         addrShortName{short: "ст.", full: "Станция", prefix: true},
	"с/п":        addrShortName{short: "с.п.", full: "Сельское поселение", prefix: true},
	"нп":         addrShortName{short: "нп.", full: "Населенный пункт", prefix: true},
	"пгт":        addrShortName{short: "пгт.", full: "Поселок городского типа", prefix: true},
	"днп":        addrShortName{short: "днп", full: "Дачное некоммерческое партнерство", prefix: true},
	"сквер":      addrShortName{short: "с-р", full: "Сквер", prefix: true},
	"рп":         addrShortName{short: "рп.", full: "Рабочий поселок", prefix: true},
	"тер.СОСН":   addrShortName{short: "тер. СОСН", full: "Территория ведения гражданами садоводства или огородничества для собственных нужд", prefix: true},
	"тракт":      addrShortName{short: "тракт", full: "Тракт", prefix: false},
	"дп":         addrShortName{short: "дп.", full: "Дачный поселок", prefix: true},
	"промзона":   addrShortName{short: "промзона", full: "Промзона", prefix: false},
	"наб":        addrShortName{short: "наб.", full: "Набережная", prefix: true},
	"рзд":        addrShortName{short: "рзд.", full: "Разъезд", prefix: true},
	"тер. ДНП":   addrShortName{short: "тер. ДНП", full: "Территория дачных некоммерческих партнерств", prefix: true},
	"ст-ца":      addrShortName{short: "ст-ца", full: "Станица", prefix: true},
	"ж/д_ст":     addrShortName{short: "ж/д ст. ", full: "Железнодорожная станция", prefix: true},
	"стр":        addrShortName{short: "стр.", full: "Строение", prefix: true},
	"уч-к":       addrShortName{short: "уч-к.", full: "Участок", prefix: true},
	"тер. СПК":   addrShortName{short: "тер. СПК", full: "Территория садоводческих потребительских кооперативов", prefix: true},
	"парк":       addrShortName{short: "парк", full: "Парк", prefix: true},
	"п/ст":       addrShortName{short: "п. ст.", full: "Поселок при станции (поселок станции)", prefix: true},
	"г-к":        addrShortName{short: "г-к", full: "Городок", prefix: true},
	"пл-ка":      addrShortName{short: "пл-ка", full: "Площадка", prefix: true},
	"у":          addrShortName{short: "улус", full: "Улус", prefix: false},
	"аул":        addrShortName{short: "аул", full: "Аул", prefix: false},
	"ж/д_рзд":    addrShortName{short: "ж/д рзд.", full: "Железнодорожный разъезд", prefix: true},
	"жт":         addrShortName{short: "жт.", full: "жт", prefix: true},
	"массив":     addrShortName{short: "массив", full: "Массив", prefix: false},
	"ост-в":      addrShortName{short: "ост-в", full: "Остров", prefix: true},
	"тер.ф.х":    addrShortName{short: "тер.ф.х.", full: "Территория фермерского хозяйства", prefix: true},
	"починок":    addrShortName{short: "п-к", full: "Починок", prefix: true},
	"сл":         addrShortName{short: "сл.", full: "Слобода", prefix: true},
	"тер. ДПК":   addrShortName{short: "тер. ДПК", full: "Территория дачных потребительских кооперативов", prefix: true},
	"ж/д_будка":  addrShortName{short: "ж/д б-ка", full: "Железнодорожная будка", prefix: true},
	"месторожд":  addrShortName{short: "месторожд.", full: "Месторождение", prefix: true},
	"казарма":    addrShortName{short: "казарма", full: "Казарма", prefix: false},
	"ф/х":        addrShortName{short: "ф.х.", full: "Фермерское хозяйство", prefix: true},
	"п/р":        addrShortName{short: "п/р", full: "Промышленный район", prefix: true},
	"тер. СНО":   addrShortName{short: "тер. СНО", full: "Территория садоводческих некоммерческих объединений граждан", prefix: true},
	"заезд":      addrShortName{short: "заезд", full: "Заезд", prefix: false},
	"спуск":      addrShortName{short: "спуск", full: "Спуск", prefix: false},
	"въезд":      addrShortName{short: "въезд", full: "Въезд", prefix: false},
	"проул":      addrShortName{short: "проулок", full: "Проулок", prefix: false},
	"остров":     addrShortName{short: "остов", full: "Остров", prefix: true},
	"ж/д_казарм": addrShortName{short: "ж/д казарма", full: "Железнодорожная казарма", prefix: true},
	"мр":         addrShortName{short: "м.р-н", full: "Муниципальный район", prefix: true},
	"п. ж/д ст":  addrShortName{short: "п. ж/д ст.", full: "Поселок при железнодорожной станции", prefix: true},
	"проулок":    addrShortName{short: "проул.", full: "Проулок", prefix: true},
	"платф":      addrShortName{short: "платф.", full: "Платформа", prefix: true},
	"тер. ОНТ":   addrShortName{short: "тер. ОНТ", full: "Территория огороднических некоммерческих товариществ", prefix: true},
	"автодорога": addrShortName{short: "автодорога", full: "Автодорога", prefix: true},
	"тер. СНП":   addrShortName{short: "тер. СНП", full: "Территория садоводческих некоммерческих партнерств", prefix: true},
	"заимка":     addrShortName{short: "з-ка", full: "Заимка", prefix: true},
	"а/я":        addrShortName{short: "а/я", full: "Абонентский ящик", prefix: true},
	"ж/д_оп":     addrShortName{short: "ж/д о.п.", full: "Железнодорожный остановочный пункт", prefix: true},
	"ферма":      addrShortName{short: "ферма", full: "Ферма", prefix: true},
	"аал":        addrShortName{short: "аал", full: "Аал", prefix: true},
	"переезд":    addrShortName{short: "пер-д", full: "Переезд", prefix: true},
	"высел":      addrShortName{short: "в-ки", full: "Выселки", prefix: true},
	"просек":     addrShortName{short: "пр-к", full: "Просек", prefix: true},
	"сп":         addrShortName{short: "с.п.", full: "Сельское поселение", prefix: true},
	"с-р":        addrShortName{short: "с-р", full: "Сквер", prefix: true},
	"обл":        addrShortName{short: "обл", full: "Область", prefix: false},
	"гп":         addrShortName{short: "гп.", full: "Городской поселок", prefix: true},
	"тер. ПК":    addrShortName{short: "тер. ПК", full: "Территория потребительских кооперативов", prefix: true},
	"ж/р":        addrShortName{short: "ж/р", full: "Жилой район", prefix: true},
	"п/о":        addrShortName{short: "п/о", full: "Почтовое отделение", prefix: true},
	"ж/д_платф":  addrShortName{short: "ж/д платф.", full: "Железнодорожная платформа", prefix: true},
	"просека":    addrShortName{short: "пр-ка", full: "Просека", prefix: true},
	"ус":         addrShortName{short: "ус.", full: "Усадьба", prefix: true},
	"кольцо":     addrShortName{short: "к-цо", full: "Кольцо", prefix: true},
	"Респ":       addrShortName{short: "респ.", full: "Республика", prefix: true},
	"н/п":        addrShortName{short: "нп.", full: "Населенный пункт", prefix: true},
	"мгстр":      addrShortName{short: "мгстр.", full: "Магистраль", prefix: true},
	"с/мо":       addrShortName{short: "с/мо", full: "с/мо", prefix: true},
	"арбан":      addrShortName{short: "арбан", full: "Арбан", prefix: true},
	"мост":       addrShortName{short: "мост", full: "Мост", prefix: true},
	"жилрайон":   addrShortName{short: "ж/р", full: "Жилой район", prefix: true},
	"пр-ка":      addrShortName{short: "пр-ка", full: "Просека", prefix: true},
	"ж/д_пост":   addrShortName{short: "ж/д пост", full: "Железнодорожный пост", prefix: true},
	"пр-к":       addrShortName{short: "пр-к", full: "Просек", prefix: true},
	"с-к":        addrShortName{short: "с-к", full: "Спуск", prefix: true},
	"кордон":     addrShortName{short: "кордон", full: "Кордон", prefix: true},
	"с/т":        addrShortName{short: "с/т", full: "Садоводческое товарищество", prefix: true},
	"тер. ДНО":   addrShortName{short: "тер. ДНО", full: "Территория дачных некоммерческих объединений граждан", prefix: true},
	"б-г":        addrShortName{short: "б-г", full: "Берег", prefix: true},
	"тер. ОНП":   addrShortName{short: "тер. ОНП", full: "Территория огороднических некоммерческих партнерств", prefix: true},
	"край":       addrShortName{short: "край", full: "Край", prefix: false},
	"кп":         addrShortName{short: "кп.", full: "Курортный поселок", prefix: true},
	"проселок":   addrShortName{short: "пр-лок", full: "Проселок", prefix: true},
	"ззд":        addrShortName{short: "ззд.", full: "Заезд", prefix: true},
	"пер-д":      addrShortName{short: "пер-д", full: "Переезд", prefix: true},
	"тер. ОПК":   addrShortName{short: "тер. ОПК", full: "Территория огороднических потребительских кооперативов", prefix: true},
	"вал":        addrShortName{short: "вал", full: "Вал", prefix: false},
	"АО":         addrShortName{short: "а.окр", full: "Автономный округ", prefix: false},
	"лпх":        addrShortName{short: "лпх.", full: "Личное подсобное хозяйство", prefix: true},
	"м-ко":       addrShortName{short: "м-ко", full: "Местечко", prefix: true},
	"берег":      addrShortName{short: "б-г", full: "Берег", prefix: true},
	"коса":       addrShortName{short: "коса", full: "Коса", prefix: false},
	"погост":     addrShortName{short: "погост", full: "Погост", prefix: false},
	"пос.рзд":    addrShortName{short: "пос.рзд.", full: "Поселок разъезд", prefix: true},
	"взв":        addrShortName{short: "взвоз", full: "Взвоз", prefix: false},
	"г.о":        addrShortName{short: "г.о.", full: "Городской округ", prefix: true},
	"к-цо":       addrShortName{short: "к-цо", full: "Кольцо", prefix: true},
	"с/а":        addrShortName{short: "с/а", full: "Сельская администрация", prefix: true},
	"сзд":        addrShortName{short: "сзд.", full: "Съезд", prefix: true},
	"тер. ОНО":   addrShortName{short: "тер. ОНО", full: "Территория огороднических некоммерческих объединений граждан", prefix: true},
	"Аобл":       addrShortName{short: "а.обл", full: "Автономная область", prefix: false},
	"Чувашия":    addrShortName{short: "- Чувашия", full: "Чувашия", prefix: false},
	"ж/д б-ка":   addrShortName{short: "ж/д б-ка", full: "Железнодорожная будка", prefix: true},
	"ж/д бл-ст":  addrShortName{short: "ж/д бл-ст", full: "Железнодорожный блокпост", prefix: true},
	"ж/д пл-ка":  addrShortName{short: "ж/д пл-ка", full: "Железнодорожная площадка", prefix: true},
	"порт":       addrShortName{short: "порт", full: "Порт", prefix: false},
	"пр-лок":     addrShortName{short: "пр-лок", full: "Проселок", prefix: true},
	"с/о":        addrShortName{short: "с/о", full: "Сельский округ", prefix: true},
}

// Форматировать название местоположения
func PrepareFullName(shortName, offName string) string {
	fullName := ""
	name, exist := shortNameList[shortName]
	if exist && name.short != "" {
		if name.prefix {
			fullName += name.short + " " + offName
		} else {
			fullName += offName + " " + name.short
		}
	} else {
		fullName = offName
	}

	return fullName
}

// Форматировать подсказку для поиска
func PrepareSuggest(suggest, shortName, offName string) string {
	name, exist := shortNameList[shortName]
	if suggest != "" {
		suggest += ", "
	}
	if exist && name.short != "" {
		suggest += name.full + " " + name.short + " " + offName
	} else {
		suggest += shortName + " " + offName
	}

	return strings.ToLower(strings.TrimSpace(suggest))
}

// Формирование строки замены названий
func prepareReplace() {
	var replaceStingAr []string
	for key, _ := range replaceList {
		key := strings.ToLower(key)
		replaceStingAr = append(replaceStingAr, key)
	}
	replaceStingAr = UniqueStringSlice(replaceStingAr)
	SortStringSliceByLength(replaceStingAr)
	replaceSting = strings.Join(replaceStingAr, "|")
}

// Заменить текст в названии местоположения
func Replace(address string) string {
	// Сформировать строку замены, если пустая
	if len(replaceSting) == 0 {
		prepareReplace()
	}

	// Поиск совпадений по подстроке
	match := "(?is)(" + replaceSting + ")"
	re := regexp.MustCompile(match)
	matched := re.FindAllString(address, 2)
	// Замена совпадений в названии
	for _, s := range matched {
		s = strings.ToLower(strings.TrimSpace(s))
		r := regexp.MustCompile("(?is)" + s)
		address = r.ReplaceAllString(address, replaceList[s])
	}

	return address
}
