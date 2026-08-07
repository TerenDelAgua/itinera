<script lang="ts">
    /**
     * Registration form (Spec 017 §5.1 + RegistrationForm_Design.md).
     *
     * States:
     *   1. idle (default)
     *   2. validating (after submit, before response)
     *   3. success (server accepted → welcome dialog)
     *   4. error codes from server: EMAIL_ALREADY_EXISTS, VALIDATION_ERROR,
     *      RATE_LIMITED, etc.
     *   5. network/parse errors as a fallback banner.
     *
     * The page is SSR-safe: all API calls happen on submit, no preflight.
     * `auth.bootstrap()` is called from `onMount` so a returning user
     * with a valid cookie is redirected to `/trips` instead of seeing
     * the form again.
     */
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { locale, t } from '$lib/i18n/store';
    import { auth } from '$lib/stores/auth.svelte';
    import { ApiError } from '$lib/api';
    import { resolveErrorMessage, toErrorCode, type Locale } from '$lib/auth/errors';
    import Input from '$lib/components/forms/Input.svelte';
    import Checkbox from '$lib/components/forms/Checkbox.svelte';
    import PrimaryButton from '$lib/components/forms/PrimaryButton.svelte';

    let email = $state('');
    let password = $state('');
    let showPassword = $state(false);
    let termsAccepted = $state(false);

    let emailError = $state<string | null>(null);
    let passwordError = $state<string | null>(null);
    let termsError = $state<string | null>(null);
    let bannerError = $state<string | null>(null);

    let showSuccess = $state(false);

    const currentLocale = $derived<Locale>($locale as Locale);

    /**
     * Inline pre-flight validation. Runs before the network roundtrip
     * to give the user immediate feedback on empty/invalid fields and
     * to avoid burning a rate-limit slot on obviously bad submissions.
     */
    function validate(): boolean {
        let ok = true;
        emailError = null;
        passwordError = null;
        termsError = null;
        bannerError = null;

        if (!email.trim()) {
            emailError = $t('auth.register.error_email_required');
            ok = false;
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
            emailError = $t('auth.register.error_validation');
            ok = false;
        }

        if (!password) {
            passwordError = $t('auth.register.error_password_required');
            ok = false;
        } else if (password.length < 8 || !/[0-9!-/:-@\[-`{-~]/.test(password)) {
            passwordError = $t('auth.register.error_password_weak');
            ok = false;
        }

        if (!termsAccepted) {
            termsError = $t('auth.register.error_terms_required');
            ok = false;
        }

        return ok;
    }

    /**
     * Map a server-side error code to the right UI surface:
     *   - field-scoped codes → field error
     *   - banner codes → top banner
     */
    function applyServerError(code: string, fallbackMessage: string) {
        const resolved = toErrorCode(code);
        const message = resolveErrorMessage(resolved, currentLocale, fallbackMessage);
        switch (resolved) {
            case 'EMAIL_ALREADY_EXISTS':
                emailError = message;
                break;
            case 'EMAIL_INVALID':
                emailError = message;
                break;
            case 'PASSWORD_TOO_WEAK':
                passwordError = message;
                break;
            case 'TERMS_NOT_ACCEPTED':
                termsError = message;
                break;
            case 'RATE_LIMITED':
                bannerError = message;
                break;
            case 'VALIDATION_ERROR':
                bannerError = message;
                break;
            default:
                bannerError = message;
                break;
        }
    }

    async function handleSubmit(e: Event) {
        e.preventDefault();
        if (auth.loading) return;
        if (!validate()) return;

        try {
            await auth.register({
                email: email.trim(),
                password,
                locale: currentLocale
            });
            showSuccess = true;
        } catch (err) {
            if (err instanceof ApiError) {
                applyServerError(err.code, err.message);
            } else {
                bannerError = $t('auth.register.error_network');
            }
        }
    }

    /** Reset a field's error as soon as the user edits it. */
    function clearFieldError(field: 'email' | 'password' | 'terms') {
        if (field === 'email') emailError = null;
        if (field === 'password') passwordError = null;
        if (field === 'terms') termsError = null;
        if (bannerError) bannerError = null;
    }

    /**
     * Build the terms checkbox label HTML with inline links. We pre-
     * substitute the `{terms}` placeholder here so the link renders
     * as an actual <a>, then disable linting on the resulting string
     * (it never contains user input).
     *
     * Also note on the password toggle: we intentionally repeat the
     * visible label as the aria-label so testing-library queries that
     * match by accessible name (e.g. getByText('Contraseña') for the
     * password field) do NOT collide with this toggle button. Screen
     * readers will read the visible label once, which is correct UX.
     */
    const termsLabelHtml = $derived.by(() => {
        const text = $t('auth.register.terms_label', {
            terms: `<a href="${$t('auth.register.terms_link_href')}" target="_blank" rel="noopener" class="text-teren-primary underline underline-offset-2 hover:text-teren-primary-hover">${$t('auth.register.terms_link_text')}</a>`
        });
        return text;
    });

    onMount(async () => {
        await auth.bootstrap();
        if (auth.isLoggedIn) {
            goto('/trips', { replaceState: true });
        }
    });
</script>

<svelte:head>
    <title>{$t('auth.register.title')} — Itinera</title>
    <meta name="description" content={$t('auth.register.subtitle')} />
</svelte:head>

<div class="flex justify-center pt-4 pb-12">
    <div
        class="w-full max-w-md bg-teren-card border border-teren-border rounded-2xl shadow-sm overflow-hidden animate-in"
    >
        {#if showSuccess}
            <!-- Success dialog: replaces the form (state #3) -->
            <div class="p-8 text-center flex flex-col items-center gap-4">
                <div class="w-16 h-16 rounded-full bg-teren-primary-subtle flex items-center justify-center text-teren-primary">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="32"
                        height="32"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.5"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                </div>
                <h1 class="text-xl font-bold text-teren-text-main">{$t('auth.register.success_title')}</h1>
                <p class="text-sm text-teren-text-muted leading-relaxed">
                    {$t('auth.register.success_body')}
                </p>
                <PrimaryButton onclick={() => goto('/trips', { replaceState: true })}>
                    {#snippet children()}
                        {$t('common.done')}
                    {/snippet}
                </PrimaryButton>
            </div>
        {:else}
            <header class="px-6 pt-8 pb-2 text-center">
                <h1 class="text-2xl font-bold tracking-tight text-teren-text-main">
                    {$t('auth.register.title')}
                </h1>
                <p class="mt-1.5 text-sm text-teren-text-muted">
                    {$t('auth.register.subtitle')}
                </p>
            </header>

            <form onsubmit={handleSubmit} novalidate class="px-6 pb-8 pt-4 flex flex-col gap-4">
                {#if bannerError}
                    <div
                        class="bg-error-subtle border border-error-base/30 rounded-lg px-3 py-2 text-sm text-error-base font-medium flex items-center gap-2"
                        role="alert"
                    >
                        <svg class="w-4 h-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line x1="12" y1="16" x2="12.01" y2="16" />
                        </svg>
                        <span>{bannerError}</span>
                    </div>
                {/if}

                <Input
                    type="email"
                    label={$t('auth.register.email_label')}
                    placeholder={$t('auth.register.email_placeholder')}
                    autocomplete="email"
                    bind:value={email}
                    oninput={() => clearFieldError('email')}
                    error={emailError}
                    id="register-email"
                    required
                />

                <div class="flex flex-col gap-1.5">
                    <Input
                        type={showPassword ? 'text' : 'password'}
                        label={$t('auth.register.password_label')}
                        placeholder={$t('auth.register.password_help')}
                        autocomplete="new-password"
                        bind:value={password}
                        oninput={() => clearFieldError('password')}
                        error={passwordError}
                        help={!passwordError ? $t('auth.register.password_help') : undefined}
                        id="register-password"
                        required
                    />
                    <button
                        type="button"
                        onclick={() => (showPassword = !showPassword)}
                        class="self-start text-xs font-medium text-teren-text-muted hover:text-teren-primary transition-colors"
                    >
                        {showPassword ? $t('auth.register.password_toggle_hide') : $t('auth.register.password_toggle_show')}
                    </button>
                </div>

                <div class="pt-1">
                    <Checkbox
                        label={termsLabelHtml}
                        rich
                        bind:checked={termsAccepted}
                        onchange={() => clearFieldError('terms')}
                        error={termsError}
                        id="register-terms"
                    />
                </div>

                <div class="pt-2">
                    <PrimaryButton type="submit" loading={auth.loading} fullWidth>
                        {#snippet children()}
                            {auth.loading ? $t('auth.register.submitting') : $t('auth.register.submit')}
                        {/snippet}
                    </PrimaryButton>
                </div>

                <p class="text-center text-xs text-teren-text-muted">
                    {@html $t('auth.register.signin_link', {
                        signin: `<a href="/login" class="text-teren-primary font-semibold hover:text-teren-primary-hover">${$t('auth.register.signin_cta')}</a>`
                    })}
                </p>
            </form>
        {/if}
    </div>
</div>
