package jolokia2_exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/influxdata/telegraf/internal/tls"
)

type Client struct {
	URL    string
	client *http.Client
	config *ClientConfig
}

type ClientConfig struct {
	ResponseTimeout time.Duration
	Username        string
	Password        string
	ProxyConfig     *ProxyConfig
	tls.ClientConfig
}

type ProxyConfig struct {
	DefaultTargetUsername string
	DefaultTargetPassword string
	Targets               []ProxyTargetConfig
}

type ProxyTargetConfig struct {
	Username string
	Password string
	URL      string
}

type ExecRequest struct {
	Mbean 		string
	Operation	string
	Arguments	[]string
}

type ExecResponse struct {
	Status 				int
	Value               string
	TimeStamp           uint32 
	RequestMbean		string
	RequestOperation	string
	RequestArguments	[]string
	RequestTarget		string
}

// Jolokia JSON request object. Example: {
//    "type":"exec",
//    "mbean":"java.lang:type=Threading",
//    "operation":"dumpAllThreads",
//    "arguments":[true,true]
// }

// wrapper structure for an EXEC json request
type jolokiaExecRequest struct {
	Type      string         `json:"type"`
	Mbean     string         `json:"mbean"`
	Operation string 		 `json:"operation,omitempty"`
	Argument  interface{}	 `json:"argument,omitempty"`
	Target    *jolokiaTarget `json:"target,omitempty"`
}

type jolokiaTarget struct {
	URL      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// Jolokia JSON response object. Example: {
//    "value":{...},
//    "status":200,
//    "request": {
//                 "type":"exec",
//                 "mbean":"java.util.logging:type=Logging",
//                 "operation":"setLoggerLevel",
//                 "arguments":["global","INFO"]
//               },
//     "timestamp":1561057459
// }

type jolokiaExecResponse struct {
	Request jolokiaExecRequest  `json:"request"`
	Value   string    		`json:"value"` // value is the execution details string
	Status  int            		`json:"status"`
	TimeStamp  	uint32          `json:"timestamp"` 
}

// create a new client
func NewClient(url string, config *ClientConfig) (*Client, error) {
	tlsConfig, err := config.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: config.ResponseTimeout,
		TLSClientConfig:       tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.ResponseTimeout,
	}

	return &Client{
		URL:    url,
		config: config,
		client: client,
	}, nil
}

// exec makes POST execute request to the Jolokia API end point to conduct corresponding operation on MBean
func (c *Client) exec(requests []ExecRequest) ([]ExecResponse, error) {
	jrequests := makeJolokiaExecRequests(requests, c.config.ProxyConfig)
	requestBody, err := json.Marshal(jrequests)
	if err != nil {
		return nil, err
	}

	requestUrl, err := formatUrl(c.URL, c.config.Username, c.config.Password)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			c.URL, resp.StatusCode, http.StatusText(resp.StatusCode), http.StatusOK, http.StatusText(http.StatusOK))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jresponses []jolokiaExecResponse
	if err = json.Unmarshal([]byte(responseBody), &jresponses); err != nil {
		return nil, fmt.Errorf("Error decoding JSON response: %s: %s", err, responseBody)
	}

	return makeExecResponses(jresponses), nil
}

// makeJolokiaExecRequests creates an array of Jolokia execute requests with extra proxy configs
func makeJolokiaExecRequests (requests []ExecRequest, proxyConfig *ProxyConfig) []jolokiaExecRequest {
    jrequests := make([]jolokiaExecRequest, 0)
	if proxyConfig == nil {
		for _, rr := range requests {
			jrequests = append(jrequests, makeJolokiaExecRequest(rr, nil))
		}
	} else {
		for _, t := range proxyConfig.Targets {
			if t.Username == "" {
				t.Username = proxyConfig.DefaultTargetUsername
			}
			if t.Password == "" {
				t.Password = proxyConfig.DefaultTargetPassword
			}

			for _, rr := range requests {
				jtarget := &jolokiaTarget{
					URL:      t.URL,
					User:     t.Username,
					Password: t.Password,
				}

				jrequests = append(jrequests, makeJolokiaExecRequest(rr, jtarget))
			}
		}
	}
	return jrequests
}

// makeJolokiaExecRequest creates a Jolokia execute request object
func makeJolokiaExecRequest (request ExecRequest, jtarget *jolokiaTarget) jolokiaExecRequest {
	jrequest := jolokiaExecRequest{
		Type:   "exec",
		Mbean:  request.Mbean,
		Operation:	request.Operation,
		Argument:	request.Arguments,
		Target: jtarget,
	}

	return jrequest
} 

// makeExecResponses creates an array of Jolokia execute response objects
func makeExecResponses(jresponses []jolokiaExecResponse) []ExecResponse {
	responses := make([]ExecResponse, 0)

	for _, jr := range jresponses {
		request := ExecRequest{
			Mbean:      jr.Request.Mbean,
			Operation:	jr.Request.Operation,
	        Arguments:	[]string{},
		}

		argValue := jr.Request.Argument
		if argValue != nil {
			argument, ok := argValue.(string)
			if ok {
				request.Arguments = []string{argument}
			} else {
				arguments, _ := argValue.([]interface{})
				request.Arguments = make([]string, len(arguments))
				for i, arg := range arguments {
					request.Arguments[i] = arg.(string)
				}
			}
		}
		
		response := ExecResponse {
			Value:             jr.Value,
			Status:            jr.Status,
			TimeStamp: 		   jr.TimeStamp,
			RequestMbean:		request.Mbean,
	        RequestOperation:	request.Operation,
	        RequestArguments:	request.Arguments,
		}

		if jtarget := jr.Request.Target; jtarget != nil {
			response.RequestTarget = jtarget.URL
		}

		responses = append(responses, response)
	}

	return responses
}

func formatUrl(configUrl, username, password string) (string, error) {
	parsedUrl, err := url.Parse(configUrl)
	if err != nil {
		return "", err
	}

	resultUrl := url.URL{
		Host:   parsedUrl.Host,
		Scheme: parsedUrl.Scheme,
	}

	if username != "" || password != "" {
		resultUrl.User = url.UserPassword(username, password)
	}

	resultUrl.Path = path.Join(parsedUrl.Path, "exec")
	resultUrl.Query().Add("ignoreErrors", "true")
	return resultUrl.String(), nil
}