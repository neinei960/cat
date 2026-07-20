import {
  buildOrderCareReportDraft,
  buildOrderCareReportPayload,
  buildOrderCareReportPetOptions,
  canGenerateOrderCareReport,
  normalizeOrderCareReportDate,
  orderCareReportBodyShapeOptions,
  orderCareReportSectionDefinitions,
} from '../src/utils/order-care-report'
import { buildOrderCareReportFileName } from '../src/utils/web-image-save'

function createSampleOrder(overrides: { pay_status?: number } = {}) {
  return {
    ID: 101,
    order_no: 'ORD-CARE-001',
    status: 1,
    pay_time: '2026-04-20 11:14:17',
    CreatedAt: '2026-04-19 09:00:00',
    customer: { ID: 1, shop_id: 1, openid: '', phone: '13800138000', nickname: '测试客户', avatar: '', gender: 0, remark: '', tags: '', total_spent: 0, visit_count: 0, last_visit_at: '', member_balance: 0, discount_rate: 1, address: '', address_detail: '', door_code: '', CreatedAt: '2026-04-19 09:00:00' },
    pet: {
      ID: 11,
      shop_id: 1,
      customer_id: 1,
      name: '福福',
      species: '猫',
      breed: '英短',
      gender: 1,
      birth_date: '2024-01-01',
      weight: 4.1,
      coat_type: '',
      coat_color: '',
      fur_level: '',
      personality: '',
      aggression: '',
      forbidden_zones: '',
      bath_frequency: '',
      neutered: true,
      avatar: '',
      care_notes: '',
      behavior_notes: '',
      status: 1,
      CreatedAt: '2026-01-01 00:00:00',
    },
    appointment: {
      pets: [
        {
          pet_id: 11,
          pet: {
            ID: 11,
            shop_id: 1,
            customer_id: 1,
            name: '福福',
            species: '猫',
            breed: '英短',
            gender: 1,
            birth_date: '2024-01-01',
            weight: 4.1,
            coat_type: '',
            coat_color: '',
            fur_level: '',
            personality: '',
            aggression: '',
            forbidden_zones: '',
            bath_frequency: '',
            neutered: true,
            avatar: '',
            care_notes: '',
            behavior_notes: '',
            status: 1,
            CreatedAt: '2026-01-01 00:00:00',
          },
        },
        {
          pet_id: 12,
          pet: {
            ID: 12,
            shop_id: 1,
            customer_id: 1,
            name: '豆豆',
            species: '猫',
            breed: '布偶',
            gender: 2,
            birth_date: '2023-05-01',
            weight: 5.2,
            coat_type: '',
            coat_color: '',
            fur_level: '',
            personality: '',
            aggression: '',
            forbidden_zones: '',
            bath_frequency: '',
            neutered: false,
            avatar: '',
            care_notes: '',
            behavior_notes: '',
            status: 1,
            CreatedAt: '2026-01-01 00:00:00',
          },
        },
      ],
    },
    pet_groups: [
      {
        pet_id: 11,
        pet_name: '福福',
        items: [
          { ID: 1, order_id: 101, item_type: 1, item_id: 1001, name: '福福 · 皮毛调理', quantity: 1, unit_price: 88, amount: 88 },
        ],
      },
      {
        pet_id: 12,
        pet_name: '豆豆',
        items: [
          { ID: 2, order_id: 101, item_type: 1, item_id: 1002, name: '福福 · Harmurry精致皮毛调理', quantity: 1, unit_price: 99, amount: 99 },
        ],
      },
    ],
    items: [
      { ID: 3, order_id: 101, item_type: 1, item_id: 1003, name: '通用 · 备用项目', quantity: 1, unit_price: 10, amount: 10 },
    ],
    ...overrides,
  }
}

