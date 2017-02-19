package magnumci

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	cc "github.com/grokify/commonchat"
	"github.com/grokify/webhook-proxy-go/src/adapters"
	"github.com/grokify/webhook-proxy-go/src/config"
	"github.com/grokify/webhook-proxy-go/src/util"
	"github.com/valyala/fasthttp"
)

const (
	DisplayName = "Magnum CI"
	HandlerKey  = "magnumci"
	IconURL     = "https://pbs.twimg.com/profile_images/433440931543388160/nZ3y7AB__400x400.png"
)

// FastHttp request handler for Semaphore CI outbound webhook
type MagnumciOutToGlipHandler struct {
	Config  config.Configuration
	Adapter adapters.Adapter
}

// FastHttp request handler constructor for Semaphore CI outbound webhook
func NewMagnumciOutToGlipHandler(cfg config.Configuration, adapter adapters.Adapter) MagnumciOutToGlipHandler {
	return MagnumciOutToGlipHandler{Config: cfg, Adapter: adapter}
}

// HandleFastHTTP is the method to respond to a fasthttp request.
func (h *MagnumciOutToGlipHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	ccMsg, err := Normalize(ctx.PostBody())

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotAcceptable)
		log.WithFields(log.Fields{
			"type":   "http.response",
			"status": fasthttp.StatusNotAcceptable,
		}).Info(fmt.Sprintf("%v request is not acceptable.", DisplayName))
		return
	}

	util.SendWebhook(ctx, h.Adapter, ccMsg)
}

func Normalize(bytes []byte) (cc.Message, error) {
	message := cc.NewMessage()
	message.IconURL = IconURL

	src, err := MagnumciOutMessageFromBytes(bytes)
	if err != nil {
		return message, err
	}

	if len(src.Title) > 0 {
		message.Activity = fmt.Sprintf("%v", src.Title)
	} else {
		message.Activity = fmt.Sprintf("%s Notification", DisplayName)
	}

	attachment := cc.NewAttachment()

	if len(src.Message) > 0 {
		if len(src.CommitURL) > 0 {
			attachment.AddField(cc.Field{
				Title: "Commit",
				Value: fmt.Sprintf("[%v](%v)", src.Message, src.CommitURL)})
		} else {
			attachment.AddField(cc.Field{
				Title: "Commit",
				Value: fmt.Sprintf("%v", src.Message)})
		}
	} else if len(src.CommitURL) > 0 {
		attachment.AddField(cc.Field{
			Title: "Commit",
			Value: fmt.Sprintf("[View Commit](%v)", src.Message)})
	}

	if len(src.Author) > 0 {
		attachment.AddField(cc.Field{
			Title: "Author",
			Value: src.Author,
			Short: true})
	}
	if len(src.DurationString) > 0 {
		attachment.AddField(cc.Field{
			Title: "Duration",
			Value: src.DurationString,
			Short: true})
	}
	if len(src.BuildURL) > 0 {
		attachment.AddField(cc.Field{
			Value: fmt.Sprintf("[View Build](%v)", src.BuildURL)})
	}

	if len(src.Title) < 1 && len(attachment.Fields) == 0 {
		return message, errors.New("Content not found")
	}

	message.AddAttachment(attachment)
	return message, nil
}

type MagnumciOutMessage struct {
	Id             int64  `json:"id,omitempty"`
	ProjectId      int64  `json:"project_id,omitempty"`
	Title          string `json:"title,omitempty"`
	Number         int64  `json:"number,omitempty"`
	Commit         string `json:"commit,omitempty"`
	Author         string `json:"author,omitempty"`
	Committer      string `json:"committer,omitempty"`
	Message        string `json:"message,omitempty"`
	Branch         string `json:"branch,omitempty"`
	State          string `json:"state,omitempty"`
	Status         string `json:"status,omitempty"`
	Result         int64  `json:"result,omitempty"`
	Duration       int64  `json:"duration,omitempty"`
	DurationString string `json:"duration_string,omitempty"`
	CommitURL      string `json:"commit_url,omitempty"`
	CompareURL     string `json:"compare_url,omitempty"`
	BuildURL       string `json:"build_url,omitempty"`
	StartedAt      string `json:"started_at,omitempty"`
	FinishedAt     string `json:"finished_at,omitempty"`
}

func MagnumciOutMessageFromBytes(bytes []byte) (MagnumciOutMessage, error) {
	msg := MagnumciOutMessage{}
	err := json.Unmarshal(bytes, &msg)
	return msg, err
}