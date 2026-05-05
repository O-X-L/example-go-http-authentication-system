<script lang="ts">
	import { Button } from '$shadcn/components/ui/button';
	import { Input } from '$shadcn/components/ui/input';
	import { Label } from '$shadcn/components/ui/label';
	import * as Card from '$shadcn/components/ui/card';
    import * as Alert from "$shadcn/components/ui/alert/index.js";
    import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
  	import { config } from '$lib/config.svelte';
  	import { authState, getGuestToken, getGuestTokenHeaders } from '$lib/auth.svelte';
    import CheckCircleIcon from "@lucide/svelte/icons/check-circle";
    import { page } from '$app/state';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';

	let email = $state('');
	let password = $state('');
	let errorMessage = $state('');
    let successTitle = $state('');
    let successMessage = $state('');
	let loading = $state(false);
	let googleOAuthLoaded = $state(false);

    function loadGoogleOAuthDependencies() {
        if (document.getElementById('google-gsi-script')) return;
        localStorage.setItem(config.StorageKeys.ConsentGoogleLogin, '1');

        loading = true;
        
        const script = document.createElement('script');
        script.id = 'google-gsi-script';
        script.src = 'https://accounts.google.com/gsi/client';
        script.async = true;
        script.defer = true;
        
        script.onload = () => {
            googleOAuthLoaded = true;
            loading = false;
            
            // @ts-ignore
            window.google.accounts.id.initialize({
                client_id: config.GoogleOAuthClientID,
                callback: handleLoginOAuthGoogle
            });
            // @ts-ignore
            window.google.accounts.id.renderButton(
                document.getElementById("google-login-button"),
                { theme: "outline", size: "large", width: 335 }
            );
        };

        script.onerror = () => {
            errorMessage = "Failed to load Google Sign-In. Please check your connection.";
            loading = false;
        };

        document.head.appendChild(script);
    }

    async function handleLoginBasic(e?: Event) {
        if (e) e.preventDefault();
        loading = true;
		errorMessage = '';

		try {
			const url = config.APIDomain + config.APILocation.LoginBasic;
			const res = await fetch(url, {
				method: 'POST',
				credentials: 'include',
				headers: getGuestTokenHeaders(),
				body: JSON.stringify({ email, password })
			});

			if (res.ok) {
				window.location.href = '/';  // reload page so that backend adds CSRF-token
			} else if (res.status == 400) {
				errorMessage = 'Please verify your email-address and password.';
			} else if (res.status == 401) {
				errorMessage = 'Invalid email or password.';
			} else if (res.status == 403) {
				errorMessage = 'Account is inactive.';
			} else {
				errorMessage = 'An unexpected error ocurred. Try later.';
			}
		} catch (err) {
			errorMessage = 'A network error occurred.';
		} finally {
			loading = false;
		}
	}

    async function handleLoginOAuthGoogle(response: any) {
        loading = true;
        errorMessage = '';
        try {
			const url = config.APIDomain + config.APILocation.LoginOAuthGoogle;
            const res = await fetch(url, {
                method: 'POST',
                credentials: 'include',
                headers: getGuestTokenHeaders(),
                // Google's response.credential is the JWT id_token expected by your backend
                body: JSON.stringify({ id_token: response.credential })
            });

            if (res.ok) {
                window.location.href = '/';
            } else if (res.status == 401 || res.status == 404) {
                errorMessage = 'Invalid Google login. Have you registered?';
            } else if (res.status == 403) {
                errorMessage = 'Account is inactive.';
            } else {
                errorMessage = 'An unexpected error ocurred with Google login.';
            }
        } catch (err) {
            errorMessage = 'A network error occurred during Google login.';
        } finally {
            loading = false;
        }
    }

    function showRedirectMessage() {
        const redirectMsg = page.url.searchParams.get('msg');
        if (!redirectMsg) {
            return;
        }
        
        if (redirectMsg == 'email_verified') {
            successTitle = 'Email verified successfully!';
            successMessage = 'Please login.';

        } else if (redirectMsg == 'password_reset_success') {
            successTitle = 'Password reset successful!';
            successMessage = 'You can now log in with your new password.';
        }

        window.history.replaceState({}, '', '/a/login');
    }

    function loadGoogleOAuthDependenciesIfConsent() {
        const consent = localStorage.getItem(config.StorageKeys.ConsentGoogleLogin);
        if (!consent) {return;}
        loadGoogleOAuthDependencies();
    }

	onMount(() => {
        if (authState.isLoggedIn) {
            goto('/');
        }
		if (!getGuestToken()) {
			errorMessage = 'Guest session could not be started. Try later.';
		}
        showRedirectMessage();
        loadGoogleOAuthDependenciesIfConsent();
    });
