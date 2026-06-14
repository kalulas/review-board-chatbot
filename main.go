package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8080"
	}
	port := flag.String("port", defaultPort, "HTTP listen port")
	flag.Parse()

	if p, err := exec.LookPath("fortune"); err == nil {
		fortunePath = p
		log.Printf("INFO: fortune found at %s; it will join the reply pool", p)
	} else {
		log.Printf("INFO: fortune not found; replies will use stored messages only")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	r := gin.Default()
	r.Use(WithSOPSignatureValidation())

	r.POST("/callback", func(ctx *gin.Context) {
		var reqSOP SOPEventCallbackReq
		if err := ctx.BindJSON(&reqSOP); err != nil {
			ctx.JSON(http.StatusInternalServerError, "something wrong")
			return
		}
		log.Printf("INFO: received event with event_type %s", reqSOP.EventType)

		switch reqSOP.EventType {
		case "event_verification":
			ctx.JSON(http.StatusOK, SOPEventVerificationResp{SeatalkChallenge: reqSOP.Event.SeatalkChallenge})
		case "message_from_bot_subscriber":
			text := reqSOP.Event.Message.Text.Content
			log.Printf("INFO: message received: %s, with employee_code: %s", text, reqSOP.Event.EmployeeCode)
			rememberMessage(text)
			reply := pickReply()
			if reply == "" {
				reply = "Hello World"
			}
			log.Printf("INFO: replying with: %s", reply)
			if err := SendMessageToUser(ctx, reply, reqSOP.Event.EmployeeCode); err != nil {
				log.Printf("ERROR: something wrong when send message to user, error: %v", err)
				ctx.JSON(http.StatusInternalServerError, "something wrong")
				return
			}
			ctx.JSON(http.StatusOK, "Success")
		default:
			log.Printf("ERROR: event %s not handled yet!", reqSOP.EventType)
			ctx.JSON(http.StatusOK, "Success")
		}
	})

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: r,
	}

	go func() {
		log.Println("starting web, listening on", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalln("failed starting web on", srv.Addr, err)
		}
	}()

	for {
		<-c
		log.Println("terminate service")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		log.Println("shutting down web on", srv.Addr)
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalln("failed shutdown server", err)
		}
		log.Println("web gracefully stopped")
		os.Exit(0)
	}
}

func WithSOPSignatureValidation() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := ctx.Request
		signature := r.Header.Get("signature")

		if signature == "" {
			ctx.JSON(http.StatusForbidden, nil)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		hasher := sha256.New()
		signingSecret := os.Getenv("SEATALK_SIGNING_SECRET")
		hasher.Write(append(body, []byte(signingSecret)...))
		targetSignature := hex.EncodeToString(hasher.Sum(nil))

		if signature != targetSignature {
			ctx.JSON(http.StatusForbidden, nil)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		ctx.Next()
	}
}

// --- reply pool ------------------------------------------------------------
// Remembers every message users send in memory and replies with a random one.
// If `fortune` is installed on the host, its output joins the pool as one more
// candidate (resolved once at startup into fortunePath).

var (
	msgStore    []string
	msgStoreMu  sync.Mutex
	fortunePath string // path to `fortune`, empty if not installed
)

func rememberMessage(text string) {
	if text == "" {
		return
	}
	msgStoreMu.Lock()
	defer msgStoreMu.Unlock()
	msgStore = append(msgStore, text)
}

// pickReply returns a random reply drawn from all remembered messages plus,
// when available, a freshly generated fortune. Returns "" if the pool is empty.
func pickReply() string {
	msgStoreMu.Lock()
	candidates := make([]string, len(msgStore))
	copy(candidates, msgStore)
	msgStoreMu.Unlock()

	if saying, ok := fortuneSaying(); ok {
		candidates = append(candidates, saying)
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[rand.Intn(len(candidates))]
}

// fortuneSaying runs the local `fortune` binary if it was found at startup.
func fortuneSaying() (string, bool) {
	if fortunePath == "" {
		return "", false
	}
	out, err := exec.Command(fortunePath).Output()
	if err != nil {
		log.Printf("WARN: fortune execution failed: %v", err)
		return "", false
	}
	saying := strings.TrimSpace(string(out))
	if saying == "" {
		return "", false
	}
	return saying, true
}

type SOPEventCallbackReq struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	TimeStamp uint64 `json:"timestamp"`
	AppID     string `json:"app_id"`
	Event     Event  `json:"event"`
}

