package services

import (
	"context"
	"fmt"

	pb "kisanlink-ecom/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	userServiceClient           pb.UserServiceClient
	roleServiceClient           pb.RoleServiceClient
	permissionServiceClient     pb.PermissionServiceClient
	connectRolePermissionClient pb.ConnectRolePermissionServiceClient

	// V2 clients
	userServiceV2Client       pb.UserServiceV2Client
	roleServiceV2Client       pb.RoleServiceV2Client
	permissionServiceV2Client pb.PermissionServiceV2Client
	conn                      *grpc.ClientConn
}

func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	// Initialize v1 clients
	userServiceClient := pb.NewUserServiceClient(conn)
	roleServiceClient := pb.NewRoleServiceClient(conn)
	permissionServiceClient := pb.NewPermissionServiceClient(conn)
	connectRolePermissionClient := pb.NewConnectRolePermissionServiceClient(conn)

	// Initialize v2 clients
	userServiceV2Client := pb.NewUserServiceV2Client(conn)
	roleServiceV2Client := pb.NewRoleServiceV2Client(conn)
	permissionServiceV2Client := pb.NewPermissionServiceV2Client(conn)

	return &GRPCClient{
		conn:                        conn,
		userServiceClient:           userServiceClient,
		roleServiceClient:           roleServiceClient,
		permissionServiceClient:     permissionServiceClient,
		connectRolePermissionClient: connectRolePermissionClient,
		userServiceV2Client:         userServiceV2Client,
		roleServiceV2Client:         roleServiceV2Client,
		permissionServiceV2Client:   permissionServiceV2Client,
	}, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// V1 Methods (keeping for backward compatibility)

func (c *GRPCClient) CreateUser(ctx context.Context, username, password string, roleIds []string) (*pb.CreateUserResponse, error) {
	req := &pb.CreateUserRequest{
		Username:    username,
		Password:    password,
		UserRoleIds: roleIds,
	}
	return c.userServiceClient.CreateUser(ctx, req)
}

func (c *GRPCClient) GetUserByID(ctx context.Context, userID string) (*pb.GetUserByIdResponse, error) {
	req := &pb.GetUserByIdRequest{Id: userID}
	return c.userServiceClient.GetUserById(ctx, req)
}

func (c *GRPCClient) GetAllUsers(ctx context.Context) (*pb.GetUserResponse, error) {
	req := &pb.GetUserRequest{}
	return c.userServiceClient.GetUser(ctx, req)
}

func (c *GRPCClient) UpdateUser(ctx context.Context, userID, username string, isValidated bool) (*pb.UpdateUserResponse, error) {
	req := &pb.UpdateUserRequest{
		Id:          userID,
		Username:    username,
		IsValidated: isValidated,
	}
	return c.userServiceClient.UpdateUser(ctx, req)
}

func (c *GRPCClient) DeleteUser(ctx context.Context, userID string) (*pb.DeleteUserResponse, error) {
	req := &pb.DeleteUserRequest{Id: userID}
	return c.userServiceClient.DeleteUser(ctx, req)
}

func (c *GRPCClient) LoginUser(ctx context.Context, username, password string) (*pb.LoginResponse, error) {
	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}
	return c.userServiceClient.Login(ctx, req)
}

// V1 Role Methods (for backward compatibility)
func (c *GRPCClient) CreateRole(ctx context.Context, name, description string) (*pb.CreateRoleResponse, error) {
	req := &pb.CreateRoleRequest{
		Name:        name,
		Description: description,
	}
	return c.roleServiceClient.CreateRole(ctx, req)
}

func (c *GRPCClient) GetAllRoles(ctx context.Context) (*pb.GetAllRolesResponse, error) {
	req := &pb.GetAllRolesRequest{}
	return c.roleServiceClient.GetAllRoles(ctx, req)
}

func (c *GRPCClient) GetRoleByID(ctx context.Context, roleID string) (*pb.GetRoleByIdResponse, error) {
	req := &pb.GetRoleByIdRequest{Id: roleID}
	return c.roleServiceClient.GetRoleById(ctx, req)
}

