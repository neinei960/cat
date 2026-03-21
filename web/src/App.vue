<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useAuthStore } from "@/store/auth";

onLaunch(() => {
  const authStore = useAuthStore();
  authStore.loadFromStorage();

  if (!authStore.isLoggedIn) {
    uni.reLaunch({ url: '/pages/login/index' });
  } else {
    uni.reLaunch({ url: '/pages/index/index' });
  }
});
onShow(() => {
  console.log("App Show");
});
onHide(() => {
  console.log("App Hide");
});
</script>
<style>
/* #ifdef H5 */
uni-page-head {
  display: none !important;
}
/* #endif */
</style>
