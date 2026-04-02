package main

// CreateUserRequest — входные данные для создания пользователя.
// Теги validate задают правила валидации через go-playground/validator.
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age"   validate:"omitempty,min=1,max=120"`
}

// UpdateUserRequest — входные данные для обновления пользователя.
type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=2,max=50"`
	Email string `json:"email" validate:"omitempty,email"`
	Age   int    `json:"age"   validate:"omitempty,min=1,max=120"`
}

// User — внутренняя модель пользователя.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitempty"`
}