func (c *GRPCClient) UpdateRole(ctx context.Context, role *pb.Role) (*pb.UpdateRoleResponse, error) {
	req := &pb.UpdateRoleRequest{Role: role}
	return c.roleServiceClient.UpdateRole(ctx, req)
}

func (c *GRPCClient) DeleteRole(ctx context.Context, roleID string) (*pb.DeleteRoleResponse, error) {
	req := &pb.DeleteRoleRequest{Id: roleID}
	return c.roleServiceClient.DeleteRole(ctx, req)
}

// V1 Permission Methods (for backward compatibility)
func (c *GRPCClient) CreatePermission(ctx context.Context, name, description string) (*pb.CreatePermissionResponse, error) {
	req := &pb.CreatePermissionRequest{
		Name:        name,
		Description: description,
	}
	return c.permissionServiceClient.CreatePermission(ctx, req)
}

func (c *GRPCClient) GetAllPermissions(ctx context.Context) (*pb.GetAllPermissionsResponse, error) {
	req := &pb.GetAllPermissionsRequest{}
	return c.permissionServiceClient.GetAllPermissions(ctx, req)
}

func (c *GRPCClient) GetPermissionByID(ctx context.Context, permissionID string) (*pb.GetPermissionByIdResponse, error) {
	req := &pb.GetPermissionByIdRequest{Id: permissionID}
	return c.permissionServiceClient.GetPermissionById(ctx, req)
}

func (c *GRPCClient) UpdatePermission(ctx context.Context, permission *pb.Permission) (*pb.UpdatePermissionResponse, error) {
	req := &pb.UpdatePermissionRequest{Permission: permission}
	return c.permissionServiceClient.UpdatePermission(ctx, req)
}

func (c *GRPCClient) DeletePermission(ctx context.Context, permissionID string) (*pb.DeletePermissionResponse, error) {
	req := &pb.DeletePermissionRequest{Id: permissionID}
	return c.permissionServiceClient.DeletePermission(ctx, req)
}

// V1 Role-Permission Connection Methods (for backward compatibility)
func (c *GRPCClient) CreateRolePermissionConnection(ctx context.Context, roleIDs, permissionIDs []string) (*pb.CreateConnRolePermissionResponse, error) {
	req := &pb.CreateConnRolePermissionRequest{
		RoleIds:       roleIDs,
		PermissionIds: permissionIDs,
	}
	return c.connectRolePermissionClient.CreateConnectRolePermission(ctx, req)
}

func (c *GRPCClient) GetAllRolePermissions(ctx context.Context) (*pb.GetConnRolePermissionallResponse, error) {
	req := &pb.GetConnRolePermissionallRequest{}
	return c.connectRolePermissionClient.GetAllRolePermission(ctx, req)
}

func (c *GRPCClient) GetRolePermissionByID(ctx context.Context, connectionID string) (*pb.GetConnRolePermissionByIdResponse, error) {
	req := &pb.GetConnRolePermissionByIdRequest{Id: connectionID}
	return c.connectRolePermissionClient.GetRolePermissionById(ctx, req)
}

func (c *GRPCClient) UpdateRolePermissionConnection(ctx context.Context, connectionID string, roleIDs, permissionIDs []string) (*pb.UpdateConnRolePermissionResponse, error) {
	req := &pb.UpdateConnRolePermissionRequest{
		Id:            connectionID,
		RoleIds:       roleIDs,
		PermissionIds: permissionIDs,
	}
	return c.connectRolePermissionClient.UpdateRolePermission(ctx, req)
}

func (c *GRPCClient) DeleteRolePermissionConnection(ctx context.Context, connectionID string) (*pb.DeleteConnRolePermissionResponse, error) {
	req := &pb.DeleteConnRolePermissionRequest{Id: connectionID}
	return c.connectRolePermissionClient.DeleteRolePermission(ctx, req)
}

// V2 Methods (enhanced functionality)

func (c *GRPCClient) CreateUserV2(ctx context.Context, username, email, fullName, password string, roleIds []string) (*pb.RegisterResponseV2, error) {
	req := &pb.RegisterRequestV2{
		Username: username,
		Email:    email,
		FullName: fullName,
		Password: password,
		RoleIds:  roleIds,
	}
	return c.userServiceV2Client.Register(ctx, req)
}