function main() {
  const fixtureOrder = createSampleOrder()
  const unpaidOrder = createSampleOrder({ pay_status: 0 })
  const paidOrder = createSampleOrder({ pay_status: 1 })

  assertEqual(canGenerateOrderCareReport(unpaidOrder), false, 'unpaid order cannot generate')
  assertEqual(canGenerateOrderCareReport({ status: 0, pet_groups: [] } as any), false, 'unpaid empty order cannot generate')
  assertEqual(canGenerateOrderCareReport(fixtureOrder), true, 'missing pay_status fixture can generate')
  assertEqual(canGenerateOrderCareReport(paidOrder), true, 'paid order can generate')
  assertEqual(buildOrderCareReportFileName('NO167', '福福'), '护理报告_NO167_福福.jpg', 'report filename')

  const petOptions = buildOrderCareReportPetOptions(fixtureOrder)
  assertEqual(petOptions.length, 2, 'pet options length')
  assertDeepEqual(petOptions.map((option) => option.petId), [11, 12], 'pet options ids')

  const draft = buildOrderCareReportDraft(fixtureOrder, 12)
  assertEqual(draft.petName, '豆豆', 'draft.petName')
  assertEqual(draft.breed, '布偶', 'draft.breed')
  assertEqual(draft.age, '2岁11月', 'draft.age')
  assertEqual(draft.careContent, 'Harmurry精致皮毛调理', 'draft.careContent')
  draft.portraitUrl = '/uploads/care-report.jpg'
  draft.weight = '5.2kg'
  draft.nextCareDate = '2026.05.20'
  draft.bodyShape = 'standard'
  draft.skin = { checks: ['normal', 'red'], note: '局部轻微泛红' }
  assertDeepEqual(buildOrderCareReportPayload(draft), {
    pet_id: 12,
    pet_name: '豆豆',
    breed: '布偶',
    gender: 'MM',
    age: '2岁11月',
    portrait_url: '/uploads/care-report.jpg',
    weight: '5.2kg',
    care_date: '2026-04-20',
    next_care_date: '2026-05-20',
    care_content: 'Harmurry精致皮毛调理',
    body_shape: 'standard',
    skin: { checks: ['normal', 'red'], note: '局部轻微泛红' },
    hair: { checks: [], note: '' },
    nails: { checks: [], note: '' },
    eyes_face: { checks: [], note: '' },
    ears: { checks: [], note: '' },
    oral: { checks: [], note: '' },
    anus: { checks: [], note: '' },
  }, 'backend payload')
  assertEqual(normalizeOrderCareReportDate('2026.04.20'), '2026-04-20', 'normalize dotted date')
  assertEqual(normalizeOrderCareReportDate('2026-04-25'), '2026-04-25', 'normalize dashed date')
  assertEqual(orderCareReportBodyShapeOptions.length, 5, 'body shape option count')
  assertEqual(orderCareReportSectionDefinitions.length, 7, 'section definition count')
  assertDeepEqual(orderCareReportSectionDefinitions.map((section) => section.key), ['skin', 'hair', 'nails', 'eyesFace', 'ears', 'oral', 'anus'], 'section keys')
  assertDeepEqual(
    Object.fromEntries(orderCareReportSectionDefinitions.map((section) => [section.key, section.options.map((option) => option.value)])),
    {
      skin: ['normal', 'dandruff', 'red', 'greasy', 'scab', 'wound'],
      hair: ['shedding', 'undercoat_many', 'dry', 'greasy', 'matting'],
      nails: ['trimmed', 'dewclaw_abnormal', 'pads_dry', 'too_long', 'wound'],
      eyesFace: ['cleaned', 'tear_many', 'eye_red', 'eye_abnormal', 'wound'],
      ears: ['cleaned', 'touch_sensitive', 'inflamed', 'earwax', 'black_earwax', 'wound'],
      oral: ['normal', 'touch_sensitive', 'tartar', 'gum_red', 'gum_swollen', 'oral_ulcer', 'bad_breath', 'dental_abnormal'],
      anus: ['normal', 'prolapse', 'red', 'inflamed'],
    },
    'section options should exactly match the printed template'
  )
}

main()

function assertEqual(actual: unknown, expected: unknown, label: string) {
  if (actual !== expected) {
    throw new Error(`${label}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`)
  }
}

function assertDeepEqual(actual: unknown, expected: unknown, label: string) {
  const actualJson = JSON.stringify(actual)
  const expectedJson = JSON.stringify(expected)
  if (actualJson !== expectedJson) {
    throw new Error(`${label}: expected ${expectedJson}, got ${actualJson}`)
  }
}
