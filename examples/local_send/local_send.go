package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grokify/gotilla/fmt/fmtutil"
	"github.com/jessevdk/go-flags"

	"github.com/grokify/chathooks/src/config"
	"github.com/grokify/chathooks/src/util"
	cc "github.com/grokify/commonchat"
	ccglip "github.com/grokify/commonchat/glip"
	ccslack "github.com/grokify/commonchat/slack"
	errs "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/grokify/chathooks/examples"

	"github.com/grokify/chathooks/src/handlers/aha"
	"github.com/grokify/chathooks/src/handlers/appsignal"
	"github.com/grokify/chathooks/src/handlers/apteligent"
	"github.com/grokify/chathooks/src/handlers/circleci"
	"github.com/grokify/chathooks/src/handlers/codeship"
	"github.com/grokify/chathooks/src/handlers/confluence"
	"github.com/grokify/chathooks/src/handlers/datadog"
	"github.com/grokify/chathooks/src/handlers/deskdotcom"
	"github.com/grokify/chathooks/src/handlers/enchant"
	"github.com/grokify/chathooks/src/handlers/gosquared"
	"github.com/grokify/chathooks/src/handlers/gosquared2"
	"github.com/grokify/chathooks/src/handlers/heroku"
	"github.com/grokify/chathooks/src/handlers/librato"
	"github.com/grokify/chathooks/src/handlers/magnumci"
	"github.com/grokify/chathooks/src/handlers/marketo"
	"github.com/grokify/chathooks/src/handlers/opsgenie"
	"github.com/grokify/chathooks/src/handlers/papertrail"
	"github.com/grokify/chathooks/src/handlers/pingdom"
	"github.com/grokify/chathooks/src/handlers/raygun"
	"github.com/grokify/chathooks/src/handlers/runscope"
	"github.com/grokify/chathooks/src/handlers/semaphore"
	"github.com/grokify/chathooks/src/handlers/statuspage"
	"github.com/grokify/chathooks/src/handlers/travisci"
	"github.com/grokify/chathooks/src/handlers/userlike"
	"github.com/grokify/chathooks/src/handlers/victorops"
)

type cliOptions struct {
	GuidOrWebhook string `short:"u" long:"url" description:"Webhook or GUID" required:"true"`
	Adapter       string `short:"a" long:"adapter" description:"Adapter" required:"true"`
	Service       string `short:"s" long:"service" description:"Service"`
}

const (
	GLIP_WEBHOOK_ENV  = "GLIP_WEBHOOK"
	SLACK_WEBHOOK_ENV = "SLACK_WEBHOOK"
)

type Sender struct {
	Adapter cc.Adapter
}

func (sender *Sender) SendCcMessage(ccMsg cc.Message, err error) {
	if err != nil {
		panic(fmt.Sprintf("Bad Test Message: %v\n", err))
	}
	var resMsg interface{}
	req, resp, err := sender.Adapter.SendMessage(ccMsg, &resMsg)
	fmt.Printf("RESPONSE_STATUS_CODE [%v]\n", resp.StatusCode())
	if err != nil {
		fmt.Printf("ERROR [%v]\n", err)
	}

	fmt.Println(string(resp.Body()))

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}