func (c *GRPCClient) GetUserByIDV2(ctx context.Context, userID string, includeRoles, includePermissions bool) (*pb.GetUserResponseV2, error) {
	req := &pb.GetUserRequestV2{
		Id:                 userID,
		IncludeRoles:       includeRoles,
		IncludePermissions: includePermissions,
	}
	return c.userServiceV2Client.GetUser(ctx, req)
}

func (c *GRPCClient) GetAllUsersV2(ctx context.Context, page, perPage int32, search, status string, roleIds []string) (*pb.GetAllUsersResponseV2, error) {
	req := &pb.GetAllUsersRequestV2{
		Page:    page,
		PerPage: perPage,
		Search:  search,
		Status:  status,
		RoleIds: roleIds,
	}
	return c.userServiceV2Client.GetAllUsers(ctx, req)
}

func (c *GRPCClient) UpdateUserV2(ctx context.Context, userID, username, email, fullName, status string, isValidated bool, roleIds []string) (*pb.UpdateUserResponseV2, error) {
	req := &pb.UpdateUserRequestV2{
		Id:          userID,
		Username:    username,
		Email:       email,
		FullName:    fullName,
		Status:      status,
		IsValidated: isValidated,
		RoleIds:     roleIds,
	}
	return c.userServiceV2Client.UpdateUser(ctx, req)
}

func (c *GRPCClient) DeleteUserV2(ctx context.Context, userID string) (*pb.DeleteUserResponseV2, error) {
	req := &pb.DeleteUserRequestV2{Id: userID}
	return c.userServiceV2Client.DeleteUser(ctx, req)
}

func (c *GRPCClient) LoginUserV2(ctx context.Context, username, password, mfaCode string) (*pb.LoginResponseV2, error) {
	req := &pb.LoginRequestV2{
		Username: username,
		Password: password,
		MfaCode:  mfaCode,
	}
	return c.userServiceV2Client.Login(ctx, req)
}

func (c *GRPCClient) RefreshTokenV2(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponseV2, error) {
	req := &pb.RefreshTokenRequestV2{RefreshToken: refreshToken}
	return c.userServiceV2Client.RefreshToken(ctx, req)
}

func (c *GRPCClient) LogoutV2(ctx context.Context, accessToken string) (*pb.LogoutResponseV2, error) {
	req := &pb.LogoutRequestV2{AccessToken: accessToken}
	return c.userServiceV2Client.Logout(ctx, req)
}

// Role V2 Methods

func (c *GRPCClient) CreateRoleV2(ctx context.Context, name, description, parentRoleID string, permissionIds []string) (*pb.CreateRoleResponseV2, error) {
	req := &pb.CreateRoleRequestV2{
		Name:          name,
		Description:   description,
		ParentRoleId:  parentRoleID,
		PermissionIds: permissionIds,
	}
	return c.roleServiceV2Client.CreateRole(ctx, req)
}

func (c *GRPCClient) GetRoleV2(ctx context.Context, roleID string, includePermissions, includeChildRoles bool) (*pb.GetRoleResponseV2, error) {
	req := &pb.GetRoleRequestV2{
		Id:                 roleID,
		IncludePermissions: includePermissions,
		IncludeChildRoles:  includeChildRoles,
	}
	return c.roleServiceV2Client.GetRole(ctx, req)
}

func (c *GRPCClient) GetAllRolesV2(ctx context.Context, page, perPage int32, search, status, parentRoleID string, includePermissions bool) (*pb.GetAllRolesResponseV2, error) {
	req := &pb.GetAllRolesRequestV2{
		Page:               page,
		PerPage:            perPage,
		Search:             search,
		Status:             status,
		ParentRoleId:       parentRoleID,
		IncludePermissions: includePermissions,
	}
	return c.roleServiceV2Client.GetAllRoles(ctx, req)
}

