package route

import (
	"encoding/base64"
	"encoding/json"
)

const (
	GET     Method       = "GET"
	POST    Method       = "POST"
	PUT     Method       = "PUT"
	PATCH   Method       = "PATCH"
	HEAD    Method       = "HEAD"
	OPTIONS Method       = "OPTIONS"
	DELETE  Method       = "DELETE"
	CONNECT Method       = "CONNECT"
	TRACE   Method       = "TRACE"
	Any     Method       = "Any"
	Pod     ProxyPattern = "Pod"
	Service ProxyPattern = "Service"
)

type (
	//method
	Method string

	//proxy pattern
	ProxyPattern string

	//rule
	Rule struct {
		//proxy method (post delete get ......)
		Method Method
		//proxy url (/api/ ....)
		AgentUrl string
		//proxy url
		ProxyUrl string
		//port
		Port int64
	}

	//rules
	Rules struct {
		//proxy pattern
		ProxyPattern ProxyPattern
		Items        []*Rule
	}
)

//marshal
func Marshal(rules *Rules) (string, error) {
	//json
	buff, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}

	//return base64
	return base64.StdEncoding.EncodeToString(buff), nil
}

//unmarshal
func Unmarshal(str string) (*Rules, error) {
	//decode base64 string
	buff, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	//json unmarshal
	r := new(Rules)
	if err := json.Unmarshal(buff, r); err != nil {
		return nil, err
	}
	return r, nil
}

func NewRules(proxyPattern ProxyPattern, rules ...*Rule) *Rules {
	r := new(Rules)
	r.ProxyPattern = proxyPattern
	for _, rule := range rules {
		r.Items = append(r.Items, rule)
	}
	return r
}

//new rule
func NewRule(method Method, agentUrl, proxyUrl string, port int64) *Rule {
	return &Rule{
		Method:   method,
		AgentUrl: agentUrl,
		ProxyUrl: proxyUrl,
		Port:     port,
	}
}
