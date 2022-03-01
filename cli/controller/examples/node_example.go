package examples

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/spf13/cobra"
)

func NewUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "user",
		Hidden: true,
	}
	cmd.AddCommand(NewUserAddCommand())
	return cmd
}

type AddUserRequest struct {
	Username string `json:"username"`
}

func NewUserAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "add",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			node := cmd.Flag("node").Value.String()
			n, err := findNode(node)
			if err != nil {
				return err
			}
			i, err := nodeInvoker(n)
			if err != nil {
				return err
			}
			return i.Do("node/addUser", &AddUserRequest{
				Username: args[0],
			}, nil)
		},
	}
	cmd.Flags().StringP("node", "", domain.DefaultNodeName, "")
	return cmd
}

func NewProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "project",
		Hidden: true,
	}
	cmd.AddCommand(NewProjectAddCommand())
	return cmd
}

type AddProjectRequest struct {
	Name string `json:"name"`
	Repo string `json:"repo"`
}

func NewProjectAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "add",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			node := cmd.Flag("node").Value.String()
			name := cmd.Flag("name").Value.String()
			repo := cmd.Flag("repo").Value.String()
			if node == "" || name == "" {
				return fmt.Errorf("must specify node and project name")
			}
			n, err := findNode(node)
			if err != nil {
				return err
			}
			i, err := nodeInvoker(n)
			if err != nil {
				return err
			}
			return i.Do("node/addProject", &AddProjectRequest{
				Name: name,
				Repo: repo,
			}, nil)
		},
	}
	cmd.Flags().StringP("node", "", domain.DefaultNodeName, "")
	cmd.Flags().StringP("name", "", "", "the name of the project")
	cmd.Flags().StringP("repo", "", "", "the project's github repo")
	return cmd
}

func nodeInvoker(node *domain.Node) (*invoke.HTTPClient, error) {
	t, err := controller.AuthToken(node)
	if err != nil {
		return nil, err
	}
	return invoke.Node(node.Endpoints.Rest, t, ui.NodeLogsSink), nil
}
