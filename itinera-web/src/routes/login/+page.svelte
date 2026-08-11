<script lang="ts">
  /**
   * States covered:
   *   1. idle (default)
   *   2. validating (after submit, before response)
   *   3. success → hard redirect to ?next=/foo (open-redirect guarded)
   *      or /trips fallback
   *   4. INVALID_CREDENTIALS (anti-enumeration: same message for
   *      "user not found" and "wrong password" — Spec §9.2)
   *   5. ACCOUNT_DELETED (403, soft-deleted user trying to sign in)
   *   6. RATE_LIMITED (banner)
   *   7. network failure (banner)
   *
   * SSR-safe: all API calls happen on submit. `auth.bootstrap()` runs
   * onMount so a returning user with valid cookies is sent to /trips
   * immediately.
   */
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { locale, t } from "$lib/i18n/store";
  import { auth } from "$lib/stores/auth.svelte";
  import { ApiError } from "$lib/api";
  import {
    resolveErrorMessage,
    toErrorCode,
    type Locale,
  } from "$lib/auth/errors";
  import PrimaryButton from "$lib/components/forms/PrimaryButton.svelte";

  let email = $state("");
  let password = $state("");
  let showPassword = $state(false);

  let emailError = $state<string | null>(null);
  let passwordError = $state<string | null>(null);
  let bannerError = $state<string | null>(null);
  /**
   * Shows the "did you recently delete your account?" hint after a
   * failed login. Set true on any credential failure, cleared when
   * the user starts editing either field.
   */
  let showPostDeleteHint = $state(false);

  const currentLocale = $derived<Locale>($locale as Locale);

  /**
   * Honour an optional `?next=/foo` query param so the user lands where
   * they were going before being redirected to login. Validates the
   * value is a same-origin path to prevent open-redirect abuse.
   */
  function safeNext(): string {
    const raw = page.url.searchParams.get("next");
    if (!raw) return "/trips";
    // Only accept same-origin paths (start with `/`, no protocol/host).
    if (!raw.startsWith("/") || raw.startsWith("//")) return "/trips";
    return raw;
  }

  function validate(): boolean {
    let ok = true;
    emailError = null;
    passwordError = null;
    bannerError = null;

    if (!email.trim()) {
      emailError = $t("auth.login.error_email_required");
      ok = false;
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      emailError = $t("auth.login.error_email_format");
      ok = false;
    }
    if (!password) {
      passwordError = $t("auth.login.error_password_required");
      ok = false;
    }
    return ok;
  }

  /**
   * Single-field validation on blur (DS §3.7). Same rationale as in
   * /register: surface "Revisa el formato del email" the moment the
   * user leaves the field rather than waiting for submit.
   */
  function validateEmailOnBlur() {
    if (!email.trim()) {
      emailError = null;
      return;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      emailError = $t("auth.login.error_email_format");
    }
  }

  /**
   * Map a server-side error code to the right UI surface.
   *
   * INVALID_CREDENTIALS is the most common error path and always goes
   * to a banner (not a field), so a wrong password and a wrong email
   * are indistinguishable. ACCOUNT_DELETED goes to a dedicated banner
   * because the user needs to know why they cannot sign in (and when
   * they can register again).
   */
  function applyServerError(code: string, fallbackMessage: string) {
    const resolved = toErrorCode(code);
    const message = resolveErrorMessage(
      resolved,
      currentLocale,
      fallbackMessage,
    );
    // All login errors are presented as banners — there is no useful
    // field-level mapping because we already pre-validated the
    // inputs and the server's anti-enumeration policy is "same
    // message whether the user exists or the password is wrong".
    bannerError = message;
    // Show the post-delete hint whenever login fails. It doesn't
    // leak account state (it's just a generic help banner) and it
    // saves the user a confused "why does my password not work?"
    // detour if they recently deleted + recreated their account.
    showPostDeleteHint = true;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (auth.loading) return;
    if (!validate()) return;

    try {
      await auth.login({
        email: email.trim(),
        password,
      });
      await auth.claimGuest();
      // Hard redirect — the auth store has populated `user`, the
      // +layout will pick it up via bootstrap on /trips.
      await goto(safeNext(), { replaceState: true });
    } catch (err) {
      if (err instanceof ApiError) {
        applyServerError(err.code, err.message);
      } else {
        bannerError = $t("auth.login.error_network");
      }
    }
  }

  /** Reset banner as soon as the user edits either field. */
  function clearBanner() {
    if (bannerError) bannerError = null;
    if (showPostDeleteHint) showPostDeleteHint = false;
  }

  onMount(() => {
    // The /+layout.svelte owns the boot-time `/auth/v2/me` probe.
    // By the time this onMount runs, `auth.user` is already
    // populated (logged-in) or `null` (guest). We just branch on
    // the synchronous `isLoggedIn` derived state — no extra fetch
    // here, no second 401 in the console.
    if (auth.isLoggedIn) {
      goto(safeNext(), { replaceState: true });
    }
  });
</script>

<svelte:head>
  <title>{$t("auth.login.title")} — Itinera</title>
  <meta name="description" content={$t("auth.login.subtitle")} />
</svelte:head>

