"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.buildOrderCareReportDraft = exports.buildOrderCareReportPetOptions = exports.canGenerateOrderCareReport = exports.normalizeOrderCareReportDate = exports.orderCareReportSectionDefinitions = exports.orderCareReportBodyShapeOptions = void 0;
exports.orderCareReportBodyShapeOptions = [
    { label: '太瘦', value: 'thin' },
    { label: '偏瘦', value: 'skinny' },
    { label: '标准', value: 'standard' },
    { label: '偏胖', value: 'chubby' },
    { label: '肥胖', value: 'obese' },
];
exports.orderCareReportSectionDefinitions = [
    {
        key: 'skin',
        label: '皮肤检查',
        options: [
            { label: '正常', value: 'normal' },
            { label: '皮屑', value: 'dandruff' },
            { label: '发红', value: 'red' },
            { label: '出油', value: 'greasy' },
            { label: '结痂', value: 'scab' },
        ],
    },
    {
        key: 'hair',
        label: '被毛检查',
        options: [
            { label: '飞毛', value: 'shedding' },
            { label: '底毛多', value: 'undercoat_many' },
            { label: '干燥', value: 'dry' },
            { label: '油腻', value: 'greasy' },
            { label: '打结', value: 'matting' },
        ],
    },
    {
        key: 'nails',
        label: '修剪指甲',
        options: [
            { label: '已剪', value: 'trimmed' },
            { label: '趾间异常', value: 'dewclaw_abnormal' },
            { label: '足垫干燥', value: 'pads_dry' },
            { label: '过长', value: 'too_long' },
            { label: '伤口', value: 'wound' },
        ],
    },
    {
        key: 'eyesFace',
        label: '眼睛及脸部',
        options: [
            { label: '已清洁', value: 'cleaned' },
            { label: '分泌物多', value: 'tear_many' },
            { label: '眼睛发红', value: 'eye_red' },
            { label: '眼貌异常', value: 'eye_abnormal' },
            { label: '伤口', value: 'wound' },
        ],
    },
    {
        key: 'ears',
        label: '清洁耳朵',
        options: [
            { label: '已清洁', value: 'cleaned' },
            { label: '讨厌触碰', value: 'touch_sensitive' },
            { label: '发红肿胀', value: 'inflamed' },
            { label: '耳垢结痂', value: 'earwax' },
            { label: '耳垢发黑', value: 'black_earwax' },
        ],
    },
    {
        key: 'oral',
        label: '口腔检查',
        options: [
            { label: '正常', value: 'normal' },
            { label: '牙结石', value: 'tartar' },
            { label: '牙龈发红', value: 'gum_red' },
            { label: '牙龈肿胀', value: 'gum_swollen' },
        ],
    },
    {
        key: 'anus',
        label: '肛门及周围',
        options: [
            { label: '正常', value: 'normal' },
            { label: '发红', value: 'red' },
            { label: '发炎', value: 'inflamed' },
            { label: '肛腺肿胀', value: 'anal_gland_swollen' },
        ],
    },
];
function compactText(value) {
    return String(value || '').trim();
}
function normalizePetName(value) {
    const text = compactText(value);
    if (!text)
        return '';
    const parts = text.split(/\s*·\s*/);
    if (parts.length > 1) {
        return parts.slice(1).join(' · ').trim();
    }
    return text;
}
function mapGender(value) {
    if (value === 1)
        return 'GG';
    if (value === 2)
        return 'MM';
    return '';
}
function formatDottedDate(value) {
    const text = compactText(value);
    if (!text)
        return '';
    const datePart = text.slice(0, 10);
    if (/^\d{4}-\d{2}-\d{2}$/.test(datePart)) {
        return datePart.replace(/-/g, '.');
    }
    if (/^\d{4}\.\d{2}\.\d{2}$/.test(datePart)) {
        return datePart;
    }
    return '';
}
function normalizeOrderCareReportDate(value) {
    const text = compactText(value);
    if (!text)
        return '';
    const normalized = formatDottedDate(text);
    if (normalized)
        return normalized.replace(/\./g, '-');
    if (/^\d{4}-\d{2}-\d{2}$/.test(text.slice(0, 10))) {
        return text.slice(0, 10);
    }
    return '';
}
exports.normalizeOrderCareReportDate = normalizeOrderCareReportDate;
function formatAgeDisplay(birthDate, referenceDate) {
    const birthText = normalizeOrderCareReportDate(birthDate);
    const referenceText = normalizeOrderCareReportDate(referenceDate);
    if (!birthText || !referenceText)
        return '';
    const birth = new Date(`${birthText}T00:00:00`);
    const reference = new Date(`${referenceText}T00:00:00`);
    if (Number.isNaN(birth.getTime()) || Number.isNaN(reference.getTime()) || birth > reference) {
        return '';
    }
    let totalMonths = (reference.getFullYear() - birth.getFullYear()) * 12 + (reference.getMonth() - birth.getMonth());
    if (reference.getDate() < birth.getDate()) {
        totalMonths--;
    }
    if (totalMonths < 0)
        return '';
    if (totalMonths >= 12) {
        const years = Math.floor(totalMonths / 12);
        const months = totalMonths % 12;
        if (months === 0)
            return `${years}岁`;
        return `${years}岁${months}月`;
    }
    if (totalMonths > 0) {
        return `${totalMonths}月`;
    }
    const days = Math.floor((reference.getTime() - birth.getTime()) / (24 * 60 * 60 * 1000));
    if (days <= 0)
        return '';
    return `${days}天`;
}
function listify(value) {
    return Array.isArray(value) ? value : [];
}
function createEmptySection() {
    return {
        checks: [],
        note: '',
    };
}
function isReportablePetId(value) {
    return typeof value === 'number' && Number.isFinite(value) && value > 0;
}
function isServiceItem(item) {
    return Number(item?.item_type || 0) === 1;
}
function collectServiceItemNames(items) {
    const list = Array.isArray(items) ? items : [];
    return list.filter(isServiceItem).map((item) => normalizePetName(item?.name)).filter(Boolean);
}
function resolveSelectedPetContext(order, petId) {
    const sourceOrder = order || {};
    if (isReportablePetId(sourceOrder.pet?.ID) && sourceOrder.pet?.ID === petId) {
        return {
            petName: compactText(sourceOrder.pet?.name),
            breed: compactText(sourceOrder.pet?.breed),
            gender: mapGender(sourceOrder.pet?.gender),
            birthDate: compactText(sourceOrder.pet?.birth_date),
            petGroup: sourceOrder.pet_groups?.find((group) => group?.pet_id === petId) || null,
        };
    }
    const appointmentPets = listify(sourceOrder.appointment?.pets);
    for (const appointmentPet of appointmentPets) {
        const resolvedPetId = appointmentPet?.pet?.ID ?? appointmentPet?.pet_id;
        if (resolvedPetId === petId) {
            return {
                petName: compactText(appointmentPet.pet?.name),
                breed: compactText(appointmentPet.pet?.breed),
                gender: mapGender(appointmentPet.pet?.gender),
                birthDate: compactText(appointmentPet.pet?.birth_date),
                petGroup: sourceOrder.pet_groups?.find((group) => group?.pet_id === petId) || null,
            };
        }
    }
    const petGroup = Array.isArray(sourceOrder.pet_groups)
        ? sourceOrder.pet_groups.find((group) => group?.pet_id === petId) || null
        : null;
    return {
        petName: compactText(petGroup?.pet_name),
        breed: '',
        gender: '',
        birthDate: '',
        petGroup,
    };
}
function mergePetOption(options, seen, petId, petName) {
    if (!isReportablePetId(petId))
        return;
    const normalizedName = compactText(petName);
    const existing = seen.get(petId);
    if (existing) {
        if (!existing.petName && normalizedName) {
            existing.petName = normalizedName;
        }
        return;
    }
    const option = {
        petId,
        petName: normalizedName,
    };
    seen.set(petId, option);
    options.push(option);
}
function getCareDateValue(order) {
    return formatDottedDate(order?.pay_time || order?.CreatedAt);
}
function canGenerateOrderCareReport(order) {
    if (!order)
        return false;
    if (Number(order.status || 0) !== 1) {
        return false;
    }
    if (Object.prototype.hasOwnProperty.call(order, 'pay_status') && Number(order.pay_status || 0) !== 1) {
        return false;
    }
    return buildOrderCareReportPetOptions(order).length > 0;
}
exports.canGenerateOrderCareReport = canGenerateOrderCareReport;
function buildOrderCareReportPetOptions(order) {
    const options = [];
    const seen = new Map();
    const sourceOrder = order || {};
    mergePetOption(options, seen, sourceOrder.pet?.ID, sourceOrder.pet?.name);
    const appointmentPets = listify(sourceOrder.appointment?.pets);
    for (const appointmentPet of appointmentPets) {
        const resolvedPetId = appointmentPet?.pet?.ID ?? appointmentPet?.pet_id;
        mergePetOption(options, seen, resolvedPetId, appointmentPet?.pet?.name);
    }
    const petGroups = listify(sourceOrder.pet_groups);
    for (const group of petGroups) {
        mergePetOption(options, seen, group?.pet_id, group?.pet_name);
    }
    mergePetOption(options, seen, sourceOrder.pet_id, sourceOrder.pet?.name);
    return options;
}
exports.buildOrderCareReportPetOptions = buildOrderCareReportPetOptions;
function buildOrderCareReportDraft(order, petId) {
    const sourceOrder = order || {};
    const selectedPet = resolveSelectedPetContext(sourceOrder, petId);
    const careDate = getCareDateValue(sourceOrder);
    const age = formatAgeDisplay(selectedPet.birthDate, careDate);
    const selectedGroup = selectedPet.petGroup;
    const sourceItems = selectedGroup?.items?.length ? selectedGroup.items : sourceOrder.items;
    const careContent = collectServiceItemNames(sourceItems).join('、');
    return {
        petId,
        petName: selectedPet.petName,
        breed: selectedPet.breed,
        gender: selectedPet.gender,
        age,
        careDate,
        nextCareDate: '',
        portraitUrl: '',
        weight: '',
        careContent,
        bodyShape: '',
        skin: createEmptySection(),
        hair: createEmptySection(),
        nails: createEmptySection(),
        eyesFace: createEmptySection(),
        ears: createEmptySection(),
        oral: createEmptySection(),
        anus: createEmptySection(),
    };
}
exports.buildOrderCareReportDraft = buildOrderCareReportDraft;
