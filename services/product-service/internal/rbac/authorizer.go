package rbac

type Authorizer struct{}

func New() *Authorizer {
	return &Authorizer{}
}

func (a *Authorizer) Authorize(role string, method string, path string) bool {
	for _, permission := range Permissions {

		if permission.Method != method {
			continue
		}

		if permission.Path != path {
			continue
		}

		for _, allowedRole := range permission.Roles {
			if allowedRole == role {
				return true
			}
		}
		return false
	}
	return false
}