/**
 * Composable for scroll-triggered animations using IntersectionObserver
 */
export function useScrollAnimation() {
  const observedElements = ref<Set<Element>>(new Set())

  const observer = ref<IntersectionObserver | null>(null)

  const initObserver = () => {
    if (import.meta.server) return

    observer.value = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('is-visible')
            // Optionally unobserve after animation
            // observer.value?.unobserve(entry.target)
          }
        })
      },
      {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
      }
    )
  }

  const observe = (el: Element | null) => {
    if (!el || observedElements.value.has(el)) return

    if (!observer.value) {
      initObserver()
    }

    observer.value?.observe(el)
    observedElements.value.add(el)
  }

  const observeAll = (selector: string = '.animate-on-scroll, .animate-on-scroll-left, .animate-on-scroll-right, .animate-on-scroll-scale') => {
    if (import.meta.server) return

    nextTick(() => {
      const elements = document.querySelectorAll(selector)
      elements.forEach((el) => observe(el))
    })
  }

  const cleanup = () => {
    observer.value?.disconnect()
    observedElements.value.clear()
  }

  onMounted(() => {
    initObserver()
  })

  onUnmounted(() => {
    cleanup()
  })

  return {
    observe,
    observeAll,
    cleanup
  }
}

/**
 * Directive for scroll animations - use v-scroll-animate on elements
 */
export const vScrollAnimate = {
  mounted(el: HTMLElement, binding: { value?: string }) {
    const animationClass = binding.value || 'animate-on-scroll'

    if (!el.classList.contains(animationClass)) {
      el.classList.add(animationClass)
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('is-visible')
          }
        })
      },
      {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
      }
    )

    observer.observe(el)

    // Store observer for cleanup
    ;(el as HTMLElement & { _scrollObserver?: IntersectionObserver })._scrollObserver = observer
  },
  unmounted(el: HTMLElement) {
    const observer = (el as HTMLElement & { _scrollObserver?: IntersectionObserver })._scrollObserver
    if (observer) {
      observer.disconnect()
    }
  }
}
