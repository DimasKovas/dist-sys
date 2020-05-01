package auth

import (
	"context"
	"errors"
	"os"

	pbauth "common/proto"

	"google.golang.org/grpc"
)

type AuthClient struct {
	client     pbauth.AuthRpcClient
	connection *grpc.ClientConn
}

func CreateAuthClient() (AuthClient, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	address := os.Getenv("AUTH_RPC_ADDRESS")

	if len(address) == 0 {
		return AuthClient{}, errors.New("Auth server address is not specified")
	}

	conn, err := grpc.Dial(address, opts...)

	if err != nil {
		return AuthClient{}, err
	}

	client := AuthClient{}
	client.connection = conn
	client.client = pbauth.NewAuthRpcClient(conn)

	return client, nil
}

type UserPermissions struct {
	Username    string
	Permissions []string
}

func (c *AuthClient) Validate(token string) (UserPermissions, error) {
	request := &pbauth.ValidateRequest{
		AccessToken: token,
	}

	response, err := c.client.Validate(context.Background(), request)

	if err != nil {
		return UserPermissions{}, err
	}

	result := UserPermissions{}
	result.Username = response.Username
	result.Permissions = response.Permissions
	return result, nil
}

var ErrForbidden = errors.New("Not enough permissions")

func (c *AuthClient) CheckPermission(token string, permission string) error {
	perms, err := c.Validate(token)
	if err != nil {
		return err
	}
	for _, p := range perms.Permissions {
		if p == permission {
			return nil
		}
	}
	return ErrForbidden
}
