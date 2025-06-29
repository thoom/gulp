package client

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// mockHTTPClient is a mock implementation of the httpClient interface for testing.
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestCheckForUpdatesWithClient(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		mockResponse   *http.Response
		mockErr        error
		wantHasUpdate  bool
		wantErr        bool
	}{
		{
			name:           "Update Available",
			currentVersion: "1.0.0",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"tag_name": "v1.1.0"}`))),
			},
			wantHasUpdate: true,
			wantErr:       false,
		},
		{
			name:           "No Update Available",
			currentVersion: "1.1.0",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"tag_name": "v1.1.0"}`))),
			},
			wantHasUpdate: false,
			wantErr:       false,
		},
		{
			name:           "API Error",
			currentVersion: "1.0.0",
			mockResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("not found")),
			},
			wantHasUpdate: false,
			wantErr:       true,
		},
		{
			name:           "SNAPSHOT version",
			currentVersion: "1.2.0-SNAPSHOT",
			wantHasUpdate:  false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				response: tt.mockResponse,
				err:      tt.mockErr,
			}
			info, err := checkForUpdatesWithClient(tt.currentVersion, mockClient)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkForUpdatesWithClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if info != nil && info.HasUpdate != tt.wantHasUpdate {
				t.Errorf("checkForUpdatesWithClient() HasUpdate = %v, want %v", info.HasUpdate, tt.wantHasUpdate)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"1.1.0", "1.0.0", true},
		{"1.0.1", "1.0.0", true},
		{"2.0.0", "1.9.9", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.1.0", false},
		{"1.0.0", "2.0.0", false},
		{"", "1.0.0", false},
		{"1.0.0", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			if got := isNewerVersion(tt.latest, tt.current); got != tt.want {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    [3]int
	}{
		{"Simple", "1.2.3", [3]int{1, 2, 3}},
		{"Two parts", "1.2", [3]int{1, 2, 0}},
		{"One part", "1", [3]int{1, 0, 0}},
		{"Invalid chars", "1.a.3", [3]int{1, 0, 3}},
		{"Empty", "", [3]int{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseVersion(tt.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
