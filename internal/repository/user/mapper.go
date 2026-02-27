package user

import "taskflow/internal/domain"

func toModel(u domain.User) UserModel {
	return UserModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

func toDomain(m UserModel) (domain.User, error) {
	return domain.NewUserFromStorage(
		m.ID,
		m.Email,
		m.PasswordHash,
		m.CreatedAt,
	)
}
