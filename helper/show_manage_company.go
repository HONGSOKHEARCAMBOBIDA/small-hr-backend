package helper

import "mysql/model"

type ManageCompany struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ShowManageCompany(user model.User) []ManageCompany {
	if user.Role.Level >= 7 {
		return []ManageCompany{
			{ID: 1, Name: "មើលបានតែមួយក្រុមហ៊ុន"},
			{ID: 2, Name: "មើលបានច្រើនក្រុមហ៊ុន"},
			{ID: 3, Name: "មើលបានគ្រប់ក្រុមហ៊ុន"},
		}
	}

	return []ManageCompany{
		{ID: 1, Name: "មើលបានតែមួយក្រុមហ៊ុន"},
		{ID: 2, Name: "មើលបានច្រើនក្រុមហ៊ុន"},
	}
}
