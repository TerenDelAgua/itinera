<script lang="ts">
  /**
   * Account settings — GDPR deletion flow.
   *
   * Three sub-states:
   *   1. `view`         — landing: shows a clear, scary description
   *                        of what will happen, with a single "Delete
   *                        account" CTA.
   *   2. `confirm`      — modal-style input. We require the user to
   *                        type their email EXACTLY. The destructive CTA stays
   *                        disabled until the typed string matches.
   *   3. `succeeded`    — soft-delete done. Show a final card and a
   *                        CTA back to / from. The server has
   *                        revoked all sessions and the access +
   *                        refresh cookies are now cleared.
   *
   * No password re-confirm: the server's `OpaqueAuth` middleware
   * already validated the session cookie for this request, so the
   * user is who they say they are. Adding a password field here
   * would only annoy a user whose session is fresh.
   */
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { t } from "$lib/i18n/store";
  import { auth } from "$lib/stores/auth.svelte";
  import { ApiError } from "$lib/api";
  import PrimaryButton from "$lib/components/forms/PrimaryButton.svelte";

  type Stage = "view" | "confirm" | "succeeded";
  let stage = $state<Stage>("view");

  let confirmInput = $state("");
  let bannerError = $state<string | null>(null);

  onMount(() => {
    // Guard: only logged-in users may reach this page. The
    // +layout.svelte's `/me` probe has populated auth.user by
    // now, so this is a synchronous read — no extra fetch.
    if (!auth.isLoggedIn) {
      goto("/login?next=/account", { replaceState: true });
    }
  });

  function enterConfirm() {
    stage = "confirm";
    confirmInput = "";
    bannerError = null;
  }

  function cancelConfirm() {
    stage = "view";
    confirmInput = "";
    bannerError = null;
  }

  async function handleDelete() {
    if (auth.loading) return;
    if (confirmInput.trim() !== auth.user?.email) return;

    try {
      await auth.deleteAccount();
      stage = "succeeded";
    } catch (err) {
      if (err instanceof ApiError) {
        bannerError = $t("auth.delete.error_generic");
      } else {
        bannerError = $t("auth.delete.error_network");
      }
    }
  }

  function backToHome() {
    goto("/", { replaceState: true });
  }
</script>

<svelte:head>
  <title>{$t("auth.delete.page_title")} — Itinera</title>
  <meta name="description" content={$t("auth.delete.subtitle")} />
</svelte:head>

<div class="flex justify-center pt-4 pb-12">
  <div
    class="w-full max-w-md bg-teren-card border border-teren-border rounded-2xl shadow-sm overflow-hidden animate-in"
  >
    <header class="px-6 pt-8 pb-2 text-center">
      <h1 class="text-2xl font-bold tracking-tight text-teren-text-main">
        {$t("auth.delete.title")}
      </h1>
      <p class="mt-1.5 text-sm text-teren-text-muted">
        {$t("auth.delete.subtitle")}
      </p>
    </header>

    <div class="mx-6 mb-6 mt-4 space-y-4">
      {#if stage === "view"}
        <!-- Stage 1: explanation card + danger CTA. We use
                     warm dark text + small destructive icon to
                     signal seriousness without being scaremongering. -->
        <div
          class="bg-input rounded-xl border border-teren-border p-4 text-sm text-teren-text-main space-y-3"
        >
          <div class="flex items-start gap-2">
            <svg
              class="w-5 h-5 shrink-0 text-teren-primary mt-0.5"
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
            <div class="space-y-2">
              <p class="font-medium">
                {$t("auth.delete.bullet1_title")}
              </p>
              <p class="text-teren-text-muted">
                {$t("auth.delete.bullet1_body")}
              </p>
              <p class="font-medium">
                {$t("auth.delete.bullet2_title")}
              </p>
              <p class="text-teren-text-muted">
                {$t("auth.delete.bullet2_body")}
              </p>
              <p class="font-medium">
                {$t("auth.delete.bullet3_title")}
              </p>
              <p class="text-teren-text-muted">
                {$t("auth.delete.bullet3_body")}
              </p>
            </div>
          </div>
        </div>

        <PrimaryButton type="button" onclick={enterConfirm} fullWidth>
          {#snippet children()}
            {$t("auth.delete.start_cta")}
          {/snippet}
        </PrimaryButton>
      {:else if stage === "confirm"}
        <!-- Stage 2: type-your-email confirmation (Spec §5.9).
                     The destructive button stays disabled until the
                     typed string equals the user's email EXACTLY
                     (case-insensitive trim). -->
        <div
          class="bg-input rounded-xl border border-teren-border overflow-hidden"
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

          <div class="px-3 py-3 border-b border-teren-border/50">
            <p class="text-sm text-teren-text-main">
              {$t("auth.delete.confirm_prompt", {
                email: auth.user?.email ?? "",
              })}
            </p>
            <input
              id="delete-confirm"
              type="email"
              placeholder={auth.user?.email ?? ""}
              bind:value={confirmInput}
              oninput={() => {
                if (bannerError) bannerError = null;
              }}
              aria-describedby="delete-confirm-hint"
              data-testid="delete-confirm-input"
              class="w-full h-11 mt-2 px-3 bg-teren-card text-sm text-teren-text-main placeholder:text-teren-text-muted/40 border border-teren-border rounded-lg focus:outline-none focus:border-teren-primary"
            />
            <p
              id="delete-confirm-hint"
              class="text-xs text-teren-text-muted mt-1 min-h-4"
            >
              {$t("auth.delete.confirm_hint")}
            </p>
          </div>

          <div class="px-3 py-3 space-y-2">
            <!-- Destructive button: red border + text
                             but transparent fill until confirmed.
                             Stays disabled until the email matches. -->
            <button
              type="button"
              disabled={confirmInput.trim().toLowerCase() !==
                (auth.user?.email ?? "").toLowerCase() || auth.loading}
              onclick={handleDelete}
              data-testid="delete-confirm-button"
              class="w-full h-11 rounded-lg font-semibold text-sm transition-colors
                                   border-2 border-error-base text-error-base
                                   hover:bg-error-subtle
                                   disabled:opacity-40 disabled:cursor-not-allowed
                                   enabled:bg-error-base enabled:text-white enabled:hover:bg-error-hover"
            >
              {auth.loading
                ? $t("auth.delete.submitting")
                : $t("auth.delete.confirm_cta")}
            </button>
            <button
              type="button"
              onclick={cancelConfirm}
              class="w-full h-11 rounded-lg font-medium text-sm text-teren-text-muted
                                   hover:text-teren-text-main hover:bg-teren-card transition-colors"
            >
              {$t("auth.delete.cancel_cta")}
            </button>
          </div>
        </div>
      {:else}
        <!-- Stage 3: success card. Server cleared the cookies
                     and revoked every session. We redirect to /
                     via the CTA. -->
        <div
          class="bg-teren-primary-subtle border border-teren-primary/30 rounded-xl p-4 text-sm text-teren-text-main"
          data-testid="delete-success"
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
                {$t("auth.delete.success_title")}
              </p>
              <p class="mt-1 text-teren-text-muted">
                {$t("auth.delete.success_body")}
              </p>
            </div>
          </div>
          <PrimaryButton type="button" onclick={backToHome} fullWidth>
            {#snippet children()}
              {$t("auth.delete.success_cta")}
            {/snippet}
          </PrimaryButton>
        </div>
      {/if}
    </div>
  </div>
</div>
