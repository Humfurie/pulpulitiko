<template>
  <div class="min-h-screen flex flex-col bg-stone-50 dark:bg-stone-950">
    <!-- Skip to main content for keyboard users -->
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:bg-green-600 focus:text-white focus:px-4 focus:py-2 focus:rounded-md focus:shadow-lg focus:outline-none focus:ring-2 focus:ring-green-400"
    >
      Skip to main content
    </a>

    <AppHeader role="banner" />

    <main id="main-content" role="main" class="flex-1">
      <slot />
    </main>

    <AppFooter role="contentinfo" />
    <SupportChat />
  </div>
</template>
