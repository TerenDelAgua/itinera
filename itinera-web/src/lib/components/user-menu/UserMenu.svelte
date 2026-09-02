<script lang="ts">
  /**
   * UserMenu — header slot for authenticated state
   *
   * Three render states, derived from `auth`:
   *   1. `auth.isLoading`         → 32px skeleton (no layout shift).
   *   2. `!auth.isLoggedIn`       → "Sign in" link to /login.
   *   3. `auth.isLoggedIn`        → 32px avatar circle. Click toggles
   *                                 a dropdown menu with:
   *                                   • Email (truncate, no interaction)
   *                                   • Account settings (→ /account)
   *                                   • Sign out (silent, optimistic).
   *
   * Keyboard support:
   *   ArrowDown / ArrowUp — navigate menuitems (wraps).
   *   Home / End          — first / last menuitem.
   *   Enter / Space       — activate the focused menuitem.
   *   Escape              — close + restore focus to avatar.
   *
   * Outside click and Escape both close the menu.
   *
   * Focus moves to the FIRST menuitem on open, and returns to the
   * avatar button on close.
   *
   * Reduced motion: animations use `fly` (Svelte transitions) with
   * a 0-duration fallback when `prefers-reduced-motion: reduce`.
   */
  import { fly, fade } from "svelte/transition";
  import { goto } from "$app/navigation";
  import { invalidateAll } from "$app/navigation";
  import { auth } from "$lib/stores/auth.svelte";
  import { getInitials } from "$lib/utils/user";
  import { t } from "$lib/i18n/store";
  // The `$t(...)` auto-subscription happens in the template below.

  let isOpen = $state(false);
  let buttonRef: HTMLButtonElement | undefined = $state();
  let menuRef: HTMLDivElement | undefined = $state();
  let menuItems: HTMLElement[] = $state([]);
  let prefersReducedMotion = $state(false);

  $effect(() => {
    if (typeof window === "undefined") return;
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    prefersReducedMotion = mq.matches;
    const handler = (e: MediaQueryListEvent) =>
      (prefersReducedMotion = e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  });

  /** Global click + keydown handlers active only while open. */
  $effect(() => {
    if (!isOpen) return;
    function onPointer(e: MouseEvent) {
      const target = e.target as Node;
      if (
        menuRef &&
        !menuRef.contains(target) &&
        buttonRef &&
        !buttonRef.contains(target)
      ) {
        isOpen = false;
      }
    }
    function onKeydown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault();
        isOpen = false;
        buttonRef?.focus();
        return;
      }
      if (e.key === "Tab") {
        // Let Tab escape the menu normally; close afterwards.
        isOpen = false;
        return;
      }
      if (e.key === "ArrowDown" || e.key === "ArrowUp") {
        e.preventDefault();
        const idx = menuItems.findIndex((el) => el === document.activeElement);
        const dir = e.key === "ArrowDown" ? 1 : -1;
        const next =
          menuItems[(idx + dir + menuItems.length) % menuItems.length];
        next?.focus();
      }
      if (e.key === "Home") {
        e.preventDefault();
        menuItems[0]?.focus();
      }
      if (e.key === "End") {
        e.preventDefault();
        menuItems[menuItems.length - 1]?.focus();
      }
    }
    document.addEventListener("click", onPointer);
    document.addEventListener("keydown", onKeydown);
    return () => {
      document.removeEventListener("click", onPointer);
      document.removeEventListener("keydown", onKeydown);
    };
  });

  /** Move focus to first menuitem on open. Tick to let the DOM settle. */
  $effect(() => {
    if (isOpen && menuItems.length > 0) {
      queueMicrotask(() => menuItems[0]?.focus());
    }
  });

  function toggle() {
    isOpen = !isOpen;
  }

  /**
   * Sign-out handler ("Optimistic state closure"):
   * the menu closes immediately, before the network call resolves.
   * The user sees an instant "logged out" state; if the server call
   * later fails, we already cleared local state and the worst case
   * is a stale session cookie that expires on its own.
   *
   * After the call we navigate to "/" and invalidate the load
   * functions so any cached data (e.g. the trips list in /trips)
   * refetches with the now-guest cookie set. Without this, the user
   * would still see the previously-fetched account trips on screen
   * until they manually reloaded.
   */
  async function handleSignOut() {
    isOpen = false;
    await auth.logout();
    // `invalidateAll` flushes the SvelteKit load cache so dependent
    // pages re-run `load` with the cleared cookies. We also `goto("/")`
    // so even pages with no server load (e.g. /trips, which fetches
    // /trips in onMount) get re-mounted and re-fetch.
    await invalidateAll();
    await goto("/", { invalidateAll: true });
  }

  function handleAccountSettingsKeydown(e: KeyboardEvent) {
    // Anchors fire click on Enter natively; Space needs explicit
    // dispatch for parity with native button behaviour.
    if (e.key === " ") {
      e.preventDefault();
      (e.currentTarget as HTMLAnchorElement).click();
    }
  }
