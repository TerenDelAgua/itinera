<script lang="ts">
    /**
     * /reset is now an alias for /forgot.
     *
     * The reset form was merged into /forgot (single-screen flow,
     * TEREN fundamentals: no extra click required to reveal the
     * code field). We redirect here so any old links / bookmarked
     * URLs still work. The email pre-fill survives the navigation
     * because /forgot reads `?email=` from the URL on mount.
     */
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import { onMount } from "svelte";

    onMount(() => {
        const email = page.url.searchParams.get("email");
        goto(`/forgot${email ? `?email=${encodeURIComponent(email)}` : ""}`, {
            replaceState: true
        });
    });
</script>

<p class="p-8 text-center text-sm text-teren-text-muted">Redirecting…</p>