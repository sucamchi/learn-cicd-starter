package auth

import (
	"net/http"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		headers    http.Header
		wantAPIKey string
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "returns API key",
			headers:    http.Header{"Authorization": []string{"ApiKey test-key"}},
			wantAPIKey: "test-key",
		},
		{
			name:    "returns error when authorization header is missing",
			headers: http.Header{},
			wantErr: ErrNoAuthHeaderIncluded,
		},
		{
			name:       "returns error for malformed authorization header",
			headers:    http.Header{"Authorization": []string{"Bearer test-key"}},
			wantErrMsg: "malformed authorization header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAPIKey(tt.headers)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("GetAPIKey() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantErrMsg != "" {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Fatalf("GetAPIKey() error = %v, want %q", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetAPIKey() unexpected error = %v", err)
			}
			if got != tt.wantAPIKey {
				t.Errorf("GetAPIKey() = %q, want %q", got, tt.wantAPIKey)
			}
		})
	}
}
