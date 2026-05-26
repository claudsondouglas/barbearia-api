// Package actions implementa a lógica de negócio para o módulo de services.
package actions

import "errors"

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceDeleted = errors.New("service has been deleted")
var ErrServiceHasActiveSchedules = errors.New("service has active schedules")
var ErrForbidden = errors.New("forbidden")
var ErrOrgNotFound = errors.New("organization not found")
var ErrInvalidPrice = errors.New("price must be greater than 0")
var ErrInvalidDuration = errors.New("duration_min must be between 1 and 240")
