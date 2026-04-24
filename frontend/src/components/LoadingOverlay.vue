<script setup lang="ts">
defineProps<{ visible: boolean; message?: string }>()
</script>

<template>
  <Teleport to="body">
    <Transition name="overlay-fade">
      <div v-if="visible" class="loading-overlay">
        <div class="loading-box">
          <div class="loading-spinner" />
          <p class="loading-message">{{ message ?? 'Loading…' }}</p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.loading-overlay {
  position: fixed;
  inset: 0;
  z-index: 300;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
}
.loading-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  background: white;
  border-radius: 16px;
  padding: 2.25rem 2.75rem;
  box-shadow: 0 20px 60px rgba(0,0,0,0.25);
}
.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e2e8f0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.75s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.loading-message {
  font-size: 0.9rem;
  font-weight: 500;
  color: #475569;
  margin: 0;
}
.overlay-fade-enter-active { transition: opacity 0.2s ease; }
.overlay-fade-leave-active { transition: opacity 0.25s ease; }
.overlay-fade-enter-from,
.overlay-fade-leave-to { opacity: 0; }
</style>
