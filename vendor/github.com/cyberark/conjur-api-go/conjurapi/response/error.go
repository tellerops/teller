package response

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/cyberark/conjur-api-go/conjurapi/logging"
)

type ConjurError struct {
	Code    int
	Message string
	Details *ConjurErrorDetails `json:"error"`
}

type ConjurErrorDetails struct {
	Message string
	Code    string
	Target  string
	Details map[string]interface{}
}

func NewConjurError(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	cerr := ConjurError{}
	cerr.Code = resp.StatusCode
	err = json.Unmarshal(body, &cerr)
	if err != nil {
		cerr.Message = strings.TrimSpace(string(body))
	}

	// If the body's empty, use the HTTP status as the message
	if cerr.Message == "" {
		cerr.Message = resp.Status
	}

	return &cerr
}

func (self *ConjurError) Error() string {
	logging.ApiLog.Debugf("self.Details: %+v, self.Message: %+v\n", self.Details, self.Message)

	var b strings.Builder

	if self.Message != "" {
		b.WriteString(self.Message + ". ")
	}

	if self.Details != nil && self.Details.Message != "" {
		b.WriteString(self.Details.Message + ".")
	}

	return b.String()
}
