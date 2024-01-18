package common

import "errors"

/**
200 - Ok
201 - Created
400 - Bad Request
401 - Unauthorized
403 - Forbidden
404 - Not Found
500 - Internal Server Error
503 - Service unavailable
*/

var (
	ErrBadRequest   = errors.New("bad request")   // return 400
	ErrUnauthorized = errors.New("unauthorized")  // return 401
	ErrForbidden    = errors.New("forbidden")     // return 403
	ErrNotFound     = errors.New("not found")     // return 404
	ErrInternalFail = errors.New("internal fail") // return 500
	ErrServiceDown  = errors.New("service down")  // return 503
	// code
	TrackingExistCode       = "TRACKING_EXIST"
	TrackingExistMessage    = "You are tracking this product!"
	TrackingSuccessMessage  = "Tracking product successfully!"
	TrackingFailCode        = "TRACKING_FAIL"
	TrackingFailMessage     = "Tracking product fail!"
	ProductNotFoundCode     = "PRODUCT_NOT_FOUND"
	ProductNotFoundMessage  = "Product not found!"
	UnTrackingSuccessCode   = "UNTRACKING_SUCCESS"
	UnTrackingSuccessMsg    = "Untracking product successfully!"
	UnTrackingFailCode      = "UNTRACKING_FAIL"
	UnTrackingFailMsg       = "Untracking product fail!"
	TrackingNotFoundCode    = "TRACKING_NOT_FOUND"
	TrackingNotFoundMsg     = "Tracking not found!"
	CanNotCreateUserCode    = "CAN_NOT_CREATE_USER"
	CanNotCreateUserMsg     = "Can not create user!"
	EmailExistCode          = "EMAIL_EXIST"
	EmailExistMsg           = "Email already exist!"
	EmailNotFoundCode       = "EMAIL_NOT_FOUND"
	EmailNotFoundMsg        = "Email not found!"
	PasswordWrongCode       = "PASSWORD_WRONG"
	PasswordWrongMsg        = "Password wrong!"
	EmailNotVerifiedCode    = "EMAIL_NOT_VERIFIED"
	EmailNotVerifiedMsg     = "Email not verified!"
	CanNotLoginNowCode      = "CAN_NOT_LOGIN_NOW"
	CanNotLoginNowMsg       = "Can not login now!"
	InternalServerErrorCode = "INTERNAL_SERVER_ERROR"
	InternalServerMsg       = "Internal server error!"
	TokenVerifyExpiredCode  = "TOKEN_VERIFY_EXPIRED"
	TokenVerifyExpiredMsg   = "Token verify expired!"
	UnauthorizedCode        = "UNAUTHORIZED"
	UnauthorizedMsg         = "Unauthorized!"
)
