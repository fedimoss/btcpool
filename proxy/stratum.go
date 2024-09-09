package proxy

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/PowPool/btcpool/bitcoin"
	"github.com/mutalisk999/bitcoin-lib/src/bigint"

	. "github.com/PowPool/btcpool/util"
)

const (
	MaxReqSize        = 1024
	Bip320Mask uint32 = 0x1fffe000
)

func (s *ProxyServer) ListenTCP() {
	timeout := MustParseDuration(s.config.Proxy.Stratum.Timeout)
	s.timeout = timeout

	addr, err := net.ResolveTCPAddr("tcp", s.config.Proxy.Stratum.Listen)
	if err != nil {
		Error.Fatalf("Error: %v", err)
	}
	server, err := net.ListenTCP("tcp", addr)
	if err != nil {
		Error.Fatalf("Error: %v", err)
	}
	defer server.Close()

	Info.Printf("Stratum listening on %s", s.config.Proxy.Stratum.Listen)
	var accept = make(chan int, s.config.Proxy.Stratum.MaxConn)

	tag := 0
	for i := 0; i < s.config.Proxy.Stratum.MaxConn; i++ {
		accept <- i
	}

	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			continue
		}
		Info.Println("Accept Stratum TCP Connection from: ", conn.RemoteAddr().String())

		_ = conn.SetKeepAlive(true)

		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

		if s.policy.IsBanned(ip) || !s.policy.ApplyLimitPolicy(ip) {
			_ = conn.Close()
			continue
		}

		tag = <-accept
		cs := &Session{conn: conn, ip: ip, shareCountInv: 0, tag: uint16(tag), isAuth: false}

		go func(cs *Session, tag int) {
			err = s.handleTCPClient(cs)
			if err != nil {
				s.removeSession(cs)
				_ = conn.Close()
			}
			accept <- tag
		}(cs, tag)
	}
}

func (s *ProxyServer) handleTCPClient(cs *Session) error {
	cs.enc = json.NewEncoder(cs.conn)
	connBuf := bufio.NewReaderSize(cs.conn, MaxReqSize)
	s.setDeadline(cs.conn)

	for {
		data, isPrefix, err := connBuf.ReadLine()
		if isPrefix {
			Error.Printf("Socket flood detected from %s", cs.ip)
			s.policy.BanClient(cs.ip)
			return err
		} else if err == io.EOF {
			Info.Printf("Client %s disconnected", cs.ip)
			s.removeSession(cs)
			_ = cs.conn.Close()
			break
		} else if err != nil {
			Error.Printf("Error reading from socket: %v", err)
			Error.Printf("Address: [%s] | Name: [%s] | IP: [%s]", cs.login, cs.id, cs.ip)
			return err
		}

		if len(data) > 1 {
			var req StratumReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				s.policy.ApplyMalformedPolicy(cs.ip)
				Error.Printf("handleTCPClient: Malformed stratum request from %s: %v", cs.ip, err)
				return err
			}

			s.setDeadline(cs.conn)
			err = cs.handleTCPMessage(s, &req)
			if err != nil {
				Error.Printf("handleTCPMessage: %v", err)
				return err
			}
		}
	}
	return nil
}

