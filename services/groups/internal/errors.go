package internal

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrBadRequest         = errors.New("bad request")
	ErrInternal           = errors.New("internal error")
	ErrInviteCodeInvalid  = errors.New("invite code is invalid")
	ErrNotOwner           = errors.New("action requires group owner")
	ErrOwnerCannotLeave   = errors.New("group owner must delete the group")
	ErrSuggestionLimit    = errors.New("suggestion limit exceeded")
	ErrSuggestionCooldown = errors.New("track is on cooldown")
	ErrNotMember          = errors.New("user is not a group member")
	ErrAlreadyMember      = errors.New("user already joined group")
	ErrSuggestionHandled  = errors.New("suggestion already handled")
)
