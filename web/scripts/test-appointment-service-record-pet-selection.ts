import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const pageSource = fs.readFileSync(path.join(root, 'src/pages/appointment/detail.vue'), 'utf8')
const repoSource = fs.readFileSync(path.resolve(root, '../server/internal/repository/service_record_repo.go'), 'utf8')

assert(
  pageSource.includes('const selectedRecordPetId = ref(0)'),
  'appointment detail should track selected service-record pet id',
)

assert(
  pageSource.includes('function openRecordForm()'),
  'appointment detail should open service-record modal through a pet-aware initializer',
)

assert(
  pageSource.includes("uni.showToast({ title: '请选择猫咪'"),
  'service record submit should require choosing a pet',
)

assert(
  !pageSource.includes('const petId = appointmentPets.value[0]?.pet_id || appt.value?.pet_id || 0'),
  'service record submit must not silently use the first appointment pet',
)

assert(
  pageSource.includes('record-pet-name') && pageSource.includes('rec.pet?.name'),
  'service record list should display the pet name for each record',
)

assert(
  repoSource.includes('Preload("Pet").Preload("Staff")') || repoSource.includes('Preload("Staff").Preload("Pet")'),
  'service record repository should preload Pet when listing appointment records',
)

console.log('appointment service record pet selection checks passed')
