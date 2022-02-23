package examples

import (
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/cli/log"
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
			return i.Do("auth/addUser", &AddUserRequest{
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
	Repo string `json:"repo"`
}

func NewProjectAddCommand() *cobra.Command {
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
			return i.Do("auth/addProject", &AddProjectRequest{
				Repo: args[0],
			}, nil)
		},
	}
	cmd.Flags().StringP("node", "", domain.DefaultNodeName, "")
	return cmd
}

func nodeInvoker(node *domain.Node) (*invoke.HTTPClient, error) {
	token, err := node.AuthToken()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return invoke.Node(node.Endpoints.Rest, token, ui.NodeLogsSink), nil
}
