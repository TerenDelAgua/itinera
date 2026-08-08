<script lang="ts">
  /**

   * States :
   *   1. idle (default)
   *   2. validating (after submit, before response)
   *   3. success (server accepted → welcome dialog)
   *   4. EMAIL_ALREADY_EXISTS / EMAIL_INVALID / PASSWORD_TOO_WEAK /
   *      TERMS_NOT_ACCEPTED → field-level error
   *   5. RATE_LIMITED / VALIDATION_ERROR → top banner
   *   6. unknown server error → top banner
   *   7. network failure → top banner
   *   8. subtitle with guest-trip count (TEREN_DESIGN_SYSTEM §3.11
   *      copy approval: "Tus N viajes se guardarán en tu cuenta").
   */
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { locale, t } from "$lib/i18n/store";
  import { auth } from "$lib/stores/auth.svelte";
  import { ApiError, apiFetch } from "$lib/api";
  import {
    resolveErrorMessage,
    toErrorCode,
    type Locale,
  } from "$lib/auth/errors";
  import Input from "$lib/components/forms/Input.svelte";
  import Checkbox from "$lib/components/forms/Checkbox.svelte";
  import PrimaryButton from "$lib/components/forms/PrimaryButton.svelte";

  let email = $state("");
  let password = $state("");
  let showPassword = $state(false);
  let termsAccepted = $state(false);

  let emailError = $state<string | null>(null);
  let passwordError = $state<string | null>(null);
  let termsError = $state<string | null>(null);
  let bannerError = $state<string | null>(null);

  let showSuccess = $state(false);

  // Issue 3: dynamic subtitle that reflects how many trips the
  // guest already has. `null` while the count is being fetched
  // (we render the loading placeholder to avoid layout shift).
  let guestTripsCount = $state<number | null>(null);

  const currentLocale = $derived<Locale>($locale as Locale);

  /**
   * Localised subtitle based on the guest's trip count.
   *   0 trips → "Save your trips to your account."
   *   1 trip  → "Your 1 trip will be saved to your account."
   *   N trips → "Your N trips will be saved to your account."
   *   loading → muted "Loading…" placeholder.
   */
  const subtitleText = $derived.by(() => {
    if (guestTripsCount === null) return $t("auth.register.subtitle_loading");
    if (guestTripsCount === 0) return $t("auth.register.subtitle_zero");
    if (guestTripsCount === 1) return $t("auth.register.subtitle_one");
    return $t("auth.register.subtitle_other", { count: guestTripsCount });
  });

  /**
   * Inline pre-flight validation. Mirrors the server's exact
   * password rule (Spec 017 §5.1: min 8 chars) — see Issue 4.
   */
  function validate(): boolean {
    let ok = true;
    emailError = null;
    passwordError = null;
    termsError = null;
    bannerError = null;

    if (!email.trim()) {
      emailError = $t("auth.register.error_email_required");
      ok = false;
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      emailError = $t("auth.register.error_validation");
      ok = false;
    }

    if (!password) {
      passwordError = $t("auth.register.error_password_required");
      ok = false;
    } else if (password.length < 8) {
      passwordError = $t("auth.register.error_password_weak");
      ok = false;
    }

    if (!termsAccepted) {
      termsError = $t("auth.register.error_terms_required");
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
    const message = resolveErrorMessage(
      resolved,
      currentLocale,
      fallbackMessage,
    );
    switch (resolved) {
      case "EMAIL_ALREADY_EXISTS":
        emailError = message;
        break;
      case "EMAIL_INVALID":
        emailError = message;
        break;
      case "PASSWORD_TOO_WEAK":
        passwordError = message;
        break;
      case "TERMS_NOT_ACCEPTED":
        termsError = message;
        break;
      case "RATE_LIMITED":
        bannerError = message;
        break;
      case "VALIDATION_ERROR":
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
        locale: currentLocale,
      });
      showSuccess = true;
    } catch (err) {
      if (err instanceof ApiError) {
        applyServerError(err.code, err.message);
      } else {
        bannerError = $t("auth.register.error_network");
      }
    }
  }

  /** Reset a field's error as soon as the user edits it. */
  function clearFieldError(field: "email" | "password" | "terms") {
    if (field === "email") emailError = null;
    if (field === "password") passwordError = null;
    if (field === "terms") termsError = null;
    if (bannerError) bannerError = null;
  }

  onMount(async () => {
    await auth.bootstrap();
    if (auth.isLoggedIn) {
      goto("/trips", { replaceState: true });
      return;
    }

    // Issue 3: load the guest's trip count so the subtitle can be
    // personalised. If the request fails (offline, server 5xx) we
    // fall back to the generic "Save your trips…" copy rather than
    // rendering a broken UI.
    try {
      const data = await apiFetch<{ trips: unknown[] }>("/trips");
      guestTripsCount = data.trips.length;
    } catch {
      guestTripsCount = 0;
    }
  });
</script>

<svelte:head>
  <title>{$t("auth.register.title")} — Itinera</title>
  <meta name="description" content={$t("auth.register.subtitle_zero")} />
</svelte:head>

