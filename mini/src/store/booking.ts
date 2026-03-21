import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useBookingStore = defineStore('booking', () => {
  const selectedService = ref<any>(null)
  const selectedStaff = ref<any>(null)
  const selectedDate = ref('')
  const selectedTime = ref('')
  const selectedPet = ref<any>(null)
  const notes = ref('')

  function reset() {
    selectedService.value = null
    selectedStaff.value = null
    selectedDate.value = ''
    selectedTime.value = ''
    selectedPet.value = null
    notes.value = ''
  }

  return { selectedService, selectedStaff, selectedDate, selectedTime, selectedPet, notes, reset }
})
