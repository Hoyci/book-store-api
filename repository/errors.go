package repository

import "fmt"

type ResourceNotFoundError struct {
	ID int
}

type InternalDatabaseError struct {
	Message string
	Err     error
}

type InsertError struct {
	Entity string
	Err    error
}

type LastInsertIDError struct {
	Err error
}

type UpdateError struct {
	Entity string
	ID     int
	Err    error
}

type DeleteError struct {
	Entity string
	ID     int
	Err    error
}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("resource with ID %d not found", e.ID)
}

func (e *InternalDatabaseError) Error() string {
	return fmt.Sprintf("internal database error: %s", e.Message)
}

func (e *InternalDatabaseError) Unwrap() error {
	return e.Err
}

func (e *InsertError) Error() string {
	return fmt.Sprintf("failed to insert entity '%s': %v", e.Entity, e.Err)
}

func (e *InsertError) Unwrap() error {
	return e.Err
}

func (e *LastInsertIDError) Error() string {
	return fmt.Sprintf("failed to fetch last inserted ID: %v", e.Err)
}

func (e *LastInsertIDError) Unwrap() error {
	return e.Err
}

func (e *UpdateError) Error() string {
	return fmt.Sprintf("failed to update entity '%s' with id '%d': %v", e.Entity, e.ID, e.Err)
}

func (e *UpdateError) Unwrap() error {
	return e.Err
}

func (e *DeleteError) Error() string {
	return fmt.Sprintf("failed to delete entity '%s' with id '%d': %v", e.Entity, e.ID, e.Err)
}

func (e *DeleteError) Unwrap() error {
	return e.Err
}
