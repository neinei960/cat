import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getPetList } from '@/api/pet'

export const usePetStore = defineStore('pet', () => {
  const pets = ref<Pet[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchList(page = 1, pageSize = 10) {
    loading.value = true
    try {
      const res = await getPetList({ page, page_size: pageSize })
      pets.value = res.data.list
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  return { pets, total, loading, fetchList }
})
