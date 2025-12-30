package src

import (
	mongoDatabase "common/database/mongo"
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// MongoDB collection name.
	COLLECTION_USERS            = "users"
	COLLECTION_COMPANIES        = "companies"
	COLLECTION_ROLES            = "roles"
	COLLECTION_PERMISSIONS      = "permissions"
	COLLECTION_USER_ROLES       = "user_roles"
	COLLECTION_ROLE_PERMISSIONS = "role_permissions"
)

// AuthStore interface.
// This interface is used to define the auth store methods.
//
// Methods:
//   - RegisterUser(ctx context.Context, user *User) (string, error): This method is used to register a new user.
//   - GetUserByID(ctx context.Context, userID string) (*User, error): This method is used to get the user by ID.
//   - GetUserByIdentifier(ctx context.Context, identifier string) (*User, error): This method is used to get the user by identifier.
//   - UpdateUser(ctx context.Context, userID string, user *User) error: This method is used to update the user.
type AuthStore interface {
	// Company
	RegisterCompany(ctx context.Context, company *Company) (string, error)
	GetCompanyByID(ctx context.Context, companyID string) (*Company, error)
	GetAllCompany(ctx context.Context) ([]Company, error)
	UpdateCompany(ctx context.Context, companyID string, company *Company) error
	DeleteCompany(ctx context.Context, companyID string) error

	// User
	RegisterUser(ctx context.Context, user *User) (string, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*User, error)
	UpdateUser(ctx context.Context, userID string, user *User) error
	DeleteUser(ctx context.Context, userID string) error

	// Admin
	CheckAdminExistence(ctx context.Context, identifier string) (bool, error)

	// Role
	RegisterRole(ctx context.Context, role *Role) (string, error)
	GetRoleByID(ctx context.Context, roleID string) (*Role, error)
	GetRoleByIdentifier(ctx context.Context, identifier string) (*Role, error)
	GetAllRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, roleID string, role *Role) error
	DeleteRole(ctx context.Context, roleID string) error

	// Permission
	RegisterPermission(ctx context.Context, permission *Permission) (string, error)
	GetPermissionByIdentifier(ctx context.Context, permissionIdentifier string) (*Permission, error)
	GetAllPermissions(ctx context.Context) ([]Permission, error)
	UpdatePermission(ctx context.Context, permissionID string, permission *Permission) error
	DeletePermission(ctx context.Context, permissionID string) error

	// RolePermission
	RegisterRolePermission(ctx context.Context, roleID string, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]RolePermission, error)
	UpdateRolePermission(ctx context.Context, roleID string, permissionID string) error
	DeleteRolePermission(ctx context.Context, roleID string, permissionID string) error
	GetAllRolePermissions(ctx context.Context) ([]RolePermission, error)
	GetRolePermission(ctx context.Context, roleID string, permissionID string) (RolePermission, error)

	// UserRole
	RegisterUserRole(ctx context.Context, userID string, roleID string) error
	UpdateUserRole(ctx context.Context, userID string, roleID string) error
	DeleteUserRole(ctx context.Context, userID string, roleID string) error
	GetAllUserRoles(ctx context.Context) ([]UserRole, error)
	GetUserRole(ctx context.Context, userID string, roleID string) (UserRole, error)
	GetUserRolesByUserID(ctx context.Context, userID string) ([]UserRole, error)
	GetUserRolesByRoleID(ctx context.Context, roleID string) ([]UserRole, error)
}

// authStore struct.
// This struct is used to implement the AuthStore interface.
//
// Attributes:
//   - usersCollection (*mongo.Collection): The users collection.
type authStore struct {
	usersCollection           *mongo.Collection
	companiesCollection       *mongo.Collection
	rolesCollection           *mongo.Collection
	permissionsCollection     *mongo.Collection
	userRolesCollection       *mongo.Collection
	rolePermissionsCollection *mongo.Collection
}

// NewAuthStore function.
// This function is used to create a new auth store.
//
// Parameters:
//   - adapter (mongoDatabase.MongoDBAdapter): The MongoDB adapter.
//
// Returns:
//   - AuthStore: The auth store.
func NewAuthStore(adapter mongoDatabase.MongoDBAdapter) AuthStore {
	userCollection := adapter.Collection(COLLECTION_USERS)
	companyCollection := adapter.Collection(COLLECTION_COMPANIES)
	roleCollection := adapter.Collection(COLLECTION_ROLES)
	permissionsCollection := adapter.Collection(COLLECTION_PERMISSIONS)
	userRolesCollection := adapter.Collection(COLLECTION_USER_ROLES)
	rolePermissionsCollection := adapter.Collection(COLLECTION_ROLE_PERMISSIONS)

	usernameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	nameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, userError := userCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{usernameIndex, emailIndex})
	if userError != nil {
		log.Fatal(userError)
	}

	_, companyError := companyCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{nameIndex})
	if companyError != nil {
		log.Fatal(companyError)
	}

	_, roleError := roleCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{nameIndex})
	if roleError != nil {
		log.Fatal(roleError)
	}

	_, permissionsError := permissionsCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{nameIndex})
	if permissionsError != nil {
		log.Fatal(permissionsError)
	}

	return &authStore{
		usersCollection:           userCollection,
		companiesCollection:       companyCollection,
		rolesCollection:           roleCollection,
		permissionsCollection:     permissionsCollection,
		userRolesCollection:       userRolesCollection,
		rolePermissionsCollection: rolePermissionsCollection,
	}
}

