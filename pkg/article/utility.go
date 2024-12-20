package article

import (
	"encoding/json"
	"log"
	"net/http"
	"rwa/pkg/utils"
)

func unmarshalBody(w http.ResponseWriter, r *http.Request, body []byte) *Article {
	dataFromBody := make(map[string]*Article)
	err := json.Unmarshal(body, &dataFromBody)
	if err != nil {
		log.Printf("unmarshal body json error: [%s]; path: [%s], method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	user, ok := dataFromBody["article"]
	if !ok {
		utils.SendErrMessage(w, r, "no article data", http.StatusBadRequest)
		return nil
	}

	return user
}
