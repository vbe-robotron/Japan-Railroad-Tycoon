package api

import (
	"log"
	"net/http"
)

// withRecover はハンドラ内の panic を 500 レスポンスへ変換する。
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("api: panic recovered: %v", rec)
				writeError(w, http.StatusInternalServerError, "PANIC", "内部エラーが発生しました")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withCORS は開発用に全オリジンを許可する CORS ヘッダを付与する。
//
// プリフライト（OPTIONS）には 204 を返して即座に終了する。
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