<div class="flex justify-center pt-4 pb-12">
  <div
    class="w-full max-w-md bg-teren-card border border-teren-border rounded-2xl shadow-sm overflow-hidden animate-in"
  >
    <header class="px-6 pt-8 pb-2 text-center">
      <h1 class="text-2xl font-bold tracking-tight text-teren-text-main">
        {$t("auth.login.title")}
      </h1>
      <p class="mt-1.5 text-sm text-teren-text-muted">
        {$t("auth.login.subtitle")}
      </p>
    </header>

    <!-- Unified Widget (§3.10): single container holding the two
             inputs + a trailing "Forgot password?" link row. -->
    <form
      onsubmit={handleSubmit}
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
            /><line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          <span>{bannerError}</span>
        </div>
      {/if}
      {#if showPostDeleteHint}
        <!--
                  Post-delete hint (Spec §5.9 §9.2):
                  Shown after a failed login. Doesn't leak account state
                  (any login failure triggers it), just tells the user
                  about the 30-day retention window if they recently
                  deleted + recreated their account.
                -->
        <div
          class="bg-teren-primary-subtle border-b border-teren-primary/30 px-3 py-2 text-sm text-teren-text-main flex items-start gap-2"
          role="status"
          aria-live="polite"
          data-testid="login-post-delete-hint"
        >
          <svg
            class="w-4 h-4 shrink-0 text-teren-primary mt-0.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            aria-hidden="true"
          >
            <circle cx="12" cy="12" r="10" /><line
              x1="12"
              y1="8"
              x2="12"
              y2="12"
            /><line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          <div>
            <p class="font-medium text-teren-text-main">
              {$t("auth.login.post_delete_hint_title")}
            </p>
            <p class="mt-0.5 text-teren-text-muted">
              {$t("auth.login.post_delete_hint_body", {
                register_link: "",
              })}<a
                href="/register"
                class="font-semibold text-teren-primary hover:text-teren-primary-hover underline-offset-2 hover:underline"
              >
                {$t("auth.login.post_delete_hint_register")}
              </a>
            </p>
          </div>
        </div>
      {/if}

      <!-- ROW 1: Email -->
      <div class="px-3 py-3 border-b border-teren-border/50">
        <label
          for="login-email"
          class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
        >
          {$t("auth.login.email_label")}
        </label>
        <input
          id="login-email"
          type="email"
          placeholder={$t("auth.login.email_placeholder")}
          autocomplete="email"
          bind:value={email}
          oninput={() => {
            if (emailError) emailError = null;
            clearBanner();
          }}
          onblur={validateEmailOnBlur}
          aria-invalid={emailError ? "true" : undefined}
          aria-describedby="login-email-error"
          class="w-full h-11 mt-1 px-1 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
        />
        <!-- min-h-[20px] keeps the row height stable whether
                     the error is rendered or not. -->
        <p
          id="login-email-error"
          class="text-xs text-error-base font-medium min-h-[20px] mt-0.5"
          role="alert"
          aria-live="polite"
        >
          {emailError ?? ""}
        </p>
      </div>

      <!-- ROW 2: Password with eye toggle + Forgot link -->
      <div class="px-3 py-3 border-b border-teren-border/50">
        <label
          for="login-password"
          class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
        >
          {$t("auth.login.password_label")}
        </label>
        <div
          class="flex items-center
                           {passwordError
            ? 'rounded border border-error-base/60 bg-error-subtle/30'
            : ''}"
        >
          <!-- The input stays as `type="text"` so the
                         browser does NOT inject its native reveal
                         icon (Chromium injects an eye into
                         `type="password"` inputs that cannot be
                         hidden via CSS). The visual masking is
                         applied via the `masked-password` class
                         which sets `-webkit-text-security: disc`. -->
          <input
            id="login-password"
            type="text"
            placeholder={$t("auth.login.password_placeholder")}
            autocomplete="current-password"
            bind:value={password}
            oninput={() => {
              if (passwordError) passwordError = null;
              clearBanner();
            }}
            aria-invalid={passwordError ? "true" : undefined}
            class="w-full h-11 mt-1 px-1 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
            class:masked-password={!showPassword}
          />
          <button
            type="button"
            onclick={() => (showPassword = !showPassword)}
            aria-label={showPassword
              ? $t("auth.login.password_toggle_hide")
              : $t("auth.login.password_toggle_show")}
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
                <line x1="1" y1="1" x2="23" y2="23"></line>
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
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                <circle cx="12" cy="12" r="3"></circle>
              </svg>
            {/if}
          </button>
        </div>
        {#if passwordError}
          <p
            id="login-password-error"
            class="text-xs text-error-base font-medium min-h-[20px] mt-0.5"
            role="alert"
            aria-live="polite"
          >
            {passwordError}
          </p>
        {/if}
        <a
          href={$t("auth.login.forgot_link_href")}
          class="inline-block mt-1.5 text-xs font-medium text-teren-text-muted hover:text-teren-primary transition-colors"
        >
          {$t("auth.login.forgot_cta")}
        </a>
      </div>

      <!-- ROW 3: Submit -->
      <div class="px-3 pb-3">
        <PrimaryButton type="submit" loading={auth.loading} fullWidth>
          {#snippet children()}
            {auth.loading
              ? $t("auth.login.submitting")
              : $t("auth.login.submit")}
          {/snippet}
        </PrimaryButton>
      </div>
    </form>

    <p class="text-center text-xs text-teren-text-muted pb-6">
      {$t("auth.login.register_link", { register: "" })}<a
        href={$t("auth.login.register_link_href")}
        class="text-teren-primary font-semibold hover:text-teren-primary-hover"
      >
        {$t("auth.login.register_cta")}
      </a>
    </p>
  </div>
</div>
