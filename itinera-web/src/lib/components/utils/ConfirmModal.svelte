<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';

  let {
    isOpen = false,
    title,
    message,
    confirmText = 'Aceptar',
    cancelText = 'Cancelar',
    isDestructive = false,
    onConfirm,
    onCancel
  }: {
    isOpen: boolean;
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    isDestructive?: boolean;
    onConfirm: () => void;
    onCancel: () => void;
  } = $props();
</script>

{#if isOpen}
  <div 
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 backdrop-blur-sm px-4" 
    transition:fade={{ duration: 200 }} 
    onclick={(e) => e.target === e.currentTarget && onCancel()}
  >
    <div 
      class="bg-teren-surface w-full max-w-sm rounded-2xl shadow-2xl overflow-hidden p-6" 
      transition:fly={{ y: 20, duration: 250, easing: cubicOut }} 
      onclick={(e) => e.stopPropagation()}
    >
      <h3 class="text-lg font-semibold text-teren-text-main tracking-tight mb-2">
        {title}
      </h3>
      <p class="text-sm text-teren-text-muted mb-6">
        {message}
      </p>
      
      <div class="flex justify-end gap-3">
        <button 
          class="px-4 py-2 text-sm font-medium text-teren-text-muted hover:text-teren-text-main rounded-lg hover:bg-gray-100 transition active:scale-95"
          onclick={onCancel}
        >
          {cancelText}
        </button>
        <button 
          class="px-4 py-2 text-sm font-medium text-white rounded-lg transition active:scale-95 shadow-sm {isDestructive ? 'bg-red-500 hover:bg-red-600 shadow-red-500/20' : 'bg-teren-primary hover:bg-teren-primary-hover shadow-teren-primary/20'}"
          onclick={onConfirm}
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}
