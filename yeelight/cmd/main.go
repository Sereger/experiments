package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sereger/experiments/yeelight/internal/controller"
	"github.com/Sereger/experiments/yeelight/internal/session"
)

func main() {
	server := new(http.Server)
	server.Addr = ":1645"
	h := &handler{
		ctl:         &controller.Controller{},
		sessStorage: session.NewStorage(),
	}

	server.Handler = h
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		fmt.Printf("signal: %s\n", sig.String())
		server.Shutdown(context.Background())
	}()

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("done")
}

type handler struct {
	ctl         *controller.Controller
	sessStorage *session.Storage
}

type (
	req struct {
		Words struct {
			Tokens []string `json:"tokens"`
		} `json:"nlu"`
		Phrase string `json:"original_utterance"`
	}
	sess struct {
		SessionID string `json:"session_id"`
		UserID    string `json:"user_id"`
		MessageID int64  `json:"message_id"`
	}
	dtoIn struct {
		Req  req    `json:"request"`
		Sess sess   `json:"session"`
		Ver  string `json:"version"`
	}

	resp struct {
		Text string `json:"text"`
		Tts  string `json:"tts"`
		End  bool   `json:"end_session"`
	}
	dtoOut struct {
		Resp resp   `json:"response"`
		Sess sess   `json:"session"`
		Ver  string `json:"version"`
	}
)

var errAnswear = resp{
	Text: "К сожалению, я не смогла понять ваш запрос",
	Tts:  "я ничего не поняла",
	End:  true,
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/webhook" {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	data := new(dtoIn)
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		h.writeErr(w, data.Sess, data.Ver)
		return
	}

	fmt.Printf("path: [%s], req: %+v", r.URL.Path, data)
	if strings.Contains(strings.ToLower(data.Req.Phrase), "навык") {
		h.writeAnswer(w, "Что мне сделать?", data.Sess, data.Ver, true)
		return
	}

	if strings.Contains(strings.ToLower(data.Req.Phrase), "спасибо") || strings.Contains(data.Req.Phrase, "отмена") {
		h.writeAnswer(w, "Всегда пожалуйста", data.Sess, data.Ver, false)
		return
	}
	if strings.Contains(data.Req.Phrase, "отмена") {
		h.writeAnswer(w, "Поняла", data.Sess, data.Ver, false)
		return
	}

	fmt.Printf("tokens: [%+v], sess: %+v\n", data.Req.Words.Tokens, h.sessStorage.ResolveSession(data.Sess.SessionID))

	err, continueSession := h.ctl.ExecuteCommand(data.Req.Words.Tokens, h.sessStorage.ResolveSession(data.Sess.SessionID))
	if err != nil {
		h.writeAnswer(w, err.Error(), data.Sess, data.Ver, continueSession)
		return
	}
	h.writeAnswer(w, "Выполнено", data.Sess, data.Ver, continueSession)
}

func (h handler) writeErr(w http.ResponseWriter, s sess, ver string) {
	resp := dtoOut{
		Resp: errAnswear,
		Sess: s,
		Ver:  ver,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h handler) writeAnswer(w http.ResponseWriter, answ string, s sess, ver string, continueSession bool) {
	resp := dtoOut{
		Resp: resp{
			Text: answ,
			Tts:  answ,
			End:  !continueSession,
		},
		Sess: s,
		Ver:  ver,
	}
	json.NewEncoder(w).Encode(resp)
}