func (cs *Session) handleTCPMessage(s *ProxyServer, req *StratumReq) error {
	Debug.Printf("handleTCPMessage, req.Method: %v", req.Method)
	Debug.Printf("handleTCPMessage, req.Params: %v", string(req.Params))

	// Handle RPC methods
	switch req.Method {

	case "mining.subscribe":
		var params []string
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			Error.Println("Malformed stratum request (mining.subscribe) params from", cs.ip)
			return err
		}
		if len(params) > 0 {
			Info.Println("mining.subscribe:", params[0])
		}

		/*
			if len(params) > 1 && params[1] != "" {
				cs.sid = params[1]
			}
		*/

		reply, errReply := s.handleSubscribeRPC(cs)
		if errReply != nil {
			return cs.sendTCPError(req.Id, errReply)
		}

		// Send subscribe response to the miner
		if err = cs.sendTCPResult(req.Id, reply, "2.0"); err != nil {
			return err
		}

		//set_difficulty
		err = cs.setDifficulty()
		if err != nil {
			Error.Printf("set difficulty error to %v@%v: %v", cs.login, cs.ip, err)
			s.removeSession(cs)
		}

		// Send mining.set_version_mask after the subscribe response
		if cs.versionMask != 0 {
			versionMask := []string{"1fffe000"}
			message := JSONPushMessage{Id: nil, Method: "mining.set_version_mask", Params: versionMask}
			m, _ := json.Marshal(&message)
			Debug.Printf("mining.set_version_mask, message: %s", string(m))
			err = cs.enc.Encode(&message)
		}

		return err

	case "mining.authorize":
		var params []string
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			Error.Println("Malformed stratum request (mining.authorize) params from", cs.ip)
			return err
		}
		reply, errReply := s.handleAuthorizeRPC(cs, params)
		if errReply != nil {
			return cs.sendTCPError(req.Id, errReply)
		}
		err = cs.sendTCPResult(req.Id, reply, "")
		if err != nil {
			return err
		}

		return err

	case "mining.submit":
		var params []string
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			Error.Println("Malformed stratum request (mining.submit) params from", cs.ip)
			return err
		}

		Debug.Printf("mining.submit, Param: %v", params)

		reply, errReply := s.handleTCPSubmitRPC(cs, params)
		if errReply != nil {
			return cs.sendTCPError(req.Id, errReply)
		}
		return cs.sendTCPResult(req.Id, reply, "2.0")

	case "mining.extranonce.subscribe":
		return cs.sendTCPResult(req.Id, true, "")

	case "mining.configure":
		var params []interface{}
		err := json.Unmarshal(req.Params, &params)
		Debug.Printf("mining.configure, Param: %v", params)
		Debug.Printf("mining.configure, error: %v", err)
		if err != nil {
			return cs.sendTCPError(req.Id, &ErrorReply{Code: 27, Message: "Illegal params"})
		}
		if len(params) < 2 {
			return cs.sendTCPError(req.Id, &ErrorReply{Code: 27, Message: "Too few params"})
		}

		reply, errReply := s.handleConfigureRPC(cs, params)
		Debug.Printf("mining.configure, reply: %v", reply)
		Debug.Printf("mining.configure, errReply: %v", errReply)
		if errReply != nil {
			return cs.sendTCPError(req.Id, errReply)
		}
		cs.versionMask = Bip320Mask
		return cs.sendTCPResult(req.Id, reply, "")

	default:
		errReply := s.handleUnknownRPC(cs, req.Method)
		return cs.sendTCPError(req.Id, errReply)
	}
}

func (cs *Session) sendTCPResult(id json.RawMessage, result interface{}, version string) error {
	cs.Lock()
	defer cs.Unlock()

	message := JSONRpcResp{Id: id, Version: version, Error: nil, Result: result}
	m, _ := json.Marshal(&message)
	Debug.Printf("sendTCPResult: %s", string(m))
	return cs.enc.Encode(&message)
}

func (cs *Session) setDifficulty() error {
	cs.Lock()
	defer cs.Unlock()

	genesisWork, err := bitcoin.GetGenesisTargetWork()
	if err != nil {
		return err
	}

	diff := TargetHexToDiff(cs.targetNextJob).Int64()
	setDiff := float64(diff) / genesisWork
	//setDiff := 1000.0

	message := JSONPushMessage{Id: nil, Method: "mining.set_difficulty", Params: []interface{}{setDiff}}
	m, _ := json.Marshal(&message)
	Debug.Printf("diff:%v,genesisWork:%v,mining.set_difficulty:%s", diff, genesisWork, string(m))
	return cs.enc.Encode(&message)
}

