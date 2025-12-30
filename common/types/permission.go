package types

type Permission struct {
	Resource string   `json:"resource" bson:"resource"`
	Actions  []string `json:"actions" bson:"actions"`
}

type CompanyPermission struct {
	CompanyID   string       `json:"companyId" bson:"companyId"`
	Permissions []Permission `json:"permissions" bson:"permissions"`
}

// Available resources
const (
	ResourceOrder     = "order"
	ResourceProduct   = "product"
	ResourceInventory = "inventory"
	ResourceCustomer  = "customer"
	ResourceReport    = "report"
	ResourceUser      = "user"
	ResourceCompany   = "company"
)

// Available actions
const (
	ActionCreate  = "create"
	ActionRead    = "read"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionList    = "list"
	ActionApprove = "approve"
	ActionExport  = "export"
)
