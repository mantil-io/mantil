package examples

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/spf13/cobra"
)

func NewNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "node",
		Hidden: true,
	}
	cmd.AddCommand(NewUserAddCommand())
	cmd.AddCommand(NewLoginCommand())
	cmd.AddCommand(NewLogoutCommand())
	return cmd
}

func NewUserAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "user-add",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			node := cmd.Flag("node").Value.String()
			user := cmd.Flag("github-username").Value.String()
			role := cmd.Flag("role").Value.String()
			n, err := findNode(node)
			if err != nil {
				return err
			}
			i, err := nodeInvoker(n)
			if err != nil {
				return err
			}
			resolveRole := func(r string) (domain.Role, error) {
				switch r {
				case "admin":
					return domain.Owner, nil
				case "user":
					return domain.Member, nil
				default:
					return -1, fmt.Errorf("unknown role")
				}
			}
			r, err := resolveRole(role)
			if err != nil {
				return err
			}
			return i.Do("node/addUser", &dto.AddUserRequest{
				Username: user,
				Role:     r,
			}, nil)
		},
	}
	cmd.Flags().StringP("node", "", domain.DefaultNodeName, "")
	cmd.Flags().StringP("github-username", "", "", "")
	cmd.Flags().StringP("role", "", "user", "")
	return cmd
}

func NewLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "login",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeURL := args[0]
			i := invoke.Node(nodeURL, "", ui.NodeLogsSink)
			var rsp dto.LoginResponse
			if err := i.Do("auth/login", nil, &rsp); err != nil {
				return err
			}
			fs, err := domain.NewSingleDeveloperWorkspaceStore()
			if err != nil {
				return err
			}
			w := fs.Workspace()
			w.AddNode(rsp.Node)
			return fs.Store()
			// buf, _ := json.Marshal(rsp)
			// fmt.Println(string(buf))
			// return nil
		},
	}
	return cmd
}

func NewLogoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "logout",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fs, err := domain.NewSingleDeveloperWorkspaceStore()
			if err != nil {
				return err
			}
			w := fs.Workspace()
			if len(w.Nodes) == 0 {
				return fmt.Errorf("no nodes avaiable")
			}
			n := w.FindNode(args[0])
			if n == nil {
				return fmt.Errorf("node not found")
			}
			n.JWT = ""
			return fs.Store()
		},
	}
	cmd.Flags().StringP("node", "", domain.DefaultNodeName, "")
	return cmd
}

func nodeInvoker(node *domain.Node) (*invoke.HTTPClient, error) {
	t, err := controller.AuthToken(node)
	if err != nil {
		return nil, err
	}
	return invoke.Node(node.Endpoints.Rest, t, ui.NodeLogsSink), nil
}