</script>

<div class="flex items-center justify-center pt-12">
    <Card.Root class="w-full max-w-sm">
        <Card.Header>
            <Card.Title class="text-2xl">Login</Card.Title>
            <Card.Description>Enter your email below to login to your account.</Card.Description>
        </Card.Header>
        <Card.Content class="grid gap-4">
            <div class="grid gap-2">
                <Label for="email">Email</Label>
                <Input id="email" type="email" placeholder="m@example.com" bind:value={email} required tabindex="1"
                    onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleLoginBasic(e)} />
            </div>
            <div class="grid gap-2">
                <div class="flex items-center justify-between">
                    <Label for="password">Password</Label>
                    <a href="/a/password_reset" class="text-sm underline-offset-4 hover:underline text-muted-foreground" tabindex="4">
                        Forgot password?
                    </a>
                </div>
                <Input id="password" type="password" bind:value={password} required tabindex="2"
                    onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleLoginBasic(e)} />
            </div>
            <div class="grid gap-2">
                <Button class="w-full" disabled={loading} onclick={handleLoginBasic} tabindex="3">
                    {loading ? "Logging in..." : "Login"}
                </Button>
            </div>

            <div class="px-6 pb-6">
				<hr class="border-border" />
			</div>

            <!-- Privacy-compliant Google Loading -->
            <div class="w-full flex justify-center min-h-10">
                {#if !googleOAuthLoaded}
                    <div>
                        <Button 
                            variant="outline" 
                            class="w-full border-dashed" 
                            disabled={loading} 
                            onclick={loadGoogleOAuthDependencies}>
                            {loading ? "Loading..." : "Login with Google"}
                        </Button>
                        <div class="text-xs mt-1">
                            <i>By clicking this button you consent to connecting to Google.
                            See: <a href="https://policies.google.com/privacy">Google Privacy Policy</a></i>
                        </div>
                    </div>
                {/if}
                <!-- The container where Google will inject its iframe once loaded -->
                <div id="google-login-button" class:hidden={!googleOAuthLoaded}></div>
            </div>

        </Card.Content>

{#if successMessage != ''}
        <Card.Footer>
            <Alert.Root class="border-green-500 text-green-700 bg-green-50 dark:border-green-500/50 dark:text-green-400 dark:bg-green-500/10">
                <CheckCircleIcon class="h-4 w-4 text-green-700! dark:text-green-400!" />
                <Alert.Title>{successTitle}</Alert.Title>
                <Alert.Description>
                    <p>{successMessage}</p>
                </Alert.Description>
            </Alert.Root>
        </Card.Footer>
{/if}

{#if errorMessage != ''}
		<Card.Footer>
			<Alert.Root variant="destructive">
				<AlertCircleIcon />
				<Alert.Title>Login failure.</Alert.Title>
				<Alert.Description>
					<p>{errorMessage}</p>
				</Alert.Description>
			</Alert.Root>
		</Card.Footer>
{/if}
		<div class="px-6 pb-6">
            <hr class="border-border" />
        </div>
		<Card.Footer>
			<Button class="w-full" variant="outline" border-dashed disabled={loading} onclick={() => {window.location.replace('/a/register')}}>
				or Register
			</Button>
		</Card.Footer>
	</Card.Root>
</div>