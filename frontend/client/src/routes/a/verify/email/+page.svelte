<script lang="ts">
	import * as Card from '$shadcn/components/ui/card';
	import { config } from '$lib/config.svelte';
	import { authState, getUserOrGuestTokenHeaders, getGuestToken } from '$lib/auth.svelte';
    import * as Alert from "$shadcn/components/ui/alert/index.js";
    import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
	import { goto } from '$app/navigation';
    import { page } from '$app/state';
    import { onMount } from 'svelte';

	let verifyToken = $state('');
	let verifyTokenID = $state('');
	let errorMessage = $state('');
	let loading = $state(false);

	async function verifyTokens() {
		loading = true;
		errorMessage = '';

		try {
			const url = config.APIDomain + config.APILocation.Verify;
			const res = await fetch(url, {
				method: 'POST',
				credentials: 'include',
				headers: getUserOrGuestTokenHeaders(),
				body: JSON.stringify({ id: verifyTokenID, token: verifyToken })
			});

			if (res.ok) {
				if (!authState.isLoggedIn) {
					window.location.replace('/a/login?msg=email_verified');
				} else {
					window.location.replace('/?msg=email_verified');
				}
			} else if (res.status == 400) {
				errorMessage = 'Invalid verification-token.';
			} else if (res.status == 401) {
				errorMessage = 'Token invalid or expired.';
			} else if (res.status == 500) {
				errorMessage = 'Unable to verify provided token. Maybe it was already verified.';
			} else {
				errorMessage = 'An unexpected error ocurred. Try later.';
			}
		} catch (err) {
			errorMessage = 'A network error occurred.';
		} finally {
			loading = false;
		}
	}

	async function loadTokenFromParams() {
		const token = page.url.searchParams.get('token');
		const tokenID = page.url.searchParams.get('id');
        if (!token || !tokenID) {
            errorMessage = 'No tokens provided!';
			return;
        }
        verifyToken = token;
		verifyTokenID = tokenID;
		await verifyTokens();
	}

	onMount(() => {
        if (authState.isLoggedIn) {
            goto('/');
        }
		if (!getGuestToken()) {
			errorMessage = 'Guest session could not be started. Try later.';
		}
		loadTokenFromParams();
    });
</script>

<div class="flex items-center justify-center pt-12">
	<Card.Root class="w-full max-w-sm border-primary/20 shadow-lg">
		<Card.Header>
			<Card.Title class="text-2xl">Verifying email-address</Card.Title>
			<Card.Description>You should be redirected to the login page in a second.</Card.Description>
		</Card.Header>
{#if loading}
		<Card.Content>
			<div class="w-full">Loading ...</div>
		</Card.Content>
{/if}
{#if errorMessage != ''}
		<Card.Footer>
			<Alert.Root variant="destructive">
				<AlertCircleIcon />
				<Alert.Title>Verification failure.</Alert.Title>
				<Alert.Description>
					<p>{errorMessage}</p>
				</Alert.Description>
			</Alert.Root>
		</Card.Footer>
{/if}
	</Card.Root>
</div>