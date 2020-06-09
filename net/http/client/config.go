/**------------------------------------------------------------**
 * @filename client/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-06-25 15:42
 * @desc     go.easytech.co - client -
 **------------------------------------------------------------**/
package client

import (
	"net/http"
	"strings"

	"go.jd100.com/medusa/log"
)

type ApiConfig struct {
	Clients map[string]*Config
}

type Config struct {
	Addr       string
	Type       string
}

func Put(cfg *Config, body []byte) *http.Response {
	payload := strings.NewReader(string(body))
	req, err := http.NewRequest("PUT", cfg.Addr, payload)
	if err != nil {
		log.Errorf("request api tutor err %v", err)
	}

	req.Header.Add("Content-Type", cfg.Type)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("do request api tutor err %v", err)
	}
	return res
}