func main() {
	opts := cliOptions{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	dirs, _, err := examples.DocsHandlersDirInfo()
	if err != nil {
		log.Fatal()
	}
	fmt.Println(strings.Join(dirs, ","))
	if len(opts.Service) > 0 {
		run(opts)
	}
	fmt.Println("DONE")
}

func run(opts cliOptions) {
	log.SetLevel(log.DebugLevel)

	fmt.Printf("LENGUID[%v]\n", len(opts.GuidOrWebhook))
	fmt.Printf("GUID [%v]\n", opts.GuidOrWebhook)
	fmt.Printf("EXAMPLE [%v]\n", opts.Service)

	if len(opts.Service) < 1 {
		panic("Usage: send_example.go -g=<GUID> -a=glip -s=raygun")
	}

	cfg := config.Configuration{
		IconBaseURL:    config.IconBaseURL,
		LogrusLogLevel: log.DebugLevel}

	err := SendMessageAdapterHandler(cfg, opts)
	if err != nil {
		log.Fatal(err)
	}
}

func SendMessageAdapterHandler(cfg config.Configuration, opts cliOptions) error {
	webhookURLOrUID := opts.GuidOrWebhook
	adapterType := opts.Adapter
	example := opts.Service

	sender := Sender{}
	if adapterType == "glip" {
		if len(webhookURLOrUID) < 1 {
			webhookURLOrUID = os.Getenv(GLIP_WEBHOOK_ENV)
			fmt.Printf("GLIP_GUID_ENV [%v]\n", webhookURLOrUID)
		}
		adapter, err := ccglip.NewGlipAdapter(webhookURLOrUID)
		if err != nil {
			return errs.Wrap(err, "Incorrect Webhook GUID or URL")
		}
		sender.Adapter = adapter
	} else if adapterType == "slack" {
		if len(webhookURLOrUID) < 1 {
			webhookURLOrUID = os.Getenv(SLACK_WEBHOOK_ENV)
			fmt.Printf("SLACK_GUID_ENV [%v]\n", webhookURLOrUID)
		}
		adapter, err := ccslack.NewSlackAdapter(webhookURLOrUID)
		if err != nil {
			return errs.Wrap(err, "Incorrect Webhook GUID or URL")
		}
		sender.Adapter = adapter
	} else {
		return errors.New("Invalid Adapter")
	}

	exampleData, err := util.NewExampleData()
	if err != nil {
		return errs.Wrap(err, fmt.Sprintf("Invalid Example Data: %v\n", err))
	}
	fmtutil.PrintJSON(exampleData)

	switch example {
	case "aha":
		source := exampleData.Data[aha.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(aha.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "appsignal":
		source := exampleData.Data[appsignal.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(appsignal.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "apteligent":
		source := exampleData.Data[apteligent.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(apteligent.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "circleci":
		sender.SendCcMessage(circleci.ExampleMessage(cfg, exampleData))
	case "codeship":
		sender.SendCcMessage(codeship.ExampleMessage(cfg, exampleData))
	case "confluence":
		source := exampleData.Data[confluence.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(confluence.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "datadog":
		sender.SendCcMessage(datadog.ExampleMessage(cfg, exampleData))
	case "deskdotcom":
		source := exampleData.Data[deskdotcom.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(deskdotcom.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "enchant":
		sender.SendCcMessage(enchant.ExampleMessage(cfg, exampleData))
	case "gosquared":
		source := exampleData.Data[gosquared.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(gosquared.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "gosquared2":
		source := exampleData.Data[gosquared.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(gosquared2.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "heroku":
		sender.SendCcMessage(heroku.ExampleMessage(cfg, exampleData))
	case "librato":
		source := exampleData.Data[librato.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(librato.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "magnumci":
		sender.SendCcMessage(magnumci.ExampleMessage(cfg, exampleData))
	case "marketo":
		source := exampleData.Data[marketo.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(marketo.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "opsgenie":
		source := exampleData.Data[opsgenie.HandlerKey]
		for i, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(opsgenie.ExampleMessage(cfg, exampleData, eventSlug))
			if i == 8 {
				time.Sleep(2000 * time.Millisecond)
			}
		}
	case "papertrail":
		source := exampleData.Data[papertrail.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(papertrail.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "pingdom":
		source := exampleData.Data[pingdom.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(pingdom.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "raygun":
		sender.SendCcMessage(raygun.ExampleMessage(cfg, exampleData))
	case "runscope":
		sender.SendCcMessage(runscope.ExampleMessage(cfg, exampleData))
	case "semaphore":
		source := exampleData.Data[semaphore.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(semaphore.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "statuspage":
		source := exampleData.Data[statuspage.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(statuspage.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "travisci":
		sender.SendCcMessage(travisci.ExampleMessage(cfg, exampleData))
	case "userlike":
		source := exampleData.Data[userlike.HandlerKey]
		for _, eventSlug := range source.EventSlugs {
			sender.SendCcMessage(userlike.ExampleMessage(cfg, exampleData, eventSlug))
		}
	case "victorops":
		sender.SendCcMessage(victorops.ExampleMessage(cfg, exampleData))
	default:
		return errors.New(fmt.Sprintf("Unknown webhook source %v\n", example))
	}
	return nil
}
