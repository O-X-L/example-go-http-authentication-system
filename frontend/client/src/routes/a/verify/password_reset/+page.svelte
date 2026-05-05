<script lang="ts">
    import { Button } from '$shadcn/components/ui/button';
    import { Input } from '$shadcn/components/ui/input';
    import { Label } from '$shadcn/components/ui/label';
    import * as Card from '$shadcn/components/ui/card';
    import * as Alert from "$shadcn/components/ui/alert/index.js";
    import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
    import { config } from '$lib/config.svelte';
    import { authState, getGuestToken, getGuestTokenHeaders } from '$lib/auth.svelte';
    import { goto } from '$app/navigation';
    import { page } from '$app/state';
    import { onMount } from 'svelte';

    let verifyToken = $state('');
    let verifyTokenID = $state('');
    let newPassword = $state('');
    let errorMessage = $state('');

    let isVerifying = $state(true);
    let tokenIsValid = $state(false);
    let isSubmitting = $state(false);

    async function verifyTokens() {
        isVerifying = true;
        errorMessage = '';
        try {
            const url = config.APIDomain + config.APILocation.Verify;
            const res = await fetch(url, {
                method: 'POST',
                credentials: 'include',
                headers: getGuestTokenHeaders(),
                body: JSON.stringify({ id: verifyTokenID, token: verifyToken })
            });

            if (res.ok) {
                tokenIsValid = true;
            } else if (res.status == 400) {
                errorMessage = 'Invalid verification token.';
            } else if (res.status == 401) {
                errorMessage = 'Token is invalid or has expired.';
            } else if (res.status == 500) {
                errorMessage = 'Unable to verify provided token. It may have already been used.';
            } else {
                errorMessage = 'An unexpected error occurred. Please try later.';
            }
        } catch (err) {
            errorMessage = 'A network error occurred while verifying the link.';
        } finally {
            isVerifying = false;
        }
    }

    async function handlePasswordResetSubmit(e: Event) {
        e.preventDefault();
        isSubmitting = true;
        errorMessage = '';

        try {
            const url = config.APIDomain + config.APILocation.PasswordResetConfirm;
            const res = await fetch(url, {
                method: 'POST',
                credentials: 'include',
                headers: getGuestTokenHeaders(),
                body: JSON.stringify({
                    token_id: verifyTokenID,
                    token: verifyToken,
                    password: newPassword
                })
            });

            if (res.ok) {
                window.location.replace('/a/login?msg=password_reset_success');
            } else if (res.status == 400) {
                errorMessage = 'Invalid request. Check your password format.';
            } else if (res.status == 401) {
                errorMessage = 'Token is invalid or expired. Please request a new link.';
            } else if (res.status == 500) {
                errorMessage = 'An internal server error occurred.';
            } else {
                errorMessage = 'An unexpected error occurred. Try later.';
            }
        } catch (err) {
            errorMessage = 'A network error occurred.';
        } finally {
            isSubmitting = false;
        }
    }

    function processTokens() {
        const token = page.url.searchParams.get('token');
        const tokenID = page.url.searchParams.get('id');

        if (!token || !tokenID) {
            errorMessage = 'No verification tokens provided in the URL!';
            isVerifying = false;
            return;
        }

        verifyToken = token;
        verifyTokenID = tokenID;
        verifyTokens();
    }

    onMount(() => {
        if (authState.isLoggedIn) {
            goto('/');
            return;
        }

        if (!getGuestToken()) {
            errorMessage = 'Guest session could not be started. Try later.';
            isVerifying = false;
            return;
        }

        processTokens();
    });
</script>

<div class="flex items-center justify-center pt-12">
    <Card.Root class="w-full max-w-sm">
        <Card.Header>
            <Card.Title class="text-2xl">Set New Password</Card.Title>
            <Card.Description>
                {#if isVerifying}
                    Verifying your secure link...
                {:else if tokenIsValid}
                    Please enter your new password.
                {:else}
                    We couldn't verify your request.
                {/if}
            </Card.Description>
        </Card.Header>

        {#if !isVerifying && tokenIsValid}
            <Card.Content>
                <form onsubmit={handlePasswordResetSubmit} class="grid gap-4">
                    <div class="grid gap-2">
                        <Label for="new-password">New Password</Label>
                        <Input id="new-password" type="password" bind:value={newPassword} required />
                    </div>
                    <Button type="submit" class="w-full" disabled={isSubmitting}>
                        {isSubmitting ? "Updating..." : "Update Password"}
                    </Button>
                </form>
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