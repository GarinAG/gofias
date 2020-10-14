package util

import (
	"regexp"
	"strings"
)

type addrShortName struct {
	short string
	full  string
}

var replaceList = map[string]string{
	"городское поселение": "г.",
}

var replaceSting = ""
var shortNameList = map[string]addrShortName{
	"г/п":        addrShortName{short: "г.", full: "Городское поселение"},
	"ул":         addrShortName{short: "ул.", full: "Улица"},
	"пер":        addrShortName{short: "пер.", full: "Переулок"},
	"д":          addrShortName{short: "д.", full: "Деревня"},
	"тер":        addrShortName{short: "тер.", full: "Территория"},
	"с":          addrShortName{short: "с.", full: "Село"},
	"тер. СНТ":   addrShortName{short: "тер. СНТ", full: "Территория садоводческих некоммерческих товариществ"},
	"снт":        addrShortName{short: "СНТ", full: "Садоводческое некоммерческое товарищество"},
	"п":          addrShortName{short: "п.", full: "Поселок"},
	"проезд":     addrShortName{short: "пр-д", full: "Проезд"},
	"кв-л":       addrShortName{short: "кв-л", full: "Квартал"},
	"гск":        addrShortName{short: "гск", full: "Гаражно-строительный кооператив"},
	"пр-д":       addrShortName{short: "пр-д", full: "Проезд"},
	"линия":      addrShortName{short: "лн.", full: "Линия"},
	"ряд":        addrShortName{short: "ряд", full: "Ряд"},
	"х":          addrShortName{short: "х.", full: "Хутор"},
	"тер. ГСК":   addrShortName{short: "тер. ГСК", full: "Территория гаражно-строительного кооператива"},
	"мкр":        addrShortName{short: "мкр.", full: "Микрорайон"},
	"туп":        addrShortName{short: "туп.", full: "Тупик"},
	"пл":         addrShortName{short: "пл.", full: "Площадь"},
	"сад":        addrShortName{short: "сад", full: "Сад"},
	"р-н":        addrShortName{short: "р-н", full: "Район"},
	"км":         addrShortName{short: "км", full: "Километр"},
	"м":          addrShortName{short: "м.", full: "Местечко"},
	"лн":         addrShortName{short: "лн.", full: "Линия"},
	"ш":          addrShortName{short: "ш.", full: "Шоссе"},
	"пр-кт":      addrShortName{short: "пр-кт", full: "Проспект"},
	"зона":       addrShortName{short: "зона", full: "Зона (массив)"},
	"ал":         addrShortName{short: "ал.", full: "Аллея"},
	"дор":        addrShortName{short: "дор.", full: "Дорога"},
	"с/с":        addrShortName{short: "с/с", full: "Сельсовет"},
	"г":          addrShortName{short: "г.", full: "Город"},
	"б-р":        addrShortName{short: "б-р", full: "Бульвар"},
	"тер. ТСН":   addrShortName{short: "тер. ТСН", full: "Территория товарищества собственников недвижимости"},
	"аллея":      addrShortName{short: "ал.", full: "Аллея"},
	"местность":  addrShortName{short: "местность", full: "Местность"},
	"ряды":       addrShortName{short: "ряды", full: "Ряды"},
	"тер. ДНТ":   addrShortName{short: "тер. ДНТ", full: "Территория дачных некоммерческих товариществ"},
	"ст":         addrShortName{short: "ст.", full: "Станция"},
	"с/п":        addrShortName{short: "с.п.", full: "Сельское поселение"},
	"нп":         addrShortName{short: "нп.", full: "Населенный пункт"},
	"пгт":        addrShortName{short: "пгт.", full: "Поселок городского типа"},
	"днп":        addrShortName{short: "днп", full: "Дачное некоммерческое партнерство"},
	"сквер":      addrShortName{short: "с-р", full: "Сквер"},
	"рп":         addrShortName{short: "рп.", full: "Рабочий поселок"},
	"тер.СОСН":   addrShortName{short: "тер. СОСН", full: "Территория ведения гражданами садоводства или огородничества для собственных нужд"},
	"тракт":      addrShortName{short: "тракт", full: "Тракт"},
	"дп":         addrShortName{short: "дп.", full: "Дачный поселок"},
	"промзона":   addrShortName{short: "промзона", full: "Промзона"},
	"наб":        addrShortName{short: "наб.", full: "Набережная"},
	"рзд":        addrShortName{short: "рзд.", full: "Разъезд"},
	"тер. ДНП":   addrShortName{short: "тер. ДНП", full: "Территория дачных некоммерческих партнерств"},
	"ст-ца":      addrShortName{short: "ст-ца", full: "Станица"},
	"ж/д_ст":     addrShortName{short: "ж/д ст. ", full: "Железнодорожная станция"},
	"стр":        addrShortName{short: "стр.", full: "Строение"},
	"уч-к":       addrShortName{short: "уч-к.", full: "Участок"},
	"тер. СПК":   addrShortName{short: "тер. СПК", full: "Территория садоводческих потребительских кооперативов"},
	"парк":       addrShortName{short: "парк", full: "Парк"},
	"п/ст":       addrShortName{short: "п. ст.", full: "Поселок при станции (поселок станции)"},
	"г-к":        addrShortName{short: "г-к", full: "Городок"},
	"пл-ка":      addrShortName{short: "пл-ка", full: "Площадка"},
	"у":          addrShortName{short: "у.", full: "Улус"},
	"аул":        addrShortName{short: "аул.", full: "Аул"},
	"ж/д_рзд":    addrShortName{short: "ж/д рзд.", full: "Железнодорожный разъезд"},
	"жт":         addrShortName{short: "жт.", full: "жт"},
	"массив":     addrShortName{short: "массив", full: "Массив"},
	"ост-в":      addrShortName{short: "ост-в", full: "Остров"},
	"тер.ф.х":    addrShortName{short: "тер.ф.х.", full: "Территория фермерского хозяйства"},
	"починок":    addrShortName{short: "п-к", full: "Починок"},
	"сл":         addrShortName{short: "сл.", full: "Слобода"},
	"тер. ДПК":   addrShortName{short: "тер. ДПК", full: "Территория дачных потребительских кооперативов"},
	"ж/д_будка":  addrShortName{short: "ж/д б-ка", full: "Железнодорожная будка"},
	"месторожд":  addrShortName{short: "месторожд.", full: "Месторождение"},
	"казарма":    addrShortName{short: "к-ма", full: "Казарма"},
	"ф/х":        addrShortName{short: "ф.х.", full: "Фермерское хозяйство"},
	"п/р":        addrShortName{short: "п/р", full: "Промышленный район"},
	"тер. СНО":   addrShortName{short: "тер. СНО", full: "Территория садоводческих некоммерческих объединений граждан"},
	"заезд":      addrShortName{short: "ззд", full: "Заезд"},
	"спуск":      addrShortName{short: "с-к", full: "Спуск"},
	"въезд":      addrShortName{short: "взд.", full: "Въезд"},
	"проул":      addrShortName{short: "проул.", full: "Проулок"},
	"остров":     addrShortName{short: "ост-в", full: "Остров"},
	"ж/д_казарм": addrShortName{short: "ж/д казарма", full: "Железнодорожная казарма"},
	"мр":         addrShortName{short: "м.р-н", full: "Муниципальный район"},
	"п. ж/д ст":  addrShortName{short: "п. ж/д ст.", full: "Поселок при железнодорожной станции"},
	"проулок":    addrShortName{short: "проул.", full: "Проулок"},
	"платф":      addrShortName{short: "платф.", full: "Платформа"},
	"тер. ОНТ":   addrShortName{short: "тер. ОНТ", full: "Территория огороднических некоммерческих товариществ"},
	"автодорога": addrShortName{short: "автодорога", full: "Автодорога"},
	"тер. СНП":   addrShortName{short: "тер. СНП", full: "Территория садоводческих некоммерческих партнерств"},
	"заимка":     addrShortName{short: "з-ка", full: "Заимка"},
	"а/я":        addrShortName{short: "а/я", full: "Абонентский ящик"},
	"ж/д_оп":     addrShortName{short: "ж/д о.п.", full: "Железнодорожный остановочный пункт"},
	"ферма":      addrShortName{short: "ферма", full: "Ферма"},
	"аал":        addrShortName{short: "аал", full: "Аал"},
	"переезд":    addrShortName{short: "пер-д", full: "Переезд"},
	"высел":      addrShortName{short: "в-ки", full: "Выселки"},
	"просек":     addrShortName{short: "пр-к", full: "Просек"},
	"сп":         addrShortName{short: "с.п.", full: "Сельское поселение"},
	"с-р":        addrShortName{short: "с-р", full: "Сквер"},
	"обл":        addrShortName{short: "обл.", full: "Область"},
	"гп":         addrShortName{short: "гп.", full: "Городской поселок"},
	"тер. ПК":    addrShortName{short: "тер. ПК", full: "Территория потребительских кооперативов"},
	"ж/р":        addrShortName{short: "ж/р", full: "Жилой район"},
	"п/о":        addrShortName{short: "п/о", full: "Почтовое отделение"},
	"ж/д_платф":  addrShortName{short: "ж/д платф.", full: "Железнодорожная платформа"},
	"просека":    addrShortName{short: "пр-ка", full: "Просека"},
	"ус":         addrShortName{short: "ус.", full: "Усадьба"},
	"кольцо":     addrShortName{short: "к-цо", full: "Кольцо"},
	"Респ":       addrShortName{short: "респ.", full: "Республика"},
	"н/п":        addrShortName{short: "нп.", full: "Населенный пункт"},
	"мгстр":      addrShortName{short: "мгстр.", full: "Магистраль"},
	"с/мо":       addrShortName{short: "с/мо", full: "с/мо"},
	"арбан":      addrShortName{short: "арбан", full: "Арбан"},
	"мост":       addrShortName{short: "мост", full: "Мост"},
	"жилрайон":   addrShortName{short: "ж/р", full: "Жилой район"},
	"пр-ка":      addrShortName{short: "пр-ка", full: "Просека"},
	"ж/д_пост":   addrShortName{short: "ж/д пост", full: "Железнодорожный пост"},
	"пр-к":       addrShortName{short: "пр-к", full: "Просек"},
	"с-к":        addrShortName{short: "с-к", full: "Спуск"},
	"кордон":     addrShortName{short: "кордон", full: "Кордон"},
	"с/т":        addrShortName{short: "с/т", full: "Садоводческое товарищество"},
	"тер. ДНО":   addrShortName{short: "тер. ДНО", full: "Территория дачных некоммерческих объединений граждан"},
	"б-г":        addrShortName{short: "б-г", full: "Берег"},
	"тер. ОНП":   addrShortName{short: "тер. ОНП", full: "Территория огороднических некоммерческих партнерств"},
	"край":       addrShortName{short: "край", full: "Край"},
	"кп":         addrShortName{short: "кп.", full: "Курортный поселок"},
	"проселок":   addrShortName{short: "пр-лок", full: "Проселок"},
	"ззд":        addrShortName{short: "ззд.", full: "Заезд"},
	"пер-д":      addrShortName{short: "пер-д", full: "Переезд"},
	"тер. ОПК":   addrShortName{short: "тер. ОПК", full: "Территория огороднических потребительских кооперативов"},
	"вал":        addrShortName{short: "вал", full: "Вал"},
	"АО":         addrShortName{short: "а.обл.", full: "Автономная область"},
	"лпх":        addrShortName{short: "лпх.", full: "Личное подсобное хозяйство"},
	"м-ко":       addrShortName{short: "м-ко", full: "Местечко"},
	"берег":      addrShortName{short: "б-г", full: "Берег"},
	"коса":       addrShortName{short: "коса", full: "Коса"},
	"погост":     addrShortName{short: "погост", full: "Погост"},
	"пос.рзд":    addrShortName{short: "пос.рзд.", full: "Поселок разъезд"},
	"взв":        addrShortName{short: "взв.", full: "Взвоз"},
	"г.о":        addrShortName{short: "г.о.", full: "Городской округ"},
	"к-цо":       addrShortName{short: "к-цо", full: "Кольцо"},
	"с/а":        addrShortName{short: "с/а", full: "Сельская администрация"},
	"сзд":        addrShortName{short: "сзд.", full: "Съезд"},
	"тер. ОНО":   addrShortName{short: "тер. ОНО", full: "Территория огороднических некоммерческих объединений граждан"},
	"Аобл":       addrShortName{short: "а.обл.", full: "Автономная область"},
	"Чувашия":    addrShortName{short: "", full: "Чувашия"},
	"ж/д б-ка":   addrShortName{short: "ж/д б-ка", full: "Железнодорожная будка"},
	"ж/д бл-ст":  addrShortName{short: "ж/д бл-ст", full: "Железнодорожный блокпост"},
	"ж/д пл-ка":  addrShortName{short: "ж/д пл-ка", full: "Железнодорожная площадка"},
	"порт":       addrShortName{short: "порт", full: "Порт"},
	"пр-лок":     addrShortName{short: "пр-лок", full: "Проселок"},
	"с/о":        addrShortName{short: "с/о", full: "Сельский округ"},
}

func PrepareFullName(shortName, offName string) string {
	fullName := ""
	name, exist := shortNameList[shortName]
	skip := false
	if exist && name.short != "" {
		if name.short == "км" {
			fullName += offName + " " + name.short
			skip = true
		} else {
			fullName += name.short + " "
		}
	}
	if !skip {
		fullName += offName
	}

	return fullName
}

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

func prepareReplace() {
	var replaceStingAr []string
	for key, _ := range replaceList {
		key := strings.ToLower(key)
		replaceStingAr = append(replaceStingAr, key)
	}
	replaceStingAr = RemoveStringsDuplicates(replaceStingAr)
	SortStringSliceByLength(replaceStingAr)
	replaceSting = strings.Join(replaceStingAr, "|")
}

func Replace(address string) string {
	if len(replaceSting) == 0 {
		prepareReplace()
	}

	match := "(?is)(" + replaceSting + ")"
	re := regexp.MustCompile(match)
	matched := re.FindAllString(address, 2)
	for _, s := range matched {
		s = strings.ToLower(strings.TrimSpace(s))
		r := regexp.MustCompile("(?is)" + s)
		address = r.ReplaceAllString(address, replaceList[s])
	}

	return address
}
