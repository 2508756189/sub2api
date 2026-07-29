<template>
  <span class="tokenport-brand" :class="{ 'with-subtitle': subtitle }">
    <span class="tokenport-mark">
      <img :src="src" alt="" />
    </span>
    <span class="tokenport-copy">
      <strong v-if="name === 'TokenPort'">Token<em>Port</em></strong>
      <strong v-else>{{ name }}</strong>
      <small v-if="subtitle">{{ subtitle }}</small>
    </span>
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  src?: string
  name?: string
  subtitle?: string
}>(), {
  src: '/tokenport-logo.svg',
  name: 'TokenPort',
  subtitle: '',
})
</script>

<style scoped>
.tokenport-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.tokenport-mark {
  display: grid;
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  place-items: center;
  overflow: hidden;
  border-radius: 9px;
}

.tokenport-mark img {
  display: block;
  width: 100%;
  height: 100%;
}

.tokenport-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.tokenport-copy strong {
  overflow: hidden;
  color: var(--color-fg, #13251f);
  font: 700 15px/1 'JetBrains Mono', 'Noto Sans SC', ui-monospace, monospace;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tokenport-copy em {
  color: var(--color-primary, var(--tp-brand));
  font-style: normal;
}

.tokenport-copy small {
  overflow: hidden;
  max-width: 240px;
  color: var(--color-faint, #71877e);
  font-size: 11px;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(.dark .tokenport-copy strong) { color: var(--color-fg, #e8f2ee); }
:global(.dark .tokenport-copy small) { color: var(--color-faint, #789087); }

@media (max-width: 520px) {
  .with-subtitle .tokenport-copy small {
    display: none;
  }
}
</style>
