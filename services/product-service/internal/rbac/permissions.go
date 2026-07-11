package rbac

import "net/http"

type Permission struct {
	Method string
	Path string

	Roles []string
}

var Permissions = []Permission{
	{
		Method: http.MethodGet,
		Path: "/products",

		Roles: []string{
			"user",
			"employee",
		},
	},
	{
		Method: http.MethodGet,
		Path: "/products/{id}",

		Roles: []string{
			"user",
			"employee",
		},
	},
	{
		Method: http.MethodPost,
		Path: "/products",

		Roles: []string{
			"employee",
		},
	},
	{
		Method: http.MethodPatch,
		Path: "/products/{id}",

		Roles: []string{
			"employee",
		},
	},
	{
		Method: http.MethodDelete,
		Path: "/products/{id}",

		Roles: []string{
			"employee",
		},
	},
}