func (c *GRPCClient) UpdateRoleV2(ctx context.Context, roleID, name, description, status, parentRoleID string, permissionIds []string) (*pb.UpdateRoleResponseV2, error) {
	req := &pb.UpdateRoleRequestV2{
		Id:            roleID,
		Name:          name,
		Description:   description,
		Status:        status,
		ParentRoleId:  parentRoleID,
		PermissionIds: permissionIds,
	}
	return c.roleServiceV2Client.UpdateRole(ctx, req)
}

func (c *GRPCClient) DeleteRoleV2(ctx context.Context, roleID string, cascade bool) (*pb.DeleteRoleResponseV2, error) {
	req := &pb.DeleteRoleRequestV2{
		Id:      roleID,
		Cascade: cascade,
	}
	return c.roleServiceV2Client.DeleteRole(ctx, req)
}

func (c *GRPCClient) AssignPermissionToRoleV2(ctx context.Context, roleID, permissionID string) (*pb.AssignPermissionToRoleResponseV2, error) {
	req := &pb.AssignPermissionToRoleRequestV2{
		RoleId:       roleID,
		PermissionId: permissionID,
	}
	return c.roleServiceV2Client.AssignPermissionToRole(ctx, req)
}

func (c *GRPCClient) RemovePermissionFromRoleV2(ctx context.Context, roleID, permissionID string) (*pb.RemovePermissionFromRoleResponseV2, error) {
	req := &pb.RemovePermissionFromRoleRequestV2{
		RoleId:       roleID,
		PermissionId: permissionID,
	}
	return c.roleServiceV2Client.RemovePermissionFromRole(ctx, req)
}

func (c *GRPCClient) GetRolePermissionsV2(ctx context.Context, roleID string) (*pb.GetRolePermissionsResponseV2, error) {
	req := &pb.GetRolePermissionsRequestV2{RoleId: roleID}
	return c.roleServiceV2Client.GetRolePermissions(ctx, req)
}

// Permission V2 Methods

func (c *GRPCClient) CreatePermissionV2(ctx context.Context, name, description, resource, effect string, actions []string) (*pb.CreatePermissionResponseV2, error) {
	req := &pb.CreatePermissionRequestV2{
		Name:        name,
		Description: description,
		Resource:    resource,
		Effect:      effect,
		Actions:     actions,
	}
	return c.permissionServiceV2Client.CreatePermission(ctx, req)
}

func (c *GRPCClient) GetPermissionV2(ctx context.Context, permissionID string) (*pb.GetPermissionResponseV2, error) {
	req := &pb.GetPermissionRequestV2{Id: permissionID}
	return c.permissionServiceV2Client.GetPermission(ctx, req)
}

func (c *GRPCClient) GetAllPermissionsV2(ctx context.Context, page, perPage int32, search, resource, effect string) (*pb.GetAllPermissionsResponseV2, error) {
	req := &pb.GetAllPermissionsRequestV2{
		Page:     page,
		PerPage:  perPage,
		Search:   search,
		Resource: resource,
		Effect:   effect,
	}
	return c.permissionServiceV2Client.GetAllPermissions(ctx, req)
}

func (c *GRPCClient) UpdatePermissionV2(ctx context.Context, permissionID, name, description, resource, effect, status string, actions []string) (*pb.UpdatePermissionResponseV2, error) {
	req := &pb.UpdatePermissionRequestV2{
		Id:          permissionID,
		Name:        name,
		Description: description,
		Resource:    resource,
		Effect:      effect,
		Status:      status,
		Actions:     actions,
	}
	return c.permissionServiceV2Client.UpdatePermission(ctx, req)
}

func (c *GRPCClient) DeletePermissionV2(ctx context.Context, permissionID string) (*pb.DeletePermissionResponseV2, error) {
	req := &pb.DeletePermissionRequestV2{Id: permissionID}
	return c.permissionServiceV2Client.DeletePermission(ctx, req)
}

func (c *GRPCClient) EvaluatePermissionV2(ctx context.Context, userID, resource, action string) (*pb.EvaluatePermissionResponseV2, error) {
	req := &pb.EvaluatePermissionRequestV2{
		UserId:   userID,
		Resource: resource,
		Action:   action,
	}
	return c.permissionServiceV2Client.EvaluatePermission(ctx, req)
}
