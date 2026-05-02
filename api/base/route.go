package base

import "example_api/route/account"

func (s *Server) setupRoutes() {
	s.Mux.HandleFunc(
		"POST /a/register/basic",
		GuestMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleRegister(s.DataStore, s.Validator),
		),
	)
	s.Mux.HandleFunc(
		"POST /a/register/oauth/google",
		GuestMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleRegisterOAuthGoogle(s.DataStore, s.Validator),
		),
	)
	s.Mux.HandleFunc(
		"POST /a/session/login/basic",
		GuestMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleLogin(s.DataStore, s.Validator),
		),
	)
	s.Mux.HandleFunc(
		"POST /a/session/login/oauth/google",
		GuestMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleLoginOAuthGoogle(s.DataStore, s.Validator),
		),
	)
	s.Mux.HandleFunc(
		"DELETE /a/session/logout",
		AuthMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleLogout(s.DataStore),
		),
	)
	s.Mux.HandleFunc(
		"DELETE /a/delete",
		AuthMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleDelete(s.DataStore),
		),
	)
	s.Mux.HandleFunc(
		"POST /a/verify",
		GuestMiddleware( // NOTE: cannot use AuthMiddleware because of password-reset..
			s.DataStore,
			s.Validator,
			account.HandleVerify(s.DataStore, s.Validator),
		),
	)
	s.Mux.HandleFunc(
		"POST /a/verify_resend",
		AuthMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleVerifyResend(s.DataStore, s.Validator),
		),
	)
	s.Mux.HandleFunc(
		"GET /a/status",
		AuthMiddleware(
			s.DataStore,
			s.Validator,
			account.HandleStatus(s.DataStore),
		),
	)
}
