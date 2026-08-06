package observability

import (
	"net/http"
)

func VersionHandler(buildInfo BuildInfoProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		vi := GetVersionInfo()
		if buildInfo != nil {
			vi.ContentGeneration = buildInfo.GetContentGeneration()
		}
		writeJSON(w, http.StatusOK, vi)
	}
}

type StaticBuildInfo struct {
	ContentGen int64
}

func (s *StaticBuildInfo) GetContentGeneration() int64 {
	if s == nil {
		return 0
	}
	return s.ContentGen
}

func NoOpBuildInfo() *StaticBuildInfo {
	return &StaticBuildInfo{}
}
