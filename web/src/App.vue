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
/* picker 弹出层 z-index 需高于筛选面板(3000) */
.uni-picker-container {
  z-index: 5000 !important;
}
.uni-picker-container .uni-picker-mask {
  z-index: 5000 !important;
}
.uni-picker-container .uni-picker-custom {
  z-index: 5001 !important;
}
/* #endif */
</style>
