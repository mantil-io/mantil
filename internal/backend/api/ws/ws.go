package ws

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mantil-io/mantil.go/pkg/proto"
	"github.com/mantil-io/mantil/internal/aws"
)

type Handler struct {
	store *store
	aws   *aws.AWS
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
	return &Handler{
		store: store,
		aws:   aws,
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
	}
	return nil
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

func (h *Handler) HandleSQSEvent(ctx context.Context, req events.SQSEvent) error {
	for _, m := range req.Records {
		if err := h.handleSQSMessage(m); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) handleSQSMessage(sm events.SQSMessage) error {
	m, err := proto.ParseMessage([]byte(sm.Body))
	if err != nil {
		return err
	}
	if m.Type != proto.Publish {
		return nil
	}
	subs, err := h.store.findSubsForSubject(m.Subject)
	if err != nil {
		return err
	}
	for _, s := range subs {
		if err := h.aws.PublishToAPIGatewayConnection(
			s.Client.Domain,
			s.Client.Stage,
			s.Client.ConnectionID,
			[]byte(sm.Body),
		); err != nil {
			return err
		}
	}
	return nil
}
