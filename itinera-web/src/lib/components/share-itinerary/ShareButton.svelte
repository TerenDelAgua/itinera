<script lang="ts">
  import { apiFetch } from "$lib/api";
  import { t } from "$lib/i18n/store";
  import { APP_URL } from "$lib/config";
  import { fly } from "svelte/transition";
  import { untrack } from "svelte";

  let {
    tripId,
    shareToken = $bindable(),
    shareEnabled = $bindable(),
  } = $props<{
    tripId: string;
    shareToken: string | null;
    shareEnabled: boolean;
  }>();

  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let copied = $state(false);
  // Local mirrors. Spec v3 §7.2 has a known gap: it uses $props without
  // $bindable, which prevents the parent from using `bind:` — but the
  // parent DOES use bind (spec §7.3). To keep both ends happy we declare
  // the props as $bindable and use `untrack` in the effect below to
  // avoid an infinite loop (prop change → set local → prop change).
  let currentToken = $state(shareToken);
  let currentEnabled = $state(shareEnabled);
  let expiresAt = $state<string | null>(null);
  let showPopover = $state(false);

  // Sync parent → child. When the parent reloads its data and the prop
  // changes (e.g. after the parent's loadAllData() following enable/disable),
  // pull the new value into the local mirror. untrack() ensures that
  // writing currentToken = shareToken does NOT trigger this effect again
  // via the bind back to the parent.
  $effect(() => {
    untrack(() => {
      currentToken = shareToken;
      currentEnabled = shareEnabled;
    });
  });

  let shareUrl = $derived(
    currentToken ? `${APP_URL}/share/${currentToken}` : "",
  );

  async function enableShare() {
    isLoading = true;
    error = null;
    try {
      const data = await apiFetch<{
        token: string;
        url: string;
        expires_at: string | null;
      }>(`/trips/${tripId}/share`, { method: "POST" });
      // Update local mirrors AND propagate to parent via $bindable.
      currentToken = data.token;
      currentEnabled = true;
      shareToken = data.token;
      shareEnabled = true;
      expiresAt = data.expires_at;
      showPopover = true;
    } catch (e: any) {
      error = e.message || $t("share.errors.generic");
    } finally {
      isLoading = false;
    }
  }

  async function disableShare() {
    isLoading = true;
    error = null;

    try {
      await apiFetch(`/trips/${tripId}/share`, { method: "DELETE" });
      currentToken = null;
      currentEnabled = false;
      shareToken = null;
      shareEnabled = false;
      showPopover = false;
    } catch (e: any) {
      error = e.message || $t("share.errors.generic");
    } finally {
      isLoading = false;
    }
  }

  async function copyLink() {
    if (!shareUrl) return;
    try {
      await navigator.clipboard.writeText(shareUrl);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      error = $t("share.errors.copy_failed");
    }
  }

  function handleclick() {
    if (!currentEnabled) enableShare();
    else showPopover = !showPopover;
  }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (!target.closest(".share-button-container")) showPopover = false;
  }
</script>

<svelte:window onclick={handleClickOutside} />

<div class="relative share-button-container">
  <button
    onclick={handleclick}
    disabled={isLoading}
    class="flex items-center gap-2 px-4 py-2 rounded-lg font-medium text-sm transition active:scale-95 disabled:opacity-50"
    class:bg-teren-primary={currentEnabled}
    class:text-white={currentEnabled}
    class:border={!currentEnabled}
    class:border-teren-border={!currentEnabled}
    class:text-teren-text-muted={!currentEnabled}
    aria-label={currentEnabled ? "Share" : "Enable share"}
  >
    {#if isLoading}
      <span class="animate-sping">⏳</span>
    {:else if currentEnabled}
      <span>🔗</span>
      <span>{$t("share.active_link")}</span>
    {:else}
      <span>🔗</span>
      <span>{$t("share.cta_share")}</span>
    {/if}
  </button>

  {#if showPopover && currentEnabled}
    <div
      transition:fly={{ y: -10, duration: 200 }}
      class="absolute right-0 top-full mt-2 w-80 rounded-xl border boder-teren-border bg-teren-surface shadow-lg p-5 z-50"
    >
      <div class="flex justify-between items-center mb-3">
        <p class="font-semibold text-sm text-teren-text-main">
          {$t("share.popover_title")}
        </p>
        <button
          onclick={() => (showPopover = false)}
          class="text-lg text-teren-text-muted hover:text-tereb-text-main"
          aria-label="Close"
        >
          ×
        </button>
      </div>

      <div
        class="rounded-lg border border-teren-border bg-teren-background p-3 mb-3 flex items-center gap-2"
      >
        <code class="text-xs flex-1 truncate text-teren-text-muted"
          >{shareUrl}</code
        >
        <button
          onclick={copyLink}
          class="text-xs font-medium px-3 py-1.5 rounded-md text-white bg-teren-primary hover:bg-teren-primary-hover transition"
        >
          {copied ? "✓" : $t("share.copy_cta")}
        </button>
      </div>

      {#if expiresAt}
        <p class="text-xs text-teren-text-muted mb-3">
          ⏰ {$t("share.expires_at", {
            date: new Date(expiresAt).toLocaleDateString(),
          })}
        </p>
      {/if}

      <button
        onclick={disableShare}
        class="w-full text-xs font-medium py-2 rounded-lg text-error-base bg-error-subtle hover:opacity-80 transition"
      >
        {$t("share.disable_cta")}
      </button>
    </div>
  {/if}
</div>
