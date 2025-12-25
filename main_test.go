package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		URL      string
		status   int
		expected int
	}{
		{"/cafe?city=moscow&count=0", http.StatusOK, 0},
		{"/cafe?city=moscow&count=1", http.StatusOK, 1},
		{"/cafe?city=moscow&count=2", http.StatusOK, 2},
		{"/cafe?city=moscow&count=100", http.StatusOK, len(cafeList["moscow"])},
		{"/cafe?city=tula&count=0", http.StatusOK, 0},
		{"/cafe?city=tula&count=1", http.StatusOK, 1},
		{"/cafe?city=tula&count=2", http.StatusOK, 2},
		{"/cafe?city=tula&count=100", http.StatusOK, len(cafeList["tula"])},
	}

	for _, test := range requests {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", test.URL, nil)
		count := 0

		handler.ServeHTTP(res, req)
		require.Equal(t, test.status, res.Code)

		// Получаем значение из тела запроса
		body := res.Body.String()
		// Удаляем пробелы в начале и конце
		bodyTrim := strings.TrimSpace(body)
		// Разделяем на слайс строк для вычисления количества элементов
		bodySplit := strings.Split(bodyTrim, ",")

		/* Если строка пустая, возвращаем 0.
		Если нет, присваиваем количество значений в слайса.*/
		if body == "" {
			count = 0
		} else {
			count = len(bodySplit)
		}

		// Прогоняем тесты
		assert.Equal(t, test.expected, count)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		searchURL string
		status    int
		expCount  int
		wantFound string
	}{
		{"/cafe?city=moscow&search=фасоль", http.StatusOK, 0, "фасоль"},
		{"/cafe?city=moscow&search=кофе", http.StatusOK, 2, "кофе"},
		{"/cafe?city=moscow&search=вилка", http.StatusOK, 1, "вилка"},
		{"/cafe?city=moscow&search=", http.StatusOK, 5, ""},
		{"/cafe?city=moscow&search=и", http.StatusOK, 3, "и"},
	}

	for _, test := range requests {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", test.searchURL, nil)
		count := 0

		handler.ServeHTTP(res, req)
		require.Equal(t, test.status, res.Code)

		// Получаем значение из тела запроса
		bodyStr := res.Body.String()
		// Удаляем пробелы в начале и конце
		bodyTrim := strings.TrimSpace(bodyStr)
		// Преведение к нижнему регистру строки
		bodyToLower := strings.ToLower(bodyTrim)
		// Разделяем на слайс строк для вычисления количества элементов
		bodySplit := strings.Split(bodyToLower, ",")

		/* Если строка пустая, возвращаем 0.
		Если нет, присваиваем количество значений в слайса.*/
		if bodyStr == "" {
			count = 0
		} else {
			// Проверка на совпадение тела запроса на наличие ключевых слов в слайсе
			for _, value := range bodySplit {
				if strings.Contains(value, test.wantFound) {
					count++
				}
			}
		}

		// Прогоняем тесты и сравниваем
		assert.Equal(t, test.expCount, count)
	}
}
