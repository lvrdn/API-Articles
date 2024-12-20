package user

import (
	"encoding/json"
	"log"
	"net/http"
	"rwa/pkg/utils"
)

func unmarshalBody(w http.ResponseWriter, r *http.Request, body []byte) *User {
	dataFromBody := make(map[string]*User)
	err := json.Unmarshal(body, &dataFromBody)
	if err != nil {
		log.Printf("unmarshal body json error: [%s]; path: [%s], method: [%s]\n", err.Error(), r.URL.Path, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	user, ok := dataFromBody["user"]
	if !ok {
		utils.SendErrMessage(w, r, "no user data", http.StatusBadRequest)
		return nil
	}

	return user
}