type SOPEventVerificationResp struct {
	SeatalkChallenge string `json:"seatalk_challenge"`
}

type Event struct {
	SeatalkChallenge string  `json:"seatalk_challenge"`
	EmployeeCode     string  `json:"employee_code"`
	Message          Message `json:"message"`
}

type Message struct {
	Tag  string      `json:"tag"`
	Text TextMessage `json:"text"`
}

type TextMessage struct {
	Content   string `json:"content"`
	PlainText string `json:"plain_text"`
}

type AppAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpireTime  uint64 `json:"expire"`
}

type SOPAuthAppResp struct {
	Code           int    `json:"code"`
	AppAccessToken string `json:"app_access_token"`
	Expire         uint64 `json:"expire"`
}

var (
	appAccessToken AppAccessToken
)

func GetAppAccessToken() AppAccessToken {
	timeNow := time.Now().Unix()

	accTokenIsEmpty := appAccessToken == AppAccessToken{}
	accTokenIsExpired := appAccessToken.ExpireTime < uint64(timeNow)

	if accTokenIsEmpty || accTokenIsExpired {
		body := []byte(fmt.Sprintf(`{"app_id": "%s", "app_secret": "%s"}`, os.Getenv("SEATALK_APP_ID"), os.Getenv("SEATALK_APP_SECRET")))

		req, err := http.NewRequest("POST", "https://openapi.seatalk.io/auth/app_access_token", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("ERROR: [GetAppAccessToken] failed to create an HTTP request: %v", err)
			return appAccessToken
		}

		req.Header.Add("Content-Type", "application/json")
		client := &http.Client{}

		res, err := client.Do(req)
		if err != nil {
			log.Printf("ERROR: [GetAppAccessToken] failed to make an HTTP call to seatalk openapi.seatalk.io: %v", err)
			return appAccessToken
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			log.Printf("ERROR: [GetAppAccessToken] got non 200 HTTP response status code: %v", err)
			return appAccessToken
		}

		resp := &SOPAuthAppResp{}
		if err := json.NewDecoder(res.Body).Decode(resp); err != nil {
			log.Printf("ERROR: [GetAppAccessToken] failed to parse response body: %v", err)
			return appAccessToken
		}

		if resp.Code != 0 {
			log.Printf("ERROR: [GetAppAccessToken] response code is not 0, error code %d, please refer to the error code documentation https://open.seatalk.io/docs/reference_server-api-error-code", resp.Code)
			return appAccessToken
		}

		appAccessToken = AppAccessToken{
			AccessToken: resp.AppAccessToken,
			ExpireTime:  resp.Expire,
		}
	}

	return appAccessToken
}

type SOPSendMessageToUser struct {
	EmployeeCode string     `json:"employee_code"`
	Message      SOPMessage `json:"message"`
}

type SOPMessage struct {
	Tag  string     `json:"tag"`
	Text SOPTextMsg `json:"text,omitempty"`
}

type SOPTextMsg struct {
	Format  int8   `json:"format"`
	Content string `json:"content"`
}

type SendMessageToUserResp struct {
	Code int `json:"code"`
}

func SendMessageToUser(ctx context.Context, message, employeeCode string) error {
	bodyJson, _ := json.Marshal(SOPSendMessageToUser{
		EmployeeCode: employeeCode,
		Message: SOPMessage{
			Tag: "text",
			Text: SOPTextMsg{
				Format:  2, //plain text message
				Content: message,
			},
		},
	})

	req, err := http.NewRequest("POST", "https://openapi.seatalk.io/messaging/v2/single_chat", bytes.NewBuffer(bodyJson))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	// Every SOP API need an authorization, to make sure that our Bot is authorized to call that API. We will use the token that we retrieved on the Step 2.
	req.Header.Add("Authorization", "Bearer "+GetAppAccessToken().AccessToken)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resp := &SendMessageToUserResp{}
	if err := json.NewDecoder(res.Body).Decode(resp); err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("something wrong, response code: %v", resp.Code)
	}

	return nil
}