<div class="flex justify-center pt-4 pb-12">
  <div
    class="w-full max-w-md bg-teren-card border border-teren-border rounded-2xl shadow-sm overflow-hidden animate-in"
  >
    {#if showSuccess}
      <!-- Success dialog: replaces the form (state #3) -->
      <div class="p-8 text-center flex flex-col items-center gap-4">
        <div
          class="w-16 h-16 rounded-full bg-teren-primary-subtle flex items-center justify-center text-teren-primary"
        >
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
        <h1 class="text-xl font-bold text-teren-text-main">
          {$t("auth.register.success_title")}
        </h1>
        <p class="text-sm text-teren-text-muted leading-relaxed">
          {$t("auth.register.success_body")}
        </p>
        <PrimaryButton onclick={() => goto("/trips", { replaceState: true })}>
          {#snippet children()}
            {$t("common.done")}
          {/snippet}
        </PrimaryButton>
      </div>
    {:else}
      <header class="px-6 pt-8 pb-2 text-center">
        <h1 class="text-2xl font-bold tracking-tight text-teren-text-main">
          {$t("auth.register.title")}
        </h1>
        <p
          class="mt-1.5 text-sm text-teren-text-muted min-h-[1.25rem]"
          data-testid="register-subtitle"
        >
          {subtitleText}
        </p>
      </header>

      <!-- Unified Widget (§3.10): a single container holding the
                 three form sections, separated by internal border-b
                 lines instead of three independent cards. The whole
                 surface reacts to focus-within / hover as one unit. -->
      <form
        onsubmit={handleSubmit}
        novalidate
        class="mx-6 mb-6 mt-4 bg-teren-surface rounded-xl border border-teren-border overflow-hidden
                       focus-within:border-teren-primary/70 focus-within:shadow-sm transition-colors"
      >
        {#if bannerError}
          <div
            class="bg-error-subtle border-b border-error-base/30 px-3 py-2 text-sm text-error-base font-medium flex items-center gap-2"
            role="alert"
          >
            <svg
              class="w-4 h-4 flex-shrink-0"
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

        <!-- ROW 1: Email (no trailing icon, no internal border
                     below because the divider between rows comes from
                     the section's border-b). -->
        <div class="px-3 py-3 border-b border-teren-border/50">
          <label
            for="register-email"
            class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
          >
            {$t("auth.register.email_label")}
          </label>
          <input
            id="register-email"
            type="email"
            placeholder={$t("auth.register.email_placeholder")}
            autocomplete="email"
            bind:value={email}
            oninput={() => clearFieldError("email")}
            aria-invalid={emailError ? "true" : undefined}
            class="w-full h-11 mt-1 px-1 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
          />
          {#if emailError}
            <p class="text-xs text-error-base font-medium" role="alert">
              {emailError}
            </p>
          {/if}
        </div>

        <!-- ROW 2: Password with eye toggle (§3.4: 44×44 tap
                     target inside the input, right-aligned). -->
        <div class="px-3 py-3 border-b border-teren-border/50">
          <label
            for="register-password"
            class="text-xs font-semibold uppercase tracking-wide text-teren-text-muted"
          >
            {$t("auth.register.password_label")}
          </label>
          <div
            class="flex items-center
                               {passwordError
              ? 'rounded border border-error-base/60 bg-error-subtle/30'
              : ''}"
          >
            <input
              id="register-password"
              type="text"
              placeholder={$t("auth.register.password_placeholder")}
              autocomplete="new-password"
              bind:value={password}
              oninput={() => clearFieldError("password")}
              aria-invalid={passwordError ? "true" : undefined}
              class="w-full h-11 mt-1 px-1 bg-transparent text-sm text-teren-text-main placeholder:text-teren-text-muted/40 focus:outline-none"
              class:masked-password={!showPassword}
            />
            <button
              type="button"
              onclick={() => (showPassword = !showPassword)}
              aria-label={showPassword
                ? $t("auth.register.password_toggle_hide")
                : $t("auth.register.password_toggle_show")}
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
            <p class="text-xs text-error-base font-medium" role="alert">
              {passwordError}
            </p>
          {:else}
            <!-- The helper text now ONLY appears below the
                             input, not duplicated as a placeholder. -->
            <p class="text-xs text-teren-text-muted mt-1">
              {$t("auth.register.password_help")}
            </p>
          {/if}
        </div>

        <!-- ROW 3: Terms (Issue 1 + Issue 2). The Checkbox component
                     handles the custom visual; here we only inject the
                     Svelte snippet that renders the two inline links. -->
        <div class="px-3 py-3">
          <Checkbox
            bind:checked={termsAccepted}
            onchange={() => clearFieldError("terms")}
            error={termsError}
            id="register-terms"
          >
            {#snippet children()}
              {$t("auth.register.terms_label_pre")}
              <a
                href={$t("auth.register.terms_link_href")}
                target="_blank"
                rel="noopener"
                class="text-teren-primary underline underline-offset-2 hover:text-teren-primary-hover"
              >
                {$t("auth.register.terms_link_text")}
              </a>
              {$t("auth.register.terms_label_conj")}
              <a
                href={$t("auth.register.privacy_link_href")}
                target="_blank"
                rel="noopener"
                class="text-teren-primary underline underline-offset-2 hover:text-teren-primary-hover"
              >
                {$t("auth.register.privacy_link_text")}
              </a>
              {$t("auth.register.terms_label_post")}
            {/snippet}
          </Checkbox>
        </div>

        <!-- ROW 4: Submit -->
        <div class="px-3 pb-3">
          <PrimaryButton type="submit" loading={auth.loading} fullWidth>
            {#snippet children()}
              {auth.loading
                ? $t("auth.register.submitting")
                : $t("auth.register.submit")}
            {/snippet}
          </PrimaryButton>
        </div>
      </form>

      <p class="text-center text-xs text-teren-text-muted pb-6">
        {$t("auth.register.signin_link", { signin: "" })}<a
          href="/login"
          class="text-teren-primary font-semibold hover:text-teren-primary-hover"
        >
          {$t("auth.register.signin_cta")}
        </a>
      </p>
    {/if}
  </div>
</div>
