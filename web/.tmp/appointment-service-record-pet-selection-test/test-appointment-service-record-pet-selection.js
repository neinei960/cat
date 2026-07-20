"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const root = process.cwd();
const pageSource = node_fs_1.default.readFileSync(node_path_1.default.join(root, 'src/pages/appointment/detail.vue'), 'utf8');
const repoSource = node_fs_1.default.readFileSync(node_path_1.default.resolve(root, '../server/internal/repository/service_record_repo.go'), 'utf8');
(0, strict_1.default)(pageSource.includes('const selectedRecordPetId = ref(0)'), 'appointment detail should track selected service-record pet id');
(0, strict_1.default)(pageSource.includes('function openRecordForm()'), 'appointment detail should open service-record modal through a pet-aware initializer');
(0, strict_1.default)(pageSource.includes("uni.showToast({ title: '请选择猫咪'"), 'service record submit should require choosing a pet');
(0, strict_1.default)(!pageSource.includes('const petId = appointmentPets.value[0]?.pet_id || appt.value?.pet_id || 0'), 'service record submit must not silently use the first appointment pet');
(0, strict_1.default)(pageSource.includes('record-pet-name') && pageSource.includes('rec.pet?.name'), 'service record list should display the pet name for each record');
(0, strict_1.default)(repoSource.includes('Preload("Pet").Preload("Staff")') || repoSource.includes('Preload("Staff").Preload("Pet")'), 'service record repository should preload Pet when listing appointment records');
console.log('appointment service record pet selection checks passed');