// Company
func (s *authStore) RegisterCompany(ctx context.Context, company *Company) (string, error) {
	tracer := otel.Tracer("auth-service")

	ctx, span := tracer.Start(ctx, "authStore.RegisterCompany")
	defer span.End()

	res, err := s.companiesCollection.InsertOne(ctx, company)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("Company name already exists")
		}
		return "", err
	}

	insertedID := res.InsertedID.(primitive.ObjectID).Hex()

	span.SetStatus(otlpcodes.Ok, "Company Registered in DB Successfully!")
	return insertedID, nil
}

func (s *authStore) GetCompanyByID(ctx context.Context, companyID string) (*Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetCompanyByID")
	defer span.End()

	companyIDHex, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	var company Company
	err = s.companiesCollection.FindOne(ctx, map[string]interface{}{"_id": companyIDHex}).Decode(&company)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Company found in DB Successfully")
	return &company, nil
}

func (s *authStore) GetAllCompany(ctx context.Context) ([]Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetAllCompany")
	defer span.End()

	var companies []Company
	cursor, err := s.companiesCollection.Find(ctx, bson.D{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &companies)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Companies found in DB Successfully")
	return companies, nil
}
func (s *authStore) UpdateCompany(ctx context.Context, companyID string, company *Company) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.UpdateCompany")
	defer span.End()

	companyIDHex, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	companyBSON, err := bson.Marshal(company)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	err = bson.Unmarshal(companyBSON, company)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.companiesCollection.UpdateOne(ctx, map[string]interface{}{"_id": companyIDHex}, map[string]interface{}{"$set": company})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Company Updated in DB Successfully!")
	return nil
}

