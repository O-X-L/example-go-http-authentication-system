<script lang="ts">
	import { Button } from '$shadcn/components/ui/button';
	import { Input } from '$shadcn/components/ui/input';
	import { Label } from '$shadcn/components/ui/label';
	import * as Card from '$shadcn/components/ui/card';
	import { config } from '$lib/config.svelte';
	import { authState, getGuestToken, getGuestTokenHeaders } from '$lib/auth.svelte';
    import * as Alert from "$shadcn/components/ui/alert/index.js";
    import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
	import { goto } from '$app/navigation';
    import { onMount } from 'svelte';

	let email = $state('');
	let password = $state('');
	let errorMessage = $state('');
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
                callback: handleRegisterOAuthGoogle
            });
            // @ts-ignore
            window.google.accounts.id.renderButton(
                document.getElementById("google-register-button"),
                { theme: "outline", size: "large", width: 335 }
            );
        };

        script.onerror = () => {
            errorMessage = "Failed to load Google Sign-In. Please check your connection.";
            loading = false;
        };

        document.head.appendChild(script);
    }

    async function handleRegisterBasic(e?: Event) {
        if (e) e.preventDefault();
		loading = true;
		errorMessage = '';

		try {
			const url = config.APIDomain + config.APILocation.RegisterBasic;
			const res = await fetch(url, {
				method: 'POST',
				credentials: 'include',
				headers: getGuestTokenHeaders(),
				body: JSON.stringify({ email, password })
			});

			if (res.ok) {
				window.location.replace('/?msg=registered_basic');
			} else if (res.status == 400) {
				errorMessage = 'Please verify your email-address and password.';
			} else if (res.status == 409) {
				errorMessage = 'This email-address is already in-use.';
			} else {
				errorMessage = 'An unexpected error ocurred. Try later.';
			}
		} catch (err) {
			errorMessage = 'A network error occurred.';
		} finally {
			loading = false;
		}
	}

	async function handleRegisterOAuthGoogle(response: any) {
        loading = true;
        errorMessage = '';
        try {
			const url = config.APIDomain + config.APILocation.RegisterOAuthGoogle;
            const res = await fetch(url, {
                method: 'POST',
                credentials: 'include',
                headers: getGuestTokenHeaders(),
                // Google's response.credential is the JWT id_token expected by your backend
                body: JSON.stringify({ id_token: response.credential })
            });

			if (res.ok) {
				window.location.replace('/?msg=registered_oauth');
			} else if (res.status == 400) {
				errorMessage = 'Please verify your email-address and password.';
			} else if (res.status == 409) {
				errorMessage = 'This email-address is already in-use.';
			} else {
				errorMessage = 'An unexpected error ocurred. Try later.';
			}
        } catch (err) {
            errorMessage = 'A network error occurred during Google login.';
        } finally {
            loading = false;
        }
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
		loadGoogleOAuthDependenciesIfConsent();
    });
</script>

<div class="flex items-center justify-center pt-12">
	<Card.Root class="w-full max-w-sm border-primary/20 shadow-lg">
		<Card.Header>
			<Card.Title class="text-2xl">Create an account</Card.Title>
			<Card.Description>Enter your details to get started.</Card.Description>
		</Card.Header>
		<Card.Content class="grid gap-4">
			<div class="grid gap-2">
				<Label for="email">Email</Label>
				<Input id="email" type="email" bind:value={email} required
					onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleRegisterBasic(e)} />
			</div>
			<div class="grid gap-2">
				<Label for="password">Password</Label>
				<Input id="password" type="password" bind:value={password} required
					onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleRegisterBasic(e)} />
			</div>
			<div class="grid gap-2">
				<Button variant="default" class="w-full" onclick={handleRegisterBasic}>Register</Button>
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
							{loading ? "Loading..." : "Register with Google"}
						</Button>
						<div class="text-xs mt-1">
							<i>By clicking this button you consent to connecting to Google.
							See: <a href="https://policies.google.com/privacy">Google Privacy Policy</a></i>
						</div>
					</div>
                {/if}
                <!-- The container where Google will inject its iframe once loaded -->
                <div id="google-register-button" class:hidden={!googleOAuthLoaded}></div>
            </div>

		</Card.Content>

{#if errorMessage != ''}
		<Card.Footer>
			<Alert.Root variant="destructive">
				<AlertCircleIcon />
				<Alert.Title>Registration failure.</Alert.Title>
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
			<Button class="w-full" variant="outline" border-dashed disabled={loading} onclick={() => {window.location.replace('/a/login')}}>
				or Login
			</Button>
		</Card.Footer>
	</Card.Root>
</div>