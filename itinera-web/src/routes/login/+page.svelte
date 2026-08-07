<script lang="ts">
  /**
   * Login form ).
   *
   * States covered:
   *   1. idle (default)
   *   2. validating (after submit, before response)
   *   3. success → redirect to /trips (or wherever the user came from)
   *   4. INVALID_CREDENTIALS (anti-enumeration: same message for
   *      "user not found" and "wrong password" — Spec §9.2)
   *   5. ACCOUNT_DELETED (403, soft-deleted user trying to sign in)
   *   6. RATE_LIMITED (banner)
   *   7. network failure (banner)
   *
   * Differences vs /register:
   *   - No terms checkbox (login doesn't need it).
   *   - No pre-flight password strength check (the server validates it).
   *   - The success path is a hard redirect rather than a dialog — the
   *     user came here to access their trips, not to celebrate.
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
  import Input from "$lib/components/forms/Input.svelte";
  import PrimaryButton from "$lib/components/forms/PrimaryButton.svelte";

  let email = $state("");
  let password = $state("");
  let showPassword = $state(false);

  let emailError = $state<string | null>(null);
  let passwordError = $state<string | null>(null);
  let bannerError = $state<string | null>(null);

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
    }
    if (!password) {
      passwordError = $t("auth.login.error_password_required");
      ok = false;
    }
    return ok;
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
  }

  onMount(async () => {
    await auth.bootstrap();
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

    <form
      onsubmit={handleSubmit}
      novalidate
      class="px-6 pb-8 pt-4 flex flex-col gap-4"
    >
      {#if bannerError}
        <div
          class="bg-error-subtle border border-error-base/30 rounded-lg px-3 py-2 text-sm text-error-base font-medium flex items-center gap-2"
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

      <Input
        type="email"
        label={$t("auth.login.email_label")}
        placeholder={$t("auth.login.email_placeholder")}
        autocomplete="email"
        bind:value={email}
        oninput={() => {
          if (emailError) emailError = null;
          clearBanner();
        }}
        error={emailError}
        id="login-email"
        required
      />

      <div class="flex flex-col gap-1.5">
        <Input
          type={showPassword ? "text" : "password"}
          label={$t("auth.login.password_label")}
          placeholder={$t("auth.login.password_placeholder")}
          autocomplete="current-password"
          bind:value={password}
          oninput={() => {
            if (passwordError) passwordError = null;
            clearBanner();
          }}
          error={passwordError}
          id="login-password"
          required
        />
        <div class="flex items-center justify-between">
          <button
            type="button"
            onclick={() => (showPassword = !showPassword)}
            class="text-xs font-medium text-teren-text-muted hover:text-teren-primary transition-colors"
          >
            {showPassword
              ? $t("auth.login.password_toggle_hide")
              : $t("auth.login.password_toggle_show")}
          </button>
          <a
            href={$t("auth.login.forgot_link_href")}
            class="text-xs font-medium text-teren-text-muted hover:text-teren-primary transition-colors"
          >
            {$t("auth.login.forgot_cta")}
          </a>
        </div>
      </div>

      <div class="pt-2">
        <PrimaryButton type="submit" loading={auth.loading} fullWidth>
          {#snippet children()}
            {auth.loading
              ? $t("auth.login.submitting")
              : $t("auth.login.submit")}
          {/snippet}
        </PrimaryButton>
      </div>

      <p class="text-center text-xs text-teren-text-muted">
        {$t("auth.login.register_link", { register: "" })}<a
          href={$t("auth.login.register_link_href")}
          class="text-teren-primary font-semibold hover:text-teren-primary-hover"
        >
          {$t("auth.login.register_cta")}
        </a>
      </p>
    </form>
  </div>
</div>
