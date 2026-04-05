import { computed, onMounted, onUnmounted, ref } from 'vue'

const screenWidth = ref(0)
let bindCount = 0

function syncScreenWidth() {
  try {
    const info = uni.getSystemInfoSync()
    screenWidth.value = info.windowWidth || 0
  } catch {
    screenWidth.value = 0
  }
}

// #ifdef H5
function handleResize() {
  screenWidth.value = window.innerWidth
}
// #endif

export function useDesktopInteraction() {
  onMounted(() => {
    syncScreenWidth()
    bindCount += 1
    if (bindCount > 1) return
    // #ifdef H5
    window.addEventListener('resize', handleResize)
    // #endif
  })

  onUnmounted(() => {
    if (bindCount <= 0) return
    bindCount -= 1
    if (bindCount > 0) return
    // #ifdef H5
    window.removeEventListener('resize', handleResize)
    // #endif
  })

  const isDesktopInteraction = computed(() => screenWidth.value >= 768)

  return {
    isDesktopInteraction,
  }
}
