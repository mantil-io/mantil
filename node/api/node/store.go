package node

import "github.com/mantil-io/mantil.go"

const (
	usersPartition    = "users"
	projectsPartition = "projects"
)

type Store struct {
	users    *mantil.KV
	projects *mantil.KV
}

func NewStore() (*Store, error) {
	users, err := mantil.NewKV(usersPartition)
	if err != nil {
		return nil, err
	}
	projects, err := mantil.NewKV(projectsPartition)
	if err != nil {
		return nil, err
	}
	return &Store{
		users:    users,
		projects: projects,
	}, nil
}

type user struct {
	Name string
}

func (s *Store) StoreUser(name string) error {
	return s.users.Put(name, &user{
		Name: name,
	})
}

func (s *Store) FindUser(name string) (*user, error) {
	u := &user{}
	if err := s.users.Get(name, u); err != nil {
		return nil, err
	}
	return u, nil
}

type project struct {
	Repo string
}

func (s *Store) StoreProject(repo string) error {
	return s.projects.Put(repo, &project{
		Repo: repo,
	})
}

func (s *Store) FindProjects() ([]*project, error) {
	projects := []*project{}
	_, err := s.projects.FindAll(&projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}
