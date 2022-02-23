package auth

import "github.com/mantil-io/mantil.go"

const (
	usersPartition    = "users"
	projectsPartition = "projects"
)

func (a *Auth) initStore() error {
	users, err := mantil.NewKV(usersPartition)
	if err != nil {
		return err
	}
	a.users = users
	projects, err := mantil.NewKV(projectsPartition)
	if err != nil {
		return err
	}
	a.projects = projects
	return nil
}

type user struct {
	Name string
}

func (a *Auth) storeUser(name string) error {
	return a.users.Put(name, &user{
		Name: name,
	})
}

func (a *Auth) findUser(name string) (*user, error) {
	u := &user{}
	if err := a.users.Get(name, u); err != nil {
		return nil, err
	}
	return u, nil
}

type project struct {
	Repo string
}

func (a *Auth) storeProject(repo string) error {
	return a.projects.Put(repo, &project{
		Repo: repo,
	})
}

func (a *Auth) findProjects() ([]*project, error) {
	projects := []*project{}
	_, err := a.projects.FindAll(&projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}