func (cs *Session) pushNewJob(params []interface{}) error {
	cs.Lock()
	defer cs.Unlock()

	//params := []interface{}{"000000aa", "ad3c695df5a484eed6d8555676b6bb5a59110446a45d74bf517ea0f6a0b07634", "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff2b034e2501046d65da6600", "182f6d696e7420627920736f6c6f6672616374616c2e696f2f00000000020000000000000000266a24aa21a9ede6581639a17e19736e14311d7cceaabfc1674398460cf53c290a59e2ff850fc7678e4995000000001976a91429e5947f66884ee245f7e571332064eafa4515b888ac00000000", []string{"8dd83b1886475b56a1c793ec34d185f187d5fde44f45009223cad1fe0789823d", "4575a258aaf19ea3324bd65ca9e6edf8c24669af3524b90f39f47d2f91255c32", "46321f21d8085b86e10177445c37c5146d74344a8793c44b293875ed935c2060", "b9c57828bc1d0fd7316d37ee3c4a9c53bb17f4e425b72948de66bd111e9dc6b0", "5a14eb9972c9042e639f6672f28061f66eb1db0d5571bc6b3ebe7af58e774603", "d3b4a7247375f5caf0ab5ee676e2bba5af82dfccceb7e9eaa199b4849412b32b", "d8425319cf07c5d85d3fd55cf1cb02e0e9bbf03b461617f190abed951b3a665e", "cfcb27b9f3806dcb42130145398bc5cca8edda70ea4b416891895b89fada0c6e", "bd8d6ad49fe48ab26eadcd1c5364a6d3a09ce1b6c93329202c5972a6244a082d", "55090461df54a8f3e4e2c28c528a360a81c72c7801cd2a34b1777a0c4e3cc4b2", "eb71f02da145dfa6eeaafcba8a9db63aa7a9d8c27acea4ced4fb124b04dd508b", "e6160561c914d5aa238ba57362688726a48360b5bc944f83d6e6b0fb2362d877"}, "20000004", "1900e1bf", "66da656c", true}

	message := JSONPushMessage{Id: nil, Method: "mining.notify", Version: "2.0", Params: params}
	//m, _ := json.Marshal(&message)
	//Debug.Printf("pushNewJob-->mining.notify: %s", string(m))
	return cs.enc.Encode(&message)
}

func (cs *Session) sendTCPError(id json.RawMessage, reply *ErrorReply) error {
	cs.Lock()
	defer cs.Unlock()

	message := JSONRpcResp{Id: id, Version: "2.0", Error: reply}
	err := cs.enc.Encode(&message)
	if err != nil {
		return err
	}
	return errors.New(reply.Message)
}

func (s *ProxyServer) setDeadline(conn *net.TCPConn) {
	_ = conn.SetDeadline(time.Now().Add(s.timeout))
}

func (s *ProxyServer) registerSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[cs] = struct{}{}
}

func (s *ProxyServer) removeSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, cs)
}

func (s *ProxyServer) broadcastNewJobs() {
	t := s.currentBlockTemplate()
	if t == nil || len(t.PrevHash) == 0 || s.isSick() {
		return
	}
	var params []interface{}

	// reverse prev hash in bytes
	var prevHash bigint.Uint256
	err := prevHash.SetHex(t.PrevHash)
	if err != nil {
		return
	}

	prevHashHex := prevHash.GetHex()
	prevHashHexStratum, err := TargetHash256StratumFormat(prevHashHex)
	if err != nil {
		return
	}

	tplJob, ok := t.BlockTplJobMap[t.lastBlkTplId]
	if !ok {
		return
	}

	//t.Version = t.Version | 0x1fffe000

	// https://stackoverflow.com/questions/44119793/why-does-json-encoding-an-empty-array-in-code-return-null
	// var MerkleBranchStratum []string
	MerkleBranchStratum := make([]string, 0)
	for _, hashHex := range tplJob.MerkleBranch {
		hashHexStratum, err := Hash256StratumFormat(hashHex)
		if err != nil {
			return
		}
		MerkleBranchStratum = append(MerkleBranchStratum, hashHexStratum)
	}

	//应用 versionMask
	//maskVersion := t.Version | defaultVersionMask
	maskVersion := t.Version

	params = append(append(append(append(append(params, t.lastBlkTplId), prevHashHexStratum), tplJob.CoinBase1), tplJob.CoinBase2), MerkleBranchStratum)
	params = append(append(append(params, fmt.Sprintf("%08x", maskVersion)), fmt.Sprintf("%08x", t.NBits)), fmt.Sprintf("%08x", tplJob.BlkTplJobTime))
	params = append(params, t.newBlkTpl)

	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	count := len(s.sessions)
	Info.Printf("Broadcasting new job to %v stratum miners", count)

	start := time.Now()
	bcast := make(chan int, 1024)
	n := 0

	for m := range s.sessions {
		if !m.isAuth {
			continue
		}

		n++
		bcast <- n

		go func(s *ProxyServer, cs *Session) {

			err := cs.pushNewJob(params)

			<-bcast
			if err != nil {
				Error.Printf("Job transmit error to %v@%v: %v", cs.login, cs.ip, err)
				s.removeSession(cs)
			} else {
				s.setDeadline(cs.conn)
			}
		}(s, m)
	}
	Info.Printf("Jobs broadcast finished %s", time.Since(start))
}
