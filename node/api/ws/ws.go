package ws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil.go/proto"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/domain"
)

type Handler struct {
	store   *store
	aws     *aws.AWS
	apiToFn map[string]string
}

func NewHandler() (*Handler, error) {
	store, err := newStore()
	if err != nil {
		return nil, err
	}
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	configEnc := os.Getenv(domain.EnvMantilConfig)
	buf, err := base64.StdEncoding.DecodeString(configEnc)
	if err != nil {
		return nil, err
	}
	c := domain.WsConfig{}
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, err
	}
	return &Handler{
		store:   store,
		aws:     aws,
		apiToFn: c.ApiToFn,
	}, nil
}

func (h *Handler) HandleApiGatewayRequest(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	client := &client{
		ConnectionID: req.RequestContext.ConnectionID,
		Domain:       req.RequestContext.DomainName,
		Stage:        req.RequestContext.Stage,
	}
	eventType := req.RequestContext.EventType
	payload := req.Body

	switch eventType {
	case "CONNECT":
		// no-op
	case "DISCONNECT":
		if err := h.disconnect(client.ConnectionID); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
	case "MESSAGE":
		if err := h.clientMessage(client, []byte(payload)); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("unknown event type")
	}

	rsp := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}
	return rsp, nil
}

func (h *Handler) disconnect(connectionID string) error {
	// remove all subscriptions and pending requests for this connection
	return h.store.removeConnection(connectionID)
}

func (h *Handler) clientMessage(client *client, payload []byte) error {
	// parse message and handle according to message type
	m, err := proto.ParseMessage(payload)
	if err != nil {
		return err
	}
	switch m.Type {
	case proto.Subscribe:
		return h.clientSubscribe(client, m.Subjects)
	case proto.Unsubscribe:
		return h.clientUnsubscribe(client.ConnectionID, m.Subjects)
	case proto.Request:
		return h.clientRequest(client, m)
	}
	return fmt.Errorf("unsupported message type")
}

func (h *Handler) clientSubscribe(client *client, subjects []string) error {
	for _, s := range subjects {
		if err := h.store.addSubscription(client, s); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) clientUnsubscribe(connectionID string, subjects []string) error {
	for _, s := range subjects {
		if err := h.store.removeSubscription(connectionID, s); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) clientRequest(client *client, m *proto.Message) error {
	if err := h.store.addRequest(client, m.Inbox); err != nil {
		return err
	}
	m.ConnectionID = client.ConnectionID
	uriParts := strings.Split(m.URI, ".")
	if len(uriParts) < 1 {
		return fmt.Errorf("function not provided in message URI")
	}
	function := uriParts[0]
	functionName := h.apiToFn[function]
	if functionName == "" {
		return fmt.Errorf("match not found for function %s in mappings %v", function, h.apiToFn)
	}
	invoker, err := mantil.NewLambdaInvoker(functionName, "")
	if err != nil {
		return err
	}
	payload, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("error marshalling proto - %v", err)
		return err
	}
	if err := invoker.CallAsync(payload); err != nil {
		return err
	}
	return nil
}

func (h *Handler) HandleBackendMessage(m proto.Message) error {
	switch m.Type {
	case proto.Response:
		return h.handleResponse(m)
	case proto.Publish:
		return h.handlePublish(m)
	}
	return fmt.Errorf("unsupported message type")
}

func (h *Handler) handleResponse(m proto.Message) error {
	r, err := h.store.findRequest(m.ConnectionID, m.Inbox)
	if err != nil {
		return err
	}
	m.ConnectionID = ""
	mp, err := m.Encode()
	if err != nil {
		return err
	}
	if err := h.aws.PublishToAPIGatewayConnection(
		r.Client.Domain,
		r.Client.Stage,
		r.Client.ConnectionID,
		mp,
	); err != nil {
		return err
	}
	return h.store.removeRequest(r)
}

func (h *Handler) handlePublish(m proto.Message) error {
	subs, err := h.store.findSubsForSubject(m.Subject)
	if err != nil {
		return err
	}
	mp, err := m.Encode()
	if err != nil {
		return err
	}
	for _, s := range subs {
		if err := h.aws.PublishToAPIGatewayConnection(
			s.Client.Domain,
			s.Client.Stage,
			s.Client.ConnectionID,
			mp,
		); err != nil {
			return err
		}
	}
	return nil
}
