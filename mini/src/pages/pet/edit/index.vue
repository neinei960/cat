<template>
  <view class="page">
    <view class="form">
      <view class="item"><text class="label">名字 *</text><input v-model="form.name" placeholder="宠物名字" class="input" /></view>
      <view class="item"><text class="label">物种 *</text>
        <picker :range="['犬','猫']" @change="(e:any)=>form.species=['犬','猫'][e.detail.value]"><view class="input">{{ form.species || '请选择' }}</view></picker>
      </view>
      <view class="item"><text class="label">品种</text><input v-model="form.breed" placeholder="如金毛" class="input" /></view>
      <view class="item"><text class="label">体重(kg)</text><input v-model="form.weight" type="digit" placeholder="0" class="input" /></view>
      <view class="item"><text class="label">毛发类型</text>
        <picker :range="['短毛','长毛','卷毛']" @change="(e:any)=>form.coat_type=['短毛','长毛','卷毛'][e.detail.value]"><view class="input">{{ form.coat_type || '请选择' }}</view></picker>
      </view>
      <view class="item"><text class="label">注意事项</text><textarea v-model="form.medical_alerts" placeholder="过敏/疾病" class="textarea" /></view>
    </view>
    <button class="btn" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '添加' }}</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getPets, createPet, updatePet } from '../../../api'
const id = ref(0); const submitting = ref(false)
const form = ref({ name:'', species:'犬', breed:'', weight:'', coat_type:'', medical_alerts:'' })

onLoad(async (q) => {
  if(q?.id) {
    id.value = parseInt(q.id)
    const r = await getPets()
    const p = (r.data||[]).find((x:any)=>x.ID===id.value)
    if(p) form.value = { name:p.name, species:p.species, breed:p.breed, weight:String(p.weight||''), coat_type:p.coat_type, medical_alerts:p.medical_alerts }
  }
})

async function onSubmit() {
  if(!form.value.name || !form.value.species) { uni.showToast({title:'请填写必填项',icon:'none'}); return }
  submitting.value = true
  try {
    const data = { ...form.value, weight: parseFloat(form.value.weight)||0 }
    if(id.value) await updatePet(id.value, data); else await createPet(data)
    uni.showToast({title:'保存成功',icon:'success'})
    setTimeout(()=>uni.navigateBack(), 500)
  } finally { submitting.value = false }
}
</script>

<style scoped>
.page{padding:24rpx;}.form{background:#fff;border-radius:16rpx;padding:8rpx 24rpx;margin-bottom:32rpx;}
.item{padding:24rpx 0;border-bottom:1rpx solid #F3F4F6;}.item:last-child{border-bottom:none;}
.label{font-size:28rpx;color:#374151;display:block;margin-bottom:12rpx;}.input{font-size:28rpx;color:#1F2937;height:60rpx;line-height:60rpx;}
.textarea{font-size:28rpx;width:100%;height:120rpx;}.btn{background:#4F46E5;color:#fff;border-radius:12rpx;font-size:30rpx;}
</style>