func (s *authStore) DeleteCompany(ctx context.Context, companyID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.DeleteCompany")
	defer span.End()

	companyIDHex, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.companiesCollection.DeleteOne(ctx, map[string]interface{}{"_id": companyIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Company Deleted from DB Successfully!")
	return nil
}

// User
func (s *authStore) RegisterUser(ctx context.Context, user *User) (string, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.RegisterUser")
	defer span.End()

	res, err := s.usersCollection.InsertOne(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("username or email already exists")
		}
		return "", err
	}

	insertedID := res.InsertedID.(primitive.ObjectID).Hex()

	span.SetStatus(otlpcodes.Ok, "User Registered in DB Successfully!")
	return insertedID, nil
}

func (s *authStore) GetUserByID(ctx context.Context, userID string) (*User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetUserByID")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	var user User
	err = s.usersCollection.FindOne(ctx, map[string]interface{}{"_id": userIDHex}).Decode(&user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Found in DB Successfully!")
	return &user, nil
}

func (s *authStore) GetUserByIdentifier(ctx context.Context, identifier string) (*User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetUserByIdentifier")
	defer span.End()

	var user User
	err := s.usersCollection.FindOne(ctx, map[string]interface{}{"$or": []map[string]interface{}{{"username": identifier}, {"email": identifier}}}).Decode(&user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Found in DB Successfully!")
	return &user, nil
}

func (s *authStore) UpdateUser(ctx context.Context, userID string, user *User) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.UpdateUser")
	defer span.End()

	userBSON, err := bson.Marshal(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	err = bson.Unmarshal(userBSON, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.usersCollection.UpdateOne(ctx, map[string]interface{}{"_id": userIDHex}, map[string]interface{}{"$set": user})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "User Updated in DB Successfully!")
	return nil
}

func (s *authStore) DeleteUser(ctx context.Context, userID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.DeleteUser")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.usersCollection.DeleteOne(ctx, map[string]interface{}{"_id": userIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "User Deleted from DB Successfully!")
	return nil
}

func (s *authStore) CheckAdminExistence(ctx context.Context, identifier string) (bool, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.CheckAdminExistence")
	defer span.End()

	count, err := s.usersCollection.CountDocuments(ctx, map[string]interface{}{"username": identifier})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return false, err
	}

	span.SetStatus(otlpcodes.Ok, "Admin Existence Checked Successfully!")
	return count > 0, nil
}

// Role
func (s *authStore) RegisterRole(ctx context.Context, role *Role) (string, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.RegisterRole")
	defer span.End()

	res, err := s.rolesCollection.InsertOne(ctx, role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("Role name already exists: ")
		}
		return "", err
	}

	insertedID := res.InsertedID.(primitive.ObjectID).Hex()

	span.SetStatus(otlpcodes.Ok, "Role Registered in DB Successfully!")
	return insertedID, nil
}

func (s *authStore) GetRoleByID(ctx context.Context, roleID string) (*Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetRoleByID")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	var role Role
	err = s.rolesCollection.FindOne(ctx, map[string]interface{}{"_id": roleIDHex}).Decode(&role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role found in DB Successfully")
	return &role, nil
}

func (s *authStore) GetRoleByIdentifier(ctx context.Context, identifier string) (*Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetRoleByIdentifier")
	defer span.End()

	var role Role
	err := s.rolesCollection.FindOne(ctx, map[string]interface{}{"identifier": identifier}).Decode(&role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role found in DB Successfully")
	return &role, nil
}

func (s *authStore) GetAllRoles(ctx context.Context) ([]Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetAllRoles")
	defer span.End()

	var roles []Role
	cursor, err := s.rolesCollection.Find(ctx, bson.D{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &roles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Roles found in DB Successfully")
	return roles, nil
}

func (s *authStore) UpdateRole(ctx context.Context, roleID string, role *Role) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.UpdateRole")
	defer span.End()

	roleBSON, err := bson.Marshal(role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	err = bson.Unmarshal(roleBSON, role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.rolesCollection.UpdateOne(ctx, map[string]interface{}{"_id": roleIDHex}, map[string]interface{}{"$set": role})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Role Updated in DB Successfully!")
	return nil
}

func (s *authStore) DeleteRole(ctx context.Context, roleID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.DeleteRole")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.rolesCollection.DeleteOne(ctx, map[string]interface{}{"_id": roleIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Role Deleted from DB Successfully!")
	return nil
}

func (s *authStore) RegisterPermission(ctx context.Context, permission *Permission) (string, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.RegisterPermission")
	defer span.End()

	permissionBSON, err := bson.Marshal(permission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return "", err
	}

	err = bson.Unmarshal(permissionBSON, permission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return "", err
	}

	_, err = s.permissionsCollection.InsertOne(ctx, permission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return "", err
	}

	insertedID := permission.ID.Hex()

	span.SetStatus(otlpcodes.Ok, "Permission Registered in DB Successfully!")
	return insertedID, nil
}

func (s *authStore) GetPermissionByIdentifier(ctx context.Context, permissionIdentifier string) (*Permission, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetPermissionByIdentifier")
	defer span.End()

	var permission Permission
	err := s.permissionsCollection.FindOne(ctx, map[string]interface{}{"identifier": permissionIdentifier}).Decode(&permission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Permission found in DB Successfully")
	return &permission, nil
}

func (s *authStore) GetAllPermissions(ctx context.Context) ([]Permission, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetAllPermissions")
	defer span.End()

	var permissions []Permission
	cursor, err := s.permissionsCollection.Find(ctx, bson.D{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &permissions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Permissions found in DB Successfully")
	return permissions, nil
}

func (s *authStore) UpdatePermission(ctx context.Context, permissionID string, permission *Permission) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.UpdatePermission")
	defer span.End()

	permissionBSON, err := bson.Marshal(permission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	err = bson.Unmarshal(permissionBSON, permission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	permissionIDHex, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.permissionsCollection.UpdateOne(ctx, map[string]interface{}{"_id": permissionIDHex}, map[string]interface{}{"$set": permission})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Permission Updated in DB Successfully!")
	return nil
}

func (s *authStore) DeletePermission(ctx context.Context, permissionID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.DeletePermission")
	defer span.End()

	permissionIDHex, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.permissionsCollection.DeleteOne(ctx, map[string]interface{}{"_id": permissionIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Permission Deleted from DB Successfully!")
	return nil
}

func (s *authStore) RegisterRolePermission(ctx context.Context, roleID string, permissionID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.RegisterRolePermission")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	permissionIDHex, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.rolePermissionsCollection.InsertOne(ctx, map[string]interface{}{"role_id": roleIDHex, "permission_id": permissionIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Role Permission Registered in DB Successfully!")
	return nil
}

func (s *authStore) GetRolePermissions(ctx context.Context, roleID string) ([]RolePermission, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetRolePermissions")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	var rolePermissions []RolePermission
	cursor, err := s.rolePermissionsCollection.Find(ctx, map[string]interface{}{"role_id": roleIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &rolePermissions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role Permissions found in DB Successfully")
	return rolePermissions, nil
}

func (s *authStore) UpdateRolePermission(ctx context.Context, roleID string, permissionID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.UpdateRolePermission")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	permissionIDHex, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.rolePermissionsCollection.UpdateOne(ctx, map[string]interface{}{"role_id": roleIDHex, "permission_id": permissionIDHex}, map[string]interface{}{"$set": map[string]interface{}{"role_id": roleIDHex, "permission_id": permissionIDHex}})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Role Permission Updated in DB Successfully!")
	return nil
}

func (s *authStore) DeleteRolePermission(ctx context.Context, roleID string, permissionID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.DeleteRolePermission")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	permissionIDHex, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.rolePermissionsCollection.DeleteOne(ctx, map[string]interface{}{"role_id": roleIDHex, "permission_id": permissionIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Role Permission Deleted from DB Successfully!")
	return nil
}

func (s *authStore) GetAllRolePermissions(ctx context.Context) ([]RolePermission, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetAllRolePermissions")
	defer span.End()

	var rolePermissions []RolePermission
	cursor, err := s.rolePermissionsCollection.Find(ctx, bson.D{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &rolePermissions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role Permissions found in DB Successfully")
	return rolePermissions, nil
}

func (s *authStore) GetRolePermission(ctx context.Context, roleID string, permissionID string) (RolePermission, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetRolePermission")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return RolePermission{}, err
	}

	permissionIDHex, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return RolePermission{}, err
	}

	var rolePermission RolePermission
	err = s.rolePermissionsCollection.FindOne(ctx, map[string]interface{}{"role_id": roleIDHex, "permission_id": permissionIDHex}).Decode(&rolePermission)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return RolePermission{}, err
	}

	span.SetStatus(otlpcodes.Ok, "Role Permission found in DB Successfully")
	return rolePermission, nil
}

func (s *authStore) RegisterUserRole(ctx context.Context, userID string, roleID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.RegisterUserRole")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.userRolesCollection.InsertOne(ctx, map[string]interface{}{"user_id": userIDHex, "role_id": roleIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "User Role Registered in DB Successfully!")
	return nil
}

func (s *authStore) UpdateUserRole(ctx context.Context, userID string, roleID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.UpdateUserRole")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.userRolesCollection.UpdateOne(ctx, map[string]interface{}{"user_id": userIDHex, "role_id": roleIDHex}, map[string]interface{}{"$set": map[string]interface{}{"user_id": userIDHex, "role_id": roleIDHex}})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "User Role Updated in DB Successfully!")
	return nil
}

func (s *authStore) DeleteUserRole(ctx context.Context, userID string, roleID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.DeleteUserRole")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	_, err = s.userRolesCollection.DeleteOne(ctx, map[string]interface{}{"user_id": userIDHex, "role_id": roleIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "User Role Deleted from DB Successfully!")
	return nil
}

func (s *authStore) GetAllUserRoles(ctx context.Context) ([]UserRole, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetAllUserRoles")
	defer span.End()

	var userRoles []UserRole
	cursor, err := s.userRolesCollection.Find(ctx, bson.D{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &userRoles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Roles found in DB Successfully")
	return userRoles, nil
}

func (s *authStore) GetUserRole(ctx context.Context, userID string, roleID string) (UserRole, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetUserRole")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return UserRole{}, err
	}

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return UserRole{}, err
	}

	var userRole UserRole
	err = s.userRolesCollection.FindOne(ctx, map[string]interface{}{"user_id": userIDHex, "role_id": roleIDHex}).Decode(&userRole)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return UserRole{}, err
	}

	span.SetStatus(otlpcodes.Ok, "User Role found in DB Successfully")
	return userRole, nil
}

func (s *authStore) GetUserRolesByUserID(ctx context.Context, userID string) ([]UserRole, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetUserRolesByUserID")
	defer span.End()

	userIDHex, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	var userRoles []UserRole
	cursor, err := s.userRolesCollection.Find(ctx, map[string]interface{}{"user_id": userIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &userRoles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Roles found in DB Successfully")
	return userRoles, nil
}

func (s *authStore) GetUserRolesByRoleID(ctx context.Context, roleID string) ([]UserRole, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authStore.GetUserRolesByRoleID")
	defer span.End()

	roleIDHex, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	var userRoles []UserRole
	cursor, err := s.userRolesCollection.Find(ctx, map[string]interface{}{"role_id": roleIDHex})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	err = cursor.All(ctx, &userRoles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Roles found in DB Successfully")
	return userRoles, nil
}