</script>

{#if auth.isLoading}
  <!--
      Skeleton: 32px circle, neutral input-bg, soft pulse. Reads as
      "we're checking your session" without hinting at guest or
      authenticated state. aria-hidden so AT skips it.
    -->
  <div
    class="w-8 h-8 rounded-full bg-input animate-pulse"
    data-testid="user-menu-skeleton"
    aria-hidden="true"
    role="presentation"
  ></div>
{:else if !auth.isLoggedIn}
  <!-- Guest: single inline "Sign in" link. -->
  <a
    href="/login"
    data-testid="user-menu-signin"
    class="text-sm font-semibold text-teren-primary hover:text-teren-primary-hover transition-colors px-3 py-1.5 rounded-lg hover:bg-teren-interactive-hover focus:outline-none focus:ring-2 focus:ring-teren-primary/30"
  >
    {$t("auth.sign_in")}
  </a>
{:else}
  <!-- Authenticated: avatar button + dropdown. The wrapper is
         `relative` so the absolutely-positioned dropdown anchors to
         the avatar regardless of header layout. -->
  <div class="relative">
    <button
      bind:this={buttonRef}
      type="button"
      aria-haspopup="menu"
      aria-expanded={isOpen}
      aria-label={$t("auth.user_menu.avatar_label", {
        email: auth.user?.email ?? "",
      })}
      data-testid="user-menu-avatar"
      onclick={toggle}
      class="w-8 h-8 rounded-full bg-teren-primary text-white font-semibold text-sm flex items-center justify-center hover:bg-teren-primary-hover focus:outline-none focus:ring-2 focus:ring-teren-primary/30 transition-colors"
    >
      {getInitials(auth.user?.email)}
    </button>

    {#if isOpen}
      <div
        bind:this={menuRef}
        role="menu"
        data-testid="user-menu-dropdown"
        transition:fly={prefersReducedMotion
          ? { duration: 0 }
          : { y: -8, duration: 200, opacity: 0 }}
        class="absolute top-12 right-4 md:right-6 bg-teren-card rounded-xl border border-teren-border shadow-lg py-2 min-w-[220px] z-50"
      >
        <div
          class="px-4 py-2 text-sm text-teren-text-muted truncate"
          title={auth.user?.email}
          data-testid="user-menu-email"
        >
          {auth.user?.email}
        </div>
        <div
          class="border-t border-teren-border my-1"
          role="separator"
          aria-orientation="horizontal"
        ></div>
        <a
          bind:this={menuItems[0]}
          href="/account"
          role="menuitem"
          tabindex="-1"
          data-testid="user-menu-account"
          onkeydown={handleAccountSettingsKeydown}
          class="block px-4 py-2 text-sm text-teren-text-main hover:bg-teren-interactive-hover focus:bg-teren-interactive-hover focus:outline-none transition-colors"
        >
          {$t("auth.user_menu.account_settings")}
        </a>
        <div
          class="border-t border-teren-border my-1"
          role="separator"
          aria-orientation="horizontal"
        ></div>
        <button
          bind:this={menuItems[1]}
          type="button"
          role="menuitem"
          tabindex="-1"
          data-testid="user-menu-signout"
          onclick={handleSignOut}
          class="w-full text-left px-4 py-2 text-sm text-teren-text-main hover:bg-teren-interactive-hover focus:bg-teren-interactive-hover focus:outline-none transition-colors"
        >
          {$t("auth.user_menu.sign_out")}
        </button>
      </div>
    {/if}
  </div>
{/if}
