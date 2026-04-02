package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStruct_Valid(t *testing.T) {
	req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Age: 25}
	errs := ValidateStruct(req)
	assert.Nil(t, errs)
}

func TestValidateStruct_MissingName(t *testing.T) {
	req := CreateUserRequest{Email: "alice@example.com"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "name", errs[0].Field)
}

func TestValidateStruct_MissingEmail(t *testing.T) {
	req := CreateUserRequest{Name: "Alice"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "email", errs[0].Field)
}

func TestValidateStruct_InvalidEmail(t *testing.T) {
	req := CreateUserRequest{Name: "Alice", Email: "not-an-email"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, "email", errs[0].Field)
}

func TestValidateStruct_NameTooShort(t *testing.T) {
	req := CreateUserRequest{Name: "A", Email: "alice@example.com"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, "name", errs[0].Field)
}

func TestValidateStruct_NameTooLong(t *testing.T) {
	longName := "A"
	for i := 0; i < 50; i++ {
		longName += "a"
	}
	req := CreateUserRequest{Name: longName, Email: "alice@example.com"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, "name", errs[0].Field)
}

func TestValidateStruct_AgeOutOfRange(t *testing.T) {
	req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Age: 200}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, "age", errs[0].Field)
}

func TestValidateStruct_AgeOptional(t *testing.T) {
	req := CreateUserRequest{Name: "Alice", Email: "alice@example.com"}
	errs := ValidateStruct(req)
	assert.Nil(t, errs)
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	req := CreateUserRequest{Name: "", Email: "bad"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.GreaterOrEqual(t, len(errs), 2)
}

func TestUpdateRequest_OptionalFields(t *testing.T) {
	req := UpdateUserRequest{Email: "new@example.com"}
	errs := ValidateStruct(req)
	assert.Nil(t, errs)
}

func TestUpdateRequest_InvalidEmail(t *testing.T) {
	req := UpdateUserRequest{Email: "bad-email"}
	errs := ValidateStruct(req)
	assert.NotNil(t, errs)
	assert.Equal(t, "email", errs[0].Field)
}
