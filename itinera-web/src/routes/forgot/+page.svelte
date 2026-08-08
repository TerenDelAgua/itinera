<script lang="ts">
    /**
     * Forgot + reset password, single-screen flow (Spec 017 §5.7 + §5.8).
     *
     * Anti-enumeration contract: the backend ALWAYS returns 202 with
     * the same generic message, regardless of whether the email exists
     * or not (Spec §9.2). Therefore the user-visible flow is:
     *   1. User enters email. (Code + new-password fields are hidden.)
     *   2. We POST /auth/v2/forgot.
     *   3. Inline reveal of the code + new-password fields with the
     *      email locked (the user has already committed to it). No
     *      intermediate screen — TEREN fundamentals say we don't
     *      require an extra click.
     *   4. User fills code + new password, click "Reset password".
     *   5. On success: success card with "Sign in" CTA.
     *
     * The success-card CTA goes to /login because the backend revokes
     * every session — we don't auto-login.
     */
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import { onMount } from "svelte";
    import { locale, t } from "$lib/i18n/store";
    import { auth } from "$lib/stores/auth.svelte";
    import { ApiError } from "$lib/api";
    import PrimaryButton from "$lib/components/forms/PrimaryButton.svelte";

    /** Pre-fill email when arriving with `?email=…`. This preserves
     *  the flow that previously sent users from /forgot (old
     *  two-screen flow) to /reset?email=…, and still works for any
     *  external links / emails that carry the pre-fill. */
    onMount(() => {
        const raw = page.url.searchParams.get("email");
        if (raw && !email) email = raw;
    });

    let email = $state("");
    let code = $state("");
    let password = $state("");
    let showPassword = $state(false);

    let emailError = $state<string | null>(null);
    let codeError = $state<string | null>(null);
    let passwordError = $state<string | null>(null);
    let bannerError = $state<string | null>(null);

    /** True after the user has requested a code — reveals the
     *  second-stage fields (code + new password). */
    let codeSent = $state(false);
    let codeSentEmail = $state("");

    /** True after a successful reset — replaces the form with a
     *  success card. */
    let succeeded = $state(false);

    /** Mirror of `passwordIsStrong` in /register (Spec 017 §5.1). */
    function passwordIsStrong(p: string): boolean {
        if (p.length < 8) return false;
        let hasDigit = false;
        let hasSymbol = false;
        for (const ch of p) {
            if (/\d/.test(ch)) hasDigit = true;
            else if (!/[a-zA-Z]/.test(ch)) hasSymbol = true;
        }
        return hasDigit && hasSymbol;
    }

    function validateEmail(): boolean {
        emailError = null;
        bannerError = null;
        if (!email.trim()) {
            emailError = $t("auth.forgot.error_email_required");
            return false;
        }
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
            emailError = $t("auth.forgot.error_email_format");
            return false;
        }
        return true;
    }

    function validateCodeAndPassword(): boolean {
        codeError = null;
        passwordError = null;
        bannerError = null;
        let ok = true;
        if (!code.trim()) {
            codeError = $t("auth.reset.error_code_required");
            ok = false;
        } else if (!/^\d{6}$/.test(code.trim())) {
            codeError = $t("auth.reset.error_code_format");
            ok = false;
        }
        if (!password) {
            passwordError = $t("auth.reset.error_password_required");
            ok = false;
        } else if (!passwordIsStrong(password)) {
            passwordError = $t("auth.reset.error_password_weak");
            ok = false;
        }
        return ok;
    }

    /** Map server error codes for the reset step. */
    function applyServerError(code: string) {
        if (code === "INVALID_RESET_TOKEN" || code === "UNAUTHENTICATED") {
            bannerError = $t("auth.reset.error_invalid_code");
        } else if (code === "WEAK_PASSWORD") {
            passwordError = $t("auth.reset.error_password_weak");
        } else if (code === "VALIDATION_ERROR") {
            bannerError = $t("auth.reset.error_validation");
        } else if (code === "RATE_LIMITED") {
            bannerError = $t("auth.reset.error_rate_limit");
        } else {
            bannerError = $t("auth.reset.error_generic");
        }
    }

    async function handleSendCode(e: Event) {
        e.preventDefault();
        if (auth.loading) return;
        if (!validateEmail()) return;
        try {
            await auth.forgotPassword({
                email: email.trim(),
                locale: $locale
            });
            codeSent = true;
            codeSentEmail = email.trim();
            bannerError = null;
        } catch (err) {
            if (err instanceof ApiError) {
                if (err.code === "RATE_LIMITED") {
                    bannerError = $t("auth.forgot.error_rate_limit");
                } else if (err.code === "VALIDATION_ERROR") {
                    emailError = $t("auth.forgot.error_email_format");
                } else {
                    bannerError = $t("auth.forgot.error_generic");
                }
            } else {
                bannerError = $t("auth.forgot.error_network");
            }
        }
    }

    /** Lets the user correct a typo in the email without losing
     *  what they already typed for code + new password. */
    function resetToEmailStep() {
        codeSent = false;
        code = "";
        password = "";
        emailError = null;
        codeError = null;
        passwordError = null;
        bannerError = null;
    }

    async function handleReset(e: Event) {
        e.preventDefault();
        if (auth.loading) return;
        if (!validateCodeAndPassword()) return;
        try {
            await auth.resetPassword({
                email: codeSentEmail,
                code: code.trim(),
                new_password: password,
                locale: $locale
            });
            succeeded = true;
        } catch (err) {
            if (err instanceof ApiError) {
                applyServerError(err.code);
            } else {
                bannerError = $t("auth.reset.error_network");
            }
        }
    }
