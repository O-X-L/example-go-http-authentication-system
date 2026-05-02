interface ConfigAPILocationType {
    LoginBasic: string
    LoginOAuthGoogle: string
    RegisterBasic: string
    RegisterOAuthGoogle: string
    Logout: string
    AccountStatus: string
    Verify: string
    VerifyEmailResend: string
}

interface ConfigStorageKeysType {
    ConsentGoogleLogin: string
}


interface ConfigType {
    APIDomain: string
    APILocation: ConfigAPILocationType
    GoogleOAuthClientID: string
    StorageKeys: ConfigStorageKeysType
}

export function isDeploymentProduction(): boolean {
    return location.hostname != 'localhost';
}

function getConfig(): ConfigType {
    let config: ConfigType = {
        APIDomain: 'api.example.oxl.app',
        APILocation: {
            LoginBasic: '/a/session/login/basic',
            LoginOAuthGoogle: '/a/session/login/oauth/google',
            RegisterBasic: '/a/register/basic',
            RegisterOAuthGoogle: '/a/register/oauth/google',
            Logout: '/a/session/logout',
            AccountStatus: '/a/status',
            Verify: '/a/verify',
            VerifyEmailResend: '/a/verify_resend',
        },
        GoogleOAuthClientID: 'xxx.apps.googleusercontent.com',
        StorageKeys: {
            ConsentGoogleLogin: 'consent_google_login',
        }
    }
    if (!isDeploymentProduction()) {
        config.APIDomain = 'http://localhost:8080';
    }
    return config
}

export const config: ConfigType = getConfig();
