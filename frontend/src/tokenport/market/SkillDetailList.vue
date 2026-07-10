<template>
  <section class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
    <h4 class="text-sm font-semibold text-gray-900 dark:text-gray-100">{{ title }}</h4>
    <ul v-if="items.length" class="mt-2 space-y-1.5 text-sm text-gray-600 dark:text-gray-300">
      <li v-for="item in items" :key="item" class="flex gap-2">
        <span class="mt-2 h-1.5 w-1.5 flex-none rounded-full bg-primary-500"></span>
        <span>{{ item }}</span>
      </li>
    </ul>
    <p v-else class="mt-2 text-sm text-gray-400">{{ emptyText }}</p>
  </section>
</template>

<script setup lang="ts">
defineProps<{ title: string; items: string[]; emptyText: string }>()
</script>