</script>

<svelte:head>
    <title>{$t("auth.forgot.title")} — Itinera</title>
    <meta name="description" content={$t("auth.forgot.subtitle")} />
</svelte:head>

<div class="flex justify-center pt-4 pb-12">
    <div
        class="w-full max-w-md bg-teren-card border border-teren-border rounded-2xl shadow-sm overflow-hidden animate-in"
    >
        <header class="px-6 pt-8 pb-2 text-center">
            <h1 class="text-2xl font-bold tracking-tight text-teren-text-main">
                {$t("auth.forgot.title")}
            </h1>
            <p class="mt-1.5 text-sm text-teren-text-muted">
                {codeSent
                    ? $t("auth.forgot.subtitle_code_sent", {
                          email: codeSentEmail
                      })
                    : $t("auth.forgot.subtitle")}
            </p>
        </header>

        {#if succeeded}
            <!-- Success card. We intentionally do NOT auto-login
                 because the backend just revoked every session. -->
            <div
                class="mx-6 mb-4 bg-teren-primary-subtle border border-teren-primary/30 rounded-xl p-4 text-sm text-teren-text-main"
                data-testid="reset-success"
                role="status"
                aria-live="polite"
            >
                <div class="flex items-start gap-2">
                    <svg
                        class="w-5 h-5 shrink-0 text-teren-primary"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <circle cx="12" cy="12" r="10" /><line
                            x1="12"
                            y1="8"
                            x2="12"
                            y2="12"
                        /><line x1="12" y1="16" x2="12.01" y2="16" />
                    </svg>
                    <div>
                        <p class="font-medium">
                            {$t("auth.reset.success_title")}
                        </p>
                        <p class="mt-1 text-teren-text-muted">
                            {$t("auth.reset.success_body")}
                        </p>
                    </div>
                </div>
                <PrimaryButton
                    type="button"
                    onclick={() => goto("/login")}
                    fullWidth
                >
                    {#snippet children()}
                        {$t("auth.reset.success_cta")}
                    {/snippet}
                </PrimaryButton>
            </div>
        {:else}
            <form
                onsubmit={codeSent ? handleReset : handleSendCode}
                novalidate
                class="mx-6 mb-6 mt-4 bg-input rounded-xl border border-teren-border overflow-hidden
                       focus-within:border-teren-primary/70 focus-within:shadow-sm transition-colors"
            >
                {#if bannerError}
                    <div
                        class="bg-error-subtle border-b border-error-base/30 px-3 py-2 text-sm text-error-base font-medium flex items-center gap-2"
                        role="alert"
                    >
                        <svg
                            class="w-4 h-4 shrink-0"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <circle cx="12" cy="12" r="10" /><line
                                x1="12"
                                y1="8"
                                x2="12"
                                y2="12"
                            /><line
                                x1="12"
                                y1="16"
                                x2="12.01"
                                y2="16"
                            />
                        </svg>
                        <span>{bannerError}</span>
                    </div>
                {/if}

                <!-- ROW 1: Email (the only always-visible field).
                     Locked once the code is sent so the user can't
                     silently change the destination of the reset. -->
                <div class="px-3 py-3 border-b border-teren-border/50">
                    <label
                        for="forgot-email"
                        class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
                    >
                        {$t("auth.forgot.email_label")}
                    </label>
                    <input
                        id="forgot-email"
                        type="email"
                        placeholder={$t("auth.forgot.email_placeholder")}
                        autocomplete="email"
                        bind:value={email}
                        readonly={codeSent}
                        oninput={() => {
                            if (emailError) emailError = null;
                            if (bannerError) bannerError = null;
                        }}
                        aria-invalid={emailError ? "true" : undefined}
                        aria-describedby="forgot-email-error"
                        class="w-full h-11 mt-1 px-1 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none read-only:text-teren-text-muted"
                    />
                    <p
                        id="forgot-email-error"
                        class="text-xs text-error-base font-medium min-h-[20px] mt-0.5"
                        role="alert"
                        aria-live="polite"
                    >
                        {emailError ?? ""}
                    </p>
                </div>

                {#if !codeSent}
                    <!-- Stage 1: only the email field + send-code CTA.
                         Hidden once the code is sent. -->
                    <div class="px-3 pb-3 pt-2">
                        <PrimaryButton
                            type="submit"
                            loading={auth.loading}
                            fullWidth
                        >
                            {#snippet children()}
                                {auth.loading
                                    ? $t("auth.forgot.submitting")
                                    : $t("auth.forgot.submit")}
                            {/snippet}
                        </PrimaryButton>
                    </div>
                {:else}
                    <!-- Stage 2: inline reveal (no intermediate screen).
                         Code + new password appear here the moment the
                         request returns, with the email locked above. -->
                    <div class="px-3 py-3 border-b border-teren-border/50">
                        <label
                            for="forgot-code"
                            class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
                        >
                            {$t("auth.reset.code_label")}
                        </label>
                        <input
                            id="forgot-code"
                            type="text"
                            inputmode="numeric"
                            autocomplete="one-time-code"
                            maxlength="6"
                            pattern="\d{6}"
                            placeholder={$t("auth.reset.code_placeholder")}
                            bind:value={code}
                            oninput={() => {
                                if (codeError) codeError = null;
                                if (bannerError) bannerError = null;
                            }}
                            aria-invalid={codeError ? "true" : undefined}
                            aria-describedby="forgot-code-error"
                            class="w-full h-11 mt-1 px-1 bg-transparent text-base font-mono tracking-[0.4em] text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
                        />
                        <p
                            id="forgot-code-error"
                            class="text-xs text-error-base font-medium min-h-[20px] mt-0.5"
                            role="alert"
                            aria-live="polite"
                        >
                            {codeError ?? ""}
                        </p>
                    </div>

                    <div class="px-3 py-3 border-b border-teren-border/50">
                        <label
                            for="forgot-password"
                            class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
                        >
                            {$t("auth.reset.password_label")}
                        </label>
                        <div class="flex items-center">
                            <input
                                id="forgot-password"
                                type="text"
                                placeholder={$t("auth.reset.password_placeholder")}
                                autocomplete="new-password"
                                bind:value={password}
                                oninput={() => {
                                    if (passwordError) passwordError = null;
                                    if (bannerError) bannerError = null;
                                }}
                                aria-invalid={passwordError ? "true" : undefined}
                                aria-describedby="forgot-password-error"
                                class="w-full h-11 mt-1 px-1 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
                                class:masked-password={!showPassword}
                            />
                            <button
                                type="button"
                                onclick={() =>
                                    (showPassword = !showPassword)}
                                aria-label={showPassword
                                    ? $t("auth.reset.password_toggle_hide")
                                    : $t("auth.reset.password_toggle_show")}
                                class="flex-shrink-0 w-11 h-11 flex items-center justify-center text-teren-text-muted hover:text-teren-primary transition-colors rounded-lg"
                            >
                                {#if showPassword}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="20"
                                        height="20"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        aria-hidden="true"
                                    >
                                        <path
                                            d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
                                        ></path>
                                        <line
                                            x1="1"
                                            y1="1"
                                            x2="23"
                                            y2="23"
                                        ></line>
                                    </svg>
                                {:else}
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="20"
                                        height="20"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        aria-hidden="true"
                                    >
                                        <path
                                            d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                                        ></path>
                                        <circle cx="12" cy="12" r="3"></circle>
                                    </svg>
                                {/if}
                            </button>
                        </div>
                        <p
                            id="forgot-password-error"
                            class="text-xs text-error-base font-medium min-h-[20px] mt-0.5"
                            role="alert"
                            aria-live="polite"
                        >
                            {passwordError ?? ""}
                        </p>
                    </div>

                    <div class="px-3 pb-3 pt-2">
                        <PrimaryButton
                            type="submit"
                            loading={auth.loading}
                            fullWidth
                        >
                            {#snippet children()}
                                {auth.loading
                                    ? $t("auth.reset.submitting")
                                    : $t("auth.reset.submit")}
                            {/snippet}
                        </PrimaryButton>
                        <button
                            type="button"
                            onclick={resetToEmailStep}
                            class="w-full mt-2 text-xs text-teren-text-muted hover:text-teren-primary underline-offset-2 hover:underline"
                        >
                            {$t("auth.forgot.change_email")}
                        </button>
                    </div>
                {/if}
            </form>
        {/if}

        <p class="text-center text-xs text-teren-text-muted pb-6">
            {$t("auth.forgot.signin_link", { signin: "" })}<a
                href="/login"
                class="text-teren-primary font-semibold hover:text-teren-primary-hover"
            >
                {$t("auth.forgot.signin_cta")}
            </a>
        </p>
    </div>
</div>