<script lang="ts">
    import { Button } from '$shadcn/components/ui/button';
    import { Input } from '$shadcn/components/ui/input';
    import { Label } from '$shadcn/components/ui/label';
    import * as Card from '$shadcn/components/ui/card';
    import * as Alert from "$shadcn/components/ui/alert/index.js";
    import AlertCircleIcon from "@lucide/svelte/icons/alert-circle";
    import CheckCircleIcon from "@lucide/svelte/icons/check-circle";
    import { config } from '$lib/config.svelte';
    import { getGuestTokenHeaders } from '$lib/auth.svelte';

    let email = $state('');
    let loading = $state(false);
    let errorMessage = $state('');
    let success = $state(false);

    async function handleResetRequest(e: Event) {
        e.preventDefault();
        loading = true;
        errorMessage = '';
        success = false;

        try {
            const url = config.APIDomain + config.APILocation.PasswordResetRequest;
            const res = await fetch(url, {
                method: 'POST',
                credentials: 'include',
                headers: getGuestTokenHeaders(),
                body: JSON.stringify({ email })
            });

            if (res.ok) {
                success = true;
            } else {
                errorMessage = 'An error occurred while processing your request.';
            }
        } catch (err) {
            errorMessage = 'A network error occurred. Please try again.';
        } finally {
            loading = false;
        }
    }
</script>

<div class="flex items-center justify-center pt-12">
    <Card.Root class="w-full max-w-sm">
        <Card.Header>
            <Card.Title class="text-2xl">Reset Password</Card.Title>
            <Card.Description>Enter your email address to request a password reset link.</Card.Description>
        </Card.Header>

        {#if !success}
        <Card.Content>
            <form onsubmit={handleResetRequest} class="grid gap-4">
                <div class="grid gap-2">
                    <Label for="email">Email</Label>
                    <Input id="email" type="email" placeholder="m@example.com" bind:value={email} required />
                </div>
                <Button type="submit" class="w-full" disabled={loading}>
                    {loading ? "Sending Request..." : "Send Reset Link"}
                </Button>
            </form>
        </Card.Content>
        {:else}
        <Card.Content>
            <Alert.Root class="border-green-500 text-green-700 bg-green-50 dark:border-green-500/50 dark:text-green-400 dark:bg-green-500/10">
                <CheckCircleIcon class="h-4 w-4 text-green-700! dark:text-green-400!" />
                <Alert.Title>Check your email</Alert.Title>
                <Alert.Description>
                    <p>If an account exists for the provided email, we have sent a password reset link.</p>
                </Alert.Description>
            </Alert.Root>
        </Card.Content>
        {/if}

        {#if errorMessage != ''}
        <Card.Footer>
            <Alert.Root variant="destructive">
                <AlertCircleIcon />
                <Alert.Title>Error</Alert.Title>
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
            <Button class="w-full" variant="outline" href="/a/login">
                Back to Login
            </Button>
        </Card.Footer>
    </Card.Root>
</div>