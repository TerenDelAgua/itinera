<script lang="ts">
import type { Snippet } from "svelte";

let { title: titleProp, subtitle: subtitleProp, 
    href: hrefProp, clickable, children }
  : { title: string; subtitle: string; 
    href: string | null; clickable: boolean; 
    children?: Snippet } = $props();

const isInteractive = () => Boolean(clickable && hrefProp);
</script>

{#if isInteractive()}
  <a
    class="card clickable"
    class:has-title={titleProp !== ""}
    href={hrefProp ?? undefined}
  >
    {#if titleProp}
      <h3 class="card-title">{titleProp}</h3>
    {/if}
    
    {#if subtitleProp}
      <p class="card-subtitle">{subtitleProp}</p>
    {/if}
    
    <div class="card-content">
      {@render children?.()}
    </div>
  </a>
{:else}
  <article
    class="card"
    class:has-title={titleProp !== ""}
  >
    {#if titleProp}
      <h3 class="card-title">{titleProp}</h3>
    {/if}
    
    {#if subtitleProp}
      <p class="card-subtitle">{subtitleProp}</p>
    {/if}
    
    <div class="card-content">
      {@render children?.()}
    </div>
  </article>
{/if}

<style>
  .card {
    display: block;
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    border: 1px solid #f0f0f0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    transition: all 0.2s ease;
    text-decoration: none;
    color: inherit;
  }
  
  .card.clickable {
    cursor: pointer;
  }
  
  .card.clickable:hover {
    box-shadow: 0 8px 24px rgba(255, 140, 66, 0.12);
    border-color: rgba(255, 140, 66, 0.3);
    transform: translateY(-2px);
  }
  
  .card-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-teren-text-main);
    margin: 0 0 0.5rem 0;
  }
  
  .card-subtitle {
    font-size: 0.875rem;
    color: var(--color-teren-text-muted);
    margin: 0 0 1rem 0;
  }
  
  .card-content {
    margin-top: 1rem;
  }
</style>
