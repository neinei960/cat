/**
 * Safe navigateBack: if there's no page history (H5 mode),
 * redirect to the index page instead of exiting the webview.
 */
export function safeBack() {
  const pages = getCurrentPages()
  if (pages.length > 1) {
    uni.navigateBack()
  } else {
    uni.reLaunch({ url: '/pages/index/index' })
  }
}
