package proxy

import "encoding/json"

type JSONRpcReq struct {
	Id     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type StratumReq struct {
	JSONRpcReq
	//Worker string `json:"worker"`
}

type JSONPushMessage struct {
	Id      json.RawMessage `json:"id"`
	Version string          `json:"jsonrpc,omitempty"`
	Method  string          `json:"method"`
	Params  interface{}     `json:"params"`
	Error   interface{}     `json:"error"`
}

type JSONRpcResp struct {
	Id      json.RawMessage `json:"id"`
	Version string          `json:"jsonrpc,omitempty"`
	Result  interface{}     `json:"result"`
	Error   interface{}     `json:"error"`
}

type SubmitReply struct {
	Status string `json:"status"`
}

type ErrorReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
