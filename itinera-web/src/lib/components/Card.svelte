<script lang="ts">
import type { Snippet } from "svelte";

let { title: title, subtitle: subtitle, 
    href: href, clickable, children }
  : { title: string; subtitle: string; 
    href: string | null; clickable: boolean; 
    children?: Snippet } = $props();

const isInteractive = () => Boolean(clickable && href);
</script>

{#if isInteractive()}
  <a
    class="card bg-teren-surface border border-teren-border {href || clickable ? 'clickable' : ''}"
    href={href ?? undefined}
    onclick={() => clickable && href ? window.location.href = href : null}

  >
   {#if title}
    <h3 class="text-xl font-bold text-teren-text-main mb-1">{title}</h3>
  {/if}
  
  {#if subtitle}
    <p class="text-sm text-teren-text-muted mb-4">{subtitle}</p>
  {/if}
    
    <div class="card-content">
      {@render children?.()}
    </div>
  </a>
{:else}
  <article
    class="card"
    class:has-title={title !== ""}
  >
    {#if title}
      <h3 class="card-title">{title}</h3>
    {/if}
    
    {#if subtitle}
      <p class="card-subtitle">{subtitle}</p>
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
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    transition: all 0.2s ease;
    text-decoration: none;
    color: inherit;
  }
  
  .card.clickable {
    cursor: pointer;
  }
  
  .card.clickable:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    border-color: var(--color-teren-border);
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
