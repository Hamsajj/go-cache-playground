package server

import (
	"fmt"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockCache struct {
	Hit      bool
	GetValue string
	SetCalls [][]string
	GetCalls []string
}

func (m *mockCache) Set(key string, value string) {
	if m.SetCalls == nil {
		m.SetCalls = make([][]string, 0, 1)
	}
	m.SetCalls = append(m.SetCalls, []string{key, value})
}
func (m *mockCache) Get(key string) (string, bool) {
	if m.GetCalls == nil {
		m.GetCalls = make([]string, 0, 1)
	}
	m.GetCalls = append(m.GetCalls, key)
	return m.GetValue, m.Hit
}

func TestServer_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		key              string
		shouldHit        bool
		expectedStatus   int
		expectedValue    string
		expectedGetCalls int
	}{
		{
			name:             "Should return 200 with value cached value",
			key:              "user-id",
			shouldHit:        true,
			expectedStatus:   http.StatusOK,
			expectedValue:    "cache_value",
			expectedGetCalls: 1,
		},
		{
			name:             "Should return 404 response because of cache miss",
			key:              "user-id",
			shouldHit:        false,
			expectedStatus:   http.StatusNotFound,
			expectedGetCalls: 1,
		},
		{
			name:             "Should return 404 because 'GET /' is an invalid route",
			key:              "",
			shouldHit:        false,
			expectedStatus:   http.StatusNotFound,
			expectedGetCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := &mockCache{
				Hit:      tt.shouldHit,
				GetValue: tt.expectedValue,
			}
			logger := zerolog.Nop()
			handler := New(&logger, cache)
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tt.key), nil)
			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, req)
			if responseRecorder.Code != tt.expectedStatus {
				t.Fatalf("Expected status code %d, got %d", tt.expectedStatus, responseRecorder.Code)
			}
			if tt.expectedValue != "" {
				if responseRecorder.Body.String() != tt.expectedValue {
					t.Errorf("Expected response body %s, got %s", tt.expectedValue, responseRecorder.Body.String())
				}
			}
			if len(cache.GetCalls) != tt.expectedGetCalls {
				t.Errorf(
					"Expected Get to be called %d times, but was called %d times",
					tt.expectedStatus,
					len(cache.GetCalls),
				)
			}
		})
	}
}

func TestServer_Post(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		key            string
		body           string
		expectedStatus int
	}{
		{
			name:           "Should return 201",
			expectedStatus: http.StatusCreated,
			key:            "user-id",
			body:           "user-value",
		},
		{
			name:           "Should return 400 because body is empty",
			expectedStatus: http.StatusBadRequest,
			key:            "user-id",
			body:           "",
		},
		{
			name:           "Should return 404 because 'POST /' is an invalid route",
			expectedStatus: http.StatusNotFound,
			key:            "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := &mockCache{}
			logger := zerolog.Nop()
			handler := New(&logger, cache)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s", tt.key), strings.NewReader(tt.body))
			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, req)
			if responseRecorder.Code != tt.expectedStatus {
				t.Fatalf("Expected status code %d, got %d", tt.expectedStatus, responseRecorder.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				setCall := cache.SetCalls
				if len(setCall) != 1 {
					t.Error("Expected cache.Set to be called once")
				}
				gotKey := setCall[0][0]
				gotValue := setCall[0][1]
				if gotKey != tt.key {
					t.Errorf(
						"Expected the cache.Set to be called with (%s, %s), but got (%s, %s)",
						tt.key,
						tt.body,
						gotKey,
						gotValue,
					)
				}
			}

		})
	}
}
