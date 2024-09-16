package response

import (
	"encoding/json"
	"net/http"
)

func ResultJSON(res http.ResponseWriter, status int, body map[string]any) {
	body["status"] = status

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(status)

	_ = json.NewEncoder(res).Encode(body)
}
