<script lang="ts">
	import { Button } from '$shadcn/components/ui/button';
	import { authState, getUserTokenHeaders } from '$lib/auth.svelte';
	import { config } from '$lib/config.svelte';
	import * as Alert from "$shadcn/components/ui/alert/index.js";
    import AlertTriangleIcon from "@lucide/svelte/icons/alert-triangle";
    import CheckCircleIcon from "@lucide/svelte/icons/check-circle";
	import { onMount } from 'svelte';
    import { page } from '$app/state';

	let { children } = $props();

	interface UserStatus {
		Email: string,
		EmailVerified: boolean,
	}

	let isLoggingOut = $state(false);
	let userStatus: UserStatus = $state({Email: "", EmailVerified: false})
    let successTitle = $state('');
    let successMessage = $state('');

	async function handleLogout() {
		isLoggingOut = true;
		try {
			const url = config.APIDomain + config.APILocation.Logout;
			const res = await fetch(url, {
				method: 'DELETE',
				credentials: 'include',
				headers: getUserTokenHeaders(),
			});

			if (res.ok) {
				authState.logout();
				window.location.href = '/a/login';

			} else if (res.status == 400 || res.status == 401) {
				// session we have is invalid
				authState.logout();
				window.location.href = '/a/login';

			} else if (res.status == 403) {
				// csrf token we have is invalid; try to pull new one from server-side-template
				window.location.reload();
			}

		} catch (err) {
			console.error('Logout failed', err);
		} finally {
			isLoggingOut = false;
		}
	}

	async function handleVerifyEmailResend() {
		try {
			const url = config.APIDomain + config.APILocation.VerifyEmailResend;
			const res = await fetch(url, {
				method: 'POST',
				credentials: 'include',
				headers: getUserTokenHeaders(),
				body: JSON.stringify({ usage: "email" })
			});

			if (res.ok) {
			}
		} catch (err) {
			console.error('Resending of email-verification failed', err);
		}
	}

	async function fetchAccountStatus() {
		try {
			const url = config.APIDomain + config.APILocation.AccountStatus;
			const res = await fetch(url, {
				method: 'GET',
				credentials: 'include',
			});

			if (res.ok) {
				const data = await res.json()
				userStatus.Email = data.email;
				userStatus.EmailVerified = data.email_verified;
			}
		} catch (err) {
			console.error('User-status fetch failed', err);
		}
	}

	function clearRedirectMessage() {
		console.log("clearing..")
		successTitle = '';
		successMessage = '';
	}

	function showRedirectMessage() {
        const redirectMsg = page.url.searchParams.get('msg');
        if (!redirectMsg) {
            return;
        }
        
        if (redirectMsg == 'registered_basic') {
            successTitle = 'Registration successful!';
            successMessage = '';

        } else if (redirectMsg == 'registered_oauth') {
            successTitle = 'Registration successful!';
            successMessage = '';

        } else if (redirectMsg == 'email_verified') {
            successTitle = 'Email verified successfully!';
            successMessage = '';
        }

        window.history.replaceState({}, '', '/');
		setTimeout(clearRedirectMessage, 5*1000);
    }

	async function userInterface() {
		fetchAccountStatus()
	}

	onMount(() => {
		showRedirectMessage();
        if (authState.isLoggedIn) {
            userInterface();
        }
    });
</script>

<div class="min-h-screen bg-background font-sans antialiased">
	<header class="border-b bg-card text-card-foreground shadow-sm">
		<div class="container mx-auto flex h-16 items-center justify-between px-4">
			<a href="/" class="text-xl font-bold tracking-tight">Example App</a>
			<nav class="flex items-center gap-4">
				{#if authState.isLoggedIn}
					<Button variant="ghost" onclick={handleLogout}>Logout</Button>
				{:else}
					<Button variant="ghost" href="/a/login">Login</Button>
					<Button href="/a/register">Register</Button>
				{/if}
			</nav>
		</div>
{#if authState.isLoggedIn && userStatus.Email && !userStatus.EmailVerified}
		<Alert.Root class="bg-orange-100 text-orange-500">
			<AlertTriangleIcon />
			<Alert.Title>Email not yet verified.</Alert.Title>
			<Alert.Description>
				<p class="text-gray-900">
					For the full functionality you have to verify your email-address!
					<Button class="inline" variant="outline" size="sm" onclick={handleVerifyEmailResend}>Resend verification email</Button>
				</p>
			</Alert.Description>
		</Alert.Root>
{/if}
{#if successMessage != ''}
		<Alert.Root class="border-green-500 text-green-700 bg-green-50 dark:border-green-500/50 dark:text-green-400 dark:bg-green-500/10">
			<CheckCircleIcon class="h-4 w-4 text-green-700! dark:text-green-400!" />
			<Alert.Title>{successTitle}</Alert.Title>
			<Alert.Description>
				<p>{successMessage}</p>
			</Alert.Description>
		</Alert.Root>
{/if}
	</header>

	<main class="container mx-auto p-6">
		{#if authState.isLoggedIn}
			Hello USER. Welcome to this example app 😊
		{/if}
		{@render children()}
	</main>
</div>
