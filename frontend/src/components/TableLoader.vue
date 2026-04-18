<script setup lang="ts">
defineProps<{ label?: string }>()
</script>

<template>
  <div class="empty-card">
    <div class="empty-icon">
      <div class="spinner" />
    </div>
    <p class="empty-title">{{ label ?? 'Loading…' }}</p>
  </div>
</template>
