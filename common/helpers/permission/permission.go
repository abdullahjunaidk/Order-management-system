package permission

import (
	"common/types"
)

// MergePermissions merges role permissions with custom permissions
func MergePermissions(rolePerms, customPerms []types.Permission) []types.Permission {
	permMap := make(map[string]map[string]bool)

	// Add role permissions
	for _, perm := range rolePerms {
		if _, exists := permMap[perm.Resource]; !exists {
			permMap[perm.Resource] = make(map[string]bool)
		}
		for _, action := range perm.Actions {
			permMap[perm.Resource][action] = true
		}
	}

	// Merge custom permissions
	for _, perm := range customPerms {
		if _, exists := permMap[perm.Resource]; !exists {
			permMap[perm.Resource] = make(map[string]bool)
		}
		for _, action := range perm.Actions {
			permMap[perm.Resource][action] = true
		}
	}

	// Convert back to slice
	var result []types.Permission
	for resource, actions := range permMap {
		actionSlice := make([]string, 0, len(actions))
		for action := range actions {
			actionSlice = append(actionSlice, action)
		}
		result = append(result, types.Permission{
			Resource: resource,
			Actions:  actionSlice,
		})
	}

	return result
}

// HasPermission checks if a permission list contains a specific resource and action
func HasPermission(permissions []types.Permission, resource, action string) bool {
	for _, perm := range permissions {
		if perm.Resource == resource {
			for _, a := range perm.Actions {
				if a == action {
					return true
				}
			}
		}
	}
	return false
}
