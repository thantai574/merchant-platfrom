package fraud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net/http"
	"orders-system/domain/entities"
	"orders-system/domain/value_objects"
	"time"
)

type repoImpl struct {
	Uri    string
	Logger *zap.Logger
}

func (r repoImpl) SaveFraud(request value_objects.FraudTransRequest) (response entities.Fraud, err error) {
	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "fraud/transaction",
		Method:   http.MethodPost,
		Headers:  nil,
		Body:     request,
		Response: &response,
	})

	return response, err
}

func (r repoImpl) GetFraud(cardNumber string) (response entities.Fraud, err error) {
	body := value_objects.FraudRequest{CardNumber: cardNumber}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "fraud",
		Method:   http.MethodPost,
		Headers:  nil,
		Body:     body,
		Response: &response,
	})

	return response, err
}

func (r repoImpl) httpRequest(request struct {
	Path     string
	Method   string
	Headers  map[string]string
	Body     interface{}
	Response interface{}
}) (err error) {
	client := new(http.Client)

	client.Timeout = time.Minute * 2

	jsonrequest, err := json.Marshal(request.Body)
	r.Logger.With(zapcore.Field{
		Key:       "request",
		Type:      zapcore.StringType,
		String:    fmt.Sprintf("%v", string(jsonrequest)),
		Interface: nil,
	}).Info("fraud_request")
	req, err := http.NewRequest(request.Method, fmt.Sprintf("%v%v", r.Uri, request.Path), bytes.NewReader(jsonrequest))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", `application/json`)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 500 {
		r.Logger.Error("FRAUD SERVICE ERROR")
		return errors.New("FRAUD SERVICE ERROR")
	}

	responseByte, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	r.Logger.With(zapcore.Field{
		Key:       "uri",
		Type:      zapcore.StringType,
		String:    fmt.Sprintf("%v%v", r.Uri, request.Path),
		Interface: nil,
	}).With(
		zapcore.Field{
			Key:       "data",
			Type:      zapcore.StringType,
			String:    string(responseByte),
			Interface: nil,
		}).Info("http_request_data")

	err = json.Unmarshal(responseByte, request.Response)
	if err != nil {
		r.Logger.With().Error("can not unmarshal response")
		return err
	}
	//Close Request
	defer func() {
		err = resp.Body.Close()
	}()

	return err
}

func NewRepoImpl(uri string, logger *zap.Logger) *repoImpl {
	return &repoImpl{
		Uri:    uri,
		Logger: logger,
	}
}
