<script setup lang="ts">
import { onLaunch } from "@dcloudio/uni-app";
import { useAuthStore } from "./store/auth";

onLaunch(() => {
  const auth = useAuthStore();
  auth.loadFromStorage();
});
</script>
<style>
page { background-color: #F5F6FA; }
</style>
