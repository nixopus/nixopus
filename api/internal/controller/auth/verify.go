package auth

import "net/http"

// SendVerificationEmail handles HTTP requests to send a verification email to a user.
//
// It expects the user to be present in the context.
//
// If the user is not present in the context, it responds with a 500 error.
//
// On successful sending of the verification email, it responds with a 200 status code and a JSON response
// indicating successful verification email sending.
func (c *AuthController) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {

}

func (c *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {

}
