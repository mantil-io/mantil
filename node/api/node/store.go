package node

import (
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/domain"
)

const (
	usersPartition = "users"
)

type Store struct {
	users  *mantil.KV
	config *mantil.KV
}

func NewStore() (*Store, error) {
	users, err := mantil.NewKV(usersPartition)
	if err != nil {
		return nil, err
	}
	config, err := mantil.NewKV(domain.NodeConfigKey)
	if err != nil {
		return nil, err
	}
	return &Store{
		users:  users,
		config: config,
	}, nil
}

type user struct {
	Name string
	Role domain.Role
}

func (s *Store) StoreUser(name string, role domain.Role) error {
	return s.users.Put(name, &user{
		Name: name,
		Role: role,
	})
}

func (s *Store) RemoveUser(name string) error {
	return s.users.Delete(name)
}

func (s *Store) FindUser(name string) (*user, error) {
	u := &user{}
	if err := s.users.Get(name, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) FindConfig() (*domain.Node, error) {
	n := &domain.Node{}
	if err := s.config.Get(domain.NodeConfigKey, n); err != nil {
		return nil, err
	}
	return n, nil
}
