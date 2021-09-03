package ws

import (
	"fmt"

	"github.com/mantil-io/mantil.go"
)

type store struct {
	subjects *mantil.KV
	subs     *mantil.KV
	requests *mantil.KV
}

func newStore() (*store, error) {
	subjects, err := mantil.NewKV("subjects")
	if err != nil {
		return nil, err
	}
	subs, err := mantil.NewKV("subs")
	if err != nil {
		return nil, err
	}
	requests, err := mantil.NewKV("requests")
	if err != nil {
		return nil, err
	}
	return &store{
		subjects: subjects,
		subs:     subs,
		requests: requests,
	}, nil
}

type client struct {
	ConnectionID string
	Domain       string
	Stage        string
}

type subscription struct {
	Client  *client
	Subject string
}

func (s *subscription) subjectsKey() string {
	return fmt.Sprintf("%s_%s", s.Subject, s.Client.ConnectionID)
}

func (s *subscription) subsKey() string {
	return fmt.Sprintf("%s_%s", s.Client.ConnectionID, s.Subject)
}

type request struct {
	Client *client
	Inbox  string
}

func (r *request) requestsKey() string {
	return fmt.Sprintf("%s_%s", r.Client.ConnectionID, r.Inbox)
}

func (s *store) addSubscription(client *client, subject string) error {
	sub := &subscription{
		Client:  client,
		Subject: subject,
	}
	if err := s.subjects.Put(sub.subjectsKey(), sub); err != nil {
		return err
	}
	if err := s.subs.Put(sub.subsKey(), sub); err != nil {
		return err
	}
	return nil
}

func (s *store) removeSubscription(connectionID, subject string) error {
	sub := &subscription{
		Client: &client{
			ConnectionID: connectionID,
		},
		Subject: subject,
	}
	if err := s.subjects.Delete(sub.subjectsKey()); err != nil {
		return err
	}
	if err := s.subs.Delete(sub.subsKey()); err != nil {
		return err
	}
	return nil
}

func (s *store) removeConnection(connectionID string) error {
	var subs []*subscription
	if _, err := s.subs.Find(&subs, mantil.FindBeginsWith, connectionID); err != nil {
		return err
	}
	for _, sub := range subs {
		if err := s.removeSubscription(sub.Client.ConnectionID, sub.Subject); err != nil {
			return err
		}
	}
	var requests []*request
	if _, err := s.requests.Find(&requests, mantil.FindBeginsWith, connectionID); err != nil {
		return err
	}
	for _, req := range requests {
		if err := s.removeRequest(req); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) findSubsForSubject(subject string) ([]subscription, error) {
	var subs []subscription
	if _, err := s.subjects.Find(&subs, mantil.FindBeginsWith, subject); err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *store) addRequest(client *client, inbox string) error {
	r := &request{
		Client: client,
		Inbox:  inbox,
	}
	return s.requests.Put(r.requestsKey(), r)
}

func (s *store) findRequest(connectionID, inbox string) (*request, error) {
	r := &request{
		Client: &client{
			ConnectionID: connectionID,
		},
		Inbox: inbox,
	}
	return r, s.requests.Get(r.requestsKey(), r)
}

func (s *store) removeRequest(r *request) error {
	return s.requests.Delete(r.requestsKey())
}
