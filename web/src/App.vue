<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useAuthStore } from "@/store/auth";
import { getShop } from "@/api/shop";
import { syncPersonalityConfigFromShop } from "@/utils/personality";

function getInitialPath() {
  // H5 刷新时保留 hash 路由，避免强制跳回工作台
  // #/pages/order/detail?id=1 -> /pages/order/detail
  // #/ -> /
  // 非 H5 端取不到 location 时，退化为根路径
  if (typeof window === 'undefined') return '/'
  const hash = window.location.hash || ''
  const raw = hash.startsWith('#') ? hash.slice(1) : hash
  const path = raw.split('?')[0] || '/'
  return path.startsWith('/') ? path : `/${path}`
}

async function refreshGlobalShopConfig() {
  try {
    const res = await getShop()
    syncPersonalityConfigFromShop(res.data)
  } catch {}
}

onLaunch(() => {
  const authStore = useAuthStore();
  authStore.loadFromStorage();
  const currentPath = getInitialPath()
  const isLoginPage = currentPath === '/pages/login/index'
  const isRootEntry = currentPath === '/' || currentPath === ''

  if (!authStore.isLoggedIn && !isLoginPage) {
    uni.reLaunch({ url: '/pages/login/index' });
  } else if (authStore.isLoggedIn && isRootEntry) {
    uni.reLaunch({ url: '/pages/index/index' });
  }
  if (authStore.isLoggedIn) {
    void refreshGlobalShopConfig()
  }
});
onShow(() => {
  console.log("App Show");
  const authStore = useAuthStore()
  if (authStore.isLoggedIn) {
    void refreshGlobalShopConfig()
  }
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
/* uni.showModal 默认 999，会被待处理面板(1500)盖住 */
.uni-mask {
  z-index: 6000 !important;
}
.uni-modal {
  z-index: 6001 !important;
}
/* #endif */

button {
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  box-sizing: border-box;
}

input,
textarea {
  text-align: center;
  box-sizing: border-box;
}

input::placeholder,
textarea::placeholder {
  text-align: center;
}

input::-webkit-input-placeholder,
textarea::-webkit-input-placeholder {
  text-align: center;
}

.uni-input-input,
.uni-textarea-textarea {
  text-align: center !important;
}

.uni-input-placeholder,
.uni-textarea-placeholder {
  width: 100%;
  text-align: center !important;
}

.search-input,
.input,
.form-input-direct,
.remark-input,
.pay-remark-input,
.parsed-note-editor,
.textarea,
.template-textarea,
.edit-input,
.add-input,
.fp-text-input,
.field-input,
.field-textarea,
.form-input-sm,
.form-textarea,
.notes-input {
  text-align: left !important;
}

.search-input::placeholder,
.input::placeholder,
.form-input-direct::placeholder,
.remark-input::placeholder,
.pay-remark-input::placeholder,
.parsed-note-editor::placeholder,
.textarea::placeholder,
.template-textarea::placeholder,
.edit-input::placeholder,
.add-input::placeholder,
.fp-text-input::placeholder,
.field-input::placeholder,
.field-textarea::placeholder,
.form-input-sm::placeholder,
.form-textarea::placeholder,
.notes-input::placeholder,
.search-input::-webkit-input-placeholder,
.input::-webkit-input-placeholder,
.form-input-direct::-webkit-input-placeholder,
.remark-input::-webkit-input-placeholder,
.pay-remark-input::-webkit-input-placeholder,
.parsed-note-editor::-webkit-input-placeholder,
.textarea::-webkit-input-placeholder,
.template-textarea::-webkit-input-placeholder,
.edit-input::-webkit-input-placeholder,
.add-input::-webkit-input-placeholder,
.fp-text-input::-webkit-input-placeholder,
.field-input::-webkit-input-placeholder,
.field-textarea::-webkit-input-placeholder,
.form-input-sm::-webkit-input-placeholder,
.form-textarea::-webkit-input-placeholder,
.notes-input::-webkit-input-placeholder {
  text-align: left !important;
}

.search-input .uni-input-input,
.input .uni-input-input,
.form-input-direct .uni-input-input,
.search-input .uni-textarea-textarea,
.input .uni-textarea-textarea,
.form-input-direct .uni-textarea-textarea,
.remark-input .uni-textarea-textarea,
.pay-remark-input .uni-textarea-textarea,
.parsed-note-editor .uni-textarea-textarea,
.textarea .uni-textarea-textarea,
.template-textarea .uni-textarea-textarea,
.edit-input .uni-input-input,
.add-input .uni-input-input,
.fp-text-input .uni-input-input,
.field-input .uni-input-input,
.field-textarea .uni-textarea-textarea,
.form-input-sm .uni-input-input,
.form-textarea .uni-textarea-textarea,
.notes-input .uni-textarea-textarea {
  text-align: left !important;
}

.search-input .uni-input-placeholder,
.input .uni-input-placeholder,
.form-input-direct .uni-input-placeholder,
.remark-input .uni-textarea-placeholder,
.pay-remark-input .uni-textarea-placeholder,
.parsed-note-editor .uni-textarea-placeholder,
.textarea .uni-textarea-placeholder,
.template-textarea .uni-textarea-placeholder,
.edit-input .uni-input-placeholder,
.add-input .uni-input-placeholder,
.fp-text-input .uni-input-placeholder,
.field-input .uni-input-placeholder,
.field-textarea .uni-textarea-placeholder,
.form-input-sm .uni-input-placeholder,
.form-textarea .uni-textarea-placeholder,
.notes-input .uni-textarea-placeholder {
  text-align: left !important;
}

.time-input,
.cap-input,
.age-input,
.discount-input,
.spec-input {
  text-align: center !important;
}

.time-input::placeholder,
.cap-input::placeholder,
.age-input::placeholder,
.discount-input::placeholder,
.spec-input::placeholder,
.time-input::-webkit-input-placeholder,
.cap-input::-webkit-input-placeholder,
.age-input::-webkit-input-placeholder,
.discount-input::-webkit-input-placeholder,
.spec-input::-webkit-input-placeholder {
  text-align: center !important;
}

.time-input .uni-input-input,
.cap-input .uni-input-input,
.age-input .uni-input-input,
.discount-input .uni-input-input,
.spec-input .uni-input-input,
.time-input .uni-input-placeholder,
.cap-input .uni-input-placeholder,
.age-input .uni-input-placeholder,
.discount-input .uni-input-placeholder,
.spec-input .uni-input-placeholder {
  text-align: center !important;
}

.input-amount,
.custom-price-input,
.addon-input,
.service-price-input {
  text-align: right !important;
}

.input-amount::placeholder,
.custom-price-input::placeholder,
.addon-input::placeholder,
.service-price-input::placeholder,
.input-amount::-webkit-input-placeholder,
.custom-price-input::-webkit-input-placeholder,
.addon-input::-webkit-input-placeholder,
.service-price-input::-webkit-input-placeholder {
  text-align: right !important;
}

.input-amount .uni-input-input,
.custom-price-input .uni-input-input,
.addon-input .uni-input-input,
.service-price-input .uni-input-input {
  text-align: right !important;
}

.input-amount .uni-input-placeholder,
.custom-price-input .uni-input-placeholder,
.addon-input .uni-input-placeholder,
.service-price-input .uni-input-placeholder {
  text-align: right !important;
}

.edit-input,
.add-input,
.fp-text-input,
.field-input,
.field-textarea,
.form-input-sm,
.form-textarea,
.notes-input,
.time-input,
.cap-input,
.age-input,
.discount-input,
.spec-input,
.input-amount,
.custom-price-input,
.addon-input,
.service-price-input {
  box-sizing: border-box;
}

.edit-input .uni-input-wrapper,
.add-input .uni-input-wrapper,
.fp-text-input .uni-input-wrapper,
.field-input .uni-input-wrapper,
.form-input-sm .uni-input-wrapper,
.time-input .uni-input-wrapper,
.cap-input .uni-input-wrapper,
.age-input .uni-input-wrapper,
.discount-input .uni-input-wrapper,
.spec-input .uni-input-wrapper,
.input-amount .uni-input-wrapper,
.custom-price-input .uni-input-wrapper,
.addon-input .uni-input-wrapper,
.service-price-input .uni-input-wrapper {
  width: 100%;
  min-height: 100%;
  display: flex;
  align-items: center;
}

.edit-input .uni-input-input,
.add-input .uni-input-input,
.fp-text-input .uni-input-input,
.field-input .uni-input-input,
.form-input-sm .uni-input-input,
.time-input .uni-input-input,
.cap-input .uni-input-input,
.age-input .uni-input-input,
.discount-input .uni-input-input,
.spec-input .uni-input-input,
.input-amount .uni-input-input,
.custom-price-input .uni-input-input,
.addon-input .uni-input-input,
.service-price-input .uni-input-input {
  width: 100%;
  min-height: 40rpx;
  line-height: 40rpx;
}

.field-textarea .uni-textarea-textarea,
.form-textarea .uni-textarea-textarea,
.notes-input .uni-textarea-textarea {
  width: 100%;
  min-height: 100%;
  box-sizing: border-box;
}

view[class^="btn-"]:not(.btn-row),
view[class*=" btn-"]:not(.btn-row),
view.btn,
view.btn-add,
view.btn-filter,
view.btn-secondary,
view.btn-primary,
view.btn-submit,
view.btn-delete,
view.btn-ghost,
view.btn-tag,
view.btn-trash,
view.btn-recharge,
view.btn-adjust,
view.btn-open-card,
view.btn-primary-sm,
view.btn-copy,
view.btn-save-schedule,
view.btn-day-off,
view.btn-reset,
view.card-action-btn,
view.record-btn,
view.pending-btn,
view.pending-select-btn,
view.pending-delete-btn,
view.pending-cancel-btn,
view.pending-filter-btn,
view.nav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  box-sizing: border-box;
}
</style>
