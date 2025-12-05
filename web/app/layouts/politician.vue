<template>
  <div class="min-h-screen flex flex-col bg-stone-50 dark:bg-stone-950">
    <AppHeader />
    <PoliticianNavbar />
    <main class="flex-1">
      <slot />
    </main>
    <AppFooter />
    <SupportChat />
  </div>
</template>
