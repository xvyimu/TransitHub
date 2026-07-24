<script setup lang="ts">
import { ConfigProvider } from 'antdv-next'
import { RouterView } from 'vue-router'
import useShadcnTheme from './shadcnTheme'

const configProps = useShadcnTheme()
</script>

<template>
  <ConfigProvider v-bind="configProps">
    <RouterView />
  </ConfigProvider>
</template>