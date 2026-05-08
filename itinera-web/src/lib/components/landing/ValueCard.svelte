<script lang="ts">
  let { icon, title, description } = $props<{
    icon: string;
    title: string;
    description: string;
  }>();
</script>

<div class="bg-teren-surface border border-teren-border border-t-white/5 rounded-xl p-6 transition-all duration-300 hover:border-teren-primary/40 hover:-translate-y-1 hover:shadow-lg hover:shadow-teren-primary/5 group">
  <div class="text-3xl mb-4 group-hover:scale-110 transition-transform duration-300">
    {icon}
  </div>
  <h3 class="text-lg font-bold text-teren-text-main mb-2">
    {title}
  </h3>
  <p class="text-sm text-teren-text-muted leading-relaxed">
    {description}
  </p>
</div>
