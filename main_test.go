package main

import (
	"fmt"
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

	type request struct {
		adr   string
		count int
		want  int
	}

	var requests []request

	testCount := []int{0, 1, 2, 100}
	cafeTotal := len(cafeList["moscow"])

	for _, v := range testCount {
		newrequest := request{}
		newrequest.adr = fmt.Sprintf("/cafe?count=%d&city=moscow", v)
		newrequest.count = v
		if v > cafeTotal {
			newrequest.want = cafeTotal
			requests = append(requests, newrequest)
			continue
		}
		if v == 0 {
			newrequest.want = 1
			requests = append(requests, newrequest)
			continue
		}
		newrequest.want = v
		requests = append(requests, newrequest)
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.adr, nil)

		handler.ServeHTTP(response, req)
		body := strings.TrimSpace(response.Body.String())

		require.Equal(t, http.StatusOK, response.Code)

		assert.Equal(t, v.want, len(strings.Split(body, ",")))
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	type request struct {
		adr       string
		search    string
		wantCount int
	}

	testSearch := []string{"фасоль", "кофе", "вилка"}
	moscowCafes := cafeList["moscow"]
	var requests []request

	for _, v := range testSearch {
		newRequest := request{}
		newRequest.adr = fmt.Sprintf("/cafe?search=%s&city=moscow", v)
		newRequest.search = v
		var found []string
		for _, cafe := range moscowCafes {
			if strings.Contains(strings.ToLower(cafe), strings.ToLower(v)) {
				found = append(found, cafe)
			}
		}
		newRequest.wantCount = len(found)
		requests = append(requests, newRequest)
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.adr, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		resCafes := strings.ToLower(response.Body.String())
		if v.wantCount != 0 {
			assert.True(t, strings.Contains(strings.ToLower(resCafes), strings.ToLower(v.search)))
		}
		assert.Equal(t, v.wantCount, strings.Count(resCafes, v.search))
	}
}
