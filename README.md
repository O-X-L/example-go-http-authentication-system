# Example - Go HTTP Authentication System

## Example Stack

* **DB**: PostgreSQL
* **API**: Golang
* **Frontend Server**: Golang
* **Frontend Client**: SvelteJS (*SvelteKit static*) + TailwindCSS + Shadcn-Components

----

## Features

### Account Registration & Login

* Via Email + Password

  * Including E-Mail verification link (*send per email to user*)

  * Option for password reset

* Via Google OAuth

  * GDPR: Google resources are only loaded if the user wants to

    The consent is saved to localStorage

### Verification Links

* Generic system for verification-link that can easily be extended

  * Implemented: Email-Verification
  * Implemented: Password Reset
  * Possible: Passwordless login
  * Possible: Account-deletion verification
  * Possible: 2FA

* Functionality to resend verification-tokens (*resend-limit is enforced*)

### Session behavior

#### Pre-Login

When users access the frontend-server they get a generic guest-session & csrf-token assigned.

Both are set as cookies.

This one is only a simple measure against dumb bots.

These guest-sessions are also limited to a single IP-address.

You can simply strip this out if you see no need for such.

#### Login

After login or registration the user gets a user-session and csrf-token assigned.

Both are set as cookies.

A user can only have a limited number of active sessions. (*MAX_CONCURRENT_SESSIONS*)

By default, a session can live up to 30 days. (*TIMEOUT_SESSION_FORCE*)

But it will be invalidated after a user is idle/offline for 36 hours. (*TIMEOUT_SESSION_IDLE*)

If a user is online, every hour the session-tokens are rotated and the idle-timeout is reset. (*TIMEOUT_SESSION_ROTATE*)

### Additional features

* Logout
* Delete account (*only API*)
* User-Status
* Expired sessions & tokens are auto-cleaned

### Security considerations

* Authentication-checks are done using a middleware.
* Passwords are hashed using Argon2 + a server-specific Pepper.
  * Warning: If the pepper is changed - all passwords will mismatch!
* Random tokens for sessions etc. are `32 byte/43 char base64`
* All User-inputs to the API are validated using [go-validator](github.com/go-playground/validator)
* As a simple anti-bot measure `Guest-Session` are required for login & register actions.
* Requiring email-verification after email+password registration.
* Timeout for transparently rotating Session- & CSRF-Tokens. (*TIMEOUT_SESSION_ROTATE | default: 1h*)
* Timeout for idle sessions. (*TIMEOUT_SESSION_IDLE | default: 36h*)
* Timeout for forced-logoff/-session-invalidation. (*TIMEOUT_SESSION_FORCE | default 30d*)
* Timeout for verification-tokens. (*default: 1h*)
* Rate-limit for resending verification-tokens. (*VERIFICATION_LINK_MAX_RESEND | default: 2*)
* Content-Security-Policy is set.
* The maximum active sessions of users are limited. (*MAX_CONCURRENT_SESSIONS | default: 5*)

**Notes**:

* You should implement rate-limits (*via reverse-proxy*) for some API-endpoints to make abuse harder

----

## Run

1. Start database:

    ```bash
    bash ./scripts/run_dev_db.sh

    # OR:

    docker run -it --rm --name example-dev-db -e POSTGRES_USER=example -e POSTGRES_PASSWORD=dev-secret -e POSTGRES_DB=example -p 127.0.0.1:5432:5432 postgres
    ```

2. Install [GoLang](https://go.dev/doc/install) and optionally [npm](https://nodejs.org/en/download)

    Or use it containerized.

3. Set the required config-env-vars & start API service:

    ```bash
    export APP_SMTP_EMAIL=your-email@example.org
    export APP_SMTP_SERVER=smtp.example.org:587
    export APP_SMTP_USER=your-email-user
    export APP_SMTP_PASSWORD=your-email-pwd
    export APP_GOOGLE_OAUTH_CLIENT_ID=xxx.apps.googleusercontent.com

    bash ./scripts/run_dev_api.sh
    ```

4. If required: Update Google OAuth Client-ID in frontend-config: `frontend/client/src/lib/config.svelte.ts`

5. Start the frontend service in another terminal:

    ```bash
    bash ./scripts/run_dev_fe.sh
    ```

----

## Config notes

* If you want to change the server/dev domain/ports - simply do a search & replace for:

    ```
    # prod
    api.example.oxl.app
    www.example.oxl.app

    # dev
    http://localhost:8080  # api
    http://localhost:8000  # frontend
    ```

* Email-Templates are placed at: `api/templates/email/`


----

## AI Usage

For transparency.

AI was used in creating this project:

* Extending the logic for new simple features (small blocks)
* Generating unit-tests
* Every change was review, reworked and verified manually
* The implementation design was created by human
