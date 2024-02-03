package JSONResponse

import (
	"net/http"

	"github.com/bitly/go-simplejson"
)

func GenerateJson(w http.ResponseWriter, key string, val interface{}) ([]byte, error) {
	payload := []byte{}
	var err error

	jsonNew := simplejson.New()
	jsonNew.Set(key, val)

	payload, err = jsonNew.MarshalJSON()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		return payload, err
	}

	w.Header().Set("Content-Type", "application/json")
	return payload, nil
}
